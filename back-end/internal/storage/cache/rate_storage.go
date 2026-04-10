package cache

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"mm/config"
	"mm/internal/client/solanarpc"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/errgroup"
)

var _ RateStorage = (*rateStorage)(nil)

var ErrDecimalNotFound = errors.New("decimal not found")
var ErrRateNotFound = errors.New("rate not found")

type RateStorage interface {
	GetUSDRate(ctx context.Context, currencyMint solana.PublicKey) (float64, error)
	GetUSDRates(ctx context.Context, currencyMints ...solana.PublicKey) (map[solana.PublicKey]float64, error)
	GetSOLRate(ctx context.Context, currencyMint solana.PublicKey) (float64, error)
	GetSOLRates(ctx context.Context, currencyMints ...solana.PublicKey) (map[solana.PublicKey]float64, error)
	GetDecimal(ctx context.Context, currencyMint solana.PublicKey) (uint8, error)
	GetDecimals(ctx context.Context, currencyMints ...solana.PublicKey) (map[solana.PublicKey]uint8, error)
}

func NewRateStorage(cfg *config.Config, rpc solanarpc.SolanaRPC) RateStorage {
	cacheTTL := cfg.App.ExchangeRateCacheTTL

	rateCache := gocache.New(cacheTTL, cacheTTL)
	decimalCache := gocache.New(gocache.NoExpiration, gocache.NoExpiration)

	storage := rateStorage{
		rateCache:    rateCache,
		decimalCache: decimalCache,
		url:          cfg.App.ExchangeRateURL,
		cacheTTL:     cacheTTL,
		rpc:          rpc,
	}

	rateCache.OnEvicted(func(key string, value interface{}) {
		_ = storage.jupiterUpdateRateAndDecimals(context.Background(), solana.SolMint)
	})

	return &storage
}

type rateStorage struct {
	rateCache    *gocache.Cache
	decimalCache *gocache.Cache
	rpc          solanarpc.SolanaRPC
	url          string
	cacheTTL     time.Duration
}

func (r *rateStorage) GetUSDRate(ctx context.Context, currencyMint solana.PublicKey) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	if rate, err := r.getUSDRate(currencyMint); err == nil {
		return rate, nil
	}

	if err := r.jupiterUpdateRateAndDecimals(ctx, currencyMint); err != nil {
		return 0, err
	}

	return r.getUSDRate(currencyMint)
}

func (r *rateStorage) GetSOLRate(ctx context.Context, currencyMint solana.PublicKey) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	if currencyMint.Equals(solana.SolMint) {
		return 1, nil
	}

	if rate, err := r.getSOLRate(currencyMint); err == nil {
		return rate, nil
	}

	if err := r.jupiterUpdateRateAndDecimals(ctx, currencyMint); err != nil {
		return 0, err
	}

	return r.getSOLRate(currencyMint)
}

func (r *rateStorage) GetDecimal(ctx context.Context, currencyMint solana.PublicKey) (uint8, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	if decimal, err := r.getDecimal(currencyMint); err == nil {
		return decimal, nil
	}

	if err := r.jupiterUpdateRateAndDecimals(ctx, currencyMint); err != nil {
		return 0, err
	}

	return r.getDecimal(currencyMint)
}

func (r *rateStorage) GetUSDRates(ctx context.Context, currencyMints ...solana.PublicKey) (map[solana.PublicKey]float64, error) {
	return getResult(r, ctx, currencyMints, r.getUSDRate)
}

func (r *rateStorage) GetSOLRates(ctx context.Context, currencyMints ...solana.PublicKey) (map[solana.PublicKey]float64, error) {
	return getResult(r, ctx, currencyMints, r.getSOLRate)
}

func (r *rateStorage) GetDecimals(ctx context.Context, currencyMints ...solana.PublicKey) (map[solana.PublicKey]uint8, error) {
	return getResult(r, ctx, currencyMints, r.getDecimal)
}

func getResult[T any](r *rateStorage, ctx context.Context, currencyMints []solana.PublicKey, getter func(solana.PublicKey) (T, error)) (map[solana.PublicKey]T, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	resultMap := make(map[solana.PublicKey]T, len(currencyMints))
	var absentMints []solana.PublicKey

	for _, mint := range currencyMints {
		if result, err := getter(mint); err == nil {
			resultMap[mint] = result
		} else {
			absentMints = append(absentMints, mint)
		}
	}

	if len(absentMints) == 0 {
		return resultMap, nil
	}

	type provider struct {
		name      string
		chunkSize int
		updateFn  func(context.Context, ...solana.PublicKey) error
	}

	providers := []provider{
		{"Jupiter", 49, r.jupiterUpdateRateAndDecimals},
		{"Solana", 100, r.solanaUpdateRateAndDecimals},
	}

	errs := make([]error, len(providers))

	var mu sync.Mutex

	for index, p := range providers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		eg, errCtx := errgroup.WithContext(ctx)
		eg.SetLimit(1)

		var nextAbsent []solana.PublicKey
		var nextAbsentMu sync.Mutex

		for chunk := range slices.Chunk(absentMints, p.chunkSize) {
			mintsChunk := chunk

			fmt.Printf("STARTING CHUNK FOR %s\n", p.name)
			eg.Go(func() error {
				if err := p.updateFn(errCtx, mintsChunk...); err != nil {
					fmt.Printf("ERROR UPDATING CHUNK (%s): %v\n", p.name, err)
					return err
				}

				for _, mint := range mintsChunk {
					if result, err := getter(mint); err == nil {
						mu.Lock()
						resultMap[mint] = result
						mu.Unlock()
					} else {
						fmt.Printf("ERROR GETTING CHUNK (%s): %v, mint: %s\n", p.name, err, mint)
						nextAbsentMu.Lock()
						nextAbsent = append(nextAbsent, mint)
						nextAbsentMu.Unlock()
					}
				}
				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			errs[index] = err
		}

		absentMints = nextAbsent
		if len(absentMints) == 0 {
			break
		}
	}

	if len(absentMints) > 0 {
		fmt.Println("DECIMALS IN FUNC", resultMap)
		return nil, errors.New("some mints were not found")
	}

	for _, err := range errs {
		if err == nil {
			return resultMap, nil
		}
	}

	return nil, errors.Join(errs...)
}

func (r *rateStorage) getFromCache(cache *gocache.Cache, key, backupKey string) (any, bool) {
	if val, ok := cache.Get(key); ok && val != nil {
		return val, true
	}
	if val, ok := cache.Get(backupKey); ok && val != nil {
		return val, true
	}
	return nil, false
}

func (r *rateStorage) saveToCache(mint solana.PublicKey, rate float64, decimals int) {
	mintStr := mint.String()

	r.rateCache.Set(fmt.Sprintf("%s_rate", mintStr), rate, gocache.DefaultExpiration)
	r.rateCache.Set(fmt.Sprintf("%s_rate_backup", mintStr), rate, r.cacheTTL*3)

	r.decimalCache.Set(fmt.Sprintf("%s_decimals", mintStr), decimals, gocache.NoExpiration)
	r.decimalCache.Set(fmt.Sprintf("%s_decimals_backup", mintStr), decimals, gocache.NoExpiration)
}

func (r *rateStorage) getUSDRate(currencyMint solana.PublicKey) (float64, error) {

	val, ok := r.getFromCache(r.rateCache, fmt.Sprintf("%s_rate", solana.SolMint.String()), fmt.Sprintf("%s_rate_backup", solana.SolMint.String()))
	if !ok {
		return 0, ErrRateNotFound
	}

	solRate := val.(float64)

	mintStr := currencyMint.String()

	if currencyMint.Equals(solana.SolMint) {
		return solRate, nil
	}

	if val, ok = r.getFromCache(r.rateCache, fmt.Sprintf("%s_rate", mintStr), fmt.Sprintf("%s_rate_backup", mintStr)); ok {
		rate := val.(float64)
		return solanarpc.TokenSOLToUSD(rate, solRate), nil
	}

	return 0, ErrRateNotFound
}

func (r *rateStorage) getSOLRate(currencyMint solana.PublicKey) (float64, error) {
	mintStr := currencyMint.String()
	if val, ok := r.getFromCache(r.rateCache, fmt.Sprintf("%s_rate", mintStr), fmt.Sprintf("%s_rate_backup", mintStr)); ok {
		return val.(float64), nil
	}
	return 0, ErrRateNotFound
}

func (r *rateStorage) getDecimal(currencyMint solana.PublicKey) (uint8, error) {
	mintStr := currencyMint.String()
	if val, ok := r.getFromCache(r.decimalCache, fmt.Sprintf("%s_decimals", mintStr), fmt.Sprintf("%s_decimals_backup", mintStr)); ok {
		return uint8(val.(int)), nil
	}
	return 0, ErrDecimalNotFound
}

func (r *rateStorage) jupiterUpdateRateAndDecimals(ctx context.Context, mints ...solana.PublicKey) error {
	fmt.Println("Jupiter Mints:", mints)
	result, err := jupiterGetRateAndDecimals(ctx, r.url, mints...)
	if err != nil {
		return err
	}

	isMintSOL := slices.Contains(mints, solana.SolMint)

	fmt.Println("JUPITER IS SOL", isMintSOL)

	var usdSolRate float64

	if !isMintSOL {
		fmt.Println("JUPITER IS NOT SOL")
		usdSolRate, err = r.GetUSDRate(ctx, solana.SolMint)
		if err != nil {
			return err
		}
	}

	for _, mint := range mints {
		usdPrice := result[mint.String()]["usdPrice"]
		decimals := result[mint.String()]["decimals"]

		if usdPrice == nil || decimals == nil {
			continue
		}

		usdPriceFloat, uOk := usdPrice.(float64)
		decimalsFloat, dOk := decimals.(float64)

		if !uOk || !dOk {
			continue
		}

		fmt.Println("JUPITER USD PRICE FLOAT", usdPriceFloat)
		fmt.Println("JUPITER DECIMALS FLOAT", decimalsFloat)

		if usdPriceFloat == 0 || decimalsFloat == 0 {
			continue
		}

		if isMintSOL {
			fmt.Println("JUPITER IS SOL")
			r.saveToCache(mint, usdPriceFloat, int(solana.SolDecimals))
		} else {
			solPriceFloat := solanarpc.TokenToSOLPrice(usdPriceFloat, usdSolRate)
			r.saveToCache(mint, solPriceFloat, int(decimalsFloat))
		}
	}
	return nil
}

func (r *rateStorage) coingeckoUpdateRateAndDecimals(ctx context.Context, mints ...solana.PublicKey) error {

	result, err := geckoGetRateAndDecimals(ctx, mints...)
	if err != nil {
		return err
	}

	usdSolRate, err := r.GetUSDRate(ctx, solana.SolMint)
	if err != nil {
		return err
	}

	for _, mint := range mints {
		fmt.Println("COINGECKO", result[mint.String()])
		usdPrice := result[mint.String()]["usdPrice"]
		decimals := result[mint.String()]["decimals"]

		fmt.Println("COINGECKO USD PRICE", usdPrice)
		fmt.Println("COINGECKO DECIMALS", decimals)

		if usdPrice == nil || decimals == nil {
			continue
		}

		usdPriceFloat, uOk := usdPrice.(float64)
		decimalsFloat, dOk := decimals.(int)

		if !uOk || !dOk {
			fmt.Println("COINGECKO Failed", usdPrice)
			fmt.Println("COINGECKO USD PRICE FLOAT", usdPriceFloat)
			fmt.Println("COINGECKO DECIMALS FLOAT", decimals.(int))

			continue
		}

		fmt.Println("COINGECKO USD PRICE FLOAT", usdPriceFloat)
		fmt.Println("COINGECKO DECIMALS FLOAT", decimalsFloat)

		if usdPriceFloat == 0 || decimalsFloat == 0 {
			continue
		}

		solPriceFloat := solanarpc.TokenToSOLPrice(usdPriceFloat, usdSolRate)

		r.saveToCache(mint, solPriceFloat, int(decimalsFloat))
	}
	return nil
}

func (r *rateStorage) solanaUpdateRateAndDecimals(ctx context.Context, mints ...solana.PublicKey) error {
	fmt.Println("Solana Mints:", mints)
	result, err := r.rpc.GetMultipleAccounts(ctx, mints...)
	if err != nil {
		return err
	}

	for index, mint := range mints {
		value := result.Value[index]
		data := value.Data.GetBinary()
		if data == nil {
			return errors.New("data is nil")
		}

		tokenMint := token.Mint{}
		if err = tokenMint.UnmarshalWithDecoder(bin.NewBinDecoder(data)); err != nil {
			return err
		}

		r.saveToCache(mint, 0.0, int(tokenMint.Decimals))
	}
	return nil
}
