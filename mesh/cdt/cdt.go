package cdt

import (
	"fmt"
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

func (t *Triangle) Vs() (v0, v1, v2 int32) {
	v0, v1, v2 = int32(t.v0), int32(t.v1), int32(t.v2)
	return
}
func (t *Triangle) Ns() (n0, n1, n2 int32) {
	n0, n1, n2 = int32(t.n0), int32(t.n1), int32(t.n2)
	return
}

type Edge struct {
	v0, v1 VertIndex
}

type CDT struct {
	MVert []mesh.Vert
	MTri  []Triangle
	MCons map[Edge]struct{} // obstacle edge

	// cache
	locator   *Locator
	mNeighbor [][]TriIndex // adjacent triangles to vert
	mDummy    []TriIndex
	Max       mesh.Vert
}

func (cdt *CDT) Init(min, max mesh.Vert, np int) {
	var (
		v0       = mesh.Vert{X: min.X, Z: min.Z}
		v1       = mesh.Vert{X: max.X, Z: min.Z}
		v2       = mesh.Vert{X: max.X, Z: max.Z}
		v3       = mesh.Vert{X: min.X, Z: max.Z}
		it0, it1 = cdt.newTriangle(), cdt.newTriangle()
	)
	cdt.Max = max
	cdt.mNeighbor = make([][]TriIndex, 4+np)
	cdt.MVert = make([]mesh.Vert, 0, 4+np)
	cdt.MVert = append(cdt.MVert, v0, v1, v2, v3)
	cdt.MTri[it0] = Triangle{v0: 0, v1: 1, v2: 2, n0: -1, n1: -1, n2: it1}
	cdt.MTri[it1] = Triangle{v0: 0, v1: 2, v2: 3, n0: it0, n1: -1, n2: -1}
	cdt.insertVertNeighbor(0, it0, it1)
	cdt.insertVertNeighbor(2, it0, it1)
	cdt.locator = new(Locator)
	cdt.MCons = make(map[Edge]struct{})
	cdt.MCons[Edge{v0: 1, v1: 0}] = struct{}{}
	cdt.MCons[Edge{v0: 2, v1: 1}] = struct{}{}
	cdt.MCons[Edge{v0: 3, v1: 2}] = struct{}{}
	cdt.MCons[Edge{v0: 0, v1: 3}] = struct{}{}
	min.X -= 2
	min.Z -= 2
	max.X += 2
	max.Z += 2
	var nx = common.Max[int32](int32(4), int32(math.Sqrt(float64(np)))/4)
	cdt.locator.Init(min, max, nx, nx)
	cdt.locator.Insert(it0, v0, v1, v2)
	cdt.locator.Insert(it1, v0, v2, v3)
}

func (cdt *CDT) DuplicateVert(v0 mesh.Vert) VertIndex {
	for i, v1 := range cdt.MVert {
		if mesh.VEqual(v0, v1) {
			return VertIndex(i)
		}
	}
	return -1
}

func (cdt *CDT) InsertVert(v mesh.Vert) VertIndex {
	var idx = cdt.DuplicateVert(v)
	if idx != -1 {
		return idx
	}
	idx = VertIndex(len(cdt.MVert))
	cdt.MVert = append(cdt.MVert, v)
	cdt.mNeighbor = append(cdt.mNeighbor, nil)
	cdt.insertVert(idx)
	return idx
}

func (cdt *CDT) InsertVerts(vs []mesh.Vert) {
	var s = len(cdt.MVert)
	for _, v := range vs {
		cdt.MVert = append(cdt.MVert, v)
	}
	for i := s; i < len(cdt.MVert); i++ {
		cdt.insertVert(VertIndex(i))
	}
}

func (cdt *CDT) insertVert(iv VertIndex) {
	var (
		it0, it1 = cdt.locatePoint(cdt.MVert[iv])
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
		affects = affects[:l]
		if ito == -1 {
			continue
		}
		if cdt.needFlip(iv, ito) {
			cdt.flip(it, ito)
			affects = append(affects, it, ito)
		}
	}
}

func (cdt *CDT) Report() {
	var find = func(it TriIndex, iv VertIndex) {
		for _, n := range cdt.mNeighbor[iv] {
			if n == it {
				return
			}
		}
		panic("wrong neighbor")
	}
	for it, t := range cdt.MTri {
		if t.v0 < 4 || t.v1 < 4 || t.v2 < 4 {
			continue
		}
		if t.v0 == t.v1 || t.v0 == t.v2 || t.v1 == t.v2 {
			panic(fmt.Sprintf("wrong tri %d[%d %d %d]", it, t.v0, t.v1, t.v2))
		}
		if t.n0 == -1 || t.n1 == -1 || t.n2 == -1 {
			panic(fmt.Sprintf("wrong tri %d(%d %d %d)", it, t.n0, t.n1, t.n2))
		}
		find(TriIndex(it), t.v0)
		find(TriIndex(it), t.v1)
		find(TriIndex(it), t.v2)
	}
}

func (cdt *CDT) report(at string, its ...TriIndex) {
	for i, it := range its {
		if it == -1 {
			continue
		}
		var t = &cdt.MTri[it]
		if t.v0 < 4 || t.v1 < 4 || t.v2 < 4 {
			continue
		}
		if t.n0 == -1 || t.n1 == -1 || t.n2 == -1 {
			panic(fmt.Sprintf("wrong tri %s: (%d %d) (%d %d %d) [%d %d %d]", at, i, it, t.n0, t.n1, t.n2, t.v0, t.v1, t.v2))
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
		iv0 = cdt.opposedVert(it0, it1)
		iv2 = cdt.opposedVert(it1, it0)
		iv1 = cdt.triNextVert(it0, iv0)
		iv3 = cdt.triNextVert(it1, iv2)
		n0  = cdt.opposedTri(it0, iv3)
		n2  = cdt.opposedTri(it0, iv1)
		n1  = cdt.opposedTri(it1, iv3)
		n3  = cdt.opposedTri(it1, iv1)
	)
	// update triangle t0 t1
	cdt.locator.Remove(it0, cdt.MVert[iv0], cdt.MVert[iv1], cdt.MVert[iv3])
	cdt.MTri[it0] = Triangle{v0: iv0, v1: iv2, v2: iv3, n0: it1, n1: n3, n2: n2}
	cdt.locator.Insert(it0, cdt.MVert[iv0], cdt.MVert[iv2], cdt.MVert[iv3])

	cdt.locator.Remove(it1, cdt.MVert[iv1], cdt.MVert[iv2], cdt.MVert[iv3])
	cdt.MTri[it1] = Triangle{v0: iv0, v1: iv1, v2: iv2, n0: n0, n1: n1, n2: it0}
	cdt.locator.Insert(it1, cdt.MVert[iv0], cdt.MVert[iv1], cdt.MVert[iv2])

	// update tri neighbor
	cdt.changeTriNeighbor(n0, it0, it1)
	cdt.changeTriNeighbor(n3, it1, it0)

	// update vert neighbor
	cdt.removeVertNeighbor(iv1, it0)
	cdt.removeVertNeighbor(iv3, it1)
	cdt.insertVertNeighbor(iv0, it1)
	cdt.insertVertNeighbor(iv2, it0)

	cdt.report("insertVertInTriangle", it0, it1, n0, n1, n2, n3)
}

func (cdt *CDT) needFlip(iv VertIndex, it TriIndex) bool {
	var (
		v          = cdt.MVert[iv]
		v0, v1, v2 = cdt.MVert[cdt.MTri[it].v0], cdt.MVert[cdt.MTri[it].v1], cdt.MVert[cdt.MTri[it].v2]
	)
	return mesh.InTriOuterCircle(v, v0, v1, v2)
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
		t          = &cdt.MTri[it]
		v0, v1, v2 = t.v0, t.v1, t.v2
		n0, n1, n2 = t.n0, t.n1, t.n2
	)
	// make new triangle t1 t2
	cdt.MTri[it1] = Triangle{v0: v1, v1: v2, v2: iv, n0: n1, n1: it2, n2: it}
	cdt.locator.Insert(it1, cdt.MVert[v1], cdt.MVert[v2], cdt.MVert[iv])
	cdt.MTri[it2] = Triangle{v0: v2, v1: v0, v2: iv, n0: n2, n1: it, n2: it1}
	cdt.locator.Insert(it2, cdt.MVert[v2], cdt.MVert[v0], cdt.MVert[iv])

	// update triangle t
	cdt.locator.Remove(it, cdt.MVert[v0], cdt.MVert[v1], cdt.MVert[v2])
	cdt.MTri[it] = Triangle{v0: v0, v1: v1, v2: iv, n0: n0, n1: it1, n2: it2}
	cdt.locator.Insert(it, cdt.MVert[v0], cdt.MVert[v1], cdt.MVert[iv])

	// update vertex's neighbor triangle
	cdt.insertVertNeighbor(iv, it, it1, it2)
	cdt.insertVertNeighbor(v0, it2)
	cdt.insertVertNeighbor(v1, it1)
	cdt.removeVertNeighbor(v2, it)
	cdt.insertVertNeighbor(v2, it1, it2)

	// change triangle neighbor
	cdt.changeTriNeighbor(n1, it, it1)
	cdt.changeTriNeighbor(n2, it, it2)
	cdt.report("insertVertInTriangle", it, it1, it2, n0, n1, n2)
	return []TriIndex{it, it1, it2}
}

/* Inserting a point on the edge between two triangles
 *    T0 (top)        v0
 *                   /|\
 *              n0 /  |  \ n3
 *               /    |    \
 *             /  T0' | T2  \
 *           v1-------v-------v3
 *             \  T3  | T1'  /
 *               \    |    /
 *              n1 \  |  / n2
 *                   \|/
 *   T1 (bottom)      v2
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
	cdt.MTri[it2] = Triangle{v0: v0, v1: iv, v2: v3, n0: it0, n1: it1, n2: n3}
	cdt.locator.Insert(it2, cdt.MVert[v0], cdt.MVert[iv], cdt.MVert[v3])
	cdt.MTri[it3] = Triangle{v0: v2, v1: iv, v2: v1, n0: it1, n1: it0, n2: n1}
	cdt.locator.Insert(it3, cdt.MVert[v2], cdt.MVert[iv], cdt.MVert[v1])

	// update triangle t0, t1
	cdt.locator.Remove(it0, cdt.MVert[v0], cdt.MVert[v1], cdt.MVert[v3])
	cdt.MTri[it0] = Triangle{v0: v0, v1: v1, v2: iv, n0: n0, n1: it3, n2: it2}
	cdt.locator.Insert(it0, cdt.MVert[v0], cdt.MVert[v1], cdt.MVert[iv])

	cdt.locator.Remove(it1, cdt.MVert[v1], cdt.MVert[v2], cdt.MVert[v3])
	cdt.MTri[it1] = Triangle{v0: v2, v1: v3, v2: iv, n0: n2, n1: it2, n2: it3}
	cdt.locator.Insert(it1, cdt.MVert[v2], cdt.MVert[v3], cdt.MVert[iv])

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

	cdt.report("insertVertInTriangle", it0, it1, it2, it3, n0, n1, n2, n3)
	return []TriIndex{it0, it1, it2, it3}
}

func (cdt *CDT) newTriangle() TriIndex {
	cdt.MTri = append(cdt.MTri, dummyTriangle)
	return TriIndex(len(cdt.MTri) - 1)
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
	if it == -1 {
		return
	}
	var t = &cdt.MTri[it]
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
	var t = &cdt.MTri[it0]
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
	var t = &cdt.MTri[it]
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
	var t = &cdt.MTri[it]
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
		var tri = cdt.MTri[i]
		return cdt.MVert[tri.v0], cdt.MVert[tri.v1], cdt.MVert[tri.v2]
	})
	switch loc {
	case LocationOutside:
		panic(fmt.Sprintf("locate point fail %f %f", pos.X, pos.Z))
	case LocationInside:
		return it0, -1
	case LocationEdge0:
		return it0, cdt.MTri[it0].n0
	case LocationEdge1:
		return it0, cdt.MTri[it0].n1
	case LocationEdge2:
		return it0, cdt.MTri[it0].n2
	}
	panic("unreachable code")
}
