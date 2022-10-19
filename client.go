package main

import (
	"encoding/binary"
	"fmt"
	"lib_chaos/sx"
	"math/rand"
	"net"
)

func main() {
	try()
}

func try() {
	var conn, err = net.Dial("tcp", "0.0.0.0:8881")
	if err != nil {
		fmt.Println(err)
		return
	}
	var send = sx.NewSx(conn)
	go func() {
		send.Start()
		send.Close()
	}()
	for {
		var s = rn()
		var header = make([]byte, 4)
		binary.BigEndian.PutUint32(header, uint32(len(s)))
		send.SendRaw(append(header, s...))
	}
}

var names = []string{"hello lily", "hello africa", "hello john", "hello johnson", "hello floyd"}

func rn() string {
	return names[rand.Intn(len(names))]
}
