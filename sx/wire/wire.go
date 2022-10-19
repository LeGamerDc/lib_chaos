package wire

import "errors"

var (
	ErrInvalidBufferSize = errors.New("proto: buffer size not match msg size when marshal")
)

type Gogo interface {
	XXX_Size() int
	XXX_Marshal(b []byte, derterministic bool) ([]byte, error)
}

func SizeCall(seq uint32, api int32, msg Gogo, extra []byte) (n int) {
	var l int
	if seq != 0 {
		n += 1 + sovWire(uint64(seq))
	}
	n += 1 + sovWire(uint64(CallType_Call))
	if api != 0 {
		n += 1 + sovWire(uint64(api))
	}
	if msg != nil {
		l = msg.XXX_Size()
		n += 1 + l + sovWire(uint64(l))
	}
	l = len(extra)
	if l > 0 {
		n += 1 + l + sovWire(uint64(l))
	}
	return
}

func EncodeCall(b []byte, seq uint32, api int32, msg Gogo, extra []byte) (err error) {
	var r int
	var lp int
	if msg != nil {
		lp = msg.XXX_Size()
	}
	r, err = marshalMsg(b, seq, CallType_Call, api, 0, lp, extra)
	if err != nil {
		return err
	}
	if lp > 0 {
		_, err = msg.XXX_Marshal(b[r:r:r+lp], false)
		if err != nil {
			return err
		}
	}
	return
}

func SizeOneWay(api int32, msg Gogo, extra []byte) (n int) {
	var l int
	if api != 0 {
		n += 1 + sovWire(uint64(api))
	}
	if msg != nil {
		l = msg.XXX_Size()
		n += 1 + l + sovWire(uint64(l))
	}
	l = len(extra)
	if l > 0 {
		n += 1 + l + sovWire(uint64(l))
	}
	return
}

func EncodeOneWay(b []byte, api int32, msg Gogo, extra []byte) (err error) {
	var r int
	var lp int
	if msg != nil {
		lp = msg.XXX_Size()
	}
	r, err = marshalMsg(b, 0, CallType_OneWay, api, 0, lp, extra)
	if err != nil {
		return err
	}
	if lp > 0 {
		_, err = msg.XXX_Marshal(b[r:r:r+lp], false)
		if err != nil {
			return err
		}
	}
	return
}

func SizeException(seq uint32, code int32) (n int) {
	if seq != 0 {
		n += 1 + sovWire(uint64(seq))
	}
	n += 1 + sovWire(uint64(CallType_Exception))
	if code != 0 {
		n += 1 + sovWire(uint64(code))
	}
	return
}

func EncodeException(b []byte, seq uint32, code int32) (err error) {
	_, err = marshalMsg(b, seq, CallType_Exception, 0, code, 0, nil)
	if err != nil {
		return err
	}
	return
}

func SizeReply(seq uint32, msg Gogo) (n int) {
	var l int
	if seq != 0 {
		n += 1 + sovWire(uint64(seq))
	}
	n += 1 + sovWire(uint64(CallType_Reply))
	if msg != nil {
		l = msg.XXX_Size()
		n += 1 + l + sovWire(uint64(l))
	}
	return
}

func EncodeReply(b []byte, seq uint32, msg Gogo) (err error) {
	var r int
	var lp int
	if msg != nil {
		lp = msg.XXX_Size()
	}
	r, err = marshalMsg(b, seq, CallType_Reply, 0, 0, lp, nil)
	if err != nil {
		return err
	}
	if lp > 0 {
		_, err = msg.XXX_Marshal(b[r:r:r+lp], false)
		if err != nil {
			return err
		}
	}
	return
}

func marshalMsg(b []byte, seq uint32, t CallType, api int32, code int32, lp int, extra []byte) (r int, err error) {
	i := len(b)
	if len(extra) > 0 {
		i -= len(extra)
		copy(b[i:], extra)
		i = encodeVarintWire(b, i, uint64(len(extra)))
		i--
		b[i] = 0x32
	}
	if lp > 0 {
		i -= lp
		r = i
		i = encodeVarintWire(b, i, uint64(lp))
		i--
		b[i] = 0x2a
	}
	if code != 0 {
		i = encodeVarintWire(b, i, uint64(code))
		i--
		b[i] = 0x20
	}
	if api != 0 {
		i = encodeVarintWire(b, i, uint64(api))
		i--
		b[i] = 0x18
	}
	if t != 0 {
		i = encodeVarintWire(b, i, uint64(t))
		i--
		b[i] = 0x10
	}
	if seq != 0 {
		i = encodeVarintWire(b, i, uint64(seq))
		i--
		b[i] = 0x8
	}
	if i != 0 {
		return 0, ErrInvalidBufferSize
	}
	return
}
