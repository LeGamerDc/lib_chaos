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

func (q *Query) Clear() {
	q.nodePool.clear()
	q.nodeQueue.clear()
}

func (q *Query) FindPath(startRef, endRef int32, startPos, endPos mesh.Vert) (success, outOfNodes bool) {
	var (
		tri          = &q.mesh.MTri[startRef]
		node         *trNode
		neighborNode *trNode
		parentNode   *trNode
		ref          int32
		neighborRef  int32
		//parentRef   int32
		idx  int32
		hash uint64

		cost, total       float64
		left, right, turn mesh.Vert
		lCut, rCut        mesh.Vert
	)
	node = q.nodePool.getNode(startRef)
	node.turn = startPos
	node.l = startPos
	node.r = startPos
	node.cost = 0
	node.total = mesh.VDist(startPos, endPos)
	node.pIdx = -1
	node.ref = startRef
	node.hash = 0
	node.status = nodeOpen
	q.nodeQueue.push(node)
	for !q.nodeQueue.empty() {
		node = q.nodeQueue.pop()
		node.status = nodeClose
		if node.ref == endRef { // Path found !
			break
		}
		//fmt.Printf("%d %f\n", node.ref, node.total)
		ref = node.ref
		tri = &q.mesh.MTri[ref]
		idx = q.nodePool.getIdx(node)
		hash = node.hash | (uint64(1) << (ref & 63))
		//parentRef = -1
		//if node.pIdx != -1 {
		//	parentRef = q.nodePool.getNodeAtIdx(node.pIdx).ref
		//}
		parentNode = nil
		if node.pIdx != -1 {
			parentNode = q.nodePool.getNodeAtIdx(node.pIdx)
		}
		//fmt.Printf("%d -> %d %f\n", parentRef, ref, node.total)
		left, right, turn = node.l, node.r, node.turn
		for l := tri.Link; l != -1; l = q.mesh.MLink[l].Next {
			neighborRef = q.mesh.MLink[l].ToRef
			if parentNode != nil && q.nodePool.exist2(parentNode, neighborRef) {
				continue
			}
			//if parentRef == neighborRef {
			//	continue
			//}
			var ll, rr, _ = q.mesh.getPortal(ref, neighborRef)
			// case 1: turn on left-right
			if _, ds := mesh.DistPtSegSqr2D(turn, left, right); ds < mesh.Eqs {
				neighborNode = q.nodePool.getNode(neighborRef)
				if neighborNode == nil {
					outOfNodes = true
					continue
				}

				cost = node.cost
				total = cost + mesh.DistPtPtThroughSeg2D(turn, endPos, ll, rr)
				if neighborNode.status == 0 || total < neighborNode.total {
					neighborNode.turn = turn
					neighborNode.l = ll
					neighborNode.r = rr
					neighborNode.cost = cost
					neighborNode.total = total
					neighborNode.pIdx = idx
					neighborNode.hash = hash
					switch neighborNode.status {
					case 0, nodeClose:
						neighborNode.status = nodeOpen
						q.nodeQueue.push(neighborNode)
					case nodeOpen:
						q.nodeQueue.fix(neighborNode)
					}
				}
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
				neighborNode = q.nodePool.getNode(neighborRef)
				if neighborNode == nil {
					outOfNodes = true
					continue
				}
				cost = node.cost + mesh.VDist(turn, left)
				total = cost + mesh.DistPtPtThroughSeg2D(left, endPos, ll, cut)
				if neighborNode.status == 0 || total < neighborNode.total {
					neighborNode.turn = left
					neighborNode.l = ll
					neighborNode.r = cut
					neighborNode.cost = cost
					neighborNode.total = total
					neighborNode.pIdx = idx
					neighborNode.hash = hash
					switch neighborNode.status {
					case 0, nodeClose:
						neighborNode.status = nodeOpen
						q.nodeQueue.push(neighborNode)
					case nodeOpen:
						q.nodeQueue.fix(neighborNode)
					}
				}
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
				neighborNode = q.nodePool.getNode(neighborRef)
				if neighborNode == nil {
					outOfNodes = true
					continue
				}
				cost = node.cost + mesh.VDist(turn, right)
				total = cost + mesh.DistPtPtThroughSeg2D(right, endPos, cut, rr)
				if neighborNode.status == 0 || total < neighborNode.total {
					neighborNode.turn = right
					neighborNode.l = cut
					neighborNode.r = rr
					neighborNode.cost = cost
					neighborNode.total = total
					neighborNode.pIdx = idx
					neighborNode.hash = hash
					switch neighborNode.status {
					case 0, nodeClose:
						neighborNode.status = nodeOpen
						q.nodeQueue.push(neighborNode)
					case nodeOpen:
						q.nodeQueue.fix(neighborNode)
					}
				}
				if finish {
					continue
				}
				rCut = cut
			}
			// case 4: handle segment contained by turn<left-right>
			neighborNode = q.nodePool.getNode(neighborRef)
			if neighborNode == nil {
				outOfNodes = true
				continue
			}
			cost = node.cost
			total = cost + mesh.DistPtPtThroughSeg2D(turn, endPos, lCut, rCut)
			if neighborNode.status == 0 || total < neighborNode.total {
				neighborNode.turn = turn
				neighborNode.l = lCut
				neighborNode.r = rCut
				neighborNode.cost = cost
				neighborNode.total = total
				neighborNode.pIdx = idx
				neighborNode.hash = hash
				switch neighborNode.status {
				case 0, nodeClose:
					neighborNode.status = nodeOpen
					q.nodeQueue.push(neighborNode)
				case nodeOpen:
					q.nodeQueue.fix(neighborNode)
				}
			}
		}
	}
	if node.ref != endRef {
		success = false
		return
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
	success = true
	return
}

//
//func (q *Query) FindPath2(startRef, endRef int32, startPos, endPos mesh.Vert) (outOfNodes bool) {
//	var (
//		node         *trNode
//		neighborNode *trNode
//		idx          int32
//		cost, total  float64
//		turn         mesh.Vert
//	)
//	node = q.nodePool.getNode2(startPos)
//	node.l = startPos
//	node.r = startPos
//	node.cost = 0
//	node.total = mesh.VDist(startPos, endPos)
//	node.pIdx = -1
//	node.ref = startRef
//	node.parentRef = -1
//	node.status = nodeOpen
//	q.nodeQueue.push(node)
//	for !q.nodeQueue.empty() {
//		node = q.nodeQueue.pop()
//		node.status = nodeClose
//		if node.ref == endRef { // path found !
//			break
//		}
//		//if node.pIdx != -1 {
//		//	fmt.Printf("%d -> %d\n", q.nodePool.getNodeAtIdx(node.pIdx).ref, node.ref)
//		//}
//		idx = q.nodePool.getIdx(node)
//		turn = node.turn
//		//fmt.Printf("(%f %f) %f %f\n", node.turn.X, node.turn.Z, node.cost, node.total)
//		var f func(left, right mesh.Vert, ref, parentRef int32)
//		f = func(left, right mesh.Vert, ref, parentRef int32) {
//			//fmt.Printf("%d, %d\n", parentRef, ref)
//			var tri = &q.mesh.MTri[ref]
//			for l := tri.Link; l != -1; l = q.mesh.MLink[l].Next {
//				var neighborRef = q.mesh.MLink[l].ToRef
//				if neighborRef == parentRef { // skip parent
//					continue
//				}
//				if neighborRef == endRef {
//					fmt.Println("fuck")
//				}
//				var ll, rr, _ = q.mesh.getPortal(ref, neighborRef)
//				// case 1: turn on left-right
//				if _, ds := mesh.DistPtSegSqr2D(turn, left, right); ds < mesh.Eqs {
//					if neighborRef == endRef {
//						node.ref = endRef
//						node.total = node.cost + mesh.VDist(turn, endPos)
//						node.status = nodeOpen
//						q.nodeQueue.push(node)
//					} else {
//						f(ll, rr, neighborRef, ref)
//					}
//					continue
//				}
//				var lCut, rCut = ll, rr
//				// case 2: handle left exceed
//				if mesh.TriArea2D(turn, left, ll) >= mesh.Eqs {
//					var (
//						cut    = rr
//						finish = true
//					)
//					if mesh.TriArea2D(turn, left, rr) < -mesh.Eqs {
//						finish = false
//						var _, t, _ = mesh.IntersectSegSeg2D(turn, left, ll, rr)
//						cut = mesh.VInter(ll, rr, t)
//					}
//					neighborNode = q.nodePool.getNode2(left)
//					if neighborNode == nil {
//						outOfNodes = true
//						continue
//					}
//					cost = node.cost + mesh.VDist(turn, left)
//					total = cost + mesh.DistPtPtThroughSeg2D(left, endPos, ll, cut)
//					if neighborNode.status == 0 || cost < neighborNode.cost {
//						neighborNode.l = ll
//						neighborNode.r = cut
//						neighborNode.cost = cost
//						neighborNode.total = total
//						neighborNode.pIdx = idx
//						neighborNode.ref = neighborRef
//						neighborNode.parentRef = ref
//						switch neighborNode.status {
//						case 0, nodeClose:
//							neighborNode.status = nodeOpen
//							q.nodeQueue.push(neighborNode)
//						case nodeOpen:
//							q.nodeQueue.fix(neighborNode)
//						}
//					}
//					if finish {
//						continue
//					}
//					lCut = cut
//				}
//				// case 3: handle right exceed
//				if mesh.TriArea2D(turn, right, rr) <= -mesh.Eqs {
//					var (
//						cut    = ll
//						finish = true
//					)
//					if mesh.TriArea2D(turn, right, ll) > mesh.Eqs {
//						finish = false
//						var _, t, _ = mesh.IntersectSegSeg2D(turn, right, ll, rr)
//						cut = mesh.VInter(ll, rr, t)
//					}
//					neighborNode = q.nodePool.getNode2(right)
//					if neighborNode == nil {
//						outOfNodes = true
//						continue
//					}
//					cost = node.cost + mesh.VDist(turn, right)
//					total = cost + mesh.DistPtPtThroughSeg2D(right, endPos, cut, rr)
//					if neighborNode.status == 0 || cost < neighborNode.cost {
//						neighborNode.l = cut
//						neighborNode.r = rr
//						neighborNode.cost = cost
//						neighborNode.total = total
//						neighborNode.pIdx = idx
//						neighborNode.ref = neighborRef
//						neighborNode.parentRef = ref
//						switch neighborNode.status {
//						case 0, nodeClose:
//							neighborNode.status = nodeOpen
//							q.nodeQueue.push(neighborNode)
//						case nodeOpen:
//							q.nodeQueue.fix(neighborNode)
//						}
//					}
//					if finish {
//						continue
//					}
//					rCut = cut
//				}
//				// case 4: handle segment contained by turn<left-right>
//				if !mesh.VEqual(lCut, rCut) {
//					if neighborRef == endRef {
//						node.ref = endRef
//						node.total = node.cost + mesh.VDist(turn, endPos)
//						node.status = nodeOpen
//						q.nodeQueue.push(node)
//					} else {
//						f(lCut, rCut, neighborRef, ref)
//					}
//				}
//			}
//		}
//		f(node.l, node.r, node.ref, node.parentRef)
//	}
//	if node.ref != endRef {
//		return true
//	}
//	// retrieve Path
//	var length = 1
//	for n := node; n != nil; n = q.nodePool.getNodeAtIdx(n.pIdx) {
//		length++
//	}
//	if cap(q.Path) < length {
//		q.Path = make([]mesh.Vert, length)
//	} else {
//		q.Path = q.Path[0:length]
//	}
//	length--
//	q.Path[length] = endPos
//	for n := node; n != nil; n = q.nodePool.getNodeAtIdx(n.pIdx) {
//		length--
//		q.Path[length] = n.turn
//	}
//	return
//}
