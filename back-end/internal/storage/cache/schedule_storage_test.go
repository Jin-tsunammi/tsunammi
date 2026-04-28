package cache

import (
	"context"
	"testing"
	"time"

	"mm/internal/model"

	"github.com/gagliardetto/solana-go"
	imcache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestJitoScheduleStorageGetSchedule_UsesBackup(t *testing.T) {
	t.Parallel()

	cache := imcache.New(time.Minute, time.Minute)
	leader := solana.NewWallet().PublicKey()
	expected := map[uint64]model.JitoSchedule{
		123: {
			NodePublicKey: leader,
			RunningJito:   true,
		},
	}
	cache.Set("schedule_backup", expected, time.Minute)

	storage := &jitoScheduleStorage{
		cache:  cache,
		logger: zap.NewNop(),
	}

	got, err := storage.GetSchedule(context.Background())
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestJitoScheduleStorageGetSchedule_BackupWrongType(t *testing.T) {
	t.Parallel()

	cache := imcache.New(time.Minute, time.Minute)
	cache.Set("schedule_backup", uint64(42), time.Minute)

	storage := &jitoScheduleStorage{
		cache:  cache,
		logger: zap.NewNop(),
	}

	got, err := storage.GetSchedule(context.Background())
	require.Error(t, err)
	require.Nil(t, got)
	require.EqualError(t, err, "can't cast schedule to map[uint64]model.JitoSchedule")
}
