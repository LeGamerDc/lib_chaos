package detour

import (
	"fmt"
	"lib_chaos/mesh"
)

type Query struct {
	nodePath  []int32
	mesh      *NavMesh
	nodePool  *dtNodePool
	nodeQueue *dtNodeQueue
}

type QueryResult interface {
	Append(x, y, z float64)
	LastPos() mesh.Vert
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

func (q *Query) FindPath(startPos, endPos mesh.Vert, result QueryResult) bool {
	var (
		startRef, endRef int32
		ok               bool
	)
	startRef, startPos, ok = q.mesh.LocatePoly(startPos)
	endRef, endPos, _ = q.mesh.LocatePoly(endPos)
	fmt.Println(startPos, endPos)
	q.clear()
	if _, _, ok = q.findPath(startRef, endRef, startPos, endPos); !ok {
		return false
	}
	return q.pullPath(startRef, endRef, startPos, endPos, result)
}

func (q *Query) appendPortal(startIndex, endIndex int, endPos mesh.Vert, result QueryResult) (ok bool) {
	var (
		startPos    = result.LastPos()
		fr, tr      int32
		fp, tp      *Poly
		left, right mesh.Vert
		k           float64
		pos         mesh.Vert
	)
	for i := startIndex; i < endIndex; i++ {
		fr = q.nodePath[i]
		tr = q.nodePath[i+1]
		fp = &q.mesh.MPoly[fr]
		tp = &q.mesh.MPoly[tr]
		if fp.AreaId == tp.AreaId {
			continue
		}
		left, right, ok = q.mesh.getPortal(fr, tr)
		if !ok {
			return
		}
		_, k, ok = mesh.IntersectSegSeg2D(startPos, endPos, left, right)
		if !ok {
			return
		}
		pos = mesh.VInter(left, right, k)
		result.Append(pos.X, pos.Y, pos.Z)
	}
	return true
}

func (q *Query) pullPath(startRef, endRef int32, startPos, endPos mesh.Vert, result QueryResult) bool {
	result.Append(startPos.X, startPos.Y, startPos.Z)
	var finalRef int32 = -1
	if len(q.nodePath) > 1 {
		var (
			portalApex, portalLeft, portalRight mesh.Vert
			left, right                         mesh.Vert
			apexIndex, leftIndex, rightIndex    int
			ok                                  bool
		)
		finalRef = q.nodePath[len(q.nodePath)-1]
		for i := 0; i < len(q.nodePath); i++ {
			if i+1 < len(q.nodePath) {
				if left, right, ok = q.mesh.getPortal(q.nodePath[i], q.nodePath[i+1]); !ok {
					return false
				}
				// portalApex on segment of portal
				if _, d := mesh.DistPtSegSqr2D(portalApex, left, right); d < mesh.Eqs {
					continue
				}
			} else {
				left, right = endPos, endPos
			}
			// handle left
			if mesh.TriArea2D(portalApex, portalLeft, left) >= 0 { // tighten left
				if mesh.VEqual(portalApex, portalLeft) || mesh.TriArea2D(portalApex, portalRight, left) < mesh.Eps*mesh.Eps {
					portalLeft = left
				} else {
					if !q.appendPortal(apexIndex, rightIndex, portalRight, result) {
						return false
					}
					portalApex = portalRight
					result.Append(portalApex.X, portalApex.Y, portalApex.Z)
					portalLeft, portalRight = portalApex, portalApex
					apexIndex, leftIndex = rightIndex, rightIndex

					//restart
					i = apexIndex
					continue
				}
			}
			// handle right
			if mesh.TriArea2D(portalApex, portalRight, right) <= 0 { //tighten right
				if mesh.VEqual(portalApex, portalRight) || mesh.TriArea2D(portalApex, portalLeft, right) > -mesh.Eps*mesh.Eps {
					portalRight = right
				} else {
					if !q.appendPortal(apexIndex, leftIndex, portalLeft, result) {
						return false
					}
					portalApex = portalLeft
					result.Append(portalApex.X, portalApex.Y, portalApex.Z)
					portalLeft, portalRight = portalApex, portalApex
					apexIndex, rightIndex = leftIndex, leftIndex

					//restart
					i = apexIndex
					continue
				}
			}
		}
		if !q.appendPortal(apexIndex, len(q.nodePath)-1, endPos, result) {
			return false
		}
	}
	if finalRef == endRef {
		result.Append(endPos.X, endPos.Y, endPos.Z)
	}
	return true
}

func (q *Query) findPath(startRef, endRef int32, startPos, endPos mesh.Vert) (
	partial, outOfNodes, success bool) {
	var (
		startNode        = q.nodePool.getNode(startRef)
		lastBestNodeCost = mesh.VDist(startPos, endPos)
		lastBestNode     = startNode
		bestNode         *dtNode
		neighborNode     *dtNode
		bestPoly         *Poly
		parentRef        int32

		cost, heuristic, total float64
	)
	*startNode = dtNode{
		pos:    startPos,
		pIdx:   -1,
		ref:    startRef,
		status: nodeOpen,
		total:  lastBestNodeCost,
	}
	q.nodeQueue.push(startNode)
	for !q.nodeQueue.empty() {
		bestNode = q.nodeQueue.pop()
		bestNode.status &= ^nodeOpen
		bestNode.status |= nodeClose
		if bestNode.ref == endRef {
			lastBestNode = bestNode
			break
		}
		bestPoly = &q.mesh.MPoly[bestNode.ref]
		parentRef = -1
		if bestNode.pIdx >= 0 {
			parentRef = q.nodePool.getNodeAtIdx(bestNode.pIdx).ref
		}
		for l := bestPoly.Link; l != -1; l = q.mesh.MLink[l].Next {
			var neighborRef = q.mesh.MLink[l].ToRef
			if neighborRef == parentRef { // skip parent
				continue
			}
			if q.mesh.MPoly[neighborRef].AreaId != bestPoly.AreaId {
				continue
			}
			neighborNode = q.nodePool.getNode(neighborRef)
			if neighborNode == nil {
				outOfNodes = true
				continue
			}
			// first visit, calculate pos
			if neighborNode.status == 0 {
				neighborNode.pos, _ = q.mesh.edgeMidPoint(bestNode.ref, neighborRef)
			}

			// calculate cost + heuristic
			if neighborRef == endRef {
				cost = bestNode.cost + mesh.VDist(bestNode.pos, neighborNode.pos) + mesh.VDist(neighborNode.pos, endPos)
				heuristic = 0
			} else {
				cost = bestNode.cost + mesh.VDist(bestNode.pos, neighborNode.pos)
				heuristic = mesh.VDist(neighborNode.pos, endPos)
			}
			total = cost + heuristic

			// skip non-optimal node
			switch neighborNode.status {
			case nodeOpen, nodeClose:
				if neighborNode.total <= total {
					continue
				}
			}

			neighborNode.pIdx = q.nodePool.getIdx(bestNode)
			neighborNode.status &= ^nodeClose
			neighborNode.cost = cost
			neighborNode.total = total

			switch neighborNode.status {
			case 0:
				neighborNode.status = nodeOpen
				q.nodeQueue.push(neighborNode)
			case nodeOpen:
				q.nodeQueue.fix(neighborNode)
			}
			if heuristic < lastBestNodeCost {
				lastBestNodeCost = heuristic
				lastBestNode = neighborNode
			}
		}
	}
	if lastBestNode.ref != endRef {
		partial = true
	}
	success = q.retrievePath(lastBestNode)
	return
}

func (q *Query) retrievePath(endNode *dtNode) bool {
	var (
		curNode = endNode
		length  = 0
	)
	for curNode != nil {
		length++
		curNode = q.nodePool.getNodeAtIdx(curNode.pIdx)
	}
	if cap(q.nodePath) < length {
		q.nodePath = make([]int32, length)
	} else {
		q.nodePath = q.nodePath[0:length]
	}

	curNode = endNode
	for i := length - 1; i >= 0; i-- {
		q.nodePath[i] = curNode.ref
		curNode = q.nodePool.getNodeAtIdx(curNode.pIdx)
	}
	return true
}
