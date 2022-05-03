package alloc

type Msg struct {
	c1, c2 *int64
	msg    interface{}
}

func (m *Msg) Close() error { // implement io.Closer
	decPage(m.c1)
	if m.c2 != nil {
		decPage(m.c2)
	}
	return nil
}
