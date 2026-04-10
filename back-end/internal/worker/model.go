package worker

import (
	"errors"
	"math/big"
	"mm/internal/model"
	"sync"
	"sync/atomic"
)

var (
	targetStoppedError = errors.New("target stopped")
	failedToSwapError  = errors.New("failed to swap")
)

type concurrentSwapTask struct {
	tasks           []*model.AsyncSwapTask
	mu              *sync.RWMutex
	remainingBudget *atomic.Pointer[big.Int]
}

func (t *concurrentSwapTask) usingJito() bool {
	if len(t.tasks) == 0 {
		return false
	}
	return t.tasks[0].UsingJito
}
