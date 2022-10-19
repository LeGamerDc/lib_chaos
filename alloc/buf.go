package alloc

import (
	"reflect"
	"unsafe"
)

type Buf struct {
	a  *Allocator
	cp *page // current page: current object malloc memory from
	pp *page // prev page: current object have malloc memory from
}

func (b *Buf) refill() {
	if b.pp != nil {
		panic("object > 2 page")
	}
	b.cp, b.pp = b.a.newPage(), b.cp
}

func alloc(b *Buf, align, size int) unsafe.Pointer {
	var (
		index int
	)
	if size > pageSize { // 1. size > page size
		return nil
	} else {
		index = _align(b.cp.off, align)
		if index+size > pageSize {
			b.refill()
			index = 0
		}
	}
	b.cp.off = index + size
	return unsafe.Pointer(&(b.cp.buf[index]))
}

func Malloc[T any](buf *Buf) *T {
	var (
		x     T
		size  = int(unsafe.Sizeof(x))
		align = int(unsafe.Alignof(x))
	)
	return (*T)(alloc(buf, align, size))
}

func MallocSlice[T any](buf *Buf, l, c int) []T {
	var (
		x     T
		size  = int(unsafe.Sizeof(x)) * c
		align = int(unsafe.Alignof(x))
		hdr   reflect.SliceHeader
	)
	hdr.Len, hdr.Cap = l, c
	hdr.Data = uintptr(alloc(buf, align, size))
	return *(*[]T)(unsafe.Pointer(&hdr))
}

// CopyString refer to strings.Clone
// user should use CopyString to assign string to pb msg,
// because struct alloc by allocator do not have
func CopyString(buf *Buf, s string) string {
	var (
		size = len(s)
		b    = MallocSlice[byte](buf, size, size)
	)
	copy(b, s)
	return *(*string)(unsafe.Pointer(&b))
}

func _align(s, a int) int {
	return (s + a - 1) &^ (a - 1)
}
