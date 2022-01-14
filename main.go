package main

import (
	"fmt"
	"unsafe"
)

type A struct {
	a int
	b byte
}

func main() {
	fmt.Println(unsafe.Sizeof(A{}))
}
