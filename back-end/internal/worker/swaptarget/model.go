package swaptarget

import (
	"errors"
	"mm/internal/model"
	"mm/internal/swapbudget"
	"sync"
)

var (
	targetStoppedError = errors.New("target stopped")
	failedToSwapError  = errors.New("failed to swap")
)

type concurrentSwapTask struct {
	tasks           []*model.AsyncSwapTask
	mu              *sync.RWMutex
	remainingBudget *swapbudget.SwapBudget
	nextTaskIndex   int
}

func (t *concurrentSwapTask) usingJito() bool {
	if len(t.tasks) == 0 {
		return false
	}
	return t.tasks[0].UsingJito
}

func (t *concurrentSwapTask) removeTasks(tasksToRemove []*model.AsyncSwapTask) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	remove := make(map[*model.AsyncSwapTask]struct{}, len(tasksToRemove))
	for _, task := range tasksToRemove {
		remove[task] = struct{}{}
	}

	active := t.tasks[:0]
	for _, task := range t.tasks {
		if _, ok := remove[task]; ok {
			continue
		}
		active = append(active, task)
	}

	t.tasks = active
	if len(t.tasks) == 0 {
		t.nextTaskIndex = 0
		return 0
	}
	t.nextTaskIndex %= len(t.tasks)

	return len(t.tasks)
}
