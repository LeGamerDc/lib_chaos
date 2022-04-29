package allocator

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	pageStructSize = 16 * 1024 // 16KB
	pageSize       = int(pageStructSize - unsafe.Sizeof(int64(0)) - unsafe.Sizeof(0))
)

type page struct {
	cnt int64
	off int
	buf [pageSize]byte
}

type Allocator struct {
	cp  *page
	buf *Buf
}

func (a *Allocator) newPage() *page {
	p := new(page)
	a.cp = p
	return p
}

func (a *Allocator) getBuf() *Buf {
	a.buf.pp = nil
	return a.buf
}

func (a *Allocator) CreateMsg(f func(*Buf) interface{}) *Msg {
	buf := a.getBuf()
	ptr := f(buf)
	atomic.AddInt64(&buf.cp.cnt, 1)
	msg := &Msg{msg: ptr, c1: &buf.cp.cnt}
	if buf.pp != nil {
		atomic.AddInt64(&buf.pp.cnt, 1)
		msg.c2 = &buf.pp.cnt
	}
	return msg
}

func decPage(x *int64) {
	if atomic.AddInt64(x, -1) == 0 {
		pagePool.Put((*page)(unsafe.Pointer(x)))
	}
}

var pagePool = sync.Pool{
	New: func() any { return new(page) },
}
