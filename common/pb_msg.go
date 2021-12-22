package common

import (
	pbBuffer "lib_chaos/pb_buffer"
	"sync/atomic"
)

type PbMsg interface {
	XXX_Size() int
	XXX_Marshal([]byte, bool) ([]byte, error)
}

type PbBufMsg struct {
	buf pbBuffer.PbBuffer
	Msg PbMsg
}

func (pb *PbBufMsg) XXX_Size() int {
	return pb.Msg.XXX_Size()
}
func (pb *PbBufMsg) XXX_Marshal(b []byte, d bool) ([]byte, error) {
	return pb.Msg.XXX_Marshal(b, d)
}
func (pb *PbBufMsg) Close() error {
	pb.buf.Destroy()
	return nil
}

type PbConcurrentBufMsg struct {
	buf *PbBufMsg
	cnt int64
}

func NewPbConcurrentBuf(msg PbMsg) *PbConcurrentBufMsg {
	return &PbConcurrentBufMsg{
		buf: &PbBufMsg{
			Msg: msg,
		},
	}
}

func (pb *PbConcurrentBufMsg) XXX_Size() int {
	return pb.buf.Msg.XXX_Size()
}

func (pb *PbConcurrentBufMsg) XXX_Marshal(b []byte, d bool) ([]byte, error) {
	return pb.buf.Msg.XXX_Marshal(b, d)
}

func (pb *PbConcurrentBufMsg) Inc(n int64) {
	atomic.AddInt64(&pb.cnt, n)
}

func (pb *PbConcurrentBufMsg) Close() error {
	if atomic.AddInt64(&pb.cnt, -1) == 0 {
		return pb.buf.Close()
	}
	return nil
}
