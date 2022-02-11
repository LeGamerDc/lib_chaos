package cdt

import (
	"fmt"
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
	var (
		sx = (max.X - min.X) / float64(nx)
		sz = (max.Z - min.Z) / float64(nz)
	)
	*l = Locator{
		min: min,
		max: max,
		nx:  nx,
		nz:  nz,
		sx:  sx,
		sz:  sz,
	}
	var grid = make([][]Grid, nx)
	for x := range grid {
		grid[x] = make([]Grid, nz)
		for z := 0; z < int(nz); z++ {
			grid[x][z] = Grid{
				min: mesh.Vert{X: min.X + float64(x)*sx, Z: min.Z + float64(z)*sz},
				max: mesh.Vert{X: min.X + float64(x+1)*sx, Z: min.Z + float64(z+1)*sz},
				tri: make(map[TriIndex]struct{}),
			}
		}
	}
	l.grid = grid
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
	)
	if x < 0 || int(x) >= len(l.grid) || z < 0 || int(z) >= len(l.grid[0]) {
		panic(fmt.Sprintf("locate fail %f %f", v.X, v.Z))
	}
	var g = l.grid[x][z]
	for i = range g.tri {
		var v0, v1, v2 = get(i)
		if loc = locateTriangle(v, v0, v1, v2); loc != LocationOutside {
			return i, loc
		}
	}
	return -1, LocationOutside
}

// test collide between rectangle(min, Max) of triangle(v0, v1, v2)
func collideTest(min, max, v0, v1, v2 mesh.Vert) bool {
	var (
		r0      = mesh.Vert{X: min.X, Z: min.Z}
		r1      = mesh.Vert{X: max.X, Z: min.Z}
		r2      = mesh.Vert{X: max.X, Z: max.Z}
		r3      = mesh.Vert{X: min.X, Z: max.Z}
		collide = cross(r0, r1, v0, v1) || cross(r0, r1, v1, v2) || cross(r0, r1, v2, v0) ||
			cross(r1, r2, v0, v1) || cross(r1, r2, v1, v2) || cross(r1, r2, v2, v0) ||
			cross(r2, r3, v0, v1) || cross(r2, r3, v1, v2) || cross(r2, r3, v2, v0) ||
			cross(r3, r0, v0, v1) || cross(r3, r0, v1, v2) || cross(r3, r0, v2, v0)
		contain = rectContain(min, max, v0) || rectContain(min, max, v1) || rectContain(min, max, v2) ||
			triContain(v0, v1, v2, r0) || triContain(v0, v1, v2, r1) || triContain(v0, v1, v2, r2) ||
			triContain(v0, v1, v2, r3)
	)
	return collide || contain
}
func cross(a, b, p, q mesh.Vert) bool {
	var s, t, ok = mesh.IntersectSegSeg2D(a, b, p, q)
	return ok && s >= 0 && s <= 1 && t >= 0 && t <= 1
}

func rectContain(min, max, v mesh.Vert) bool {
	return v.X >= min.X && v.X <= max.X && v.Z >= min.Z && v.Z <= max.Z
}

func triContain(v0, v1, v2, v mesh.Vert) bool {

	var (
		d0  = mesh.VCrossXz(mesh.VSub(v1, v0), mesh.VSub(v, v0))
		d1  = mesh.VCrossXz(mesh.VSub(v2, v1), mesh.VSub(v, v1))
		d2  = mesh.VCrossXz(mesh.VSub(v0, v2), mesh.VSub(v, v2))
		pos = d0 > 0 || d1 > 0 || d2 > 0
		neg = d0 < 0 || d1 < 0 || d2 < 0
	)
	return !(pos && neg)
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
