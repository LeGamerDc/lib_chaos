package pathfinding

import (
	"math"
)

const (
	eps      float64 = 0.01
	eqs      float64 = 0.1 // sqrt of eps
	bigFloat         = 10000000.0
)

var NilVert = Vert{X: -1, Y: -1, Z: -1}

type Vert struct {
	X, Y, Z float64
}

func vDist(a, b Vert) float64 {
	var dx = b.X - a.X
	var dy = b.Y - a.Y
	var dz = b.Z - a.Z
	return float64(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func vEqual(a, b Vert) bool {
	var (
		dx = a.X - b.X
		dy = a.Y - b.Y
		dz = a.Z - b.Z
	)
	return dx*dx+dy*dy+dz*dz < eps*eps
}

func vSub(a, b Vert) Vert {
	return Vert{
		X: a.X - b.X,
		Y: a.Y - b.Y,
		Z: a.Z - b.Z,
	}
}

func vSqr(v Vert) float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// take y from (ab x ac)
func triArea2D(a, b, c Vert) float64 {
	var (
		abx = b.X - a.X
		abz = b.Z - a.Z
		acx = c.X - a.X
		acz = c.Z - a.Z
	)
	return acx*abz - abx*acz
}

// dist from pt to segment p-q, (q-p)*t is the mapping point
func distPtSegSqr2D(pt, p, q Vert) (t float64, ds float64) {
	var (
		pqx = q.X - p.X
		pqz = q.Z - p.Z
		dx  = pt.X - p.X
		dz  = pt.Z - p.Z
		d   = pqx*pqx + pqz*pqz
	)
	t = pqx*dx + pqz*dz
	if d > 0 {
		t /= d
	}
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	dx = p.X - pt.X + t*pqx
	dz = p.Z - pt.Z + t*pqz
	return t, dx*dx + dz*dz
}

// intersect between two segment
func intersectSegSeg2D(ap, aq, bp, bq Vert) (s, t float64, ok bool) {
	var (
		u = vSub(aq, ap)
		v = vSub(bq, bp)
		w = vSub(ap, bp)
		d = vCrossXz(u, v)
	)
	if math.Abs(d) < eps*eps {
		return
	}
	s = vCrossXz(v, w) / d
	t = vCrossXz(u, w) / d
	return s, t, true
}

// cross product of a,b on plane xz
func vCrossXz(a, b Vert) float64 {
	return a.X*b.Z - a.Z*b.X
}

func vInter(a, b Vert, k float64) Vert {
	return Vert{
		X: a.X + (b.X-a.X)*k,
		Y: a.Y + (b.Y-a.Y)*k,
		Z: a.Z + (b.Z-a.Z)*k,
	}
}

func vHeightOnTriangle(p, a, b, c Vert) (h float64, ok bool) {
	var (
		ab = vSub(b, a)
		ac = vSub(c, a)
		ap = vSub(p, a)
		d  = ab.Z*ac.X - ab.X*ac.Z
	)
	if d*d < eps*eps {
		return
	}
	var (
		u = ab.Z*ap.X - ab.X*ap.Z
		v = ap.Z*ac.X - ap.X*ac.Z
	)
	if d < 0 {
		d, u, v = -d, -u, -v
	}
	if u >= 0 && v >= 0 && (u+v) <= d {
		h = a.Y + (ab.Y*u+ac.Y*v)/d
		return h, true
	}
	return
}

func nextPow2(x uint32) uint32 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}

// https://gist.github.com/badboy/6267743
func hash64(ix int64) int32 {
	var x = uint64(ix)
	x = (^x) + (x << 18)
	x = x ^ (x >> 31)
	x = x * 21
	x = x ^ (x >> 11)
	x = x + (x << 6)
	x = x ^ (x >> 22)
	return int32(x)
}
func hash32(ix int32) int32 {
	var x = uint32(ix)
	x += ^(x << 15)
	x ^= x >> 10
	x += x << 3
	x ^= x >> 6
	x += ^(x << 11)
	x ^= x >> 16
	return int32(x)
}

func memset(a []int32, x int32) {
	if len(a) == 0 {
		return
	}
	a[0] = x
	for bp := 1; bp < len(a); bp <<= 1 {
		copy(a[bp:], a[:bp])
	}
}
