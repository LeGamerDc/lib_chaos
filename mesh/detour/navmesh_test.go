package detour

import (
	"fmt"
	"lib_chaos/mesh"
	"testing"
)

type Path []mesh.Vert

func (p *Path) Append(x, y, z float64) {
	*p = append(*p, mesh.Vert{
		X: x,
		Y: y,
		Z: z,
	})
}

func (p *Path) LastPos() mesh.Vert {
	return (*p)[len(*p)-1]
}

func initMesh() *NavMesh {
	var nav = new(NavMesh)
	nav.MVert = make([]mesh.Vert, 36)
	for r := 0; r < 6; r++ {
		for c := 0; c < 6; c++ {
			nav.MVert[r*6+c] = mesh.Vert{
				X: float64(c),
				Z: float64(r),
			}
		}
	}
	nav.MPoly = make([]Poly, 25)
	for r := 0; r < 5; r++ {
		for c := 0; c < 5; c++ {
			var p = Poly{
				Link:    0,
				VertCnt: 4,
				AreaId:  0,
				Vs:      [6]int32{},
			}
			p.Vs[0] = int32((r+1)*6 + c)
			p.Vs[1] = int32((r+1)*6 + c + 1)
			p.Vs[2] = int32(r*6 + c + 1)
			p.Vs[3] = int32(r*6 + c)

			var next int32 = -1
			if c > 0 {
				nav.MLink = append(nav.MLink, Link{
					ToRef: int32(r*5 + c - 1),
					Next:  next,
					Edge:  3,
				})
				next = int32(len(nav.MLink) - 1)
			}
			if c < 4 {
				nav.MLink = append(nav.MLink, Link{
					ToRef: int32(r*5 + c + 1),
					Next:  next,
					Edge:  1,
				})
				next = int32(len(nav.MLink) - 1)
			}
			if r > 0 {
				nav.MLink = append(nav.MLink, Link{
					ToRef: int32((r-1)*5 + c),
					Next:  next,
					Edge:  2,
				})
				next = int32(len(nav.MLink) - 1)
			}
			if r < 4 {
				nav.MLink = append(nav.MLink, Link{
					ToRef: int32((r+1)*5 + c),
					Next:  next,
					Edge:  0,
				})
				next = int32(len(nav.MLink) - 1)
			}

			p.Link = next
			nav.MPoly[r*5+c] = p
		}
	}
	nav.MPoly[4].AreaId = 100
	nav.MPoly[8].AreaId = 100
	nav.MPoly[12].AreaId = 100
	nav.MPoly[16].AreaId = 100
	nav.MPoly[23].AreaId = 100
	return nav
}

func TestHeight(t *testing.T) {
	var (
		v0 = mesh.Vert{X: 0, Z: 0}
		v1 = mesh.Vert{X: 1, Z: 0}
		v2 = mesh.Vert{X: 0, Z: 1}
		p  = mesh.Vert{X: 0.2, Y: 1, Z: 0.2}
	)
	fmt.Println(mesh.VHeightOnTriangle(p, v0, v1, v2))
}

func TestNavMesh_LocatePoly(t *testing.T) {
	var nav = initMesh()
	fmt.Println(nav.LocatePoly(mesh.Vert{
		X: 0.5,
		Y: 0.5,
	}))
}

func TestNavMesh(t *testing.T) {
	var nav = initMesh()
	var q = NewQuery(nav, 100)
	var p Path
	fmt.Println(q.FindPath(mesh.Vert{
		X: 0.5,
		Z: 0.5,
	}, mesh.Vert{
		X: 4.5,
		Z: 4.5,
	}, &p))

	fmt.Println(p)
}
