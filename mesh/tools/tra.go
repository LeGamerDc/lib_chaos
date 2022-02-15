package main

import (
	"fmt"
	"lib_chaos/mesh"
	"lib_chaos/mesh/tra"
	"math/rand"
)

func main() {
	var t = read()
	var nav = new(tra.NavMesh)
	nav.MVert = make([]mesh.Vert, len(t.MVert))
	for i, v := range t.MVert {
		nav.MVert[i] = v
	}
	nav.MTri = make([]tra.Tri, len(t.MTri))
	for i, tr := range t.MTri {
		var v0, v1, v2 = tr.Vs()
		nav.MTri[i] = tra.Tri{
			Ref:     int32(i),
			Link:    -1,
			GroupId: 0,
			Vs:      [3]int32{v0, v1, v2},
		}
	}
	for i, tr := range t.MTri {
		var n0, n1, n2 = tr.Ns()
		if n0 != -1 {
			nav.InsertEdge(int32(i), n0, 0)
		}
		if n1 != -1 {
			nav.InsertEdge(int32(i), n1, 1)
		}
		if n2 != -1 {
			nav.InsertEdge(int32(i), n2, 2)
		}
	}
	var (
		it0        = rand.Intn(len(t.MTri))
		it1        = rand.Intn(len(t.MTri))
		sa, sb, sc = t.MTri[it0].Vs()
		ta, tb, tc = t.MTri[it1].Vs()
		v0         = mesh.TriRandomPoint(t.MVert[sa], t.MVert[sb], t.MVert[sc])
		v1         = mesh.TriRandomPoint(t.MVert[ta], t.MVert[tb], t.MVert[tc])
		q          = tra.NewQuery(nav, 200000)
	)
	fmt.Printf("from %d to %d\n", it0, it1)
	fmt.Println(q.FindPath(int32(it0), int32(it1), v0, v1))
	var buf = nav.Svg(t.Max, v0, v1, q.Path)
	//var buf = nav.Svg(t.Max, v0, v1, nil)
	fmt.Print(buf.String())
}
