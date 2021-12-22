package main

import (
	"fmt"
	"io"
	"lib_chaos/common"
	"lib_chaos/pb_buffer"
	"time"
)

type Person struct {
	name  string
	age   int
	house *House
}

type House struct {
	addr string
	size int
}

func (p *Person) String() string {
	if p.house != nil {
		return fmt.Sprintf("name: %s, age %d, house:[%s]", p.name, p.age, p.house)
	} else {
		return fmt.Sprintf("name: %s, age %d, homeless", p.name, p.age)
	}
}

func (h *House) String() string {
	return fmt.Sprintf("addr: %s, size: %d", h.addr, h.size)
}

func (h *Person) XXX_Size() int { return 0 }
func (h *Person) XXX_Marshal([]byte, bool) ([]byte, error) {
	return []byte(h.String()), nil
}

func main() {
	var buf = new(pbBuffer.PbBuffer)
	var p = pbBuffer.Malloc[Person](buf)
	p.name = pbBuffer.CopyString(buf, "john")
	p.age = 22
	p.house = pbBuffer.Malloc[House](buf)
	p.house.addr = pbBuffer.CopyString(buf, "天府新区-兴隆湖-xx街-21-1")
	p.house.size = 169

	var c = common.NewPbConcurrentBuf(p)
	c.Inc(4)
	for i := 1; i <= 4; i++ {
		var ch = make(chan common.PbMsg, 1)
		go func(x int) {
			time.Sleep(time.Second * time.Duration(x))
			m := <-ch
			data, _ := m.XXX_Marshal(nil, false)
			fmt.Printf("g %d: %s\n", x, string(data))
			if mc, ok := m.(io.Closer); ok {
				_ = mc.Close()
			}
		}(i)
		ch <- c
	}
	time.Sleep(time.Second * 6)
}
