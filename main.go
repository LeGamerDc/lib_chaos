package main

import (
	"fmt"
	"lib_chaos/mesh"
)

func main() {
	var a = mesh.Vert{X: 0, Z: 1}
	var p = mesh.Vert{X: 1, Z: 0}
	var q = mesh.Vert{X: 2, Z: 1}

	fmt.Println(mesh.RotatePtThroughSeg2D(a, p, q))
}
