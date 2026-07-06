package concurrency

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

type Backoff struct {
	mux         sync.Mutex
	min         time.Duration
	max         time.Duration
	current     time.Duration
	backoffRate float64
	recoverRate float64
}

func (b *Backoff) Backoff() {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.current == 0 {
		b.current = b.min
		return
	}

	b.current = time.Duration(float64(b.current) * b.backoffRate)
	if b.current > b.max {
		b.current = b.max
	}
}

func (b *Backoff) Recover() {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.current = time.Duration(float64(b.current) * b.recoverRate)
	if b.current < b.min {
		b.current = 0
	}
}

func (b *Backoff) Wait() (time.Duration, error) {
	return b.WaitContext(context.Background())
}

func (b *Backoff) WaitContext(ctx context.Context) (time.Duration, error) {
	current := b.getCurrent()

	if current == 0 {
		return 0, nil
	}

	// #nosec G404 (jitter doesn't need a secure rng)
	waitTime := rand.N(current)
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return waitTime, ctx.Err()
	case <-timer.C:
		return waitTime, nil
	}
}

func (b *Backoff) getCurrent() time.Duration {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.current
}

type Option func(o *Backoff)

func WithRecoverRate(rate float64) Option {
	return func(o *Backoff) {
		o.recoverRate = rate
	}
}

func WithBackoffRate(rate float64) Option {
	return func(o *Backoff) {
		o.backoffRate = rate
	}
}

func NewBackoff(minBackoff, maxBackoff time.Duration, options ...Option) *Backoff {
	ret := &Backoff{
		min:         minBackoff,
		max:         maxBackoff,
		backoffRate: 2.0,
		recoverRate: 0.5,
	}

	for _, option := range options {
		option(ret)
	}
	return ret
}
