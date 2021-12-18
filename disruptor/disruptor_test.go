package disruptor

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkDisruptor(b *testing.B) {
	fmt.Println(b.N)
	var (
		one = int64(b.N / 4)
		sum = one * 4

		cnt int64 = 0
	)
	var cond = sync.WaitGroup{}
	cond.Add(1)

	var writers, reader = NewQueue(4, 1024, func(lo int64, hi int64) {
		cnt += hi - lo + 1
		if cnt >= sum {
			cond.Done()
		}
	})
	go reader.Read()
	for i := 0; i < 4; i++ {
		i := i
		go func() {
			writer := writers[i]
			for i := int64(0); i < one; i++ {
				f, _ := writer.Reserve(1)
				writer.Commit(1, f)
			}
		}()
	}

	cond.Wait()
}
