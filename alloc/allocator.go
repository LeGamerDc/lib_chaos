package alloc

import (
	"reflect"
	"sync"
	"unsafe"
)

const (
	pageStructSize = 512 * 1024 // 64KB
	pageSize       = uintptr(pageStructSize - unsafe.Sizeof((*page)(nil)) - unsafe.Sizeof(0))
)

// page is struct supply unmanaged memory alloc.
type page struct {
	next *page
	off  uintptr
	buf  [pageSize]byte
}

// Allocator use cp as current page allocator
// used store allocation full page
type Allocator struct {
	cp   *page // cp is never nil
	used *page
}

func NewAllocator() *Allocator {
	return &Allocator{
		cp: getPage(),
	}
}

func (a *Allocator) Free() {
	putPage(a.cp)
	for p := a.used; p != nil; {
		next := p.next
		putPage(p)
		p = next
	}
}

func (a *Allocator) refill() *page {
	var p *page
	p, a.cp = a.cp, getPage()
	if p != nil {
		a.used, p.next = p, a.used
	}
	return a.cp
}

func (a *Allocator) alloc(align, size uintptr) unsafe.Pointer {
	if size > pageSize { // case 1. size > page size
		return nil
	}
	p := a.cp
	index := _align(p.off, align)
	if index+size > pageSize {
		p = a.refill()
		index = 0
	}
	p.off = index + size
	return unsafe.Pointer(&p.buf[index])
}

func Malloc[T any](a *Allocator) *T {
	var (
		x     T
		size  = unsafe.Sizeof(x)
		align = unsafe.Alignof(x)
	)
	return (*T)(a.alloc(align, size))
}

func MallocSlice[T any](a *Allocator, l, c int) []T {
	var (
		x     T
		size  = unsafe.Sizeof(x) * uintptr(c)
		align = unsafe.Alignof(x)
		hdr   reflect.SliceHeader
	)
	hdr.Len, hdr.Cap = l, c
	hdr.Data = uintptr(a.alloc(align, size))
	return *(*[]T)(unsafe.Pointer(&hdr))
}

func CopyString(a *Allocator, s string) string {
	var (
		size = len(s)
		b    = MallocSlice[byte](a, size, size)
	)
	copy(b, s)
	return *(*string)(unsafe.Pointer(&b))
}

func _align(s, a uintptr) uintptr {
	return (s + a - 1) &^ (a - 1)
}

var pagePool = sync.Pool{
	New: func() any { return new(page) },
}

func getPage() *page {
	p := pagePool.Get().(*page)
	p.off = 0
	return p
}

func putPage(p *page) {
	if p != nil {
		p.next = nil
		pagePool.Put(p)
	}
}
