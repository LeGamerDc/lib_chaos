package tra

import (
	"lib_chaos/mesh"
)

type Query struct {
	mesh *NavMesh

	nodePool  *trNodePool
	nodeQueue *trNodeQueue
	Path      []mesh.Vert
}

func NewQuery(mesh *NavMesh, size int32) *Query {
	var q = new(Query)
	q.mesh = mesh
	q.nodePool = newNodePool(size)
	q.nodeQueue = newNodeQueue(size)
	return q
}

func (q *Query) clear() {
	q.nodePool.clear()
	q.nodeQueue.clear()
}

func (q *Query) FindPath(startRef, endRef int32, startPos, endPos mesh.Vert) (outOfNodes bool) {
	var (
		tri          = &q.mesh.MTri[startRef]
		node         *trNode
		neighborNode *trNode
		ref          int32
		neighborRef  int32
		idx          int32

		cost              float64
		left, right, turn mesh.Vert
		lCut, rCut        mesh.Vert
	)
	for l := tri.Link; l != -1; l = q.mesh.MLink[l].Next {
		neighborRef = q.mesh.MLink[l].ToRef
		left, right, _ = q.mesh.getPortal(startRef, neighborRef)
		node = q.nodePool.getNode()
		*node = trNode{
			turn:  startPos,
			l:     left,
			r:     right,
			cost:  0,
			total: mesh.DistPtPtThroughSeg2D(startPos, endPos, left, right),
			pIdx:  -1,
			ref:   neighborRef,
		}
		q.nodeQueue.push(node)
	}
	for !q.nodeQueue.empty() {
		node = q.nodeQueue.pop()
		if node.ref == endRef { // Path found !
			break
		}
		ref = node.ref
		tri = &q.mesh.MTri[ref]
		idx = q.nodePool.getIdx(node)
		//if node.pIdx == -1 {
		//	fmt.Printf("-1 -> %d\n", ref)
		//} else {
		//	fmt.Printf("%d -> %d\n", q.nodePool.getNodeAtIdx(node.pIdx).ref, ref)
		//}
		left, right, turn = node.l, node.r, node.turn
		for l := tri.Link; l != -1; l = q.mesh.MLink[l].Next {
			neighborRef = q.mesh.MLink[l].ToRef
			if q.nodePool.exist(idx, neighborRef) { // skip walked Path
				continue
			}
			var ll, rr, _ = q.mesh.getPortal(ref, neighborRef)
			// case 1: turn on left-right
			if _, ds := mesh.DistPtSegSqr2D(turn, left, right); ds < mesh.Eqs {
				neighborNode = q.nodePool.getNode()
				if neighborNode == nil {
					outOfNodes = true
					goto END
				}
				*neighborNode = trNode{
					turn:  turn,
					l:     ll,
					r:     rr,
					cost:  node.cost,
					total: node.cost + mesh.DistPtPtThroughSeg2D(turn, endPos, ll, rr),
					pIdx:  idx,
					ref:   neighborRef,
				}
				q.nodeQueue.push(neighborNode)
				continue
			}
			lCut, rCut = ll, rr
			// case 2: handle left exceed
			if mesh.TriArea2D(turn, left, ll) >= mesh.Eqs {
				var (
					cut    = rr
					finish = true
				)
				if mesh.TriArea2D(turn, left, rr) < -mesh.Eqs {
					finish = false
					var _, t, _ = mesh.IntersectSegSeg2D(turn, left, ll, rr)
					cut = mesh.VInter(ll, rr, t)
				}
				neighborNode = q.nodePool.getNode()
				if neighborNode == nil {
					outOfNodes = true
					goto END
				}
				cost = node.cost + mesh.VDist(turn, left)
				*neighborNode = trNode{
					turn:  left,
					l:     ll,
					r:     cut,
					cost:  cost,
					total: cost + mesh.DistPtPtThroughSeg2D(left, endPos, ll, cut),
					pIdx:  idx,
					ref:   neighborRef,
				}
				q.nodeQueue.push(neighborNode)
				if finish {
					continue
				}
				lCut = cut
			}
			// case 3: handle right exceed
			if mesh.TriArea2D(turn, right, rr) <= -mesh.Eqs {
				var (
					cut    = ll
					finish = true
				)
				if mesh.TriArea2D(turn, right, ll) > mesh.Eqs {
					finish = false
					var _, t, _ = mesh.IntersectSegSeg2D(turn, right, ll, rr)
					cut = mesh.VInter(ll, rr, t)
				}
				neighborNode = q.nodePool.getNode()
				if neighborNode == nil {
					outOfNodes = true
					goto END
				}
				cost = node.cost + mesh.VDist(turn, right)
				*neighborNode = trNode{
					turn:  right,
					l:     cut,
					r:     rr,
					cost:  cost,
					total: cost + mesh.DistPtPtThroughSeg2D(right, endPos, cut, rr),
					pIdx:  idx,
					ref:   neighborRef,
				}
				q.nodeQueue.push(neighborNode)
				if finish {
					continue
				}
				rCut = cut
			}
			// case 4: handle segment contained by turn<left-right>
			neighborNode = q.nodePool.getNode()
			if neighborNode == nil {
				outOfNodes = true
				goto END
			}
			*neighborNode = trNode{
				turn:  turn,
				l:     lCut,
				r:     rCut,
				cost:  node.cost,
				total: node.cost + mesh.DistPtPtThroughSeg2D(turn, endPos, lCut, rCut),
				pIdx:  idx,
				ref:   neighborRef,
			}
			q.nodeQueue.push(neighborNode)
		}
	}
END:
	if outOfNodes || node.ref != endRef {
		return true
	}
	// retrieve Path
	var length = 1
	for n := node; n != nil; n = q.nodePool.getNodeAtIdx(n.pIdx) {
		length++
	}
	if cap(q.Path) < length {
		q.Path = make([]mesh.Vert, length)
	} else {
		q.Path = q.Path[0:length]
	}
	length--
	q.Path[length] = endPos
	for n := node; n != nil; n = q.nodePool.getNodeAtIdx(n.pIdx) {
		length--
		q.Path[length] = n.turn
	}
	return
}
