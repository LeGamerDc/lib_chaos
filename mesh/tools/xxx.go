package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
)

func main() {
	process(load())
}

type Pos struct {
	X float64 `json:"x"`
	Z float64 `json:"z"`
}

type Polygon struct {
	Vert         []Pos `json:"Vertexs"`
	PolygonIndex int   `json:"PolygonIndex"`
}

type Map struct {
	MapWidth  float64   `json:"MapWidth"`
	MapHeight float64   `json:"MapHeight"`
	StartX    float64   `json:"StartX"`
	StartZ    float64   `json:"StartZ"`
	Mesh      []Polygon `json:"NavMeshPolygons"`
}

func load() *Map {
	var data, e = ioutil.ReadFile("area_1.json")
	if e != nil {
		panic(e)
	}
	var m Map
	e = json.Unmarshal(data, &m)
	if e != nil {
		panic(e)
	}
	return &m
}

func process(m *Map) {
	var h hash
	var e edgeCnt
	e.Init()
	h.Init(m.StartX, m.StartZ, m.MapWidth, m.MapHeight, 100.0)
	for _, p := range m.Mesh {
		var last, first = -1, -1
		for _, v := range p.Vert {
			id := h.Get(v.X, v.Z, true)
			if first == -1 {
				first = id
			}
			if last != -1 {
				if e.Insert(last, id) {
					fmt.Println(p.PolygonIndex)
				}
			}
			last = id
		}
		if e.Insert(last, first) {
			fmt.Println(p.PolygonIndex)
		}
	}
	fmt.Println(len(h.pos), e.Cnt())
	for _, p := range h.pos {
		fmt.Printf("%f %f\n", p.x, p.z)
	}
	for k, v := range e.m {
		if v%2 == 1 {
			fmt.Printf("%d %d\n", k.a, k.b)
		}
	}
}

type edge struct {
	a, b int
}

type edgeCnt struct {
	m map[edge]int
}

func (e *edgeCnt) Init() {
	e.m = make(map[edge]int)
}

func (e *edgeCnt) Cnt() int {
	var cnt int
	for _, v := range e.m {
		if v%2 == 1 {
			cnt++
		}
	}
	return cnt
}

func (e *edgeCnt) Insert(a, b int) bool {
	if a > b {
		a, b = b, a
	} else if a == b {
		return false
	}
	e.m[edge{a: a, b: b}]++
	return a == 5108 && b == 5109
}

type item struct {
	x, z float64
	idx  int
}

type slot struct {
	k []item
}

type hash struct {
	m   [][]slot
	pos []item

	sx, sz, x, z, gap float64
}

func (h *hash) Get(x, z float64, insert bool) int {
	x, z = round(x), round(z)
	ix, iz := h.find(x, z)
	for _, i := range h.m[ix][iz].k {
		if i.x == x && i.z == z {
			return i.idx
		}
	}
	if !insert {
		return -1
	}
	idx := len(h.pos)
	h.pos = append(h.pos, item{x: x, z: z, idx: idx})
	h.m[ix][iz].k = append(h.m[ix][iz].k, item{x: x, z: z, idx: idx})
	return idx
}

func (h *hash) Init(sx, sz, x, z float64, gap float64) {
	sx = round(sx - gap)
	sz = round(sz - gap)
	x = round(x + gap)
	z = round(z + gap)
	h.sx, h.sz, h.x, h.z, h.gap = sx, sz, x, z, gap
	h.m = make([][]slot, int(x/gap))
	for i := range h.m {
		h.m[i] = make([]slot, int(z/gap))
	}
}

func (h *hash) find(x, z float64) (int, int) {
	return int((x - h.sx) / h.gap), int((z - h.sz) / h.gap)
}

func round(x float64) float64 {
	return math.Round(x*8.0) / 8.0
}
