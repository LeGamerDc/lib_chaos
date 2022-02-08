package cdt

import (
	"lib_chaos/common"
	"lib_chaos/mesh"
	"math"
)

type (
	VertIndex int32
	TriIndex  int32
)

type Triangle struct {
	v0, v1, v2 VertIndex // vert index, counter clock-wise in x-z plane
	n0, n1, n2 TriIndex  // neighbor triangle index
}
type Edge struct {
	v0, v1 VertIndex
}

type CDT struct {
	mVert []mesh.Vert
	mTri  []Triangle
	mCons []Edge // constrained edge

	// cache
	locator   *Locator
	mNeighbor [][]TriIndex // adjacent triangles to vert
}

func (cdt *CDT) Init(min, max mesh.Vert, np int) {
	var (
		v0       = mesh.Vert{X: min.X, Z: min.Z}
		v1       = mesh.Vert{X: max.X, Z: min.Z}
		v2       = mesh.Vert{X: max.X, Z: max.Z}
		v3       = mesh.Vert{X: min.X, Z: max.Z}
		it0, it1 = cdt.newTriangle(), cdt.newTriangle()
	)
	cdt.mNeighbor = make([][]TriIndex, 4+np)
	cdt.mVert = make([]mesh.Vert, 0, 4+np)
	cdt.mVert = append(cdt.mVert, v0, v1, v2, v3)
	cdt.mTri[it0] = Triangle{v0: 0, v1: 1, v2: 2, n0: -1, n1: -1, n2: it1}
	cdt.mTri[it1] = Triangle{v0: 0, v1: 2, v2: 3, n0: it0, n1: -1, n2: -1}
	cdt.insertVertNeighbor(0, it0, it1)
	cdt.insertVertNeighbor(2, it0, it1)
	cdt.locator = new(Locator)
	cdt.locator.Init(min, max, 100, 100)
}

func (cdt *CDT) insertVert(iv VertIndex) {
	var (
		it0, it1 = cdt.locatePoint(cdt.mVert[iv])
		affects  []TriIndex
	)
	if it1 == -1 {
		affects = cdt.insertVertInTriangle(iv, it0)
	} else {
		affects = cdt.insertVertOnEdge(iv, it0, it1)
	}
	for len(affects) > 0 {
		var (
			l   = len(affects) - 1
			it  = affects[l]
			ito = cdt.opposedTri(it, iv)
		)
		affects = affects[0:l]
		if ito == -1 {
			continue
		}
		if cdt.needFlip(iv, ito) {
			cdt.flip(it, ito)
			affects = append(affects, it, ito)
		}
	}
}

/* Flip edge between T0 and T1:
 *
 *                v3         | - old edge
 *               /|\         ~ - new edge
 *              / | \
 *          n2 /  T0 \ n3
 *            /   |   \
 *           /    |    \
 *    T0`-> v0~~~~~~~~~v2 <- T1`
 *           \    |    /
 *            \   |   /
 *          n0 \  T1 / n1
 *              \ | /
 *               \|/
 *                v1
 */
func (cdt *CDT) flip(it0, it1 TriIndex) {
	var (
		iv1 = cdt.opposedVert(it1, it0)
		iv3 = cdt.opposedVert(it0, it1)
		iv0 = cdt.triNextVert(it0, iv3)
		iv2 = cdt.triNextVert(it1, iv1)
		n2  = cdt.opposedTri(it0, iv2)
		n3  = cdt.opposedTri(it0, iv0)
		n0  = cdt.opposedTri(it1, iv2)
		n1  = cdt.opposedTri(it1, iv0)
	)
	// update triangle t0 t1
	cdt.locator.Remove(it0, cdt.mVert[iv0], cdt.mVert[iv2], cdt.mVert[iv3])
	cdt.mTri[it0] = Triangle{v0: iv0, v1: iv1, v2: iv3, n0: n0, n1: it1, n2: n2}
	cdt.locator.Insert(it0, cdt.mVert[iv0], cdt.mVert[iv1], cdt.mVert[iv3])

	cdt.locator.Remove(it1, cdt.mVert[iv0], cdt.mVert[iv1], cdt.mVert[iv2])
	cdt.mTri[it1] = Triangle{v0: iv1, v1: iv2, v2: iv3, n0: n1, n1: n3, n2: it0}
	cdt.locator.Insert(it1, cdt.mVert[iv1], cdt.mVert[iv2], cdt.mVert[iv3])

	// update neighbor
	cdt.changeTriNeighbor(n0, it1, it0)
	cdt.changeTriNeighbor(n3, it0, it1)
	cdt.removeVertNeighbor(iv0, it1)
	cdt.removeVertNeighbor(iv2, it0)
	cdt.insertVertNeighbor(iv1, it0)
	cdt.insertVertNeighbor(iv3, it1)
}

func (cdt *CDT) needFlip(iv VertIndex, it TriIndex) bool {
	var (
		v          = cdt.mVert[iv]
		v0, v1, v2 = cdt.mVert[cdt.mTri[it].v0], cdt.mVert[cdt.mTri[it].v1], cdt.mVert[cdt.mTri[it].v2]
		a, b, c    = mesh.VDist(v0, v1), mesh.VDist(v1, v2), mesh.VDist(v2, v0)
		mid        = (a + b + c) * 0.5
		r          = (a * b * c * 0.25) / math.Sqrt(mid*(mid-a)*(mid-b)*(mid-c))
		ma, mc     = mesh.VInter(v0, v1, 0.5), mesh.VInter(v0, v2, 0.5)
		va, vc     = mesh.VSub(v1, v0), mesh.VSub(v2, v0)
		ta, tc     = mesh.Vert{X: va.Z, Y: va.Y, Z: -va.X}, mesh.Vert{X: vc.Z, Y: vc.Y, Z: -vc.X}
		na, nc     = mesh.VAdd(va, ta), mesh.VAdd(vc, tc)
		s, _, _    = mesh.IntersectSegSeg2D(ma, na, mc, nc)
		center     = mesh.VInter(ma, na, s)
	)
	return mesh.VDist(center, v) <= r-mesh.Eps
}

/* Insert point into triangle: split into 3 triangles:
*                      v2
*                    / | \
*                   /  |  \ <-- original triangle (t)
*                  /   |   \
*              n2 /    |    \ n1
*                /newT2|newT1\
*               /      v      \
*              /    __/ \__    \
*             /  __/       \__  \
*            / _/      t0     \_ \
*          v0 ___________________ v1
*                     n0
 */
func (cdt *CDT) insertVertInTriangle(iv VertIndex, it TriIndex) []TriIndex {
	var (
		it1        = cdt.newTriangle()
		it2        = cdt.newTriangle()
		t          = &cdt.mTri[it]
		v0, v1, v2 = t.v0, t.v1, t.v2
		n0, n1, n2 = t.n0, t.n1, t.n2
	)
	// make new triangle t1 t2
	cdt.mTri[it1] = Triangle{v0: v1, v1: v2, v2: iv, n0: n1, n1: it2, n2: it}
	cdt.locator.Insert(it1, cdt.mVert[v1], cdt.mVert[v2], cdt.mVert[iv])
	cdt.mTri[it2] = Triangle{v0: v2, v1: v0, v2: iv, n0: n2, n1: it, n2: it1}
	cdt.locator.Insert(it2, cdt.mVert[v2], cdt.mVert[v0], cdt.mVert[iv])

	// update triangle t
	cdt.locator.Remove(it, cdt.mVert[v0], cdt.mVert[v1], cdt.mVert[v2])
	cdt.mTri[it] = Triangle{v0: v0, v1: v1, v2: iv, n0: n0, n1: it1, n2: it2}
	cdt.locator.Insert(it, cdt.mVert[v0], cdt.mVert[v1], cdt.mVert[iv])

	// update vertex's neighbor triangle
	cdt.insertVertNeighbor(iv, it, it1, it2)
	cdt.insertVertNeighbor(v0, it2)
	cdt.insertVertNeighbor(v1, it1)
	cdt.removeVertNeighbor(v2, it)
	cdt.insertVertNeighbor(v2, it1, it2)

	// change triangle neighbor
	cdt.changeTriNeighbor(n1, it, it1)
	cdt.changeTriNeighbor(n2, it, it2)
	return []TriIndex{it, it1, it2}
}

/* Inserting a point on the edge between two triangles
 *    T1 (top)        v0
 *                   /|\
 *              n0 /  |  \ n3
 *               /    |    \
 *             /  T0' | T2  \
 *           v1-------v-------v3
 *             \  T3  | T1'  /
 *               \    |    /
 *              n1 \  |  / n2
 *                   \|/
 *   T2 (bottom)      v2
 */
func (cdt *CDT) insertVertOnEdge(iv VertIndex, it0, it1 TriIndex) []TriIndex {
	var (
		it2 = cdt.newTriangle()
		it3 = cdt.newTriangle()
		v0  = cdt.opposedVert(it0, it1)
		v1  = cdt.triNextVert(it0, v0)
		v2  = cdt.opposedVert(it1, it0)
		v3  = cdt.triNextVert(it1, v2)
		n0  = cdt.opposedTri(it0, v3)
		n1  = cdt.opposedTri(it1, v3)
		n2  = cdt.opposedTri(it1, v1)
		n3  = cdt.opposedTri(it0, v1)
	)
	// make new triangle t2 t3
	cdt.mTri[it2] = Triangle{v0: v0, v1: iv, v2: v3, n0: it0, n1: it1, n2: n3}
	cdt.locator.Insert(it2, cdt.mVert[v0], cdt.mVert[iv], cdt.mVert[v3])
	cdt.mTri[it3] = Triangle{v0: v2, v1: iv, v2: v1, n0: it1, n1: it0, n2: n1}
	cdt.locator.Insert(it3, cdt.mVert[v2], cdt.mVert[iv], cdt.mVert[v1])

	// update triangle t0, t1
	cdt.locator.Remove(it0, cdt.mVert[v0], cdt.mVert[v1], cdt.mVert[v3])
	cdt.mTri[it0] = Triangle{v0: v0, v1: v1, v2: iv, n0: n0, n1: it3, n2: it2}
	cdt.locator.Insert(it0, cdt.mVert[v0], cdt.mVert[v1], cdt.mVert[iv])

	cdt.locator.Remove(it1, cdt.mVert[v1], cdt.mVert[v2], cdt.mVert[v3])
	cdt.mTri[it1] = Triangle{v0: v2, v1: v3, v2: iv, n0: n2, n1: it2, n2: it3}
	cdt.locator.Insert(it1, cdt.mVert[v2], cdt.mVert[v3], cdt.mVert[iv])

	// update vertex's neighbor triangle
	cdt.insertVertNeighbor(iv, it0, it1, it2, it3)
	cdt.insertVertNeighbor(v0, it2)
	cdt.insertVertNeighbor(v2, it3)
	cdt.removeVertNeighbor(v1, it1)
	cdt.insertVertNeighbor(v1, it3)
	cdt.removeVertNeighbor(v3, it0)
	cdt.insertVertNeighbor(v3, it2)

	// change triangle neighbor
	cdt.changeTriNeighbor(n1, it1, it3)
	cdt.changeTriNeighbor(n3, it0, it2)

	return []TriIndex{it0, it1, it2, it3}
}

func (cdt *CDT) newTriangle() TriIndex {
	cdt.mTri = append(cdt.mTri, Triangle{})
	return TriIndex(len(cdt.mTri) - 1)
}
func (cdt *CDT) insertVertNeighbor(iv VertIndex, its ...TriIndex) {
	cdt.mNeighbor[iv] = append(cdt.mNeighbor[iv], its...)
}
func (cdt *CDT) removeVertNeighbor(iv VertIndex, its ...TriIndex) {
	for _, it := range its {
		cdt.mNeighbor[iv] = common.EraseOnce(cdt.mNeighbor[iv], it)
	}
}
func (cdt *CDT) changeTriNeighbor(it TriIndex, from, to TriIndex) {
	var t = &cdt.mTri[it]
	switch from {
	case t.n0:
		t.n0 = to
	case t.n1:
		t.n1 = to
	case t.n2:
		t.n2 = to
	}
}
func (cdt *CDT) opposedVert(it0, it1 TriIndex) VertIndex {
	var t = &cdt.mTri[it0]
	switch it1 {
	case t.n0:
		return t.v2
	case t.n1:
		return t.v0
	case t.n2:
		return t.v1
	}
	return -1
}
func (cdt *CDT) opposedTri(it TriIndex, iv VertIndex) TriIndex {
	var t = &cdt.mTri[it]
	switch iv {
	case t.v0:
		return t.n1
	case t.v1:
		return t.n2
	case t.v2:
		return t.n0
	}
	return -1
}
func (cdt *CDT) triNextVert(it TriIndex, iv VertIndex) VertIndex {
	var t = &cdt.mTri[it]
	switch iv {
	case t.v0:
		return t.v1
	case t.v1:
		return t.v2
	case t.v2:
		return t.v0
	}
	return -1
}
func (cdt *CDT) locatePoint(pos mesh.Vert) (it0, it1 TriIndex) {
	var loc int32
	it0, loc = cdt.locator.Locate(pos, func(i TriIndex) (v0, v1, v2 mesh.Vert) {
		var tri = cdt.mTri[i]
		return cdt.mVert[tri.v0], cdt.mVert[tri.v1], cdt.mVert[tri.v2]
	})
	switch loc {
	case LocationOutside:
		panic("locate point fail")
	case LocationInside:
		return it0, -1
	case LocationEdge0:
		return it0, cdt.mTri[it0].n0
	case LocationEdge1:
		return it0, cdt.mTri[it0].n1
	case LocationEdge2:
		return it0, cdt.mTri[it0].n2
	}
	panic("unreachable code")
}
