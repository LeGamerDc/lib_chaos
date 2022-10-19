package detour

import "lib_chaos/mesh"

const maxVertPerPoly = 6

type Poly struct {
	Link    int32 // first Link
	VertCnt int32
	AreaId  int32
	Vs      [maxVertPerPoly]int32
}

type Link struct {
	ToRef int32 // connect which poly [must exist]
	Next  int32 // Next Link [-1 if no Next]
	Edge  int32 // which Edge this Link belongs, 0->(0,1) 1->(1,2)... n-1->(n-1,0)
}

type NavMesh struct {
	MVert []mesh.Vert
	MPoly []Poly
	MLink []Link
}

func (n *NavMesh) InsertEdge(i, j, e int32) {
	var l = len(n.MLink)
	n.MLink = append(n.MLink, Link{
		ToRef: j,
		Next:  n.MPoly[i].Link,
		Edge:  e,
	})
	n.MPoly[i].Link = int32(l)
}

func (n *NavMesh) getPortal(fromRef, toRef int32) (left, right mesh.Vert, ok bool) {
	for l := n.MPoly[fromRef].Link; l != -1; l = n.MLink[l].Next {
		if n.MLink[l].ToRef == toRef {
			var poly = &n.MPoly[fromRef]
			var link = &n.MLink[l]
			var v0 = n.MVert[poly.Vs[link.Edge]]
			var v1 = n.MVert[poly.Vs[(link.Edge+1)%poly.VertCnt]]
			return v0, v1, true
		}
	}
	return mesh.NilVert, mesh.NilVert, false
}

func (n *NavMesh) edgeMidPoint(fromRef, toRef int32) (p mesh.Vert, ok bool) {
	var (
		v0, v1 mesh.Vert
	)
	v0, v1, ok = n.getPortal(fromRef, toRef)
	if !ok {
		return
	}
	return mesh.Vert{
		X: (v0.X + v1.X) * 0.5,
		Y: (v0.Y + v1.Y) * 0.5,
		Z: (v0.Z + v1.Z) * 0.5,
	}, true
}

func (n *NavMesh) GetPolyHeight(poly *Poly, p mesh.Vert) (height float64, ok bool) {
	var v0 = n.MVert[poly.Vs[0]]
	for i := int32(1); i < poly.VertCnt-1; i++ {
		var (
			v1 = n.MVert[poly.Vs[i]]
			v2 = n.MVert[poly.Vs[i+1]]
		)
		if height, ok = mesh.VHeightOnTriangle(p, v0, v1, v2); ok {
			return
		}
	}
	return 0, false
}

func (n *NavMesh) closestPointOnPoly(poly *Poly, p mesh.Vert) mesh.Vert {
	var (
		l, r mesh.Vert
		bd   = mesh.BigFloat
		bt   float64
	)
	for i := int32(0); i < poly.VertCnt; i++ {
		var (
			v0   = n.MVert[poly.Vs[i]]
			v1   = n.MVert[poly.Vs[(i+1)%poly.VertCnt]]
			t, d = mesh.DistPtSegSqr2D(p, v0, v1)
		)
		if d < bd {
			bd, bt = d, t
			l, r = v0, v1
		}
	}
	return mesh.VInter(l, r, bt)
}

func (n *NavMesh) LocatePoly(p mesh.Vert) (polyRef int32, pt mesh.Vert, onPoly bool) {
	var (
		bd = mesh.BigFloat
	)
	for i := 0; i < len(n.MPoly); i++ {
		if h, ok := n.GetPolyHeight(&n.MPoly[i], p); ok {
			return int32(i), mesh.Vert{
				X: p.X,
				Y: h,
				Z: p.Z,
			}, true
		}
		c := n.closestPointOnPoly(&n.MPoly[i], p)
		d := mesh.VSqr(mesh.VSub(p, c))
		if d < bd {
			polyRef = int32(i)
			pt = c
		}
	}
	return
}
