package worker

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

const WaitersCount = 1_000_000

type CondMonitor struct {
	slot atomic.Uint64
	cond *sync.Cond
}

func NewCondMonitor() *CondMonitor {
	return &CondMonitor{cond: sync.NewCond(&sync.Mutex{})}
}

func (m *CondMonitor) Update(s uint64) {
	m.slot.Store(s)
	m.cond.L.Lock()
	m.cond.Broadcast()
	m.cond.L.Unlock()
}

func (m *CondMonitor) WaitUntil(target uint64) {
	if m.slot.Load() >= target {
		return
	}
	m.cond.L.Lock()
	for m.slot.Load() < target {
		m.cond.Wait()
	}
	m.cond.L.Unlock()
}

type MapMonitor struct {
	mu       sync.Mutex
	waiters  map[uint64]chan struct{}
	lastSlot uint64
}

func NewMapMonitor() *MapMonitor {
	return &MapMonitor{waiters: make(map[uint64]chan struct{})}
}

func (m *MapMonitor) Update(s uint64) {
	m.mu.Lock()
	m.lastSlot = s
	for t, ch := range m.waiters {
		if t <= s {
			close(ch)
			delete(m.waiters, t)
		}
	}
	m.mu.Unlock()
}

func (m *MapMonitor) WaitUntil(target uint64) {
	m.mu.Lock()
	if m.lastSlot >= target {
		m.mu.Unlock()
		return
	}
	ch, ok := m.waiters[target]
	if !ok {
		ch = make(chan struct{})
		m.waiters[target] = ch
	}
	m.mu.Unlock()
	<-ch
}

type SpinMonitor struct {
	slot atomic.Uint64
}

func (m *SpinMonitor) Update(s uint64) {
	m.slot.Store(s)
}

func (m *SpinMonitor) WaitUntil(target uint64) {
	for {
		if m.slot.Load() >= target {
			return
		}
		runtime.Gosched()
	}
}

func Benchmark_Method1_SyncCond(b *testing.B) {
	mon := NewCondMonitor()
	targetSlot := uint64(1)

	for i := 0; i < b.N; i++ {
		mon.slot.Store(0)
		var wg sync.WaitGroup
		wg.Add(WaitersCount)

		for j := 0; j < WaitersCount; j++ {
			go func() {
				mon.WaitUntil(targetSlot)
				wg.Done()
			}()
		}

		runtime.Gosched()

		mon.Update(targetSlot)

		wg.Wait()
	}
}

func Benchmark_Method2_MapChannels(b *testing.B) {
	mon := NewMapMonitor()
	targetSlot := uint64(1)

	for i := 0; i < b.N; i++ {
		mon.mu.Lock()
		mon.lastSlot = 0
		mon.mu.Unlock()

		var wg sync.WaitGroup
		wg.Add(WaitersCount)

		for j := 0; j < WaitersCount; j++ {
			go func() {
				mon.WaitUntil(targetSlot)
				wg.Done()
			}()
		}

		runtime.Gosched()
		mon.Update(targetSlot)
		wg.Wait()
	}
}

func Benchmark_Method3_Spinlock(b *testing.B) {
	mon := &SpinMonitor{}
	targetSlot := uint64(1)

	for i := 0; i < b.N; i++ {
		mon.slot.Store(0)
		var wg sync.WaitGroup
		wg.Add(WaitersCount)

		for j := 0; j < WaitersCount; j++ {
			go func() {
				mon.WaitUntil(targetSlot)
				wg.Done()
			}()
		}

		runtime.Gosched()
		mon.Update(targetSlot)
		wg.Wait()
	}
}

func BenchmarkScaling(b *testing.B) {
	counts := []int{100, 1000, 10000, 100000, 1000000}

	for _, count := range counts {
		b.Run(fmt.Sprintf("SyncCond-%d", count), func(b *testing.B) {
			mon := NewCondMonitor()
			targetSlot := uint64(1)

			for i := 0; i < b.N; i++ {
				mon.slot.Store(0)
				var wg sync.WaitGroup
				wg.Add(count)

				for j := 0; j < count; j++ {
					go func() {
						mon.WaitUntil(targetSlot)
						wg.Done()
					}()
				}

				runtime.Gosched()
				mon.Update(targetSlot)
				wg.Wait()
			}
		})

		b.Run(fmt.Sprintf("MapChan-%d", count), func(b *testing.B) {
			mon := NewMapMonitor()
			targetSlot := uint64(1)

			for i := 0; i < b.N; i++ {
				mon.mu.Lock()
				mon.lastSlot = 0
				mon.mu.Unlock()

				var wg sync.WaitGroup
				wg.Add(count)

				for j := 0; j < count; j++ {
					go func() {
						mon.WaitUntil(targetSlot)
						wg.Done()
					}()
				}

				runtime.Gosched()
				mon.Update(targetSlot)
				wg.Wait()
			}
		})

		b.Run(fmt.Sprintf("Spinlock-%d", count), func(b *testing.B) {
			mon := &SpinMonitor{}
			targetSlot := uint64(1)

			for i := 0; i < b.N; i++ {
				mon.slot.Store(0)
				var wg sync.WaitGroup
				wg.Add(count)

				for j := 0; j < count; j++ {
					go func() {
						mon.WaitUntil(targetSlot)
						wg.Done()
					}()
				}

				runtime.Gosched()
				mon.Update(targetSlot)
				wg.Wait()
			}
		})
	}
}

const SlotsRange = 50

func Benchmark_Realistic_50Slots(b *testing.B) {
	counts := []int{100, 1000, 10_000, 100_000}

	for _, count := range counts {
		b.Run(fmt.Sprintf("SyncCond-%d", count), func(b *testing.B) {
			mon := NewCondMonitor()

			for i := 0; i < b.N; i++ {
				mon.slot.Store(0)
				var wg sync.WaitGroup
				wg.Add(count)

				var startWg sync.WaitGroup
				startWg.Add(count)

				for j := 0; j < count; j++ {
					target := uint64((j % SlotsRange) + 1)
					go func(t uint64) {
						startWg.Done()
						mon.WaitUntil(t)
						wg.Done()
					}(target)
				}

				startWg.Wait()

				for s := 1; s <= SlotsRange; s++ {
					mon.Update(uint64(s))
					runtime.Gosched()
				}

				wg.Wait()
			}
		})

		b.Run(fmt.Sprintf("MapChan-%d", count), func(b *testing.B) {
			mon := NewMapMonitor()

			for i := 0; i < b.N; i++ {
				mon.mu.Lock()
				mon.lastSlot = 0
				mon.waiters = make(map[uint64]chan struct{})
				mon.mu.Unlock()

				var wg sync.WaitGroup
				wg.Add(count)

				var startWg sync.WaitGroup
				startWg.Add(count)

				for j := 0; j < count; j++ {
					target := uint64((j % SlotsRange) + 1)
					go func(t uint64) {
						startWg.Done()
						mon.WaitUntil(t)
						wg.Done()
					}(target)
				}

				startWg.Wait()

				for s := 1; s <= SlotsRange; s++ {
					mon.Update(uint64(s))
					runtime.Gosched()
				}

				wg.Wait()
			}
		})
	}
}
