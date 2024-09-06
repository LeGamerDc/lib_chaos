package main

import (
	"fmt"
)

type PbHome struct {
	Addr string
	Size int
}

type PbPerson struct {
	Name string
	Age  *int
	Home *PbHome
}

const (
	heapArenaBytes         = 4 << 20
	userArenaChunkBytesMax = 8 << 20
	userArenaChunkBytes    = uintptr(int64(userArenaChunkBytesMax-heapArenaBytes)&(int64(userArenaChunkBytesMax-heapArenaBytes)>>63) + heapArenaBytes) // min(userArenaChunkBytesMax, heapArenaBytes)
)

func main() {
	//var ac alloc.Allocator
	//ac.Init()
	//
	//msg := ac.CreateMsg(func(buf *alloc.Buf) interface{} {
	//    p := alloc.Malloc[PbPerson](buf)
	//    p.Name = "john"
	//    p.Age = alloc.Malloc[int](buf)
	//    *p.Age = 18
	//    p.Home = alloc.Malloc[PbHome](buf)
	//    *p.Home = PbHome{
	//        Addr: "xxx-xxx-xx",
	//        Size: 125,
	//    }
	//    return p
	//})
	//// do marshal with msg
	//_ = msg.Close()
	fmt.Println(userArenaChunkBytes / 1024)
}
