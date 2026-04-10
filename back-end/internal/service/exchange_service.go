package service

import (
	"context"
	"mm/internal/model"
	"mm/internal/storage/repository"
	"mm/pkg/apperrors"
)

type ExchangeService struct {
	ExchangeRepository *repository.ExchangeRepository
}

func NewExchangeService(exchangeRepository *repository.ExchangeRepository) *ExchangeService {
	return &ExchangeService{ExchangeRepository: exchangeRepository}
}

func (s *ExchangeService) GetExchanges(ctx context.Context) ([]model.Exchange, error) {
	res, err := s.ExchangeRepository.FindAll(ctx)
	if err != nil {
		return nil, apperrors.Internal("failed to get exchanges", err)
	}
	return res, nil
}
