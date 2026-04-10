package service

import (
	"context"
	"mm/internal/client/solanarpc"
	"mm/internal/model"
	"mm/internal/storage/cache"
)

type JitoService struct {
	SolanaRPC      solanarpc.SolanaRPC
	ValidatorCache cache.JitoValidatorStorage
}

func (s *JitoService) GetSchedule(ctx context.Context, start, end uint64) (map[uint64]model.JitoSchedule, error) {
	validators, err := s.ValidatorCache.Get(ctx)

	if err != nil {
		return nil, err
	}

	leaders, err := s.SolanaRPC.GetSlotLeadersWithRange(ctx, start, end)
	if err != nil {
		return nil, err
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

	return schedule, nil
}
