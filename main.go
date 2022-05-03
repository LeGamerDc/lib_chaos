package main

import (
	"fmt"
	"lib_chaos/alloc"
	"unsafe"
)

func main() {
	fmt.Println(size[alloc.Msg]())
}

func size[T any]() int {
	var x T
	return int(unsafe.Sizeof(x))
}
