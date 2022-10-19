package sparse_set

import (
    "lib_chaos/common"
    "math/bits"
    "sync"
    "unsafe"
)

const (
    tagMask int64 = 255
    tagBit        = 8
)

func IdTag(id int64) int32 {
    return int32(id & tagMask)
}

type SparseArray[T any] struct {
    chucks   []*Array[T]
    freeBits []uint64

    chuckSize int32
}

func MakeSparseArray[T any](chuckSize int32) *SparseArray[T] {
    return &SparseArray[T]{
        chuckSize: chuckSize,
    }
}

func (s *SparseArray[T]) grow() {
    var base = int32(len(s.chucks)) * s.chuckSize
    s.chucks = append(s.chucks, MakeArray[T](s.chuckSize, base))
    if len(s.freeBits)*64 < len(s.chucks) {
        s.freeBits = append(s.freeBits, 1)
    } else {
        p := len(s.chucks) - 1
        i, j := p/64, p%64
        s.freeBits[i] |= 1 << j
    }
}

func (s *SparseArray[T]) free() (*Array[T], int32) {
    for i, b := range s.freeBits {
        if b != 0 {
            j := bits.TrailingZeros64(b)
            p := i*64 + j
            if p >= len(s.chucks) {
                return nil, -1
            }
            return s.chucks[p], int32(p)
        }
    }
    return nil, -1
}

func (s *SparseArray[T]) Get(i int64) *T {
    var idx = _idx(i)
    idx.index >>= tagBit
    var p = idx.index / s.chuckSize
    idx.index %= s.chuckSize
    return s.chucks[p].Get(idx)
}

func (s *SparseArray[T]) Place(check int32, tag int32) (int64, *T) {
    var arr, p = s.free()
    if p == -1 {
        s.grow()
        arr, p = s.free()
    }
    var idx, place = arr.Place(check)
    if arr.Full() {
        s.setFull(p)
    }
    idx.index = _tag(idx.index, arr.base, tag)
    return _id(idx), place
}

func (s *SparseArray[T]) Set(check int32, tag int32, v T) int64 {
    var arr, p = s.free()
    if p == -1 {
        s.grow()
        arr, p = s.free()
    }
    var idx = arr.Set(check, v)
    if arr.Full() {
        s.setFull(p)
    }
    idx.index = _tag(idx.index, arr.base, tag)
    return _id(idx)
}

func (s *SparseArray[T]) Remove(i int64) bool {
    var idx = _idx(i)
    idx.index >>= tagBit
    var p = idx.index / s.chuckSize
    idx.index %= s.chuckSize
    if s.chucks[p].Remove(idx) {
        if s.chucks[p].free == 1 {
            s.setFree(p)
        }
        return true
    }
    return false
}

func (s *SparseArray[T]) Foreach(f func(*T)) {
    for _, a := range s.chucks {
        a.Foreach(f)
    }
}

func (s *SparseArray[T]) ParForeach(f func(*T), threads int64) {
    var wg sync.WaitGroup
    var g = common.SimplePool(threads)
    for _, c := range s.chucks {
        wg.Add(1)
        c := c
        g(func() {
            c.Foreach(f)
            wg.Done()
        })
    }
    wg.Wait()
}

func (s *SparseArray[T]) setFull(p int32) {
    var i, j = p / 64, p % 64
    s.freeBits[i] &= ^(1 << j)
}

func (s *SparseArray[T]) setFree(p int32) {
    var i, j = p / 64, p % 64
    s.freeBits[i] |= 1 << j
}

func _idx(i int64) Index {
    return *(*Index)(unsafe.Pointer(&i))
}
func _id(i Index) int64 {
    return *(*int64)(unsafe.Pointer(&i))
}
func _tag(x int32, base int32, tag int32) int32 {
    return ((x + base) << tagBit) + tag
}
