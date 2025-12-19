package concurrency

import (
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
		return true
	}
	return false
}
