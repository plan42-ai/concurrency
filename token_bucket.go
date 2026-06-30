package concurrency

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

type TokenBucket struct {
	mux            sync.Mutex
	tokens         float64
	maxTokens      float64
	refillRate     float64
	refillDuration time.Duration
	lastRefill     time.Time
}

func NewTokenBucket(maxTokens float64, refillRate float64, refillDuration time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		refillDuration: refillDuration,
		lastRefill:     time.Now(),
	}
}

func (tb *TokenBucket) Take(amount float64) bool {
	ok, _ := tb.take(amount)
	return ok
}

func (tb *TokenBucket) take(amount float64) (bool, time.Duration) {
	tb.mux.Lock()
	defer tb.mux.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	refills := float64(elapsed) / float64(tb.refillDuration)
	tb.tokens += refills * tb.refillRate
	tb.lastRefill = now
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	if tb.tokens >= amount {
		tb.tokens -= amount
		return true, 0
	}

	return false, time.Duration((amount - tb.tokens) * (float64(tb.refillDuration) / tb.refillRate))
}

func (tb *TokenBucket) Wait(ctx context.Context, amount float64, jitterPct float64) error {
	ok, timeToWait := tb.take(amount)
	if ok {
		return nil
	}
	timer := time.NewTimer(withJitter(timeToWait, jitterPct))
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			ok, timeToWait = tb.take(amount)
			if ok {
				return nil
			}
			timer.Reset(withJitter(timeToWait, jitterPct))
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func withJitter(wait time.Duration, pct float64) time.Duration {
	extraWait := time.Duration(float64(wait) * pct)
	if extraWait <= 0 {
		return wait
	}
	// #nosec G404: Use of weak random number generator
	//     We don't need a cryptographically strong RNG here... we just need jitter to make sure we don't storm when
	//     recovering from throttles
	return wait + rand.N(extraWait)
}

func (tb *TokenBucket) WaitTimeout(amount float64, d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return tb.Wait(ctx, amount, 0.0)
}

func (tb *TokenBucket) WaitTimeoutContext(ctx context.Context, amount float64, d time.Duration) error {
	return tb.WaitTimeoutContextWithJitter(ctx, amount, 0.0, d)
}

func (tb *TokenBucket) WaitTimeoutContextWithJitter(ctx context.Context, amount float64, jitterPct float64, d time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return tb.Wait(ctx, amount, jitterPct)
}
