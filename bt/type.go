package bt

import (
	"unsafe"
)

type (
	BvTime int32
	BvNode unsafe.Pointer
	StNode unsafe.Pointer
)

const (
	Success = int32(iota)
	Failure
	Running
)

// behavior
type BehaviorKind int32

const (
	BvWait = BehaviorKind(iota)
	BvWaitForever
	BvAction
	BvFail
	BvAlwaysSucceed
	BvSelect
	BvIf
	BvSequence
	BvWhile
	BvWhenAll
	BvWhenAny
)

type BvNodeWait struct {
	Head BehaviorKind
	Secs BvTime
}
type BvNodeWaitForever struct {
	Head BehaviorKind
}
type BvNodeAction struct {
	Head BehaviorKind
	fn   ActionFn
}
type BvNodeFail struct {
	Head  BehaviorKind
	inner BvNode
}
type BvNodeAlwaysSucceed struct {
	Head  BehaviorKind
	inner BvNode
}
type BvNodeSelect struct {
	Head BehaviorKind
	todo []BvNode
}
type BvNodeIf struct {
	Head             BehaviorKind
	cond, succ, fail BvNode
}
type BvNodeSequence struct {
	Head BehaviorKind
	todo []BvNode
}
type BvNodeWhile struct {
	Head BehaviorKind
	cond BvNode
	todo []BvNode
}
type BvNodeWhenAll struct {
	Head BehaviorKind
	todo []BvNode
}
type BvNodeWhenAny struct {
	Head BehaviorKind
	todo []BvNode
}

// behavior end

// state
type StateKind int32

const (
	StAction = StateKind(iota)
	StFail
	StAlwaysSucceed
	StWait
	StWaitForever
	StIf
	StSelect
	StSequence
	StWhile
	StWhenAll
	StWhenAny
)

type StNodeAction struct {
	Head StateKind
	fn   ActionFn
}
type StNodeFail struct {
	Head  StateKind
	state StNode
}
type StNodeAlwaysSucceed struct {
	Head  StateKind
	state StNode
}
type StNodeWait struct {
	Head StateKind
	end  BvTime
}
type StNodeWaitForever struct {
	Head StateKind
}
type StNodeIf struct {
	Head       StateKind
	Status     int32
	succ, fail BvNode
	state      StNode
}
type StNodeSelect struct {
	Head StateKind
	idx  int32
	seq  []BvNode
	cur  StNode
}
type StNodeSequence struct {
	Head StateKind
	idx  int32
	seq  []BvNode
	cur  StNode
}
type StNodeWhile struct {
	Head  StateKind
	idx   int32
	cond  StNode
	todo  []BvNode
	state StNode
}
type StNodeWhenAll struct {
	Head   StateKind
	states []StNode
}
type StNodeWhenAny struct {
	Head   StateKind
	states []StNode
}

type ActionFn func() int32
