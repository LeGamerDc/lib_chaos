package ecs

import "unsafe"

// check must be positive to ensure Index.Id to be positive
// embed 8 bit in index lower bits

type Index struct {
	index, check int32
}

func (i *Index) Id() int64 {
	return *(*int64)(unsafe.Pointer(i))
}

type Item struct {
	check int32
	next  int32
	data  T
}

type Array struct {
	slots []Item

	capacity int32
	free     int32

	first int32
	base  int32
}

func MakeArray(size int32, base int32) *Array {
	var a = new(Array)
	a.capacity = size
	a.free = size
	a.base = base
	a.slots = make([]Item, size)
	a.first = 0
	for i := int32(0); i < size; i++ {
		a.slots[i] = Item{
			check: -1,
			next:  i + 1,
		}
	}
	return a
}

func (a *Array) Get(i Index) *T {
	var v = a.slots[i.index]
	if v.check == i.check {
		return &v.data
	}
	return nil
}

func (a *Array) Place(check int32) (Index, *T) {
	if a.free <= 0 {
		panic("no slot")
	}
	var pos = a.first
	a.first = a.slots[pos].next
	a.slots[pos].check = check
	a.free--
	return Index{index: pos, check: check}, &a.slots[pos].data
}

func (a *Array) Set(check int32, v T) Index {
	if a.free <= 0 {
		panic("no slot")
	}
	var pos = a.first
	a.first = a.slots[pos].next
	a.slots[pos] = Item{
		check: check,
		data:  v,
	}
	a.free--
	return Index{index: pos, check: check}
}

func (a *Array) Remove(i Index) bool {
	if a.slots[i.index].check == i.check {
		a.slots[i.index] = Item{
			check: -1,
			next:  a.first,
		}
		a.first = i.index
		a.free++
		return true
	}
	return false
}

func (a *Array) Foreach(f func(*T)) {
	for i := range a.slots {
		if a.slots[i].check >= 0 {
			f(&a.slots[i].data)
		}
	}
}

func (a *Array) Size() int {
	return int(a.capacity - a.free)
}

func (a *Array) Free() int {
	return int(a.free)
}

func (a *Array) Cap() int {
	return int(a.capacity)
}

func (a *Array) Full() bool {
	return a.free <= 0
}
