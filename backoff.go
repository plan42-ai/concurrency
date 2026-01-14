package concurrency

import (
	"context"
	"math/rand/v2"
	"time"
)

type Backoff struct {
	min     time.Duration
	max     time.Duration
	current time.Duration
	timer   *time.Timer
}

func (b *Backoff) Backoff() {
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
	b.current /= 2
	if b.current < b.min {
		b.current = 0
	}
}

func (b *Backoff) Wait() error {
	return b.WaitContext(context.Background())
}

func (b *Backoff) WaitContext(ctx context.Context) error {
	if b.current == 0 {
		return nil
	}

	// #nosec G404 (jitter doesn't need a secure rng)
	waitTime := rand.N(b.current)

	if b.timer == nil {
		b.timer = time.NewTimer(waitTime)
	} else {
		b.timer.Reset(waitTime)
	}
	defer b.timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.timer.C:
		return nil
	}
}

func (b *Backoff) WaitChannel() <-chan time.Time {
	var waitDuration time.Duration

	if b.current != 0 {
		// #nosec G404 (jitter doesn't need a secure rng)
		waitDuration = rand.N(b.current)
	}

	if b.timer == nil {
		b.timer = time.NewTimer(waitDuration)
	} else {
		b.timer.Reset(waitDuration)
	}
	return b.timer.C
}

func (b *Backoff) StopTimer() {
	if b.timer != nil {
		b.timer.Stop()
	}
}

func NewBackoff(minBackoff, maxBackoff time.Duration) *Backoff {
	return &Backoff{
		min: minBackoff,
		max: maxBackoff,
	}
}
