package main

import (
	"fmt"
	"lib_chaos/allocator"
	"unsafe"
)

func main() {
	fmt.Println(size[allocator.Msg]())
}

func size[T any]() int {
	var x T
	return int(unsafe.Sizeof(x))
}
