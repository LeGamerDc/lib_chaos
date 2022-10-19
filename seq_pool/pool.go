package seq_pool

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	// sleep 设定为produce消耗掉 1/3 buffer的时间
	sleep = time.Millisecond
)

type Config struct {
	QueueSize  int64
	WorkerSize int
}

type Pool[T any] struct {
	produce func(*T)
	consume func(*T)
	c       Config

	queue      []Msg[T]
	prepareIdx int64
	produceIdx int64
	consumeIdx int64
	mask       int64

	// stat
	produceWaitCnt int64
	consumeWaitCnt int64
}

type Msg[T any] struct {
	done int32
	ptr  *T
}

func (t *Msg[T]) wait() int64 {
	i := int64(1)
	for ; atomic.LoadInt32(&t.done) == 0; i++ {
		time.Sleep(sleep)
	}
	return i
}

func wrap[T any](f func(*T)) func(*T) {
	return func(t *T) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()
		f(t)
	}
}

func NewPool[T any](p, c func(*T), config Config) *Pool[T] {
	pool := &Pool[T]{
		//produce: p,
		//consume: c,
		produce:    wrap(p),
		consume:    wrap(c),
		c:          config,
		prepareIdx: -1,
	}
	pool.queue = make([]Msg[T], config.QueueSize)
	pool.mask = config.QueueSize - 1
	// produce
	for i := 0; i < config.WorkerSize; i++ {
		go pool.produceF()
	}
	// consume
	go pool.consumeF()
	return pool
}

func (p *Pool[T]) produceWait(idx int64) int64 {
	for i := 0; i < 4; i++ {
		if idx <= atomic.LoadInt64(&p.prepareIdx)+1 {
			return int64(i + 1)
		}
	}
	for i := 0; i < 4; i++ {
		time.Sleep(sleep * (1 << i))
		if idx <= atomic.LoadInt64(&p.prepareIdx)+1 {
			return int64(i + 5)
		}
	}
	i := int64(9)
	for ; idx > atomic.LoadInt64(&p.prepareIdx)+1; i++ {
		time.Sleep(sleep * 20)
	}
	return i
}

func (p *Pool[T]) produceF() {
	for {
		// atomic add is expensive here
		newIdx := atomic.AddInt64(&p.produceIdx, 1)
		if newIdx > atomic.LoadInt64(&p.prepareIdx)+1 {
			// slow path
			atomic.AddInt64(&p.produceWaitCnt, p.produceWait(newIdx))
		}
		m := &p.queue[(newIdx-1)&p.mask]
		p.produce(m.ptr)
		atomic.StoreInt32(&m.done, 1)
	}
}

func (p *Pool[T]) consumeWait() int64 {
	for i := 0; i < 4; i++ {
		if p.consumeIdx <= atomic.LoadInt64(&p.prepareIdx) {
			return int64(i + 1)
		}
	}
	for i := 0; i < 4; i++ {
		time.Sleep(sleep * (1 << i))
		if p.consumeIdx <= atomic.LoadInt64(&p.prepareIdx) {
			return int64(i + 5)
		}
	}
	i := int64(9)
	for ; p.consumeIdx > atomic.LoadInt64(&p.prepareIdx); i++ {
		time.Sleep(sleep * 20)
	}
	return i
}

func (p *Pool[T]) consumeF() {
	for {
		if p.consumeIdx > atomic.LoadInt64(&p.prepareIdx) { // no prepare data
			// slow path
			p.consumeWaitCnt += p.consumeWait()
		}
		m := &p.queue[p.consumeIdx&p.mask]
		if atomic.LoadInt32(&m.done) == 0 {
			// slow path
			p.consumeWaitCnt += m.wait()
		}
		p.consume(m.ptr)
		m.ptr = nil
		p.consumeIdx++
	}
}

func (p *Pool[T]) Dispatch(t *T) bool {
	if p.prepareIdx >= p.consumeIdx+p.mask {
		return false
	}
	p.queue[(p.prepareIdx+1)&p.mask] = Msg[T]{done: 0, ptr: t}
	atomic.AddInt64(&p.prepareIdx, 1)
	return true
}
