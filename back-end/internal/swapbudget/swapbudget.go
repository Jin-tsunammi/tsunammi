package swapbudget

import (
	"math/big"
	"mm/internal/common"
	"mm/internal/swaperror"
	"sync"
)

type SwapBudget struct {
	mu        sync.Mutex
	remaining *big.Int
}

func NewSwapBudget(init *big.Int) *SwapBudget {
	return &SwapBudget{
		remaining: new(big.Int).Set(init),
	}
}

func (b *SwapBudget) Reserve(min, max *big.Int) (*big.Int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.remaining.Sign() <= 0 {
		return nil, swaperror.BudgetExceededError
	}

	amount, err := common.SelectTxAmountInRange(
		b.remaining,
		min,
		max,
	)
	if err != nil {
		return nil, err
	}

	next := new(big.Int).Sub(b.remaining, amount)
	if next.Sign() < 0 {
		return nil, swaperror.BudgetExceededError
	}

	b.remaining = next
	return new(big.Int).Set(amount), nil

}

func (b *SwapBudget) Release(a *big.Int) {
	if a == nil || a.Sign() <= 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.remaining = new(big.Int).Add(b.remaining, a)
}

func (b *SwapBudget) Remaining() *big.Int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return new(big.Int).Set(b.remaining)
}

func (b *SwapBudget) Store(a *big.Int) {
	if a == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.remaining = new(big.Int).Set(a)
}
