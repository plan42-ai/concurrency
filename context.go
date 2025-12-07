package concurrency

import (
	"context"
	"sync"
	"time"
)

// ContextGroup is similar to a sync.WaitGroup, except that it supports cancellation via a shared context. It is useful
// for implementing components that support both Graceful shutdown (with timeouts) and forced cancellation.
// Its use is subject to the following constraints:
//
//  1. After Wait(), WaitTimeout() or Close() is called, shutdown will be triggered when the counter reaches zero, and all goroutines blocked on Wait will be released.
//  2. Calling Add() after shutdown will panic.
//  3. To avoid a resource leak, you must arrange to call Cancel() (directly, or via Close()).
type ContextGroup struct {
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	notifyShutdown sync.Once
	shutdown       chan struct{}
}

func NewContextGroup() *ContextGroup {
	ctx, cancel := context.WithCancel(context.Background())
	return &ContextGroup{
		ctx:      ctx,
		cancel:   cancel,
		shutdown: make(chan struct{}),
	}
}

// Add adds delta, which may be negative, to the [ContextGroup] counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics. Similar, calling Add once shutdown is complete will also panic.
func (c *ContextGroup) Add(delta int) {
	select {
	case <-c.shutdown:
		panic("ContextGroup: Add called after shutdown is complete.")
	default:
		c.wg.Add(delta)
	}
}

// Done decrements the [ContextGroup] counter by one.
func (c *ContextGroup) Done() {
	c.wg.Done()
}

// Cancel cancels the [ContextGroup] context.
func (c *ContextGroup) Cancel() {
	c.cancel()
}

// Context returns the shared context of the [ContextGroup].
func (c *ContextGroup) Context() context.Context {
	return c.ctx
}

// ensureNotifyShutdown triggers shutdown tracking. Calling ensureNotifyShutdown more than once is a no-op.
func (c *ContextGroup) ensureNotifyShutdown() {
	c.notifyShutdown.Do(
		func() {
			go NotifyShutdown(c.shutdown, &c.wg)
		},
	)
}

// Wait blocks until the [ContextGroup] counter is zero.
func (c *ContextGroup) Wait() {
	_ = c.WaitContext(context.Background())
}

// WaitContext blocks until the [ContextGroup] counter is zero or the provided context is completed.
// If the context is completed before the counter reaches zero, it returns the context's error.
func (c *ContextGroup) WaitContext(ctx context.Context) error {
	c.ensureNotifyShutdown()
	select {
	case <-c.shutdown:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WaitTimeout blocks until the [ContextGroup] counter is zero or the provided duration has elapsed.
// If the duration elapses before the counter reaches zero, it returns a context deadline exceeded error.
func (c *ContextGroup) WaitTimeout(d time.Duration) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return c.WaitContext(ctx)
}

// Close is equivalent to running Cancel() followed by Wait().
func (c *ContextGroup) Close() error {
	c.ensureNotifyShutdown()
	c.cancel()

	<-c.shutdown
	return nil
}
