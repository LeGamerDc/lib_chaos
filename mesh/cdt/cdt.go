package cdt

import "lib_chaos/mesh"

type edge struct {
	from, to int32
}

type CDT struct {
	mVert []mesh.Vert
	mEdge []edge

	// setting
	min, max mesh.Vert
}

func (c *CDT) initBound() {

}

func (c *CDT) insertVert() {

}
