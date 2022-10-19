package sparse_set

import "unsafe"

// check must be positive to ensure int64(Index) to be positive
// embed 8 bit in Index.index lower bits

type Index struct {
    index, check int32
}

func (i *Index) Id() int64 {
    return *(*int64)(unsafe.Pointer(i))
}

type Item[T any] struct {
    check int32
    next  int32
    data  T
}

type Array[T any] struct {
    slots []Item[T]

    capacity int32
    free     int32

    first int32
    base  int32
}

func MakeArray[T any](size int32, base int32) *Array[T] {
    var a = new(Array[T])
    a.capacity = size
    a.free = size
    a.base = base
    a.slots = make([]Item[T], size)
    a.first = 0
    for i := int32(0); i < size; i++ {
        a.slots[i] = Item[T]{
            check: -1,
            next:  i + 1,
        }
    }
    return a
}

func (a *Array[T]) Get(i Index) *T {
    var v = a.slots[i.index]
    if v.check == i.check {
        return &v.data
    }
    return nil
}

func (a *Array[T]) Place(check int32) (Index, *T) {
    if a.free <= 0 {
        panic("no slot")
    }
    var pos = a.first
    a.first = a.slots[pos].next
    a.slots[pos].check = check
    a.free--
    return Index{index: pos, check: check}, &a.slots[pos].data
}

func (a *Array[T]) Set(check int32, v T) Index {
    if a.free <= 0 {
        panic("no slot")
    }
    var pos = a.first
    a.first = a.slots[pos].next
    a.slots[pos] = Item[T]{
        check: check,
        data:  v,
    }
    a.free--
    return Index{index: pos, check: check}
}

func (a *Array[T]) Remove(i Index) bool {
    if a.slots[i.index].check == i.check {
        a.slots[i.index] = Item[T]{
            check: -1,
            next:  a.first,
        }
        a.first = i.index
        a.free++
        return true
    }
    return false
}

func (a *Array[T]) Foreach(f func(*T)) {
    for i := range a.slots {
        if a.slots[i].check >= 0 {
            f(&a.slots[i].data)
        }
    }
}

func (a *Array[T]) Size() int {
    return int(a.capacity - a.free)
}

func (a *Array[T]) Free() int {
    return int(a.free)
}

func (a *Array[T]) Cap() int {
    return int(a.capacity)
}

func (a *Array[T]) Full() bool {
    return a.free <= 0
}
