package concurrency

import "sync"

func NotifyShutdown(done chan<- struct{}, wgs ...*sync.WaitGroup) {
	for _, wg := range wgs {
		wg.Wait()
	}
	close(done)
}
