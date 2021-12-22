package pbBuffer

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

const (
	pageSize = 1024
)

type PbBuffer struct {
	buf [][]byte
}

func alloc(buf *PbBuffer, align, size int) unsafe.Pointer {
	var (
		l     = len(buf.buf)
		page  *[]byte
		index int
	)
	if size > pageSize { // size > page size
		buf.buf = append(buf.buf, make([]byte, 0, size))
		page = &buf.buf[l]
	} else {
		for i := 0; i < l; i++ {
			index = _align(len(buf.buf[i]), align)
			//fmt.Println(len(buf.buf[i]), align, index)
			if index+size <= pageSize { // find page to place
				page = &buf.buf[i]
			}
		}
		if page == nil { // new page to place
			buf.buf = append(buf.buf, pagePool.Get().([]byte))
			page = &buf.buf[l]
			index = 0
		}
	}
	*page = (*page)[0 : index+size]
	return unsafe.Pointer(&(*page)[index])
}

func Malloc[T any](buf *PbBuffer) *T {
	var (
		x     T
		size  = int(unsafe.Sizeof(x))
		align = int(unsafe.Alignof(x))
	)
	return (*T)(alloc(buf, align, size))
}

func MallocSlice[T any](buf *PbBuffer, l, c int) []T {
	var (
		x     T
		size  = int(unsafe.Sizeof(x)) * c
		align = int(unsafe.Alignof(x))
		hdr   = Malloc[reflect.SliceHeader](buf)
	)
	hdr.Len, hdr.Cap = l, c
	hdr.Data = uintptr(alloc(buf, align, size))
	return *(*[]T)(unsafe.Pointer(hdr))
}

// CopyString reference strings.Clone
func CopyString(buf *PbBuffer, s string) string {
	var (
		size = len(s)
		b    = MallocSlice[byte](buf, size, size)
	)
	copy(b, s)
	return *(*string)(unsafe.Pointer(&b))
}

func (buf *PbBuffer) Destroy() {
	for _, b := range buf.buf {
		if cap(b) == pageSize {
			pagePool.Put(b)
		}
	}
}

func (buf *PbBuffer) Explain() string {
	var builder = strings.Builder{}
	builder.WriteByte('[')
	for _, b := range buf.buf {
		builder.WriteString(strconv.Itoa(len(b)) + ":" + strconv.Itoa(cap(b)) + " ")
	}
	builder.WriteByte(']')
	return builder.String()
}

func _align(s, a int) int {
	return (s + a - 1) &^ (a - 1)
}

var pagePool *sync.Pool

func init() {
	pagePool = &sync.Pool{
		New: func() any {
			return make([]byte, 0, pageSize)
		},
	}
}
