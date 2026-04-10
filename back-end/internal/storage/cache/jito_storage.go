package cache

import (
	"context"
	"errors"
	"mm/config"
	"mm/internal/client/jito"
	"mm/internal/client/solanarpc"
	"time"

	"github.com/gagliardetto/solana-go"
	imcache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var _ JitoValidatorStorage = (*jitoValidatorStorage)(nil)

type JitoValidatorStorage interface {
	Get(ctx context.Context) (map[solana.PublicKey]struct{}, error)
}

type jitoValidatorStorage struct {
	cache     *imcache.Cache
	jito      *jito.Client
	solanaRPC solanarpc.SolanaRPC
	logger    *zap.Logger
	ttl       time.Duration
}

func NewJitoValidatorStorage(cfg *config.Config, jito *jito.Client, solanaRPC solanarpc.SolanaRPC, logger *zap.Logger) JitoValidatorStorage {
	cacheTTL := cfg.Jito.ValidatorInterval

	cache := imcache.New(cacheTTL, cacheTTL)

	storage := jitoValidatorStorage{cache: cache, jito: jito, solanaRPC: solanaRPC, logger: logger, ttl: cacheTTL}

	cache.OnEvicted(func(key string, value interface{}) {
		_ = storage.update(context.Background(), cacheTTL)
	})

	return &storage
}

func (s *jitoValidatorStorage) Get(ctx context.Context) (map[solana.PublicKey]struct{}, error) {

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		data, ok := s.cache.Get("validators")
		if !ok {
			data, ok = s.cache.Get("validators_backup")
			if !ok {
				err := s.update(ctx, s.ttl)
				if err != nil {
					return nil, err
				}
				data, ok = s.cache.Get("validators")
				if !ok {
					return nil, errors.New("validators not found")
				}
			}
		}

		keys, ok := data.(map[solana.PublicKey]struct{})

		if !ok {
			return nil, errors.New("can't cast validators to []solana.PublicKey")
		}

		return keys, nil
	}
}

func (s *jitoValidatorStorage) update(ctx context.Context, ttl time.Duration) error {
	var err error

	defer func() {
		if err != nil {
			s.logger.Error("error update validators", zap.Error(err))
		}
	}()

	validators, err := s.jito.GetValidatorsCurrentEpoch(ctx)
	if err != nil {
		return err
	}

	result := make(map[solana.PublicKey]struct{})
	errs := make([]error, 0, len(validators))

	var key solana.PublicKey

	for _, validator := range validators {
		key, err = solana.PublicKeyFromBase58(validator.IdentityAccount)

		if err != nil {
			errs = append(errs, err)
		} else {
			result[key] = struct{}{}
		}
	}

	s.cache.Set("validators", result, ttl)
	s.cache.Set("validators_backup", result, ttl*12/10)

	return nil
}
