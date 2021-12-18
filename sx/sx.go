package sx

import (
    "github.com/bytedance/gopkg/lang/mcache"
    "lib_chaos/sx/wire"
    "net"
    "runtime"
    "time"
)

type Message struct {
    raw  []byte
    gogo wire.Gogo
    recycle bool
}

type sx struct {
    // control
    q chan *Message
    quit chan struct{}

    // data
    conn net.Conn
    buf  []*Message
    data [][]byte
}

func NewSx(conn net.Conn) *sx {
    return &sx{
        q: make(chan *Message, 1024),
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

func (s *sx) Send(data []byte) {
    var m = messagePool.Get().(*Message)
    m.raw = data
    s.q <- m
}

func getData(m *Message) []byte {
    if len(m.raw) > 0 {
        return m.raw
    }
    var s = m.gogo.XXX_Size()
    var data = mcache.Malloc(s+4)
    _, _ = m.gogo.XXX_Marshal(data, false)
    m.recycle = true
    m.raw = data
    return data[:s+4]
}

func (s *sx) loop() (quit bool) {
    runtime.Gosched()
Loop:
    for {
        select {
        case m, ok := <- s.q:
            if !ok {
                quit = true
                break Loop
            }
            s.buf = append(s.buf, m)
        default:
            break Loop
        }
    }

    var err error
    if len(s.buf) == 1 {
        var data = getData(s.buf[0])
        _, err = s.conn.Write(data)
    } else {
        for _, m := range s.buf {
            s.data = append(s.data, getData(m))
        }
        var tmp = s.data
        _, err = (*net.Buffers)(&tmp).WriteTo(s.conn)

        // clean s.data
        for i:=0; i<len(s.data); i++ {
            s.data[i] = nil
        }
        s.data = s.data[0:0]
    }
    // clean s.buf
    for _, m := range s.buf {
        if m.recycle {
            mcache.Free(m.raw)
        }
        messagePool.Put(m)
    }
    for i:=0; i<len(s.buf); i++ {
        s.buf[i] = nil
    }
    s.buf = s.buf[0:0]

    if err != nil {
        return true
    }
    return
}

func (s *sx) Close() {
    close(s.q)
    select {
    case <- time.After(time.Second * 5):
    case <- s.quit:
    }
}

//func (s *sx) Start2() {
//    defer close(s.quit)
//    for m := range s.q {
//        _, err := s.conn.Write(m.raw)
//        if err != nil {
//            return
//        }
//    }
//}
//
//func (s *sx) Start3() {
//    defer close(s.quit)
//    for m := range s.q {
//        if s.loop2(m.raw) {
//            break
//        }
//    }
//}
//func (s *sx) loop2(first []byte) (quit bool) {
//    runtime.Gosched()
//    var buf *p.ByteBuffer
//Loop:
//    for {
//        select {
//        case m, ok := <- s.q:
//            if !ok {
//                quit = true
//                break Loop
//            }
//            if buf == nil {
//                buf = pool.Get()
//                _, _ = buf.Write(first)
//            }
//            _, _ = buf.Write(m.raw)
//        default:
//            break Loop
//        }
//    }
//    var err error
//    if buf != nil {
//        _, err = s.conn.Write(buf.Bytes())
//        pool.Put(buf)
//    } else {
//        _, err = s.conn.Write(first)
//    }
//    if err != nil {
//        return true
//    }
//    return
//}
//var pool = p.Pool{}