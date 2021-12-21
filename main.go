package main

import (
	"lib_chaos/ecs"
	"math/rand"
)

const (
	SquareSuffix = 3
)

type Square struct {
	id       int64
	troopCnt int64
}

func main() {
	var (
		sqs  = ecs.MakeSparseArray[Square](1024)
		tick int32
		m    = map[int64]int64{}
	)
	for tick = 0; tick < 10; tick++ {
		for i := 0; i < 999; i++ {
			id, sq := sqs.Place(tick, SquareSuffix)
			sq.troopCnt = rand.Int63n(1000)
			sq.id = id
			m[id] = sq.troopCnt
		}
	}
	for id, cnt := range m {
		sq := sqs.Get(id)
		if sq == nil || sq.troopCnt != cnt {
			panic("not equal")
		}
	}
	sqs.Foreach(func(sq *Square) {
		if cnt, ok := m[sq.id]; !ok || cnt != sq.troopCnt {
			panic("not equal")
		}
	})
}
