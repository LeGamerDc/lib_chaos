package detour

import (
	"container/heap"
	"lib_chaos/mesh"
	"unsafe"
)

type nodeStatus uint32

const (
	nodeOpen  nodeStatus = 1
	nodeClose nodeStatus = 2
)

type dtNode struct {
	pos    mesh.Vert
	cost   float64    // g
	total  float64    // g+h
	pIdx   int32      // parent node
	ref    int32      // which poly
	status nodeStatus // node status
}

type dtNodePool struct {
	mNode                       []dtNode // size = maxNodes
	mNext                       []int32  // size = maxNodes
	mFirst                      []int32  // size = hashSize
	maxNodes, hashSize, nodeCnt int32
}

func newNodePool(maxNodes int32) *dtNodePool {
	var p = new(dtNodePool)
	p.maxNodes = maxNodes
	p.hashSize = int32(nextPow2(uint32(maxNodes / 4)))
	p.mNode = make([]dtNode, p.maxNodes)
	p.mNext = make([]int32, p.maxNodes)
	p.mFirst = make([]int32, p.hashSize)
	memset(p.mFirst, -1)
	memset(p.mNext, -1)
	return p
}

func (p *dtNodePool) clear() {
	memset(p.mFirst, -1)
	p.nodeCnt = 0
}

func (p *dtNodePool) findNode(pid int32) *dtNode {
	var i = p.mFirst[hash32(pid)&(p.hashSize-1)]
	for i != -1 {
		if p.mNode[i].ref == pid {
			return &p.mNode[i]
		}
		i = p.mNext[i]
	}
	return nil
}

func (p *dtNodePool) getIdx(n *dtNode) int32 {
	return int32((uintptr(unsafe.Pointer(n)) - uintptr(unsafe.Pointer(&p.mNode[0]))) / unsafe.Sizeof(dtNode{}))
}

func (p *dtNodePool) getNodeAtIdx(idx int32) *dtNode {
	if idx >= 0 {
		return &p.mNode[idx]
	}
	return nil
}

func (p *dtNodePool) getNode(pid int32) *dtNode {
	var bucket = hash32(pid) & (p.hashSize - 1)
	var i = p.mFirst[bucket]
	for i != -1 {
		if p.mNode[i].ref == pid {
			return &p.mNode[i]
		}
		i = p.mNext[i]
	}
	if p.nodeCnt >= p.maxNodes {
		return nil
	}
	i = p.nodeCnt
	p.nodeCnt++

	p.mNode[i] = dtNode{
		ref: pid,
	}
	p.mNext[i] = p.mFirst[bucket]
	p.mFirst[bucket] = i
	return &p.mNode[i]
}

////////////////////////// node queue //////////////////////////////

type nodes []*dtNode

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
	*n = append(*n, x.(*dtNode))
}

func (n *nodes) Pop() interface{} {
	var l = len(*n) - 1
	var x = (*n)[l]
	(*n)[l] = nil
	*n = (*n)[:l]
	return x
}

type dtNodeQueue struct {
	mHeap nodes
}

func newNodeQueue(size int32) *dtNodeQueue {
	var q = new(dtNodeQueue)
	q.mHeap = make([]*dtNode, 0, size+1)
	return q
}

func (q *dtNodeQueue) push(n *dtNode) {
	heap.Push(&q.mHeap, n)
}

func (q *dtNodeQueue) top() *dtNode {
	return q.mHeap[0]
}

func (q *dtNodeQueue) pop() *dtNode {
	return heap.Pop(&q.mHeap).(*dtNode)
}

func (q *dtNodeQueue) empty() bool {
	return len(q.mHeap) == 0
}

func (q *dtNodeQueue) fix(n *dtNode) {
	for i, x := range q.mHeap {
		if x == n {
			heap.Fix(&q.mHeap, i)
			return
		}
	}
}

func (q *dtNodeQueue) clear() {
	q.mHeap = q.mHeap[0:0]
}
