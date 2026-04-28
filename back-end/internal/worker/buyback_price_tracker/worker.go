package buybackpricetracker

import (
	"context"
	"mm/config"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/pricing"
	"mm/internal/storage/repository"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/zap"
)

type PoolPrice struct {
	Key       string    `json:"key"`
	PoolID    string    `json:"pool_id"`
	ProgramID string    `json:"program_id"`
	TokenIn   string    `json:"token_in"`
	TokenOut  string    `json:"token_out"`
	Price     string    `json:"price"`
	UpdatedAt time.Time `json:"updated_at"`
}

type priceQuery struct {
	Key      string
	PoolID   string
	TokenIn  string
	TokenOut string
}

type BuybackPriceTracker struct {
	logger   *zap.Logger
	rpc      solanarpc.SolanaRPC
	repo     *repository.BuybackRepository
	interval time.Duration

	mu     sync.RWMutex
	prices map[string]map[string]PoolPrice // program_id -> key -> price

	stopCh chan struct{}
	once   sync.Once
}

func NewBuybackPriceTracker(
	logger *zap.Logger,
	rpcClient solanarpc.SolanaRPC,
	buybackRepo *repository.BuybackRepository,
	cfg *config.Config,
) *BuybackPriceTracker {
	return &BuybackPriceTracker{
		logger:   logger,
		rpc:      rpcClient,
		repo:     buybackRepo,
		interval: cfg.App.PriceMonitorInterval,
		prices:   make(map[string]map[string]PoolPrice),
		stopCh:   make(chan struct{}),
	}
}

func (w *BuybackPriceTracker) Start(_ context.Context) error {
	go w.loop()
	return nil
}

func (w *BuybackPriceTracker) Stop(_ context.Context) error {
	w.once.Do(func() {
		close(w.stopCh)
	})
	return nil
}

func (w *BuybackPriceTracker) GetPrice(poolID, tokenIn, tokenOut string) (PoolPrice, bool, error) {
	key := composeTrackKey(poolID, tokenIn, tokenOut)

	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, pricesByKey := range w.prices {
		if price, ok := pricesByKey[key]; ok {
			return price, true, nil
		}
	}

	return PoolPrice{}, false, nil
}

func (w *BuybackPriceTracker) loop() {
	interval := w.interval
	if interval <= 0 {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	_ = w.refresh()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.refresh(); err != nil {
				w.logger.Warn("buyback price tracker refresh failed", zap.Error(err))
			}
		}
	}
}

func (w *BuybackPriceTracker) refresh() error {
	queries, err := w.loadQueries(context.Background())
	if err != nil {
		return err
	}

	if len(queries) == 0 {
		w.mu.Lock()
		w.prices = map[string]map[string]PoolPrice{}
		w.mu.Unlock()
		return nil
	}

	poolKeysByID := make(map[string]solana.PublicKey)
	orderedPoolIDs := make([]string, 0)
	orderedPoolKeys := make([]solana.PublicKey, 0)

	for _, q := range queries {
		if _, ok := poolKeysByID[q.PoolID]; ok {
			continue
		}

		poolPK, parseErr := solana.PublicKeyFromBase58(q.PoolID)
		if parseErr != nil {
			w.logger.Debug("skip invalid pool id", zap.String("pool_id", q.PoolID), zap.Error(parseErr))
			continue
		}

		poolKeysByID[q.PoolID] = poolPK
		orderedPoolIDs = append(orderedPoolIDs, q.PoolID)
		orderedPoolKeys = append(orderedPoolKeys, poolPK)
	}

	if len(orderedPoolKeys) == 0 {
		w.mu.Lock()
		w.prices = map[string]map[string]PoolPrice{}
		w.mu.Unlock()
		return nil
	}

	accountsRes, err := w.rpc.GetMultipleAccountsWithNoLimits(context.Background(), orderedPoolKeys...)
	if err != nil {
		return err
	}

	flatAccounts := make([]*rpc.Account, 0, len(orderedPoolKeys))
	for _, chunk := range accountsRes {
		flatAccounts = append(flatAccounts, chunk.Value...)
	}

	poolAccounts := make(map[string]*rpc.Account, len(orderedPoolIDs))
	for i, poolID := range orderedPoolIDs {
		if i < len(flatAccounts) {
			poolAccounts[poolID] = flatAccounts[i]
		}
	}

	next := make(map[string]map[string]PoolPrice)
	now := time.Now().UTC()

	for _, q := range queries {
		poolPK, ok := poolKeysByID[q.PoolID]
		if !ok {
			continue
		}

		poolAccount := poolAccounts[q.PoolID]
		if poolAccount == nil || poolAccount.Data == nil {
			continue
		}

		tokenInPK, parseInErr := solana.PublicKeyFromBase58(q.TokenIn)
		if parseInErr != nil {
			w.logger.Debug("skip invalid token_in mint", zap.String("token_in", q.TokenIn), zap.Error(parseInErr))
			continue
		}

		tokenOutPK, parseOutErr := solana.PublicKeyFromBase58(q.TokenOut)
		if parseOutErr != nil {
			w.logger.Debug("skip invalid token_out mint", zap.String("token_out", q.TokenOut), zap.Error(parseOutErr))
			continue
		}

		price, calcErr := pricing.CalculatePoolPrice(
			context.Background(),
			w.rpc,
			poolAccount,
			poolPK,
			tokenInPK,
			tokenOutPK,
		)
		if calcErr != nil {
			w.logger.Debug("failed to calculate pool price",
				zap.String("key", q.Key),
				zap.String("pool_id", q.PoolID),
				zap.Error(calcErr),
			)
			continue
		}

		programID := poolAccount.Owner.String()
		if _, ok := next[programID]; !ok {
			next[programID] = make(map[string]PoolPrice)
		}

		next[programID][q.Key] = PoolPrice{
			Key:       q.Key,
			PoolID:    q.PoolID,
			ProgramID: programID,
			TokenIn:   q.TokenIn,
			TokenOut:  q.TokenOut,
			Price:     price.RatString(),
			UpdatedAt: now,
		}
	}

	w.mu.Lock()
	w.prices = next
	w.mu.Unlock()

	return nil
}

func (w *BuybackPriceTracker) loadQueries(ctx context.Context) ([]priceQuery, error) {
	campaigns, err := w.repo.GetActiveWithTargets(ctx)
	if err != nil {
		return nil, err
	}

	queriesByKey := make(map[string]priceQuery)

	for _, campaign := range campaigns {
		for _, target := range campaign.Targets {
			tokenIn, tokenOut := getTargetDirectionMints(campaign.TokenMint, target.Type)
			key := composeTrackKey(campaign.PoolID, tokenIn, tokenOut)
			queriesByKey[key] = priceQuery{
				Key:      key,
				PoolID:   campaign.PoolID,
				TokenIn:  tokenIn,
				TokenOut: tokenOut,
			}
		}
	}

	queries := make([]priceQuery, 0, len(queriesByKey))
	for _, q := range queriesByKey {
		queries = append(queries, q)
	}

	return queries, nil
}

func composeTrackKey(poolID, tokenIn, tokenOut string) string {
	return poolID + "|" + tokenIn + "|" + tokenOut
}

func getTargetDirectionMints(tokenMint string, targetType model.BuybackCampaignTargetType) (tokenIn, tokenOut string) {
	switch targetType {
	case model.BuybackCampaignTargetTypeBuy:
		return solana.WrappedSol.String(), tokenMint
	default:
		return tokenMint, solana.WrappedSol.String()
	}
}
