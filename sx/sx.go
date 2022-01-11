package sx

import (
	"fmt"
	"lib_chaos/common"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

// sx gather []byte and send by writev to reduce syscall.
// sx could cooperate with mcache in reuse []byte, you should
// send *Message and use Message.ctl to tell sx how to reuse []byte.

type Message struct {
	raw []byte
	ctl int32 // 0: no recycle, 1: direct recycle, 2: ref-count recycle
	ref int32
}

type sx struct {
	q    chan *Message
	quit chan struct{}

	conn   net.Conn
	buf    []*Message
	_block [][]byte
}

func NewSx(conn net.Conn) *sx {
	return &sx{
		q:    make(chan *Message, 1024),
		quit: make(chan struct{}),
		conn: conn,
	}
}

func (s *sx) Start() {
	defer close(s.quit)
	for m := range s.q {
		s.buf = append(s.buf, m)
		if s.loop() {
			break
		}
	}
}

func (s *sx) SendRaw(raw []byte) {
	s.q <- &Message{
		raw: raw,
	}
}

func (s *sx) SendMsg(m *Message) {
	s.q <- m
}

func (s *sx) Close() {
	close(s.q)
	select {
	case <-time.After(time.Second * 5):
	case <-s.quit:
	}
}

func (s *sx) loop() (quit bool) {
	runtime.Gosched()
Loop:
	for {
		select {
		case m, ok := <-s.q:
			if !ok {
				quit = true
				break Loop
			}
			s.buf = append(s.buf, m)
		default:
			break Loop
		}
	}
	var e error
	if len(s.buf) == 1 {
		_, e = s.conn.Write(s.buf[0].raw)
	} else {
		for _, m := range s.buf {
			s._block = append(s._block, m.raw)
		}
		var tmp = s._block // avoid update s._block
		_, e = (*net.Buffers)(&tmp).WriteTo(s.conn)
		for i := 0; i < len(s._block); i++ { // compiler optimized
			s._block[i] = nil
		}
		s._block = s._block[0:0]
	}
	for i := 0; i < len(s.buf); i++ {
		switch s.buf[i].ctl {
		case 0: // do nothing
		case 1: // recycle
			common.Free(s.buf[i].raw)
		case 2:
			if atomic.AddInt32(&s.buf[i].ref, -1) == 0 { // recycle
				common.Free(s.buf[i].raw)
			}
		}
		s.buf[i] = nil
	}
	s.buf = s.buf[0:0]
	if e != nil {
		fmt.Printf("sx write fail: %s", e.Error())
	}
	return quit || e != nil
}
