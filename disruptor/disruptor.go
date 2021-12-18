package disruptor

import (
	"fmt"
	"math"
	"sync/atomic"
)

const (
	InitBarrier int64 = -1
)

type Barrier struct {
	idx [8]int64
}

func NewBarrier() *Barrier {
	b := new(Barrier)
	b.idx[0] = InitBarrier
	return b
}

func (b *Barrier) Inc(count int64) int64 {
	return atomic.AddInt64(&b.idx[0], count)
}

func (b *Barrier) Load() int64 {
	return b.idx[0]
}

func (b *Barrier) Store(x int64) {
	atomic.StoreInt64(&b.idx[0], x)
}

func (b *Barrier) CAS(v, old int64) bool {
	return atomic.CompareAndSwapInt64(&b.idx[0], old, v)
}

type compositeBarrier struct {
	bs []*Barrier
}

func (c *compositeBarrier) Load() int64 {
	var min = int64(math.MaxInt64)
	for _, b := range c.bs {
		if seq := b.Load(); seq < min {
			min = seq
		}
	}
	return min
}

type Writer struct {
	reserve *Barrier
	commit  *Barrier
	reader  *Barrier

	capacity int64
}

func (w *Writer) Print() {
	fmt.Printf("res: %d, com: %d, red: %d\n", w.reserve.idx[0],
		w.commit.idx[0], w.reader.idx[0])
}

func (w *Writer) Reserve(count int64) (front, spin int64) {
	front = w.reserve.Inc(count)
	for spin = int64(0); front-w.capacity > w.reader.Load(); spin++ {
		//if spin & 0xf == 0xf {
		//    time.Sleep(time.Millisecond)
		//}
	}
	return front, spin
}

func (w *Writer) Commit(cnt, seq int64) int64 {
	var spin int64
	for ; !w.commit.CAS(seq, seq-cnt); spin++ {
		//if spin & 0xf == 0xf {
		//    time.Sleep(time.Millisecond)
		//}
	}
	return spin
}

type Reader struct {
	current  *Barrier
	upstream *Barrier
	consumer func(lower, upper int64)
}

func (r *Reader) Read() {
	var lower, upper int64
	var current = InitBarrier
	for {
		lower = current + 1
		upper = r.upstream.Load()
		if lower <= upper {
			r.consumer(lower, upper)
			current = upper
			r.current.Store(current)
		} else {
			//time.Sleep(time.Second)
		}
	}
}

func NewQueue(n int, capacity int64, consumer func(int64, int64)) (
	[]*Writer, *Reader) {
	var (
		rb = NewBarrier()
		wb = NewBarrier()
		cb = NewBarrier()
	)

	writer := make([]*Writer, 0, n)
	for i := 0; i < n; i++ {
		w := &Writer{
			reserve:  wb,
			commit:   cb,
			reader:   rb,
			capacity: capacity,
		}
		writer = append(writer, w)
	}

	reader := &Reader{
		upstream: cb,
		consumer: consumer,
		current:  rb,
	}
	return writer, reader
}

func NewSingleQueue(capacity int64, consumer func(int64, int64)) (*Writer, *Reader) {
	var (
		rb = NewBarrier()
		wb = NewBarrier()
		cb = NewBarrier()
	)
	w := &Writer{
		reserve:  wb,
		commit:   cb,
		reader:   rb,
		capacity: capacity,
	}

	r := &Reader{
		current:  rb,
		upstream: cb,
		consumer: consumer,
	}
	return w, r
}
