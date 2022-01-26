package tr

import (
	"container/heap"
	"lib_chaos/mesh"
	"unsafe"
)

type trNode struct {
	turn  mesh.Vert // turn point
	l, r  mesh.Vert // l & r directly visible to turn
	cost  float64   // g
	total float64   // g+h
	pIdx  int32     // parent node
	ref   int32     // which tri
}

type trNodePool struct {
	mNode            []trNode
	maxNode, nodeCnt int32
}

func newNodePool(maxNode int32) *trNodePool {
	return &trNodePool{
		mNode:   make([]trNode, maxNode),
		maxNode: maxNode,
		nodeCnt: 0,
	}
}

func (p *trNodePool) clear() {
	p.nodeCnt = 0
}

func (p *trNodePool) getNode() *trNode {
	if p.nodeCnt >= p.maxNode {
		return nil
	}
	var ptr = &p.mNode[p.nodeCnt]
	p.nodeCnt++
	return ptr
}

func (p *trNodePool) getNodeAtIdx(idx int32) *trNode {
	if idx >= 0 {
		return &p.mNode[idx]
	}
	return nil
}

func (p *trNodePool) exist(idx, ref int32, stop func(int32) bool) bool {
	for idx >= 0 {
		var node = p.mNode[idx]
		if stop(node.ref) {
			return false
		}
		if node.ref == ref {
			return true
		}
	}
	return false
}

func (p *trNodePool) getIdx(n *trNode) int32 {
	return int32((uintptr(unsafe.Pointer(n)) - uintptr(unsafe.Pointer(&p.mNode[0]))) / unsafe.Sizeof(trNode{}))
}

////////////////////////// node queue //////////////////////////////

type nodes []*trNode

func (n *nodes) Len() int {
	return len(*n)
}

func (n *nodes) Less(i, j int) bool {
	return (*n)[i].total < (*n)[j].total
}

func (n *nodes) Swap(i, j int) {
	(*n)[i], (*n)[j] = (*n)[j], (*n)[i]
}

func (n *nodes) Push(x interface{}) {
	*n = append(*n, x.(*trNode))
}

func (n *nodes) Pop() interface{} {
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
	heap.Push(&q.mHeap, n)
}

func (q *trNodeQueue) top() *trNode {
	return q.mHeap[0]
}

func (q *trNodeQueue) pop() *trNode {
	return heap.Pop(&q.mHeap).(*trNode)
}

func (q *trNodeQueue) empty() bool {
	return len(q.mHeap) == 0
}

func (q *trNodeQueue) fix(n *trNode) {
	for i, x := range q.mHeap {
		if x == n {
			heap.Fix(&q.mHeap, i)
			return
		}
	}
}

func (q *trNodeQueue) clear() {
	q.mHeap = q.mHeap[0:0]
}
