package main

import (
	"fmt"
	"lib_chaos/mesh"
)

func main() {
	var (
		turn  = mesh.Vert{X: 7504.927913, Z: 8515.870712}
		left  = mesh.Vert{X: 7347.875000, Z: 8860.000000}
		right = mesh.Vert{X: 7394.000000, Z: 8566.750000}
		ll    = mesh.Vert{X: 7347.875000, Z: 8860.000000}
		rr    = mesh.Vert{X: 7388.375000, Z: 8566.625000}
	)
	fmt.Println(mesh.DistPtSegSqr2D(turn, left, right))
	fmt.Println(mesh.TriArea2D(turn, left, ll))
	fmt.Println(mesh.TriArea2D(turn, left, rr))
	fmt.Println(mesh.TriArea2D(turn, right, rr))
	fmt.Println(mesh.TriArea2D(turn, right, ll))
}
