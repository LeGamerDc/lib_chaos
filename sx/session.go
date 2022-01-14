package sx

import (
	"fmt"
	"net"
)

// Handler : client or server might use session (like send msg, close, etc.)
// client could use session give by Start Callback
// server could use session in Handle Callback
type Handler interface {
	Start(*Session)
	Handle([]byte, *Session)
	Close()
}

type Session struct {
	conn  *net.TCPConn
	s     *sx
	r     *rx
	close func(session *Session) // ntf of self close
}

func _session(conn *net.TCPConn, close func(*Session)) *Session {
	return &Session{
		conn:  conn,
		close: close,
	}
}

func (s *Session) Send(raw []byte) {
	s.s.SendRaw(raw)
}

func (s *Session) SendOwn(raw []byte) {
	s.s.SendMsg(&Message{
		raw: raw,
		ctl: 1,
	})
}

func (s *Session) SendMsg(m *Message) {
	s.s.SendMsg(m)
}

func (s *Session) Close() {
	if s.conn.Close() == nil {
		s.s.Close()
	}
}

func (s *Session) Start(h Handler) {
	s.s = NewSx(s.conn)
	s.r = NewRx(s.conn)
	go func() {
		if e := s.r.Start(func(raw []byte) {
			h.Handle(raw, s)
		}); e != nil {
			fmt.Println("session stop: ", e)
		}
		s.close(s)
	}()
	go s.s.Start()
}
