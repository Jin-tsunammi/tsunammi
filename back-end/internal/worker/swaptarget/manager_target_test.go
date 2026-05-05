package swaptarget

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"mm/internal/model"
	"mm/internal/swapbudget"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type campaignRepoStub struct {
	mu             sync.Mutex
	updateCalls    int
	lastStatus     string
	lastCampaignID uuid.UUID
}

func (s *campaignRepoStub) UpdateStatusByID(_ context.Context, status string, campaignID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateCalls++
	s.lastStatus = status
	s.lastCampaignID = campaignID
	return nil
}

func (s *campaignRepoStub) UpdateDoneIfNoPendingTransactions(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

func TestRunTarget_SetsErrorStatusWhenExecuteTargetFails(t *testing.T) {
	t.Parallel()

	repo := &campaignRepoStub{}
	campaignID := uuid.New()

	manager := &SwapTargetManager{
		logger:             zap.NewNop(),
		data:               make(map[uuid.UUID]*concurrentSwapTask),
		stop:               make(map[uuid.UUID]chan struct{}),
		slots:              make(map[uint64]chan struct{}),
		mu:                 &sync.RWMutex{},
		muSlot:             &sync.RWMutex{},
		muStop:             &sync.RWMutex{},
		campaignRepository: repo,
	}

	remainingBudget := swapbudget.NewSwapBudget(big.NewInt(1))

	manager.runTarget(campaignID, 1, time.Second, 2*time.Second, remainingBudget)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	require.Equal(t, 1, repo.updateCalls)
	require.Equal(t, model.StatusError, repo.lastStatus)
	require.Equal(t, campaignID, repo.lastCampaignID)
}

func TestChunkRotatesTasks(t *testing.T) {
	t.Parallel()

	manager := &SwapTargetManager{
		computeUnitLimit: 200_000,
	}
	wallets := []*solana.Wallet{
		solana.NewWallet(),
		solana.NewWallet(),
		solana.NewWallet(),
	}
	concurrentTask := &concurrentSwapTask{
		mu: &sync.RWMutex{},
		tasks: []*model.AsyncSwapTask{
			newTestSwapTask(wallets[0].PrivateKey),
			newTestSwapTask(wallets[1].PrivateKey),
			newTestSwapTask(wallets[2].PrivateKey),
		},
	}
	stop := make(chan struct{})

	_, _, firstTasks, err := manager.chunk(1, 4, stop, concurrentTask)
	require.NoError(t, err)
	_, _, secondTasks, err := manager.chunk(1, 4, stop, concurrentTask)
	require.NoError(t, err)
	_, _, thirdTasks, err := manager.chunk(1, 4, stop, concurrentTask)
	require.NoError(t, err)

	require.Equal(t, wallets[0].PublicKey(), firstTasks[0][0].PrivateKey.PublicKey())
	require.Equal(t, wallets[1].PublicKey(), secondTasks[0][0].PrivateKey.PublicKey())
	require.Equal(t, wallets[2].PublicKey(), thirdTasks[0][0].PrivateKey.PublicKey())
}

func newTestSwapTask(privateKey solana.PrivateKey) *model.AsyncSwapTask {
	return &model.AsyncSwapTask{
		MinTransactionsAmount: 1,
		MaxTransactionsAmount: 1,
		SourceTokenDecimals:   solana.SolDecimals,
		PrivateKey:            privateKey,
	}
}
