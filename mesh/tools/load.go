package main

import (
	"fmt"
	"lib_chaos/mesh"
	"lib_chaos/mesh/cdt"
	"math"
)

func read() *cdt.CDT {
	var (
		nv, ne int
		c      cdt.CDT
		ps     []mesh.Vert
		min    = mesh.Vert{X: 999999, Z: 999999}
		max    = mesh.Vert{X: -999999, Z: -999999}
	)
	_, _ = fmt.Scanf("%d %d", &nv, &ne)
	for i := 0; i < nv; i++ {
		var x, z float64
		_, e := fmt.Scan(&x, &z)
		if e != nil {
			panic(fmt.Sprintf("v %d %s", i, e.Error()))
		}
		//x *= 100
		//z *= 100
		ps = append(ps, mesh.Vert{X: x, Z: z})
		min.X = math.Min(min.X, x)
		min.Z = math.Min(min.Z, z)
		max.X = math.Max(max.X, x)
		max.Z = math.Max(max.Z, z)
	}
	min.X -= 200
	min.Z -= 200
	max.X = max.X + 200
	max.Z = max.Z + 200
	c.Init(min, max, nv)
	c.InsertVerts(ps)
	c.Report()
	for i := 0; i < ne; i++ {
		var a, b cdt.VertIndex
		_, e := fmt.Scan(&a, &b)
		if e != nil {
			panic(fmt.Sprintf("e %d %s", i, e.Error()))
		}
		a += 4
		b += 4
		c.InsertEdge(b, a)
		//c.Report()
		//fmt.Printf("[e %d]\n", i)
	}
	c.Culling()
	c.Shrink()
	return &c
}
