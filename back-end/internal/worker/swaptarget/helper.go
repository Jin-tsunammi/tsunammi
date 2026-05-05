package swaptarget

import (
	"context"
	"mm/internal/client/jito"
	"mm/internal/client/raydium"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/swaperror"
	"mm/pkg/apperrors"
	"sync/atomic"
	"time"

	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (m *SwapTargetManager) addData(campaignID uuid.UUID, task *concurrentSwapTask) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[campaignID] = task
}

func (m *SwapTargetManager) getData(campaignID uuid.UUID) (*concurrentSwapTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	task, ok := m.data[campaignID]
	if !ok {
		return nil, apperrors.Internal("target not found")
	}
	return task, nil
}

func (m *SwapTargetManager) removeData(campaignID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, campaignID)
}

func (m *SwapTargetManager) addSlot(slot uint64) {
	m.muSlot.Lock()
	defer m.muSlot.Unlock()
	m.slots[slot] = make(chan struct{})
}

// notifySlotsUpTo closes wait channels for all slots up to slot.
func (m *SwapTargetManager) notifySlotsUpTo(slot uint64) {
	m.muSlot.Lock()
	defer m.muSlot.Unlock()

	for waitingSlot, ch := range m.slots {
		if waitingSlot <= slot {
			close(ch)
			delete(m.slots, waitingSlot)
		}
	}
}

func (m *SwapTargetManager) getSlotChan(slot uint64) (chan struct{}, error) {
	m.muSlot.RLock()
	defer m.muSlot.RUnlock()
	channel, ok := m.slots[slot]
	if !ok {
		return nil, apperrors.Internal("slot not found")
	}
	return channel, nil
}

func (m *SwapTargetManager) addStop(campaignID uuid.UUID) {
	m.muStop.Lock()
	defer m.muStop.Unlock()
	if _, ok := m.stop[campaignID]; !ok {
		m.stop[campaignID] = make(chan struct{})
	}
}

func (m *SwapTargetManager) removeStop(campaignID uuid.UUID) {
	m.muStop.Lock()
	defer m.muStop.Unlock()
	if ch, exists := m.stop[campaignID]; exists {
		close(ch)
		delete(m.stop, campaignID)
	}
}

func (m *SwapTargetManager) getStopChan(campaignID uuid.UUID) (chan struct{}, error) {
	m.muStop.RLock()
	defer m.muStop.RUnlock()
	channel, ok := m.stop[campaignID]
	if !ok {
		return nil, apperrors.Internal("stop channel not found")
	}
	return channel, nil

}

func (m *SwapTargetManager) listenAndUpdateCurrentSlot(ctx context.Context) error {
	if m.solanaRPC == nil {
		return nil
	}

	backoff := 500 * time.Millisecond
	for {
		slotUpdate, err := m.solanaWs.SubscribeToSlotUpdate(ctx)
		if err != nil {
			m.logger.Error("failed to subscribe to slot updates", zap.Error(err))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			backoff *= 2
			if backoff > 10*time.Second {
				backoff = 10 * time.Second
			}
			continue
		}

		backoff = 500 * time.Millisecond
	recvLoop:
		for {
			var result *ws.SlotResult

			select {
			case <-ctx.Done():
				slotUpdate.Unsubscribe()
				return ctx.Err()
			default:
				result, err = slotUpdate.Recv(ctx)
				if err != nil {
					if ctx.Err() != nil {
						slotUpdate.Unsubscribe()
						return ctx.Err()
					}

					m.logger.Error("slotUpdate.Recv failed", zap.Error(err))
					atomic.AddUint64(&m.currentSlot, 1)
					slotUpdate.Unsubscribe()
					break recvLoop
				}
				m.notifySlotsUpTo(result.Slot)
				m.logger.Debug("current slot ", zap.Uint64("slot", result.Slot))
				atomic.StoreUint64(&m.currentSlot, result.Slot)
			}
		}
	}
}

func (m *SwapTargetManager) updateLatestBlockhash(ctx context.Context, blockhashInterval time.Duration) error {

	ticker := time.NewTicker(blockhashInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			blockhash, err := m.solanaRPC.GetLatestBlockhash(ctx)
			if err != nil {
				continue
			}

			m.latestBlockhash.Store(blockhash)
		}
	}
}

func (m *SwapTargetManager) cleanLatestBlockhash(ctx context.Context) error {

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			m.latestBlockhash.Store(nil)
		}
	}

}

func (m *SwapTargetManager) chunk(parallelTransactionsAmount, batchSize int, stop <-chan struct{}, concurrentTask *concurrentSwapTask) (configs [][]raydium.TWAPConfig, params [][]raydium.SwapParams, tasks [][]*model.AsyncSwapTask, err error) {
	size := parallelTransactionsAmount/batchSize + 1

	configs = make([][]raydium.TWAPConfig, 0, size)
	params = make([][]raydium.SwapParams, 0, size)
	tasks = make([][]*model.AsyncSwapTask, 0, size)

	currentConfig := make([]raydium.TWAPConfig, 0, jito.BundleLimit)
	currentParam := make([]raydium.SwapParams, 0, jito.BundleLimit)
	currentTask := make([]*model.AsyncSwapTask, 0, jito.BundleLimit)

	for i := 0; i < parallelTransactionsAmount; i++ {
		select {
		case <-stop:
			return nil, nil, nil, targetStoppedError
		default:
			task := func() *model.AsyncSwapTask {
				concurrentTask.mu.Lock()
				defer concurrentTask.mu.Unlock()

				if len(concurrentTask.tasks) == 0 {
					return nil
				}

				index := concurrentTask.nextTaskIndex % len(concurrentTask.tasks)
				concurrentTask.nextTaskIndex = (concurrentTask.nextTaskIndex + 1) % len(concurrentTask.tasks)

				return concurrentTask.tasks[index]
			}()
			if task == nil {
				return nil, nil, nil, swaperror.ErrInsufficientFunds
			}

			currentConfig = append(currentConfig, raydium.TWAPConfig{
				MinTransactionsAmount:         solanarpc.ToAtomicUnit(task.MinTransactionsAmount, task.SourceTokenDecimals),
				MaxTransactionsAmount:         solanarpc.ToAtomicUnit(task.MaxTransactionsAmount, task.SourceTokenDecimals),
				SlippageBPS:                   task.Slippage,
				ComputeUnitLimit:              m.computeUnitLimit,
				ComputeUnitPriceMicroLamports: task.PriorityFeeMLP,
			})

			currentParam = append(currentParam, raydium.SwapParams{
				UserWallet: task.PrivateKey.PublicKey(),
				PoolID:     task.PoolID,

				InputTokenMint:  task.SourceTokenMint,
				OutputTokenMint: task.DestTokenMint,
				UserSourceToken: task.SourceAddress,
				UserDestToken:   task.DestAddress,
			})

			t := task

			currentTask = append(currentTask, t)

			if len(currentTask) == batchSize {
				configs = append(configs, currentConfig)
				params = append(params, currentParam)
				tasks = append(tasks, currentTask)
				currentConfig = make([]raydium.TWAPConfig, 0, jito.BundleLimit)
				currentParam = make([]raydium.SwapParams, 0, jito.BundleLimit)
				currentTask = make([]*model.AsyncSwapTask, 0, jito.BundleLimit)

			}
		}
	}

	if len(currentTask) > 0 {
		configs = append(configs, currentConfig)
		params = append(params, currentParam)
		tasks = append(tasks, currentTask)
	}

	return configs, params, tasks, nil
}
