package alloc

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	pageStructSize = 16 * 1024 // 16KB
	pageSize       = int(pageStructSize - unsafe.Sizeof(int64(0)) - unsafe.Sizeof(0))
)

// page is struct supply unmanaged memory alloc.
// it behaves like a `reference count pointer`, when page.cnt dec to zero,
// the page will be recycled to a page pool.
// however, if the cnt are never dec to zero(due to program bugs), page will
// be `gc`ed when no object keep reference of this page.
type page struct {
	cnt int64
	off int
	buf [pageSize]byte
}

type Allocator struct {
	cp  *page
	buf *Buf
}

// Init Allocator must Init before use
func (a *Allocator) Init() {
	p := a.newPage()
	a.buf = &Buf{
		a:  a,
		cp: p,
		pp: nil,
	}
}

// allocator will occupy page's cnt like any other object
// to avoid allocator.cp been put to pagePool
func (a *Allocator) newPage() *page {
	if a.cp != nil {
		decPage(&a.cp.cnt)
	}
	p := pagePool.Get().(*page)
	p.cnt, p.off = 1, 0
	a.cp = p
	return p
}

func (a *Allocator) getBuf() *Buf {
	a.buf.pp = nil
	return a.buf
}

// CreateMsg will supply a Buf for malloc, user invoke CreateMsg
// and send a function to use Buf to create pb_msg
// 1. user `MUST` make sure all object in Msg have the same lifetime
// 2. user `MUST NOT` call CreateMsg in parallel
// 3. user `MUST` make sure every Msg do not use more than one page
func (a *Allocator) CreateMsg(f func(*Buf) interface{}) *Msg {
	buf := a.getBuf()
	ptr := f(buf)
	// below code run same as:
	// msg := &Msg{msg: ptr, c1: &buf.cp.cnt}
	// however, since Msg have the same lifetime with its internal msg
	// we could create Msg on unmanaged memory too.
	msg := Malloc[Msg](buf)
	*msg = Msg{msg: ptr, c1: &buf.cp.cnt}
	atomic.AddInt64(&buf.cp.cnt, 1)
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
