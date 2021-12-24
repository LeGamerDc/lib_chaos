package ecs

const _maxTick = 128

type TimeManager interface {
	Insert(int32, InvokeF)
}
type InvokeF func(ctx TimeManager)
type node struct {
	next   int32
	invoke InvokeF
}
type TimeWheel struct {
	slots []node
	free  []int32

	cycle            [_maxTick]int32
	curTick, maxTick int32
}

func NewTimeWheel(n int, m int32) *TimeWheel {
	var tw = &TimeWheel{
		slots:   make([]node, n),
		free:    make([]int32, n),
		curTick: 0,
		maxTick: m,
	}
	for i := 0; i < n; i++ {
		tw.free[i] = int32(i)
	}
	return tw
}

func (tw *TimeWheel) Tick() {
	tw.tick(tw.curTick)
	tw.curTick = (tw.curTick + 1) % tw.maxTick
}
func (tw *TimeWheel) tick(t int32) {
	var n = tw.cycle[t]
	for n != -1 {
		var node = &tw.slots[n]
		node.invoke(tw)
		node.invoke = nil
		tw.free = append(tw.free, n)
		n = node.next
	}
	tw.cycle[t] = -1
}
func (tw *TimeWheel) Insert(when int32, f InvokeF) {
	if when <= 0 || when >= tw.maxTick {
		// not enough tick
		return
	}
	var (
		idx = (tw.curTick + when) % tw.maxTick
		n   = tw.getNode()
	)
	if n == -1 {
		// not enough node
		return
	}
	tw.slots[n] = node{
		next:   tw.cycle[idx],
		invoke: f,
	}
	tw.cycle[idx] = n
}
func (tw *TimeWheel) getNode() int32 {
	var l = len(tw.free) - 1
	if l < 0 {
		return -1
	}
	var n = tw.free[l]
	tw.free = tw.free[:l]
	return n
}
