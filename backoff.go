package concurrency

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

type Backoff struct {
	mux     sync.Mutex
	min     time.Duration
	max     time.Duration
	current time.Duration
}

func (b *Backoff) Backoff() {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.current == 0 {
		b.current = b.min
		return
	}

	b.current *= 2
	if b.current > b.max {
		b.current = b.max
	}
}

func (b *Backoff) Recover() {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.current /= 2
	if b.current < b.min {
		b.current = 0
	}
}

func (b *Backoff) Wait() error {
	return b.WaitContext(context.Background())
}

func (b *Backoff) WaitContext(ctx context.Context) error {
	current := b.getCurrent()

	if current == 0 {
		return nil
	}

	// #nosec G404 (jitter doesn't need a secure rng)
	waitTime := rand.N(current)
	timer := time.NewTimer(waitTime)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (b *Backoff) getCurrent() time.Duration {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.current
}

func NewBackoff(minBackoff, maxBackoff time.Duration) *Backoff {
	return &Backoff{
		min: minBackoff,
		max: maxBackoff,
	}
}
