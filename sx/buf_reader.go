package sx

import (
	"io"
)

const (
	bufSize = 1024 * 8 // 8 KB
)

type zcReader struct {
	buf      []byte
	rd       io.Reader
	r, w, rp int
}

func newZcReader(rd io.Reader) *zcReader {
	return &zcReader{
		buf: make([]byte, bufSize),
		rd:  rd,
	}
}

func (r *zcReader) Read(n int) (data []byte, err error) {
	if r.w-r.r >= n { // enough data
		r.rp = r.r + n
		return r.buf[r.r : r.r+n], nil
	}
	if bufSize-r.r >= n { // enough buffer
		if err = r.fill(n); err != nil {
			return
		}
		r.rp = r.r + n
		return r.buf[r.r : r.r+n], nil
	}
	if bufSize >= n { // enough buffer if move data
		copy(r.buf, r.buf[r.r:r.w])
		r.w -= r.r
		r.r = 0
		if err = r.fill(n); err != nil {
			return
		}
		r.rp = r.r + n
		return r.buf[r.r : r.r+n], nil
	}
	// data size > buffer
	data = make([]byte, n)
	copy(data, r.buf[r.r:r.w])
	_, err = io.ReadFull(r.rd, data[r.w-r.r:])
	r.r, r.w, r.rp = 0, 0, 0
	return
}

func (r *zcReader) Confirm() {
	r.r = r.rp
}

func (r *zcReader) fill(n int) (err error) {
	var cnt int
	for r.w-r.r < n {
		cnt, err = r.rd.Read(r.buf[r.w:])
		if err != nil {
			return
		}
		r.w += cnt
	}
	return
}
