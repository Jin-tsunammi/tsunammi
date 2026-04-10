package worker

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const gReadNumber = 1000

func getWriterSleep() time.Duration {
	ms := 340 + rand.Intn(61) // 340 + [0..50]
	return time.Duration(ms) * time.Millisecond
}

func BenchmarkMutexSwap(b *testing.B) {
	var mu sync.Mutex
	var val int64
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				newVal := rand.Int63()
				mu.Lock()
				val = newVal
				mu.Unlock()
				time.Sleep(getWriterSleep())
			}
		}
	}()

	b.SetParallelism(gReadNumber)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			_ = val
			mu.Unlock()
		}
	})
	close(done)
}

func BenchmarkRWMutexSwap(b *testing.B) {
	var rw sync.RWMutex
	var val int64
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				newVal := rand.Int63()
				rw.Lock()
				val = newVal
				rw.Unlock()
				time.Sleep(getWriterSleep())
			}
		}
	}()

	b.SetParallelism(gReadNumber)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rw.RLock()
			_ = val
			rw.RUnlock()
		}
	})
	close(done)
}

func BenchmarkAtomicSwap(b *testing.B) {
	var val int64
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				newVal := rand.Int63()
				atomic.SwapInt64(&val, newVal)
				time.Sleep(getWriterSleep())
			}
		}
	}()

	b.SetParallelism(gReadNumber)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = atomic.LoadInt64(&val)
		}
	})
	close(done)
}
