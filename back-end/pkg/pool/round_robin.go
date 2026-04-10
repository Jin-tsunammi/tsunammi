package pool

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"sync"
	"time"
)

type CloseableRoundRobin[T io.Closer] struct {
	*RoundRobin[T]
}

func NewCloseableRoundRobin[T io.Closer](resources []T, cooldown time.Duration) *CloseableRoundRobin[T] {
	pool := NewPool[T](resources, cooldown)
	return &CloseableRoundRobin[T]{RoundRobin: pool}
}

func (c *CloseableRoundRobin[T]) Close() error {
	errs := make([]error, 0, len(c.resources))

	for _, resource := range c.resources {
		errs = append(errs, resource.Close())
	}

	return errors.Join(errs...)
}

type RoundRobin[T any] struct {
	available chan T
	recycle   chan item[T]
	cooldown  time.Duration
	stopCh    chan struct{}
	wg        sync.WaitGroup
	resources []T
}

type item[T any] struct {
	val     T
	readyAt time.Time
}

func NewPool[T any](resources []T, cooldown time.Duration) *RoundRobin[T] {

	p := &RoundRobin[T]{
		available: make(chan T, len(resources)),
		recycle:   make(chan item[T], len(resources)),
		cooldown:  cooldown,
		stopCh:    make(chan struct{}),
		resources: resources,
	}

	perm := rand.Perm(len(resources))
	for _, i := range perm {
		p.available <- resources[i]
	}
	return p
}

func (p *RoundRobin[T]) Start(ctx context.Context) error {

	p.wg.Add(1)
	go p.maintenanceLoop()
	return nil

}

func (p *RoundRobin[T]) Stop(ctx context.Context) error {
	close(p.stopCh)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *RoundRobin[T]) Get(ctx context.Context) (T, error) {

	var zero T

	select {
	case res, ok := <-p.available:
		if !ok {
			return zero, errors.New("pool is closed")
		}
		p.recycle <- item[T]{val: res, readyAt: time.Now().Add(p.cooldown)}
		return res, nil
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-p.stopCh:
		return zero, errors.New("pool is stopping")
	default:
		return zero, errors.New("pool is empty")
	}
}

func (p *RoundRobin[T]) maintenanceLoop() {
	defer p.wg.Done()

	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		select {
		case <-p.stopCh:
			timer.Stop()
			return

		case it := <-p.recycle:
			if d := time.Until(it.readyAt); d > 0 {
				timer.Reset(d)
				select {
				case <-timer.C:
				case <-p.stopCh:
					timer.Stop()
					return
				}
			}

			select {
			case p.available <- it.val:
			case <-p.stopCh:
				return
			}
		}
	}
}
