package main

import (
	"fmt"
	"unsafe"
)

const (
	pageStructSize = 16 * 1024 // 16KB
	pageSize       = int(pageStructSize - unsafe.Sizeof(int64(0)) - unsafe.Sizeof(0))
)

type page struct {
	cnt int64
	off int
	buf [pageSize]byte
}

func main() {
	var p = page{}
	fmt.Println(unsafe.Sizeof(p))
}
