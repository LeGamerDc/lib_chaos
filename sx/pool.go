package sx

import "sync"

var messagePool = sync.Pool{
	New: func() interface{} {
		return new(Message)
	},
}
