package tra

import "lib_chaos/mesh"

type Tri struct {
	Ref     int32 // self Tri index
	Link    int32 // first Link index
	GroupId int32
	Vs      [3]int32 // vertex of triangle, in clock-wise
}

type Link struct {
	ToRef int32 // connect which tri
	Next  int32 // next Link index [-1 if last]
	Edge  int32 // which Edge this Link belongs, 0->(0,1) 1->(1,2) 2->(2,0)
}

type NavMesh struct {
	MVert []mesh.Vert
	MTri  []Tri
	MLink []Link
	cache *Tr
}

func (nav *NavMesh) InsertEdge(t0, t1 int32, e int32) {
	var l = len(nav.MLink)
	nav.MLink = append(nav.MLink, Link{
		ToRef: t1,
		Next:  nav.MTri[t0].Link,
		Edge:  e,
	})
	nav.MTri[t0].Link = int32(l)
}

func (nav *NavMesh) Ns(it int32) (ns []int32) {
	var l = nav.MTri[it].Link
	for l != -1 {
		ns = append(ns, nav.MLink[l].ToRef)
		l = nav.MLink[l].Next
	}
	return
}

func (nav *NavMesh) countEdge(i int32) (n int) {
	var l = nav.MTri[i].Link
	for ; l != -1; l = nav.MLink[l].Next {
		n++
	}
	return n
}

func (nav *NavMesh) getPortal(fromRef, toRef int32) (left, right mesh.Vert, ok bool) {
	for l := nav.MTri[fromRef].Link; l != -1; l = nav.MLink[l].Next {
		if nav.MLink[l].ToRef == toRef {
			var tri = &nav.MTri[fromRef]
			switch nav.MLink[l].Edge {
			case 0:
				left, right = nav.MVert[tri.Vs[0]], nav.MVert[tri.Vs[1]]
			case 1:
				left, right = nav.MVert[tri.Vs[1]], nav.MVert[tri.Vs[2]]
			case 2:
				left, right = nav.MVert[tri.Vs[2]], nav.MVert[tri.Vs[0]]
			}
			return left, right, true
		}
	}
	return mesh.NilVert, mesh.NilVert, false
}

func (nav *NavMesh) FindMesh(v mesh.Vert) int32 {
	for id, t := range nav.MTri {
		var (
			v0 = nav.MVert[t.Vs[0]]
			v1 = nav.MVert[t.Vs[1]]
			v2 = nav.MVert[t.Vs[2]]
		)
		if _, ok := mesh.VHeightOnTriangle(v, v0, v1, v2); ok {
			return int32(id)
		}
	}
	return -1
}
