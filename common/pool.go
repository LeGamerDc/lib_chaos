package common

import "sync/atomic"

func SimplePool(worker int64) func(f func()) {
	var idle = worker
	wait := make(chan struct{}, 1)
	return func(f func()) {
		if atomic.AddInt64(&idle, -1) < 0 {
			<-wait
		}
		go func() {
			f()
			if atomic.AddInt64(&idle, 1) <= 0 {
				wait <- struct{}{}
			}
		}()
	}
}
