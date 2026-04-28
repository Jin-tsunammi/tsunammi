package cache

import (
	"context"
	"errors"
	"mm/config"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"time"

	imcache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var _ JitoScheduleStorage = (*jitoScheduleStorage)(nil)

type JitoScheduleStorage interface {
	GetSchedule(ctx context.Context) (map[uint64]model.JitoSchedule, error)
}

func NewJitoScheduleStorage(cfg *config.Config, jitoValidator JitoValidatorStorage, rpc solanarpc.SolanaRPC, logger *zap.Logger) JitoScheduleStorage {
	cacheTTL := cfg.Jito.ScheduleTTL

	cache := imcache.New(cacheTTL, cacheTTL)

	storage := jitoScheduleStorage{cache: cache, jitoValidator: jitoValidator, rpcSolana: rpc, logger: logger}

	cache.OnEvicted(func(key string, value interface{}) {
		_ = storage.update(context.Background(), cacheTTL)
	})

	return &storage
}

type jitoScheduleStorage struct {
	cache         *imcache.Cache
	jitoValidator JitoValidatorStorage
	rpcSolana     solanarpc.SolanaRPC
	logger        *zap.Logger
}

func (j *jitoScheduleStorage) GetSchedule(ctx context.Context) (map[uint64]model.JitoSchedule, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:

		data, ok := j.cache.Get("schedule")
		if !ok {
			data, ok = j.cache.Get("schedule_backup")
			if !ok {
				return nil, errors.New("schedule not found")
			}

		}

		schedule, ok := data.(map[uint64]model.JitoSchedule)

		if !ok {
			return nil, errors.New("can't cast schedule to map[uint64]model.JitoSchedule")
		}
		return schedule, nil
	}
}

func (j *jitoScheduleStorage) update(ctx context.Context, ttl time.Duration) error {
	slotTime, err := j.rpcSolana.GetAverageSlotTime(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			j.logger.Error("error update schedule", zap.Error(err))
		}
	}()

	average1HSlotTime, exists := slotTime[time.Hour]

	if !exists {
		return errors.New("slot time not found")
	}

	slotCount := uint64(ttl / average1HSlotTime)

	slotCount = slotCount * 5 / 4

	slot, err := j.rpcSolana.GetCurrentSlot(ctx)
	if err != nil {
		return err
	}

	start := slot
	end := slot + slotCount

	validators, err := j.jitoValidator.Get(ctx)

	if err != nil {
		return err
	}

	leaders, err := j.rpcSolana.GetSlotLeadersWithRange(ctx, start, end)
	if err != nil {
		return err
	}

	schedule := make(map[uint64]model.JitoSchedule, len(leaders)+1)

	var i uint64

	for i = 0; i < uint64(len(leaders)); i++ {
		leader := leaders[i]
		_, ok := validators[leader]

		schedule[start+i] = model.JitoSchedule{
			NodePublicKey: leader,
			RunningJito:   ok,
		}
	}

	j.cache.Set("schedule", schedule, ttl)
	j.cache.Set("schedule_backup", schedule, ttl*6/5)

	return nil
}
