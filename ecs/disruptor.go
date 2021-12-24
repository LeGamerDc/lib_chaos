package ecs

import (
	"lib_chaos/common"
	"sync/atomic"
)

// Do not use disruptor lib anymore, golang chan is already very fast.
// in my last benchmark, chan vs disruptor receive is 12ns vs 11 ns.

const InitBarrier int64 = -1

type barrier struct {
	_   [64]struct{} // for use in go 1.19
	idx [8]int64
}

func _barrier() *barrier {
	var b = new(barrier)
	b.idx[0] = InitBarrier
	return b
}
func (b *barrier) inc(cnt, max int64) (int64, bool) {
	for {
		var old = atomic.LoadInt64(&b.idx[0])
		if old+cnt <= max {
			if atomic.CompareAndSwapInt64(&b.idx[0], old, old+cnt) {
				return old + cnt, true
			}
		} else {
			return 0, false
		}
	}
}
func (b *barrier) load() int64 {
	return b.idx[0]
}
func (b *barrier) store(x int64) {
	atomic.StoreInt64(&b.idx[0], x)
}
func (b *barrier) cas(old, v int64) bool {
	return atomic.CompareAndSwapInt64(&b.idx[0], old, v)
}

type Writer struct {
	reserve *barrier
	commit  *barrier
	reader  *barrier

	capacity int64
}

func (w *Writer) TryReserve(cnt int64) (seq int64, ok bool) {
	seq, ok = w.reserve.inc(cnt, w.reader.load()+w.capacity)
	return
}
func (w *Writer) Commit(seq, cnt int64) int64 {
	var spin int64
	for ; !w.commit.cas(seq-cnt, seq); spin++ {
	}
	return spin
}

type Reader struct {
	reader   *barrier
	commit   *barrier
	current  int64
	consumer func(lower, upper int64)
}

func (r *Reader) Read(max int64) (cnt int64) {
	var (
		lower = r.current + 1
		upper = r.commit.load()
	)
	if lower <= upper {
		upper = common.Min(upper, lower+max-1)
		r.consumer(lower, upper)
		r.current = upper
		r.reader.store(r.current)
		return upper - lower + 1
	}
	return 0
}

func CreateDisruptor(capacity int64, consumer func(int64, int64)) (*Reader, *Writer) {
	var (
		rb = _barrier()
		wb = _barrier()
		cb = _barrier()
	)
	return &Reader{
			reader:   rb,
			commit:   cb,
			current:  InitBarrier,
			consumer: consumer,
		}, &Writer{
			reserve:  wb,
			commit:   cb,
			reader:   rb,
			capacity: capacity,
		}
}
