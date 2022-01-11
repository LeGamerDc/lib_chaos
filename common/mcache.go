package common

import (
	"math/bits"
	"sync"
)

// we only cache most used []byte

func Malloc(size, capacity int) (ret []byte) {
	if capacity < size {
		capacity = size
	}
	var c = sizeClass(capacity)
	if c >= 0 {
		ret = caches[c].Get().([]byte)
		return ret[:size]
	}
	return make([]byte, size, capacity)
}

func Free(buf []byte) {
	if idx := support(cap(buf)); idx >= 0 {
		caches[idx].Put(buf)
	}
}

func init() {
	for i, size := range class {
		size := size
		caches[i] = sync.Pool{New: func() interface{} {
			return make([]byte, 0, size)
		}}
	}
}

const (
	smallSizeMin    = 8
	smallSizeMax    = 1024
	largeSizeDiv    = 1024
	largeSizeMax    = 4096
	_NumSizeClasses = 67
)

var class = []int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 3072, 4096}
var caches [_NumSizeClasses]sync.Pool

func roundUp(n, a int) int {
	return ((n + a - 1) / a) * a
}

func sizeClass(n int) (c int) {
	if n <= smallSizeMin {
		return 0
	}
	if n <= smallSizeMax {
		return bits.Len64(uint64(n-1)) - 3
	}
	if n <= largeSizeMax {
		return support(roundUp(n, largeSizeDiv))
	}
	return -1
}

func support(n int) (idx int) {
	switch n {
	case 8:
		idx = 0
	case 16:
		idx = 1
	case 32:
		idx = 2
	case 64:
		idx = 3
	case 128:
		idx = 4
	case 256:
		idx = 5
	case 512:
		idx = 6
	case 1024:
		idx = 7
	case 2048:
		idx = 8
	case 3072:
		idx = 9
	case 4096:
		idx = 10
	default:
		idx = -1
	}
	return
}
