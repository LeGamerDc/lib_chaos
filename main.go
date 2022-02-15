package main

import (
	"fmt"
	"lib_chaos/mesh"
)

func main() {
	var (
		a = mesh.Vert{X: 0, Z: 1}
		b = mesh.Vert{X: 1, Z: 1}
		c = mesh.Vert{X: 1, Z: 0}
	)
	fmt.Println(mesh.TriArea2D(a, b, c))
}
