package main

import (
	"fmt"
	"lib_chaos/mesh"
	"lib_chaos/mesh/tra"
	"math/rand"
	"time"
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
	rand.Seed(time.Now().UnixNano())
	//var (
	//	it0        = rand.Intn(len(t.MTri))
	//	it1        = rand.Intn(len(t.MTri))
	//	sa, sb, sc = t.MTri[it0].Vs()
	//	ta, tb, tc = t.MTri[it1].Vs()
	//	v0         = mesh.TriRandomPoint(t.MVert[sa], t.MVert[sb], t.MVert[sc])
	//	v1         = mesh.TriRandomPoint(t.MVert[ta], t.MVert[tb], t.MVert[tc])
	//	//it0 = 16632
	//	//it1 = 17868
	//	//v0  = mesh.Vert{X: 4446.748171, Z: 8255.833410}
	//	//v1  = mesh.Vert{X: 5608.595476, Z: 9318.090223}
	//	//v0  = mesh.Vert{X: 5096.496378800867, Z: 9434.413936456034}
	//	//v1  = mesh.Vert{X: 5116.997891797, Z: 9597.49099981785}
	//	//it0 = nav.FindMesh(v0)
	//	//it1 = nav.FindMesh(v1)
	//	q = tra.NewQuery(nav, 32768)
	//)
	////fmt.Printf("from %d to %d\n", it0, it1)
	//q.Clear()
	//q.FindPath(int32(it0), int32(it1), v0, v1)
	////s := time.Now()
	////fmt.Println()
	////fmt.Println(time.Since(s))
	//var buf = nav.Svg(t.Max, v0, v1, q.Path)
	////var buf = nav.Svg(t.Max, v0, v1, nil)
	//fmt.Print(buf.String())

	var (
		total time.Duration
		max   time.Duration
		cnt   int
		q     = tra.NewQuery(nav, 32768)
	)
	for i := 0; i < 1000; i++ {
		var (
			it0        = rand.Intn(len(t.MTri))
			it1        = rand.Intn(len(t.MTri))
			sa, sb, sc = t.MTri[it0].Vs()
			ta, tb, tc = t.MTri[it1].Vs()
			v0         = mesh.TriRandomPoint(t.MVert[sa], t.MVert[sb], t.MVert[sc])
			v1         = mesh.TriRandomPoint(t.MVert[ta], t.MVert[tb], t.MVert[tc])
		)
		s := time.Now()
		q.Clear()
		//fmt.Printf("%d %d (%f %f) (%f %f)\n", it0, it1, v0.X, v0.Z, v1.X, v1.Z)
		if ok, o := q.FindPath(int32(it0), int32(it1), v0, v1); !ok {
			fmt.Println("fail", o)
		} else {
			used := time.Since(s)
			total += used
			cnt++
			if used > max {
				max = used
			}
			//fmt.Println(cnt)
		}
	}
	fmt.Println(cnt, total/time.Duration(cnt), max)
}
