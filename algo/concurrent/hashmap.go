package concurrent

import (
	"hash/fnv"
	"math/rand"
	"sync"
	"unsafe"

	"golang.org/x/exp/maps"
)

type Key interface {
	Hash() uint32
	comparable
}

type bucket[K Key, V any] struct {
	rw sync.RWMutex
	m  map[K]V
}

func (b *bucket[K, V]) find(k K) (v V, ok bool) {
	b.rw.RLock()
	defer b.rw.RUnlock()

	v, ok = b.m[k]
	return
}

func (b *bucket[K, V]) set(k K, v V) {
	b.rw.Lock()
	defer b.rw.Unlock()

	b.m[k] = v
}

func (b *bucket[K, V]) del(k K) {
	b.rw.Lock()
	defer b.rw.Unlock()

	delete(b.m, k)
}

func (b *bucket[K, V]) del2(k K) (V, bool) {
	b.rw.Lock()
	defer b.rw.Unlock()

	v, ok := b.m[k]
	delete(b.m, k)
	return v, ok
}

func (b *bucket[K, V]) del3(k K, f func(V) bool) (V, bool) {
	b.rw.Lock()
	defer b.rw.Unlock()

	v, ok := b.m[k]
	if ok && f(v) {
		delete(b.m, k)
		return v, true
	}
	return zero[V](), false
}

func (b *bucket[K, V]) del4(m map[K]V) {
	b.rw.Lock()
	defer b.rw.Unlock()

	maps.Copy(m, b.m)
	maps.Clear(b.m)
}

func (b *bucket[K, V]) findOrSet(k K, f func() V) (v V, found bool) {
	b.rw.RLock()
	v, found = b.m[k]
	if found {
		b.rw.RUnlock()
		return
	}
	b.rw.RUnlock()

	b.rw.Lock()
	defer b.rw.Unlock()
	if v, found = b.m[k]; found { // double check
		return
	}
	v = f()
	b.m[k] = v
	return v, false
}

// findAndDo 查询bucket中是否存在k，如果存在则执行do函数。
// 注意：findAndDo会锁住整个bucket，因此do函数的内容不能阻塞并且需要
// 在短时间内完成。如果do函数不能确定在短时间内完成，则用户应该使用
// find函数获取v，并自行设置锁。
func (b *bucket[K, V]) findAndDo(k K, do func(V)) bool {
	b.rw.Lock()
	defer b.rw.Unlock()
	v, found := b.m[k]
	if found {
		do(v)
		return true
	}
	return false
}

// findOrSetAndDo 查询bucket中是否存在k，如果不存在则通过set函数创建，
// 如果存在k或者创建成功，则执行do函数。
// 注意：findOrSetAndDo 会锁住整个bucket，因此do函数的内容不能阻塞并且需要
// 在短时间内完成。如果do函数不能确定在短时间内完成，则用户应该使用
// find函数获取v，并自行设置锁。
func (b *bucket[K, V]) findOrSetAndDo(k K, set func() (V, bool), do func(V)) bool {
	b.rw.Lock()
	defer b.rw.Unlock()
	if v, found := b.m[k]; found {
		do(v)
		return true
	} else {
		var ok bool
		v, ok = set()
		if ok {
			b.m[k] = v
			do(v)
			return true
		}
		return false
	}
}

func (b *bucket[K, V]) filter(f func(K, V) (remove, stop bool)) bool {
	b.rw.Lock()
	defer b.rw.Unlock()

	for k, v := range b.m {
		r, s := f(k, v)
		if r {
			delete(b.m, k)
		}
		if s {
			return true
		}
	}
	return false
}

func (b *bucket[K, V]) iterate(f func(K, V) bool) bool {
	b.rw.RLock()
	defer b.rw.RUnlock()

	for k, v := range b.m {
		if f(k, v) {
			return true
		}
	}
	return false
}

type HashMap[K Key, V any] struct {
	maxHash int
	buckets []*bucket[K, V]
}

func NewHashMap[K Key, V any](n int) *HashMap[K, V] {
	h := &HashMap[K, V]{
		maxHash: n,
		buckets: make([]*bucket[K, V], n),
	}
	for i := 0; i < n; i++ {
		h.buckets[i] = &bucket[K, V]{m: make(map[K]V)}
	}
	return h
}

func (h *HashMap[K, V]) Get(k K) (v V, ok bool) {
	idx := k.Hash() % uint32(h.maxHash)
	return h.buckets[idx].find(k)
}

func (h *HashMap[K, V]) Set(k K, v V) {
	idx := k.Hash() % uint32(h.maxHash)
	h.buckets[idx].set(k, v)
}

func (h *HashMap[K, V]) Delete(k K) {
	idx := k.Hash() % uint32(h.maxHash)
	h.buckets[idx].del(k)
}

// LoadAndDelete 读取k对应的v，并删除k，返回读取到的v 和是否读取成功。
func (h *HashMap[K, V]) LoadAndDelete(k K) (V, bool) {
	idx := k.Hash() % uint32(h.maxHash)
	return h.buckets[idx].del2(k)
}

// LoadAndDeleteIf 当k对应的v满足条件f时，读取v，并删除掉这个k。
func (h *HashMap[K, V]) LoadAndDeleteIf(k K, f func(V) bool) (V, bool) {
	idx := k.Hash() % uint32(h.maxHash)
	return h.buckets[idx].del3(k, f)
}

// LoadAndDeleteAll 删除所有key并返回删除的kv对
func (h *HashMap[K, V]) LoadAndDeleteAll() map[K]V {
	m := make(map[K]V, h.N())
	for _, b := range h.buckets {
		b.del4(m)
	}
	return m
}

// GetOrSet 读取k，如果不存在，则通过 f 设置 kv。
func (h *HashMap[K, V]) GetOrSet(k K, f func() V) (v V, found bool) {
	idx := k.Hash() % uint32(h.maxHash)
	return h.buckets[idx].findOrSet(k, f)
}

// GetAndDo 查询bucket中是否存在k，如果存在则执行do函数。
// 注意：GetAndDo 会锁住整个 bucket，因此do函数的内容不能阻塞并且需要
// 在短时间内完成。如果do函数不能确定在短时间内完成，则用户应该使用
// find函数获取v，并自行设置锁。
func (h *HashMap[K, V]) GetAndDo(k K, do func(V)) bool {
	idx := k.Hash() % uint32(h.maxHash)
	return h.buckets[idx].findAndDo(k, do)
}

// GetOrSetAndDo 查询bucket中是否存在k，如果不存在则通过set函数创建，
// 如果存在k或者创建成功，则执行do函数。
// 注意：GetOrSetAndDo 会锁住整个bucket，因此do函数的内容不能阻塞并且需要
// 在短时间内完成。如果do函数不能确定在短时间内完成，则用户应该使用
// find函数获取v，并自行设置锁。
func (h *HashMap[K, V]) GetOrSetAndDo(k K, set func() (V, bool), do func(V)) bool {
	idx := k.Hash() % uint32(h.maxHash)
	return h.buckets[idx].findOrSetAndDo(k, set, do)
}

// Iterate 遍历所有kv，并执行 f 函数，当函数返回true时或遍历结束时结束。
func (h *HashMap[K, V]) Iterate(f func(K, V) (stop bool)) {
	start := rand.Intn(h.maxHash)
	end := start + h.maxHash
	for i := start; i < end; i++ {
		if h.buckets[i%h.maxHash].iterate(f) {
			return
		}
	}
}

// IterateBucket 遍历指定bucket，函数参数与Iterate类似，但只遍历idx这一个bucket。
// 用于分批遍历
func (h *HashMap[K, V]) IterateBucket(idx int, f func(K, V) (stop bool)) {
	if idx >= 0 && idx < h.maxHash {
		h.buckets[idx].iterate(f)
	}
}

// Filter 过滤map中的一些kv。
func (h *HashMap[K, V]) Filter(f func(K, V) (remove, stop bool)) {
	for _, b := range h.buckets {
		if b.filter(f) {
			return
		}
	}
}

// N 返回hashmap中的总kv对
func (h *HashMap[K, V]) N() (n int) {
	for _, b := range h.buckets {
		n += len(b.m)
	}
	return
}

func HashMapLen[K Key, V any](h *HashMap[K, V]) int {
	l := 0
	for _, b := range h.buckets {
		l += len(b.m)
	}
	return l
}

type StrKey string

func (s StrKey) Hash() uint32 {
	h := fnv.New32()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func ConvertStrKey(ids []string) []StrKey {
	return *(*[]StrKey)(unsafe.Pointer(&ids))
}

type Int64Key int64

func (i Int64Key) Hash() uint32 {
	h := hash64(int64(i))
	return uint32(h)
}

type Int32Key int32

func (i Int32Key) Hash() uint32 {
	h := hash32(int32(i))
	return uint32(h)
}

// https://gist.github.com/badboy/6267743
func hash64(ix int64) int32 {
	var x = uint64(ix)
	x = (^x) + (x << 18)
	x = x ^ (x >> 31)
	x = x * 21
	x = x ^ (x >> 11)
	x = x + (x << 6)
	x = x ^ (x >> 22)
	return int32(x)
}

func hash32(ix int32) int32 {
	var x = uint32(ix)
	x += ^(x << 15)
	x ^= x >> 10
	x += x << 3
	x ^= x >> 6
	x += ^(x << 11)
	x ^= x >> 16
	return int32(x)
}

func zero[T any]() (t T) {
	return
}
