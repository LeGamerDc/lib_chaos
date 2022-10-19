package main

import (
	"fmt"
)

func main() {
	var t = read()
	var buf = t.Svg()
	fmt.Print(buf.String())
}
