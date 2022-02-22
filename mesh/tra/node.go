package tra

import (
	"lib_chaos/mesh"
	"unsafe"
)

type nodeStatus uint32

const (
	nodeOpen  nodeStatus = 1
	nodeClose nodeStatus = 2
)

type trNode struct {
	turn   mesh.Vert  // turn point
	l, r   mesh.Vert  // l & r directly visible to turn
	cost   float64    // g
	total  float64    // g+h
	hash   uint64     // path hash
	pIdx   int32      // parent node
	ref    int32      // which tri
	status nodeStatus // node status
}

type trNodePool struct {
	mNode    []trNode
	mNext    []int32
	mFirst   []int32
	maxNode  int32
	nodeCnt  int32
	hashSize int32
}

func newNodePool(maxNode int32) *trNodePool {
	var p = &trNodePool{
		mNode:    make([]trNode, maxNode),
		mNext:    make([]int32, maxNode),
		mFirst:   make([]int32, maxNode),
		hashSize: int32(nextPow2(uint32(maxNode / 4))),
		maxNode:  maxNode,
		nodeCnt:  0,
	}
	memset(p.mFirst, -1)
	memset(p.mNext, -1)
	return p
}

func (p *trNodePool) clear() {
	memset(p.mFirst, -1)
	p.nodeCnt = 0
}

func (p *trNodePool) getNode2(v mesh.Vert) *trNode {
	var (
		ix, iz = int32(v.X), int32(v.Z)
		bucket = hash64((int64(ix)<<32)|int64(iz)) & (p.hashSize - 1)
		i      = p.mFirst[bucket]
	)
	for i != -1 {
		if mesh.VEqual(p.mNode[i].turn, v) {
			return &p.mNode[i]
		}
		i = p.mNext[i]
	}
	if p.nodeCnt >= p.maxNode {
		return nil
	}
	i = p.nodeCnt
	p.nodeCnt++
	p.mNode[i] = trNode{turn: v}
	p.mNext[i] = p.mFirst[bucket]
	p.mFirst[bucket] = i
	return &p.mNode[i]
}

func (p *trNodePool) getNode(ref int32) *trNode {
	var (
		bucket = hash32(ref) & (p.hashSize - 1)
		i      = p.mFirst[bucket]
	)
	for i != -1 {
		if p.mNode[i].ref == ref {
			return &p.mNode[i]
		}
		i = p.mNext[i]
	}
	if p.nodeCnt >= p.maxNode {
		return nil
	}
	i = p.nodeCnt
	p.nodeCnt++
	p.mNode[i] = trNode{ref: ref}
	p.mNext[i] = p.mFirst[bucket]
	p.mFirst[bucket] = i
	return &p.mNode[i]
}

func (p *trNodePool) getNodeAtIdx(idx int32) *trNode {
	if idx >= 0 {
		return &p.mNode[idx]
	}
	return nil
}

func (p *trNodePool) exist2(node *trNode, ref int32) bool {
	var h = uint64(1) << (ref & 63)
	for {
		if node.ref == ref {
			return true
		}
		if node.hash&h == 0 {
			break
		}
		if node.pIdx >= 0 {
			node = &p.mNode[node.pIdx]
		} else {
			break
		}
	}
	return false
}

func (p *trNodePool) exist(idx, ref int32) int32 {
	var h = uint64(1) << (ref & 63)
	for idx >= 0 {
		var node = p.mNode[idx]
		if node.ref == ref {
			return idx
		}
		if node.hash&h == 0 {
			break
		}
		idx = node.pIdx
	}
	return -1
}

func (p *trNodePool) getIdx(n *trNode) int32 {
	return int32((uintptr(unsafe.Pointer(n)) - uintptr(unsafe.Pointer(&p.mNode[0]))) / unsafe.Sizeof(trNode{}))
}

////////////////////////// node queue //////////////////////////////

type nodes []*trNode

func (n *nodes) Pop() *trNode {
	var l = len(*n) - 1
	var x = (*n)[l]
	(*n)[l] = nil
	*n = (*n)[:l]
	return x
}

type trNodeQueue struct {
	mHeap nodes
}

func newNodeQueue(size int32) *trNodeQueue {
	var q = new(trNodeQueue)
	q.mHeap = make([]*trNode, 0, size+1)
	return q
}

func (q *trNodeQueue) push(n *trNode) {
	//fmt.Printf("%d %f (%f %f) (%f %f)\n", n.ref, n.total, n.l.X, n.l.Z, n.r.X, n.r.Z)
	q.mHeap = append(q.mHeap, n)
	up(q, len(q.mHeap)-1)
}

func (q *trNodeQueue) top() *trNode {
	return q.mHeap[0]
}

func (q *trNodeQueue) pop() *trNode {
	n := len(q.mHeap) - 1
	q.mHeap[0], q.mHeap[n] = q.mHeap[n], q.mHeap[0]
	down(q, 0, n)
	return (&q.mHeap).Pop()
}

func (q *trNodeQueue) empty() bool {
	return len(q.mHeap) == 0
}

func (q *trNodeQueue) fix(n *trNode) {
	for i, x := range q.mHeap {
		if x == n {
			_fix(q, i)
			return
		}
	}
}

func (q *trNodeQueue) clear() {
	q.mHeap = q.mHeap[0:0]
}

// heap

func _fix(h *trNodeQueue, i int) {
	if !down(h, i, len(h.mHeap)) {
		up(h, i)
	}
}

func up(h *trNodeQueue, j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || h.mHeap[i].total <= h.mHeap[j].total {
			break
		}
		h.mHeap[i], h.mHeap[j] = h.mHeap[j], h.mHeap[i]
		j = i
	}
}

func down(h *trNodeQueue, i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.mHeap[j2].total < h.mHeap[j1].total {
			j = j2 // = 2*i + 2  // right child
		}
		if h.mHeap[i].total <= h.mHeap[j].total {
			break
		}
		h.mHeap[i], h.mHeap[j] = h.mHeap[j], h.mHeap[i]
		i = j
	}
	return i > i0
}
