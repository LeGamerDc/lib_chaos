package main

import (
	"fmt"
	"lib_chaos/mesh"
)

func locateTriangle(p, a, b, c mesh.Vert) int32 {
	var (
		ab = mesh.VSub(b, a)
		ac = mesh.VSub(c, a)
		ap = mesh.VSub(p, a)
		d  = ab.Z*ac.X - ab.X*ac.Z
	)
	if d*d < mesh.Eqs {
		return -1
	}
	var (
		u = ab.Z*ap.X - ab.X*ap.Z
		v = ap.Z*ac.X - ap.X*ac.Z
	)
	if d < 0 {
		d, u, v = -d, -u, -v
	}
	if u >= 0 && v >= 0 && (u+v) <= d {
		if u < mesh.Eps {
			return 0
		}
		if v < mesh.Eps {
			return 2
		}
		if u+v > d-mesh.Eps {
			return 1
		}
		return -2
	}
	return -1
}

func main() {
	var (
		a = mesh.Vert{}
		b = mesh.Vert{X: 1}
		c = mesh.Vert{Z: 1}
		o = mesh.Vert{X: 0.8, Z: 0.5}
	)
	fmt.Println(locateTriangle(o, a, b, c))
}
