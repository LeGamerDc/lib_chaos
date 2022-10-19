package sx

import (
	"encoding/binary"
	"net"
)

type rx struct {
	rd zcReader
}

func NewRx(conn net.Conn) *rx {
	return &rx{rd: *newZcReader(conn)}
}

func (p *rx) Start(handle func([]byte)) (err error) {
	var data []byte
	var size uint32
	for {
		p.rd.Confirm()
		data, err = p.rd.Read(4)
		if err != nil {
			return
		}
		size = binary.BigEndian.Uint32(data[:4])
		p.rd.Confirm()
		data, err = p.rd.Read(int(size))
		if err != nil {
			return
		}
		handle(data)
	}
}
