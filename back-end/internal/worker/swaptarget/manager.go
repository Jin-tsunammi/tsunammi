package swaptarget

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"mm/config"
	"mm/internal/client/jito"
	pumpAMM "mm/internal/client/pumpfun/amm"
	pumpBonding "mm/internal/client/pumpfun/bonding"
	"mm/internal/client/raydium"
	"mm/internal/client/raydium/ammv4"
	raydiumamm "mm/internal/client/raydium/ammv4/ammv4_client"
	"mm/internal/client/raydium/cpmm"
	raydiumcpswap "mm/internal/client/raydium/cpmm/cpmm_client"
	"mm/internal/client/solanarpc"
	"mm/internal/client/solanaws"
	"mm/internal/model"
	"mm/internal/storage/cache"
	"mm/internal/storage/repository"
	"mm/internal/swapbudget"
	"mm/internal/swaperror"
	"mm/internal/swaptxlog"
	"mm/pkg/apperrors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	transactionsLimit = 45 * time.Second
)

type SwapTargetManager struct {
	logger             *zap.Logger
	raydiumCPMMClient  *cpmm.Client
	raydiumAMMClient   *ammv4.Client
	pumpAMMClient      *pumpAMM.Client
	pumpBondingClient  *pumpBonding.Client
	solanaRPC          solanarpc.SolanaRPC
	solanaWs           *solanaws.Client
	schedule           cache.JitoScheduleStorage
	decimalCache       cache.RateStorage
	data               map[uuid.UUID]*concurrentSwapTask
	stop               map[uuid.UUID]chan struct{}
	slots              map[uint64]chan struct{}
	closeCh            chan struct{}
	mu                 *sync.RWMutex
	muSlot             *sync.RWMutex
	muStop             *sync.RWMutex
	currentSlot        uint64
	latestBlockhash    atomic.Pointer[solana.Hash]
	padding            uint64
	computeUnitLimit   uint32
	campaignRepository campaignRepository
	transactionRepo    *repository.SwapTransactionRepository
}

type campaignRepository interface {
	UpdateStatusByID(ctx context.Context, status model.SwapStatus, campaignID uuid.UUID) error
	UpdateDoneIfNoPendingTransactions(ctx context.Context, campaignID uuid.UUID) (bool, error)
}

func NewSwapTargetManager(logger *zap.Logger, raydiumCPMMClient *cpmm.Client, raydiumAMMClient *ammv4.Client, pumpAMMClient *pumpAMM.Client, pumpBondingClient *pumpBonding.Client, solanaRPC solanarpc.SolanaRPC, schedule cache.JitoScheduleStorage, cfg *config.Config, solanaWs *solanaws.Client, campaignRepo *repository.SwapCampaignRepository, transactionRepo *repository.SwapTransactionRepository, decimalCache cache.RateStorage) *SwapTargetManager {
	return &SwapTargetManager{
		logger:             logger,
		raydiumCPMMClient:  raydiumCPMMClient,
		raydiumAMMClient:   raydiumAMMClient,
		pumpAMMClient:      pumpAMMClient,
		pumpBondingClient:  pumpBondingClient,
		solanaRPC:          solanaRPC,
		solanaWs:           solanaWs,
		schedule:           schedule,
		decimalCache:       decimalCache,
		data:               make(map[uuid.UUID]*concurrentSwapTask, 100),
		stop:               make(map[uuid.UUID]chan struct{}, 100),
		slots:              make(map[uint64]chan struct{}, 100),
		mu:                 &sync.RWMutex{},
		muSlot:             &sync.RWMutex{},
		muStop:             &sync.RWMutex{},
		closeCh:            make(chan struct{}),
		currentSlot:        0,
		padding:            cfg.Jito.SlotPadding,
		computeUnitLimit:   cfg.App.ComputeUnitLimit,
		campaignRepository: campaignRepo,
		transactionRepo:    transactionRepo,
	}
}

func (m *SwapTargetManager) AddTarget(
	ctx context.Context,
	minTimeBetweenTransactions, maxTimeBetweenTransactions time.Duration,
	campaignID uuid.UUID,
	parallelTransactionsAmount int,
	remainingBudget *swapbudget.SwapBudget,
	data []*model.AsyncSwapTask,
) error {
	concurrentTask := &concurrentSwapTask{
		tasks:           data,
		mu:              &sync.RWMutex{},
		remainingBudget: remainingBudget,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.addData(campaignID, concurrentTask)
		m.addStop(campaignID)
	}

	go func() {
		m.runTarget(campaignID, parallelTransactionsAmount, minTimeBetweenTransactions, maxTimeBetweenTransactions, remainingBudget)
	}()

	return nil

}

func (m *SwapTargetManager) runTarget(campaignID uuid.UUID, parallelTransactionsAmount int, minTimeBetweenTransactions, maxTimeBetweenTransactions time.Duration, remainingBudget *swapbudget.SwapBudget) {
	err := m.executeTarget(campaignID, 0, parallelTransactionsAmount, minTimeBetweenTransactions, maxTimeBetweenTransactions, remainingBudget)

	m.removeData(campaignID)
	m.removeStop(campaignID)

	ctx := context.Background()

	switch {
	case err == nil:
		if statusErr := m.campaignRepository.UpdateStatusByID(ctx, model.SwapStatusTargetCompleted, campaignID); statusErr != nil {
			m.logger.Error("failed to update campaign status", zap.Error(statusErr))
		}
	case errors.Is(err, targetStoppedError):
		if statusErr := m.campaignRepository.UpdateStatusByID(ctx, model.SwapStatusStop, campaignID); statusErr != nil {
			m.logger.Error("failed to update campaign status", zap.Error(statusErr))
		}
	case errors.Is(err, swaperror.BudgetExceededError):
		if statusErr := m.campaignRepository.UpdateStatusByID(ctx, model.SwapStatusBudgetDone, campaignID); statusErr != nil {
			m.logger.Error("failed to update campaign status", zap.Error(statusErr))
		}
	case errors.Is(err, swaperror.ErrInsufficientFunds), errors.Is(err, pumpBonding.NotEnoughTokensToSellError):
		if statusErr := m.campaignRepository.UpdateStatusByID(ctx, model.SwapStatusInsufficientFunds, campaignID); statusErr != nil {
			m.logger.Error("failed to update campaign status", zap.Error(statusErr))
		}
	default:
		m.logger.Error("execute target failed", zap.Error(err))
		if statusErr := m.campaignRepository.UpdateStatusByID(ctx, model.SwapStatusError, campaignID); statusErr != nil {
			m.logger.Error("failed to update campaign status", zap.Error(statusErr))
		}
	}
}

func (m *SwapTargetManager) DeleteTarget(ctx context.Context, campaignID uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.removeData(campaignID)
		m.removeStop(campaignID)
		return nil
	}

}

func (m *SwapTargetManager) controlThread(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	rpcWithRetries := solanarpc.WithRetries(m.solanaRPC, 3)

	slot, err := rpcWithRetries.GetCurrentSlot(ctx)

	if err != nil {
		return
	}

	atomic.StoreUint64(&m.currentSlot, slot)

	eg, errctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return m.listenAndUpdateCurrentSlot(errctx)
	})

	eg.Go(func() error {
		return m.updateLatestBlockhash(ctx, cfg.App.BlockhashInterval)
	})

	waitFinished := make(chan error, 1)

	go func() {
		if err := eg.Wait(); err != nil {
			waitFinished <- err
		}
	}()

	select {
	case <-m.closeCh:
		return
	case err = <-waitFinished:
		if err != nil {
			m.logger.Error("control thread error", zap.Error(err))
		}
	}

}

func (m *SwapTargetManager) close() {
	close(m.closeCh)
}

func (m *SwapTargetManager) executeTarget(
	campaignID uuid.UUID,
	slot uint64,
	parallelTransactionsAmount int,
	minTimeBetweenTransactions,
	maxTimeBetweenTransactions time.Duration,
	remainingBudget *swapbudget.SwapBudget,
) (err error) {
	const batchSize = jito.BundleLimit - 1
	nextSlot := slot

	for {
		concurrentTask, taskErr := m.getTaskByCampaignID(campaignID)
		if taskErr != nil {
			return taskErr
		}
		useJito := len(concurrentTask.tasks) > 0 && concurrentTask.tasks[0].UsingJito

		m.logger.Info("starting target", zap.Uint64("slot", nextSlot))
		stop, err := m.getStopChan(campaignID)
		if err != nil {
			return err
		}

		if nextSlot == 0 && useJito {
			var targetSlot uint64

			schedule, sErr := m.schedule.GetSchedule(context.Background())
			if sErr != nil {
				return sErr
			}

			targetSlot = m.getJitoSlot(schedule, m.currentSlot+m.padding*2+1, m.currentSlot+1000)

			m.logger.Info("target slot when current slot > target slot", zap.Uint64("slot", targetSlot))

			if targetSlot != 0 {
				m.addSlot(targetSlot)

				slotChan, err := m.getSlotChan(targetSlot)
				if err != nil {
					return err
				}
				m.logger.Info("waiting for assigned slot", zap.Uint64("slot", targetSlot))

				select {
				case <-stop:
					err = targetStoppedError
				case <-slotChan:
					m.logger.Info("assigned slot", zap.Uint64("slot", targetSlot))
				}

				if err != nil {
					m.logger.Info("cycle finished", zap.Uint64("slot", nextSlot), zap.Any("err", err))
					return err
				}
			}
		} else if nextSlot > m.currentSlot {
			channel, chErr := m.getSlotChan(nextSlot)
			if chErr != nil {
				m.addSlot(nextSlot)

				channel, chErr = m.getSlotChan(nextSlot)
				if chErr != nil {
					return chErr
				}
			}

			select {
			case <-stop:
				err = targetStoppedError
			case <-channel:
			}

			if err != nil {
				m.logger.Info("cycle finished", zap.Uint64("slot", nextSlot), zap.Any("err", err))
				return err
			}
		} else {
			m.logger.Info("slot already reached or passed, executing immediately", zap.Uint64("requested_slot", nextSlot), zap.Uint64("current_slot", m.currentSlot))
		}

		select {
		case <-stop:
			err = targetStoppedError
		default:
		}
		if err != nil {
			m.logger.Info("cycle finished", zap.Uint64("slot", nextSlot), zap.Any("err", err))
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), transactionsLimit)

		configs, params, tasks, err := m.chunk(parallelTransactionsAmount, batchSize, stop, concurrentTask)
		if err != nil {
			cancel()
			m.logger.Info("cycle finished", zap.Uint64("slot", nextSlot), zap.Any("err", err))
			return err
		}

		go func() {
			<-stop
			cancel()
		}()

		errs := make([]error, len(configs))
		wg := &sync.WaitGroup{}
		semaphore := make(chan struct{}, 10)

		wg.Add(len(configs))

		for i := range configs {
			go func(idx int) {
				defer wg.Done()

				semaphore <- struct{}{}
				defer func() {
					<-semaphore
				}()

				var localErr error
				var results []swaptxlog.Result

				switch tasks[idx][0].PoolProgramID {
				case raydiumcpswap.ProgramID:
					results, localErr = m.raydiumCPMMClient.Swap(ctx, campaignID, nil, remainingBudget, tasks[idx], params[idx], configs[idx], &m.latestBlockhash)
				case raydiumamm.ProgramID:
					results, localErr = m.raydiumAMMClient.Swap(ctx, campaignID, nil, remainingBudget, tasks[idx], params[idx], configs[idx], &m.latestBlockhash)
				case pumpAMM.ProgramID:
					results, localErr = m.pumpAMMClient.Swap(ctx, campaignID, nil, remainingBudget, tasks[idx], params[idx], configs[idx], &m.latestBlockhash)
				case pumpBonding.ProgramID:
					results, localErr = m.pumpBondingClient.Swap(ctx, campaignID, nil, remainingBudget, tasks[idx], params[idx], configs[idx], &m.latestBlockhash)
				}

				for _, result := range results {
					if logErr := swaptxlog.LogSwapTransaction(ctx, result.Err, campaignID, nil, result.Params, m.transactionRepo, m.logger); logErr != nil {
						localErr = errors.Join(localErr, logErr)
					}
				}

				if localErr != nil {
					errs[idx] = localErr
				}
			}(i)
		}

		wg.Wait()
		cancel()

		err = nil
		for idx, swapErr := range errs {
			if errors.Is(swapErr, swaperror.ErrInsufficientFunds) || errors.Is(swapErr, pumpBonding.NotEnoughTokensToSellError) {
				if remainingTasks := concurrentTask.removeTasks(tasks[idx]); remainingTasks > 0 {
					m.logger.Info(
						"removed insufficiently funded swap tasks",
						zap.Int("removed", len(tasks[idx])),
						zap.Int("remaining", remainingTasks),
					)
					errs[idx] = nil
					continue
				}
				err = swaperror.ErrInsufficientFunds
				break
			}
			if errors.Is(swapErr, swaperror.BudgetExceededError) {
				err = swaperror.BudgetExceededError
				break
			}
		}

		if err == nil {
			if joinedErr := errors.Join(errs...); joinedErr != nil {
				m.logger.Error("failed to swap", zap.Error(joinedErr))
				err = apperrors.Internal("failed to swap", failedToSwapError, joinedErr)
			}
		}

		m.logger.Info("cycle finished", zap.Uint64("slot", nextSlot), zap.Any("err", err))

		if m.isTerminalTargetError(err) {
			if errors.Is(err, raydium.PriceIsAlreadyReachedError) {
				return nil
			}
			return err
		}
		if err == nil {
			budget := remainingBudget.Remaining()
			if budget == nil || budget.Sign() <= 0 {
				return swaperror.BudgetExceededError
			}
		}

		if useJito {
			targetSlot, scheduleErr := m.getNextJitoSlot(minTimeBetweenTransactions, maxTimeBetweenTransactions)
			if scheduleErr != nil {
				return scheduleErr
			}
			if targetSlot == 0 {
				return fmt.Errorf("invalid 0 schedule")
			}

			nextSlot = targetSlot - m.padding
			m.logger.Info("schedule to target slot", zap.Uint64("slot", targetSlot))
			continue
		}

		delay := minTimeBetweenTransactions
		if maxTimeBetweenTransactions > minTimeBetweenTransactions {
			delay += time.Duration(rand.Int63n(int64(maxTimeBetweenTransactions - minTimeBetweenTransactions)))
		}
		if delay <= 0 {
			delay = time.Second
		}

		timer := time.NewTimer(delay)
		select {
		case <-stop:
			timer.Stop()
			return targetStoppedError
		case <-timer.C:
			nextSlot = 0
		}
	}
}

func (m *SwapTargetManager) isTerminalTargetError(err error) bool {
	return errors.Is(err, raydium.PriceIsAlreadyReachedError) ||
		errors.Is(err, targetStoppedError) ||
		errors.Is(err, swaperror.BudgetExceededError) ||
		errors.Is(err, swaperror.ErrInsufficientFunds) ||
		errors.Is(err, pumpBonding.NotEnoughTokensToSellError)
}

func (m *SwapTargetManager) getNextJitoSlot(minTimeBetweenTransactions, maxTimeBetweenTransactions time.Duration) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rpcWithRetries := solanarpc.WithRetries(m.solanaRPC, 3)
	slotTimes, err := rpcWithRetries.GetAverageSlotTime(ctx)
	if err != nil {
		return 0, err
	}

	average5mSlotTime := slotTimes[5*time.Minute]

	minSlot := uint64(minTimeBetweenTransactions/average5mSlotTime) + m.currentSlot + 1
	maxSlot := uint64(maxTimeBetweenTransactions/average5mSlotTime) + m.currentSlot + 1

	schedule, err := m.schedule.GetSchedule(ctx)
	if err != nil {
		return 0, err
	}

	if targetSlot := m.getJitoSlot(schedule, minSlot, maxSlot); targetSlot != 0 {
		return targetSlot, nil
	}

	return m.getJitoSlot(schedule, maxSlot+1, maxSlot+1000), nil
}

func (m *SwapTargetManager) getJitoSlot(schedule map[uint64]model.JitoSchedule, start, end uint64) uint64 {

	var index uint64

	for index = start; index <= end; index++ {
		if block, ok := schedule[index]; ok {
			if block.RunningJito {
				return index
			}
		}
	}

	return 0

}

func (m *SwapTargetManager) getTaskByCampaignID(campaignID uuid.UUID) (*concurrentSwapTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.data[campaignID]
	if !ok {
		return nil, apperrors.Internal("target not found")
	}
	return task, nil
}

func (m *SwapTargetManager) UpdateTarget(campaignID uuid.UUID, campaign *model.SwapCampaign) (err error) {
	data, err := m.getData(campaignID)
	if err != nil {
		return err
	}

	data.mu.Lock()

	for index := range data.tasks {
		task := data.tasks[index]

		task.GoalPrice = campaign.GoalPrice.GetBigRat()

		task.Slippage = campaign.SlippageBPS

		task.MinTransactionsAmount = campaign.MinTransactionsBudget
		task.MaxTransactionsAmount = campaign.MaxTransactionsBudget

		task.TransactionSpeed = campaign.TransactionSpeed
	}

	mint, err := solana.PublicKeyFromBase58(campaign.TokenMintFrom)

	if err != nil {
		return err
	}

	decimals, err := m.decimalCache.GetDecimals(context.Background(), mint)

	if err != nil {
		return err
	}

	decimal, ok := decimals[mint]

	if !ok {
		return apperrors.Internal("mint not found")
	}

	data.remainingBudget.Store(new(big.Int).SetUint64(solanarpc.ToAtomicUnit(campaign.Budget, decimal)))

	return nil
}
