package main

import "lib_chaos/alloc"

type PbHome struct {
    Addr string
    Size int
}

type PbPerson struct {
    Name string
    Age  *int
    Home *PbHome
}

func main() {
    var ac alloc.Allocator
    ac.Init()

    msg := ac.CreateMsg(func(buf *alloc.Buf) interface{} {
        p := alloc.Malloc[PbPerson](buf)
        p.Name = "john"
        p.Age = alloc.Malloc[int](buf)
        *p.Age = 18
        p.Home = alloc.Malloc[PbHome](buf)
        *p.Home = PbHome{
            Addr: "xxx-xxx-xx",
            Size: 125,
        }
        return p
    })
    // do marshal with msg
    _ = msg.Close()
}
