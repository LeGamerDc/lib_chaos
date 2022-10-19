package cdt

import (
	"lib_chaos/common"
	"lib_chaos/mesh"
)

var dummyTriangle = Triangle{v0: -1, v1: -1, v2: -1, n0: -1, n1: -1, n2: -1}

func (cdt *CDT) Shrink() {
	var (
		m  = make(map[TriIndex]TriIndex)
		id TriIndex
	)
	m[-1] = -1
	for it := range cdt.MTri {
		var t = &cdt.MTri[it]
		if t.v0 >= 0 { // not dummy
			m[TriIndex(it)] = id
			id++
		}
	}
	for it := range cdt.MTri {
		if slot, ok := m[TriIndex(it)]; ok {
			cdt.MTri[slot] = cdt.MTri[it]
			var t = &cdt.MTri[slot]
			t.n0 = m[t.n0]
			t.n1 = m[t.n1]
			t.n2 = m[t.n2]
		}
	}
	cdt.MTri = cdt.MTri[:id]
}

func (cdt *CDT) Culling() {
	var (
		depth = make(map[TriIndex]int)
		seeds = make(map[TriIndex]struct{})
		dfs   func(index TriIndex, d int)
		dd    int
	)
	dfs = func(it TriIndex, d int) {
		depth[it] = d
		var t = &cdt.MTri[it]
		if _, ok := depth[t.n0]; !ok && t.n0 >= 0 {
			if !cdt.isConstrained(t.v0, t.v1) {
				//fmt.Printf("[%d %d]", it, t.n0)
				dfs(t.n0, d)
			} else {
				seeds[t.n0] = struct{}{}
			}
		}
		if _, ok := depth[t.n1]; !ok && t.n1 >= 0 {
			if !cdt.isConstrained(t.v1, t.v2) {
				//fmt.Printf("[%d %d]", it, t.n1)
				dfs(t.n1, d)
			} else {
				seeds[t.n1] = struct{}{}
			}
		}
		if _, ok := depth[t.n2]; !ok && t.n2 >= 0 {
			if !cdt.isConstrained(t.v2, t.v0) {
				dfs(t.n2, d)
			} else {
				seeds[t.n2] = struct{}{}
			}
		}
	}
	seeds[cdt.mNeighbor[0][0]] = struct{}{}
	for len(seeds) > 0 {
		var ts = seeds
		seeds = make(map[TriIndex]struct{})
		for it := range ts {
			if _, ok := depth[it]; !ok {
				dfs(it, dd)
			}
		}
		dd++
	}
	for it := range cdt.MTri {
		if depth[TriIndex(it)]%2 == 0 {
			cdt.makeDummy(TriIndex(it))
		}
	}
}

func (cdt *CDT) isConstrained(iv0, iv1 VertIndex) bool {
	if _, ok := cdt.MCons[Edge{v0: iv0, v1: iv1}]; ok {
		return true
	}
	if _, ok := cdt.MCons[Edge{v0: iv1, v1: iv0}]; ok {
		return true
	}
	return false
}

func (cdt *CDT) fixEdge(ia, ib VertIndex, ob bool) {
	//fmt.Printf("(%d %d) %v\n", ia, ib, ob)
	if ob {
		cdt.MCons[Edge{v0: ia, v1: ib}] = struct{}{}
	}
}

func (cdt *CDT) InsertEdge(ia, ib VertIndex) {
	//for e := range cdt.MCons {
	//	var (
	//		x, y     = cdt.MVert[e.v0], cdt.MVert[e.v1]
	//		a, b     = cdt.MVert[ia], cdt.MVert[ib]
	//		s, t, ok = mesh.IntersectSegSeg2D(a, b, x, y)
	//	)
	//	if ok && s >= mesh.Eps && s <= 1-mesh.Eps && t >= mesh.Eps && t <= 1-mesh.Eps {
	//		var (
	//			c  = mesh.VInter(a, b, s)
	//			ic = cdt.InsertVert(c)
	//		)
	//		cdt.InsertEdge(ia, ic)
	//		cdt.InsertEdge(ic, ib)
	//		return
	//	}
	//}
	cdt.insertEdge(ia, ib, true)
}

func (cdt *CDT) insertEdge(ia, ibb VertIndex, ob bool) {
	var ib = ibb
	if ia == ib { // bad param
		return
	}
	if cdt.vertexShareTri(ia, ib) != -1 { // already edge
		cdt.fixEdge(ia, ib, ob)
		return
	}
	var (
		va         = cdt.MVert[ia]
		vb         = cdt.MVert[ib]
		it, il, ir = cdt.head(ia, cdt.mNeighbor[ia], cdt.MVert[ia], cdt.MVert[ib])
		iv         = ia
		crossTri   = []TriIndex{it}
		left       = []VertIndex{il}
		right      = []VertIndex{ir}
	)
	if it == -1 {
		cdt.fixEdge(ia, il, ob)
		cdt.insertEdge(il, ib, ob)
		return
	}
	for {
		var (
			itop = cdt.opposedTri(it, iv)
			ivop = cdt.opposedVert(itop, it)
			vop  = cdt.MVert[ivop]
		)
		crossTri = append(crossTri, itop)
		it = itop
		var cxz = mesh.VCrossXz(mesh.VSub(vb, va), mesh.VSub(vop, va))
		if cxz > mesh.Eqs { // left
			left = append(left, ivop)
			iv = il
			il = ivop
		} else if cxz < -mesh.Eqs { // right
			right = append(right, ivop)
			iv = ir
			ir = ivop
		} else {
			ib = ivop
		}
		if ivop == ib {
			break
		}
	}
	for _, i := range crossTri {
		cdt.makeDummy(i)
	}
	// triangulate pseudo-polygons both sides
	common.Reverse(right)
	var (
		itl = cdt.triangulatePseudoPolygon(ia, ib, left)
		itr = cdt.triangulatePseudoPolygon(ib, ia, right)
	)
	cdt.changeTriNeighbor(itl, -1, itr)
	cdt.changeTriNeighbor(itr, -1, itl)
	cdt.reconstructNeighbor(left)
	cdt.reconstructNeighbor(right)
	cdt.report("InsertEdge", itl, itr)
	cdt.fixEdge(ia, ib, ob)
	if ib != ibb {
		cdt.insertEdge(ib, ibb, ob)
	}
}

func (cdt *CDT) reconstructNeighbor(points []VertIndex) {
	if len(points) < 3 {
		return
	}
	var f = func(ia, ib VertIndex) {
		var ts = common.ToSlice(common.Intersection(common.ToSet(cdt.mNeighbor[ia]), common.ToSet(cdt.mNeighbor[ib])))
		if len(ts) == 2 {
			cdt.updateOpposedNeighbor(ts[0], ts[1], ia, ib)
			cdt.updateOpposedNeighbor(ts[1], ts[0], ia, ib)
		} else {
			panic("fuck 1")
		}
	}
	for i := 2; i < len(points); i++ {
		if points[i] == points[i-2] {
			f(points[i], points[i-1])
		}
	}
}

func (cdt *CDT) vertexShareTri(ia, ib VertIndex) TriIndex {
	for _, it0 := range cdt.mNeighbor[ia] {
		for _, it1 := range cdt.mNeighbor[ib] {
			if it0 == it1 {
				return it0
			}
		}
	}
	return -1
}

func (cdt *CDT) head(ia VertIndex, neighbor []TriIndex, a, b mesh.Vert) (it TriIndex, il, ir VertIndex) {
	for _, it = range neighbor {
		ir = cdt.triNextVert(it, ia)
		il = cdt.triNextVert(it, ir)
		var (
			pl = cdt.MVert[il]
			pr = cdt.MVert[ir]
			cl = mesh.VCrossXz(mesh.VSub(b, a), mesh.VSub(pl, a))
			cr = mesh.VCrossXz(mesh.VSub(b, a), mesh.VSub(pr, a))
		)
		if cr < -mesh.Eqs {
			if cl > mesh.Eqs {
				return it, il, ir
			} else if cl > -mesh.Eqs {
				return -1, il, ir
			}
		}
	}
	panic("can not find head")
}

func (cdt *CDT) makeDummy(it TriIndex) {
	var t = &cdt.MTri[it]
	cdt.removeVertNeighbor(t.v0, it)
	cdt.removeVertNeighbor(t.v1, it)
	cdt.removeVertNeighbor(t.v2, it)

	cdt.changeTriNeighbor(t.n0, it, -1)
	cdt.changeTriNeighbor(t.n1, it, -1)
	cdt.changeTriNeighbor(t.n2, it, -1)
	*t = dummyTriangle
	cdt.mDummy = append(cdt.mDummy, it)
}

func (cdt *CDT) triangulatePseudoPolygon(ia, ib VertIndex, points []VertIndex) TriIndex {
	if len(points) == 0 {
		return cdt.vertexShareTri(ia, ib)
	}
	var (
		ic, l, r = cdt.findDelaunayPoint(ia, ib, points)
		lt       = cdt.triangulatePseudoPolygon(ia, ic, l)
		rt       = cdt.triangulatePseudoPolygon(ic, ib, r)
		it       = cdt.newTriangle2()
	)
	cdt.MTri[it] = Triangle{v0: ia, v1: ib, v2: ic, n0: -1, n1: rt, n2: lt}
	// update triangle neighbor
	if lt != -1 {
		if len(l) == 0 {
			cdt.updateOpposedNeighbor(lt, it, ia, ic)
		} else {
			cdt.updateNeighbor(lt, it, ia)
		}
	}
	if rt != -1 {
		if len(r) == 0 {
			cdt.updateOpposedNeighbor(rt, it, ib, ic)
		} else {
			cdt.updateNeighbor(rt, it, ic)
		}
	}
	// update vertex neighbor
	cdt.insertVertNeighbor(ia, it)
	cdt.insertVertNeighbor(ib, it)
	cdt.insertVertNeighbor(ic, it)
	//cdt.report(fmt.Sprintf("triangulatePseudoPolygon %d %d %d", len(l), len(r), it), lt, rt)
	return it
}

func (cdt *CDT) updateOpposedNeighbor(it, ito TriIndex, ia, ib VertIndex) {
	var t = &cdt.MTri[it]
	if cmp(ia, ib, t.v0, t.v1) {
		t.n0 = ito
	} else if cmp(ia, ib, t.v1, t.v2) {
		t.n1 = ito
	} else if cmp(ia, ib, t.v2, t.v0) {
		t.n2 = ito
	} else {
		panic("fuck")
	}
}

func cmp(a0, b0, a1, b1 VertIndex) bool {
	return (a0 == a1 && b0 == b1) || (a0 == b1 && b0 == a1)
}

func (cdt *CDT) updateNeighbor(it, ito TriIndex, ia VertIndex) {
	var t = &cdt.MTri[it]
	switch ia {
	case t.v0:
		t.n0 = ito
	case t.v1:
		t.n1 = ito
	case t.v2:
		t.n2 = ito
	}
}

func (cdt *CDT) newTriangle2() TriIndex {
	if l := len(cdt.mDummy); l > 0 {
		var it = cdt.mDummy[l-1]
		cdt.mDummy = cdt.mDummy[:l-1]
		return it
	}
	return cdt.newTriangle()
}

func (cdt *CDT) findDelaunayPoint(ia, ib VertIndex, points []VertIndex) (mid VertIndex, l, r []VertIndex) {
	var (
		a = cdt.MVert[ia]
		b = cdt.MVert[ib]
		i = 0
		c = cdt.MVert[points[0]]
	)
	for idx, iv := range points[1:] {
		var v = cdt.MVert[iv]
		if mesh.InTriOuterCircle(v, a, b, c) {
			i = idx + 1
			c = cdt.MVert[iv]
		}
	}
	return points[i], points[:i], points[i+1:]
}
