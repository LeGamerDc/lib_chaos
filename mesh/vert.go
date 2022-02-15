package mesh

import (
	"lib_chaos/common"
	"math"
	"math/rand"
)

type Vert struct {
	X, Y, Z float64
}

const (
	Eps      = 0.0001
	Eqs      = Eps * Eps // sqr of Eps
	BigFloat = 100000000.0
)

var NilVert = Vert{X: -1, Y: -1, Z: -1}

func VDist(a, b Vert) float64 {
	var dx = b.X - a.X
	var dy = b.Y - a.Y
	var dz = b.Z - a.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

func VEqual(a, b Vert) bool {
	var (
		dx = a.X - b.X
		dy = a.Y - b.Y
		dz = a.Z - b.Z
	)
	return dx*dx+dy*dy+dz*dz < Eps*Eps
}

func VSub(a, b Vert) Vert {
	return Vert{
		X: a.X - b.X,
		Y: a.Y - b.Y,
		Z: a.Z - b.Z,
	}
}

func VAdd(a, b Vert) Vert {
	return Vert{
		X: a.X + b.X,
		Y: a.Y + b.Y,
		Z: a.Z + b.Z,
	}
}

func VMul(a Vert, k float64) Vert {
	return Vert{
		X: k * a.X,
		Y: k * a.Y,
		Z: k * a.Z,
	}
}

func VSqr(v Vert) float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func VSqrt2D(v Vert) float64 {
	return math.Sqrt(v.X*v.X + v.Z*v.Z)
}

// TriArea2D take y from (ab x ac)
func TriArea2D(a, b, c Vert) float64 {
	var (
		abx = b.X - a.X
		abz = b.Z - a.Z
		acx = c.X - a.X
		acz = c.Z - a.Z
	)
	return acx*abz - abx*acz
}

// DistPtSegSqr2D dist from pt to segment p-q, (q-p)*t is the mapping point
func DistPtSegSqr2D(pt, p, q Vert) (t float64, ds float64) {
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

// IntersectSegSeg2D intersect between two segment
func IntersectSegSeg2D(ap, aq, bp, bq Vert) (s, t float64, ok bool) {
	var (
		u = VSub(aq, ap)
		v = VSub(bq, bp)
		w = VSub(ap, bp)
		d = VCrossXz(u, v)
	)
	if math.Abs(d) < Eps*Eps {
		return
	}
	s = VCrossXz(v, w) / d
	t = VCrossXz(u, w) / d
	return s, t, true
}

// ReflectPtThroughSeg2D rotate pt around segment
func ReflectPtThroughSeg2D(a, p, q Vert) (b Vert) {
	var (
		pq  = VSub(q, p)
		r   = VSqrt2D(pq)
		sin = pq.X / r
		cos = pq.Z / r
	)
	b = VSub(a, p)                              // translation -p
	b.X, b.Z = cos*b.X+sin*b.Z, sin*b.X-cos*b.Z // rotation pq^-1, then reflection
	b.X, b.Z = cos*b.X-sin*b.Z, sin*b.X+cos*b.Z // rotation pq
	return VAdd(b, p)                           // translation p
}

// DistPtPtThroughSeg2D dist from a to b through segment p-q
func DistPtPtThroughSeg2D(a, b Vert, p, q Vert) float64 {
	var s, t, ok = IntersectSegSeg2D(a, b, p, q)
	if !ok || s < 0 || s > 1 { // a b on same side of p-q
		b = ReflectPtThroughSeg2D(b, p, q)
		s, t, ok = IntersectSegSeg2D(a, b, p, q) // use reflect b over seg p-q
	}
	if !ok || (t >= 0 && t <= 1) {
		return VDist(a, b)
	}
	if t < 0.1 {
		return VDist(a, p) + VDist(p, b)
	} else {
		return VDist(a, q) + VDist(q, b)
	}
}

// VCrossXz cross product of a,b on plane xz
func VCrossXz(a, b Vert) float64 {
	return a.X*b.Z - b.X*a.Z
}

func VInter(a, b Vert, k float64) Vert {
	return Vert{
		X: a.X + (b.X-a.X)*k,
		Y: a.Y + (b.Y-a.Y)*k,
		Z: a.Z + (b.Z-a.Z)*k,
	}
}

func TriRandomPoint(a, b, c Vert) Vert {
	var (
		ab     = VSub(b, a)
		ac     = VSub(c, a)
		ub, uc = rand.Float64(), rand.Float64()
	)
	if ub+uc >= 1 {
		ub = 1 - ub
		uc = 1 - uc
	}
	return VAdd(a, VAdd(VMul(ab, ub), VMul(ac, uc)))
}

func VHeightOnTriangle(p, a, b, c Vert) (h float64, ok bool) {
	var (
		ab = VSub(b, a)
		ac = VSub(c, a)
		ap = VSub(p, a)
		d  = ab.Z*ac.X - ab.X*ac.Z
	)
	if d*d < Eps*Eps {
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
		h = a.Y + (ab.Y*v+ac.Y*u)/d
		return h, true
	}
	return
}

func TriBox2D(a, b, c Vert) (min, max Vert) {
	min.X = common.MinN(a.X, b.X, c.X)
	min.Z = common.MinN(a.Z, b.Z, c.Z)
	max.X = common.MaxN(a.X, b.X, c.X)
	max.Z = common.MaxN(a.Z, b.Z, c.Z)
	return
}

func InTriOuterCircle(v, v0, v1, v2 Vert) bool {
	var (
		ma, mc  = VInter(v0, v1, 0.5), VInter(v0, v2, 0.5)
		va, vc  = VSub(v1, v0), VSub(v2, v0)
		ta, tc  = Vert{X: va.Z, Y: va.Y, Z: -va.X}, Vert{X: vc.Z, Y: vc.Y, Z: -vc.X}
		na, nc  = VAdd(ma, ta), VAdd(mc, tc)
		s, _, _ = IntersectSegSeg2D(ma, na, mc, nc)
		center  = VInter(ma, na, s)
	)
	return VDist(center, v) <= VDist(center, v0)-Eps
}
