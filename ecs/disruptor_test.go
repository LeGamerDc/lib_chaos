package ecs

import (
	"fmt"
	"testing"
)

const P = 1

func BenchmarkLoop(b *testing.B) {
	var ch = make(chan struct{}, 1)
	for i := 0; i < b.N; i++ {
		ch <- struct{}{}
		<-ch
	}
}

func BenchmarkChannel(b *testing.B) {
	var ch = make(chan struct{}, 1024)
	for i := 0; i < P; i++ {
		go func() {
			for i := 0; i < b.N; i++ {
				ch <- struct{}{}
			}
		}()
	}
	b.ResetTimer()
	var tot = b.N * P
	for i := 0; i < tot; i++ {
		<-ch
	}
}

func BenchmarkDisruptor(b *testing.B) {
	var r, w = CreateDisruptor(1024, func(_, _ int64) {})
	for i := 0; i < P; i++ {
		go func() {
			var (
				cmt int64
				ok  bool
			)
			for i := 0; i < b.N; i++ {
				for {
					cmt, ok = w.TryReserve(1)
					if ok {
						break
					}
				}
				w.Commit(cmt, 1)
			}
		}()
	}
	b.ResetTimer()
	var tot = int64(b.N * P)
	var fail int
	for tot > 0 {
		var x = r.Read(1)
		tot -= x
		if x == 0 {
			fail++
		}
	}
	fmt.Println("fail: ", fail)
}
