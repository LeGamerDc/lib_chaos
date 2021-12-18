package skill_system

type K = int64
type V = int32

type hashNode struct {
	next int32
	val  V
}

type HashLink struct {
	nodes []hashNode
	free  []int32

	first map[K]int32
}

func NewHashLink(n int) *HashLink {
	var h = &HashLink{
		nodes: make([]hashNode, n),
		free:  make([]int32, n),
		first: make(map[K]int32),
	}
	for i := 0; i < n; i++ {
		h.free[i] = int32(n - i - 1)
	}
	return h
}

func (h *HashLink) Insert(k K, v V) bool {
	var (
		n        = h.getNode()
		next, ok = h.first[k]
	)
	if !ok {
		next = -1
	}
	if n == -1 {
		return false
	}
	h.nodes[n] = hashNode{
		next: next,
		val:  v,
	}
	h.first[k] = n
	return true
}

func (h *HashLink) Remove(k K, v V) {
	var f, ok = h.first[k]
	if !ok {
		return
	}
	var fp = &f
	for *fp != -1 {
		var node = &h.nodes[*fp]
		if v == node.val {
			h.free = append(h.free, *fp)
			*fp = node.next
		}
		fp = &node.next
	}
	if f == -1 {
		delete(h.first, k)
	} else {
		h.first[k] = f
	}
}

// remove k's v if v match f(v) is true
func (h *HashLink) RemoveMatch(k K, match func(V) bool) {
	var f, ok = h.first[k]
	if !ok {
		return
	}
	var fp = &f
	for *fp != -1 {
		var node = &h.nodes[*fp]
		if match(node.val) {
			h.free = append(h.free, *fp)
			*fp = node.next
		}
		fp = &node.next
	}
	if f == -1 {
		delete(h.first, k)
	} else {
		h.first[k] = f
	}
}
func (h *HashLink) RemoveAll(k K) {
	var f, ok = h.first[k]
	if !ok {
		return
	}
	for f != -1 {
		h.free = append(h.free, f)
		f = h.nodes[f].next
	}
	delete(h.first, k)
}
func (h *HashLink) Remain() int {
	return len(h.free)
}

// internal
func (h *HashLink) getNode() int32 {
	var l = len(h.free) - 1
	if l < 0 {
		return -1
	}
	var n = h.free[l]
	h.free = h.free[:l]
	return n
}
