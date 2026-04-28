package buyback

import (
	"context"
	"mm/config"
	pump_amm "mm/internal/client/pumpfun/amm"
	pumpBonding "mm/internal/client/pumpfun/bonding"
	"mm/internal/client/raydium/ammv4"
	"mm/internal/client/raydium/cpmm"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/internal/storage/secret"
	buybackpricetracker "mm/internal/worker/buyback_price_tracker"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultManagerPollInterval = 5 * time.Second
	defaultCampaignTick        = 1 * time.Second
)

type campaignRuntime struct {
	cancel context.CancelFunc
}

type CampaignManager struct {
	logger            *zap.Logger
	buybackRepo       *repository.BuybackRepository
	buybackTxRepo     *repository.BuybackTransactionRepository
	projectRepo       *repository.ProjectRepository
	priceProvider     *buybackpricetracker.BuybackPriceTracker
	keyStorage        *secret.KeyStorage
	dexProviders      map[model.SwapProviderID]model.DexProvider
	raydiumCPMMClient *cpmm.Client
	raydiumAMMClient  *ammv4.Client
	pumpAMMClient     *pump_amm.Client
	pumpBondingClient *pumpBonding.Client

	solanaRPC    solanarpc.SolanaRPC
	pollInterval time.Duration
	campaignTick time.Duration

	mu      sync.Mutex
	workers map[uuid.UUID]campaignRuntime
	stopCh  chan struct{}
	once    sync.Once
}

func NewSmartBuybackManager(
	logger *zap.Logger,
	buybackCampaignRepo *repository.BuybackRepository,
	buybackTxRepo *repository.BuybackTransactionRepository,
	projectRepo *repository.ProjectRepository,
	keyStorage *secret.KeyStorage,
	raydiumCPMMClient *cpmm.Client,
	raydiumAMMClient *ammv4.Client,
	pumpAMMClient *pump_amm.Client,
	pumpBondingClient *pumpBonding.Client,
	priceTracker *buybackpricetracker.BuybackPriceTracker,
	solanaRPC solanarpc.SolanaRPC,
	dexProviders map[model.SwapProviderID]model.DexProvider,
	cfg *config.Config,
) *CampaignManager {
	poll := cfg.App.PriceMonitorInterval
	if poll <= 0 {
		poll = defaultManagerPollInterval
	}

	tick := poll / 2
	if tick <= 0 {
		tick = defaultCampaignTick
	}

	return &CampaignManager{
		logger:        logger,
		buybackRepo:   buybackCampaignRepo,
		buybackTxRepo: buybackTxRepo,
		projectRepo:   projectRepo,
		keyStorage:    keyStorage,
		solanaRPC:     solanaRPC,
		priceProvider: priceTracker,
		pollInterval:  poll,
		campaignTick:  tick,
		workers:       make(map[uuid.UUID]campaignRuntime),
		stopCh:        make(chan struct{}),
		dexProviders:  dexProviders,

		raydiumCPMMClient: raydiumCPMMClient,
		raydiumAMMClient:  raydiumAMMClient,
		pumpAMMClient:     pumpAMMClient,
		pumpBondingClient: pumpBondingClient,
	}
}

func (m *CampaignManager) Start(_ context.Context) error {
	m.logger.Info("buyback manager started", zap.Duration("poll", m.pollInterval), zap.Duration("tick", m.campaignTick))
	go m.loop()
	return nil
}

func (m *CampaignManager) Stop(_ context.Context) error {
	m.once.Do(func() {
		close(m.stopCh)
	})
	return nil
}

func (m *CampaignManager) loop() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	_ = m.syncCampaigns(context.Background())

	for {
		select {
		case <-m.stopCh:
			m.stopAllCampaigns()
			return
		case <-ticker.C:
			_ = m.syncCampaigns(context.Background())
		}
	}
}

func (m *CampaignManager) syncCampaigns(ctx context.Context) error {
	campaigns, err := m.buybackRepo.GetActiveWithTargets(ctx)
	if err != nil {
		m.logger.Warn("buyback manager sync failed", zap.Error(err))
		return err
	}

	active := make(map[uuid.UUID]struct{}, len(campaigns))
	for _, c := range campaigns {
		active[c.ID] = struct{}{}

		m.mu.Lock()
		_, exists := m.workers[c.ID]
		if !exists {
			runCtx, cancel := context.WithCancel(context.Background())
			m.workers[c.ID] = campaignRuntime{cancel: cancel}
			go m.runCampaign(runCtx, c.ID)
			m.logger.Info("campaign worker started", zap.String("campaign_id", c.ID.String()))
		}
		m.mu.Unlock()
	}

	m.mu.Lock()
	toStop := make([]campaignRuntime, 0)
	for id, rt := range m.workers {
		if _, ok := active[id]; ok {
			continue
		}
		toStop = append(toStop, rt)
		delete(m.workers, id)
		m.logger.Info("campaign worker stopped", zap.String("campaign_id", id.String()))
	}
	m.mu.Unlock()

	for _, rt := range toStop {
		rt.cancel()
	}

	return nil
}

func (m *CampaignManager) stopAllCampaigns() {
	m.mu.Lock()
	toStop := make([]campaignRuntime, 0, len(m.workers))
	for _, rt := range m.workers {
		toStop = append(toStop, rt)
	}
	m.workers = make(map[uuid.UUID]campaignRuntime)
	m.mu.Unlock()

	for _, rt := range toStop {
		rt.cancel()
	}
}

func (m *CampaignManager) runCampaign(ctx context.Context, campaignID uuid.UUID) {
	ticker := time.NewTicker(m.campaignTick)
	defer ticker.Stop()

	eval := func() {
		campaign, err := m.buybackRepo.GetActiveWithTargetsByID(context.Background(), campaignID)
		if err != nil {
			m.logger.Warn("failed to load campaign",
				zap.String("campaign_id", campaignID.String()),
				zap.Error(err),
			)
			return
		}
		if campaign.Status != model.SmartBuybackCampaignStatusActive {
			return
		}

		now := time.Now().UTC()
		selected := m.selectTargets(campaign, now)

		for _, target := range selected {
			err := m.dispatchBatch(ctx, campaign.SmartBuybackCampaign, target)
			if err != nil {
				m.logger.Warn("buyback batch dispatch failed",
					zap.String("campaign_id", campaign.ID.String()),
					zap.String("target_id", target.ID.String()),
					zap.Error(err),
				)
			}
		}

		m.logger.Info("campaign tick evaluated",
			zap.String("campaign_id", campaignID.String()),
			zap.Int("targets_total", len(campaign.Targets)),
			zap.Int("targets_selected", len(selected)),
		)
	}

	eval()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			eval()
		}
	}
}
