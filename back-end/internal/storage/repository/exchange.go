package repository

import (
	"mm/internal/model"
	"mm/pkg/repository"
)

type ExchangeRepository struct {
	repository.Generic[model.Exchange, uint64]
}

func NewExchangeRepository(genericRepository repository.Generic[model.Exchange, uint64]) *ExchangeRepository {
	return &ExchangeRepository{Generic: genericRepository}
}
