package cdt

import (
	"lib_chaos/mesh"
)

const (
	LocationOutside = iota
	LocationInside
	LocationEdge0
	LocationEdge1
	LocationEdge2
)

type Grid struct {
	min, max mesh.Vert
	tri      map[TriIndex]struct{}
}

type Locator struct {
	min, max mesh.Vert // bound
	nx, nz   int32
	sx, sz   float64
	grid     [][]Grid
}

func (l *Locator) Init(min, max mesh.Vert, nx, nz int32) {
	*l = Locator{
		min: min,
		max: max,
		nx:  nx,
		nz:  nz,
		sx:  (max.X - min.X) / float64(nx),
		sz:  (max.Z - min.Z) / float64(nz),
	}
	var grid = make([][]Grid, nx)
	for x := range grid {
		grid[x] = make([]Grid, nz)
	}
}

func (l *Locator) pos(v mesh.Vert) (x, z int32) {
	x = int32((v.X - l.min.X) / l.sx)
	z = int32((v.Z - l.min.Z) / l.sz)
	return
}

func (l *Locator) Insert(id TriIndex, v0, v1, v2 mesh.Vert) {
	var (
		min, max   = mesh.TriBox2D(v0, v1, v2)
		minX, minZ = l.pos(min)
		maxX, maxZ = l.pos(max)
	)
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			var g = l.grid[x][z]
			if collideTest(g.min, g.max, v0, v1, v2) {
				g.tri[id] = struct{}{}
			}
		}
	}
}

func (l *Locator) Remove(id TriIndex, v0, v1, v2 mesh.Vert) {
	var (
		min, max   = mesh.TriBox2D(v0, v1, v2)
		minX, minZ = l.pos(min)
		maxX, maxZ = l.pos(max)
	)
	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			var g = l.grid[x][z]
			delete(g.tri, id)
		}
	}
}

func (l *Locator) Locate(v mesh.Vert, get func(index TriIndex) (v0, v1, v2 mesh.Vert)) (i TriIndex, loc int32) {
	var (
		x, z = l.pos(v)
		g    = l.grid[x][z]
	)
	for i = range g.tri {
		var v0, v1, v2 = get(i)
		if loc = locateTriangle(v, v0, v1, v2); loc != LocationOutside {
			return i, loc
		}
	}
	return -1, LocationOutside
}

// test collide between rectangle(min, max) of triangle(v0, v1, v2)
func collideTest(min, max, v0, v1, v2 mesh.Vert) bool {
	var (
		r0 = mesh.Vert{X: min.X, Z: min.Z}
		r1 = mesh.Vert{X: max.X, Z: min.Z}
		r2 = mesh.Vert{X: max.X, Z: max.Z}
		r3 = mesh.Vert{X: min.X, Z: max.Z}
	)
	return cross(r0, r1, v0, v1) || cross(r0, r1, v1, v2) || cross(r0, r1, v2, v0) ||
		cross(r1, r2, v0, v1) || cross(r1, r2, v1, v2) || cross(r1, r2, v2, v0) ||
		cross(r2, r3, v0, v1) || cross(r2, r3, v1, v2) || cross(r2, r3, v2, v0) ||
		cross(r3, r0, v0, v1) || cross(r3, r0, v1, v2) || cross(r3, r0, v2, v0)
}
func cross(a, b, p, q mesh.Vert) bool {
	var s, t, ok = mesh.IntersectSegSeg2D(a, b, p, q)
	return ok && s >= 0 && s <= 1 && t >= 0 && t <= 1
}

func locateTriangle(p, a, b, c mesh.Vert) int32 {
	var (
		ab = mesh.VSub(b, a)
		ac = mesh.VSub(c, a)
		ap = mesh.VSub(p, a)
		d  = ab.Z*ac.X - ab.X*ac.Z
	)
	if d*d < mesh.Eqs {
		return LocationOutside
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
			return LocationEdge0
		}
		if v < mesh.Eps {
			return LocationEdge2
		}
		if u+v > d-mesh.Eps {
			return LocationEdge1
		}
		return LocationInside
	}
	return LocationOutside
}
