package common

import "sync"

var class = []int{8, 16, 24, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192, 208, 224, 240, 256, 288, 320, 352, 384, 416, 448, 480, 512, 576, 640, 704, 768, 896, 1024, 1152, 1280, 1408, 1536, 1792, 2048, 2304, 2688, 3072, 3200, 3456, 4096, 4864, 5376, 6144, 6528, 6784, 6912, 8192, 9472, 9728, 10240, 10880, 12288, 13568, 14336, 16384, 18432, 19072, 20480, 21760, 24576, 27264, 28672, 32768}
var caches [_NumSizeClasses]sync.Pool

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

func GetSmallSizeTable() []int {
	return smallSizeTable
}

func GetLargeSizeTable() []int {
	return largeSizeTable
}

func init() {
	for i, size := range class {
		size := size
		caches[i] = sync.Pool{New: func() interface{} {
			return make([]byte, 0, size)
		}}
	}
	var (
		c = 0
	)
	for s := 0; s <= smallSizeMax; s += smallSizeDiv {
		if support(s) >= 0 {
			smallSizeTable = append(smallSizeTable, support(s))
			c = support(s) + 1
		} else {
			smallSizeTable = append(smallSizeTable, c)
		}
	}
	c = 0
	for s := smallSizeMax; s <= largeSizeMax; s += largeSizeDiv {
		if support(s) >= 0 {
			largeSizeTable = append(largeSizeTable, support(s))
			c = support(s) + 1
		} else {
			largeSizeTable = append(largeSizeTable, c)
		}
	}
}

const (
	smallSizeDiv    = 8
	smallSizeMax    = 1024
	largeSizeDiv    = 128
	largeSizeMax    = 32768
	_NumSizeClasses = 67
)

var smallSizeTable []int
var largeSizeTable []int

func divRoundUp(n, a int) int {
	return (n + a - 1) / a
}

func sizeClass(n int) (c int) {
	if n <= smallSizeMax {
		return smallSizeTable[divRoundUp(n, smallSizeDiv)]
	}
	if n <= largeSizeMax {
		return largeSizeTable[divRoundUp(n-smallSizeMax, largeSizeDiv)]
	}
	return -1
}

func support(n int) int {
	var idx = -1
	switch n {
	case 8:
		idx = 0
	case 16:
		idx = 1
	case 24:
		idx = 2
	case 32:
		idx = 3
	case 48:
		idx = 4
	case 64:
		idx = 5
	case 80:
		idx = 6
	case 96:
		idx = 7
	case 112:
		idx = 8
	case 128:
		idx = 9
	case 144:
		idx = 10
	case 160:
		idx = 11
	case 176:
		idx = 12
	case 192:
		idx = 13
	case 208:
		idx = 14
	case 224:
		idx = 15
	case 240:
		idx = 16
	case 256:
		idx = 17
	case 288:
		idx = 18
	case 320:
		idx = 19
	case 352:
		idx = 20
	case 384:
		idx = 21
	case 416:
		idx = 22
	case 448:
		idx = 23
	case 480:
		idx = 24
	case 512:
		idx = 25
	case 576:
		idx = 26
	case 640:
		idx = 27
	case 704:
		idx = 28
	case 768:
		idx = 29
	case 896:
		idx = 30
	case 1024:
		idx = 31
	case 1152:
		idx = 32
	case 1280:
		idx = 33
	case 1408:
		idx = 34
	case 1536:
		idx = 35
	case 1792:
		idx = 36
	case 2048:
		idx = 37
	case 2304:
		idx = 38
	case 2688:
		idx = 39
	case 3072:
		idx = 40
	case 3200:
		idx = 41
	case 3456:
		idx = 42
	case 4096:
		idx = 43
	case 4864:
		idx = 44
	case 5376:
		idx = 45
	case 6144:
		idx = 46
	case 6528:
		idx = 47
	case 6784:
		idx = 48
	case 6912:
		idx = 49
	case 8192:
		idx = 50
	case 9472:
		idx = 51
	case 9728:
		idx = 52
	case 10240:
		idx = 53
	case 10880:
		idx = 54
	case 12288:
		idx = 55
	case 13568:
		idx = 56
	case 14336:
		idx = 57
	case 16384:
		idx = 58
	case 18432:
		idx = 59
	case 19072:
		idx = 60
	case 20480:
		idx = 61
	case 21760:
		idx = 62
	case 24576:
		idx = 63
	case 27264:
		idx = 64
	case 28672:
		idx = 65
	case 32768:
		idx = 66
	}
	return idx
}
