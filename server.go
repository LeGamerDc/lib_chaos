package main

import (
	"fmt"
	"lib_chaos/sx"
	"net"
	"time"
)

func main() {
	var (
		ln  net.Listener
		err error
	)
	ln, err = net.Listen("tcp", "0.0.0.0:8881")
	if err != nil {
		fmt.Println(err)
		return
	}
	go start()
	for {
		var conn net.Conn
		conn, err = ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			var port = sx.NewRx(conn)
			var cnt int
			var last = coarseTime
			var err = port.Start(func(data []byte) {
				//fmt.Println(string(data))
				cnt++
				if coarseTime-last >= 4 {
					last = coarseTime
					fmt.Println(cnt / 4)
					cnt = 0
				}
			})
			fmt.Println(err)
		}(conn)
	}
}

var coarseTime int64

func start() {
	for t := range time.NewTicker(time.Second).C {
		coarseTime = t.Unix()
	}
}
