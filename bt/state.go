package bt

import (
	"fmt"
)

func StateNew(behavior BvNode, now BvTime) (state StNode) {
	//fmt.Printf("bv %d\n", *(*BehaviorKind)(behavior))
	switch k := *(*BehaviorKind)(behavior); k {
	case BvWait:
		node := (*BvNodeWait)(behavior)
		return StNode(&StNodeWait{
			Head: StWait,
			end:  now + node.Secs,
		})
	case BvWaitForever:
		return StNode(&StNodeWaitForever{
			Head: StWaitForever,
		})
	case BvAction:
		node := (*BvNodeAction)(behavior)
		return StNode(&StNodeAction{
			Head: StAction,
			fn:   node.fn,
		})
	case BvFail:
		node := (*BvNodeFail)(behavior)
		return StNode(&StNodeFail{
			Head:  StFail,
			state: StateNew(node.inner, now),
		})
	case BvAlwaysSucceed:
		node := (*BvNodeAlwaysSucceed)(behavior)
		return StNode(&StNodeAlwaysSucceed{
			Head:  StAlwaysSucceed,
			state: StateNew(node.inner, now),
		})
	case BvIf:
		node := (*BvNodeIf)(behavior)
		return StNode(&StNodeIf{
			Head:   StIf,
			Status: Running,
			succ:   node.succ,
			fail:   node.fail,
			state:  StateNew(node.cond, now),
		})
	case BvSelect:
		node := (*BvNodeSelect)(behavior)
		return StNode(&StNodeSelect{
			Head: StSelect,
			idx:  0,
			seq:  node.todo,
			cur:  StateNew(node.todo[0], now),
		})
	case BvSequence:
		node := (*BvNodeSequence)(behavior)
		return StNode(&StNodeSequence{
			Head: StSequence,
			idx:  0,
			seq:  node.todo,
			cur:  StateNew(node.todo[0], now),
		})
	case BvWhile:
		node := (*BvNodeWhile)(behavior)
		return StNode(&StNodeWhile{
			Head:  StWhile,
			idx:   0,
			cond:  StateNew(node.cond, now),
			todo:  node.todo,
			state: StateNew(node.todo[0], now),
		})
	case BvWhenAll:
		node := (*BvNodeWhenAll)(behavior)
		return StNode(&StNodeWhenAll{
			Head:   StWhenAll,
			states: behavior2State(node.todo, now),
		})
	case BvWhenAny:
		node := (*BvNodeWhenAny)(behavior)
		return StNode(&StNodeWhenAny{
			Head:   StWhenAny,
			states: behavior2State(node.todo, now),
		})
	default:
		panic(fmt.Sprintf("unknown bv %d", k))
	}
}

func Event(state StNode, now BvTime) int32 {
	//fmt.Printf("exec %d\n", *(*StateKind)(state))
	switch k := *(*StateKind)(state); k {
	case StAction:
		node := (*StNodeAction)(state)
		return node.fn()
	case StFail:
		node := (*StNodeFail)(state)
		switch Event(node.state, now) {
		case Running:
			return Running
		case Failure:
			return Success
		case Success:
			return Failure
		}
		panic("unreachable")
	case StAlwaysSucceed:
		node := (*StNodeAlwaysSucceed)(state)
		switch Event(node.state, now) {
		case Running:
			return Running
		case Failure, Success:
			return Success
		}
		panic("unreachable")
	case StWait:
		node := (*StNodeWait)(state)
		if now >= node.end {
			return Success
		}
		return Running
	case StWaitForever:
		return Running
	case StIf:
		node := (*StNodeIf)(state)
		for {
			switch node.Status {
			case Running:
				switch Event(node.state, now) {
				case Running:
					return Running
				case Success:
					node.state = StateNew(node.succ, now)
					node.Status = Success
				case Failure:
					node.state = StateNew(node.fail, now)
					node.Status = Failure
				}
			case Success, Failure:
				return Event(node.state, now)
			}
		}
	case StSelect:
		node := (*StNodeSelect)(state)
		return sequence(true, node.seq, &node.idx, &node.cur, now)
	case StSequence:
		node := (*StNodeSequence)(state)
		return sequence(false, node.seq, &node.idx, &node.cur, now)
	case StWhile:
		node := (*StNodeWhile)(state)
		switch k := Event(node.cond, now); k {
		case Success, Failure:
			return k
		}
		for {
			switch k := Event(node.state, now); k {
			case Failure, Running:
				return k
			case Success:
				node.idx++
				if int(node.idx) >= len(node.todo) {
					node.idx = 0
				}
				node.state = StateNew(node.todo[node.idx], now)
			}
		}
	case StWhenAll:
		node := (*StNodeWhenAll)(state)
		return when(false, node.states, now)
	case StWhenAny:
		node := (*StNodeWhenAny)(state)
		return when(true, node.states, now)
	default:
		panic(fmt.Sprintf("unknown state %d", k))
	}
}

func when(any bool, states []StNode, now BvTime) int32 {
	var full, stop = Success, Failure
	if any {
		full, stop = Failure, Success
	}
	finished := 0
	for i, state := range states {
		if state != nil {
			k := Event(state, now)
			if k == Running {
				continue
			} else if k == stop {
				return stop
			}
			states[i] = nil
		}
		finished++
	}
	if finished == len(states) {
		return full
	}
	return Running
}

func sequence(sel bool, seq []BvNode, idx *int32, cursor *StNode, now BvTime) int32 {
	var (
		status, inv_status = Failure, Success // select
	)
	if !sel {
		status, inv_status = Success, Failure // sequence
	}
	for {
		var s = Event(*cursor, now)
		if s == Running {
			return Running
		} else if s == inv_status {
			return inv_status
		} else {
			if int(*idx)+1 == len(seq) { // finished
				return status
			}
			*idx++
			*cursor = StateNew(seq[*idx], now)
		}
	}
}

func behavior2State(bvs []BvNode, now BvTime) (sts []StNode) {
	sts = make([]StNode, 0, len(bvs))
	for _, bv := range bvs {
		sts = append(sts, StateNew(bv, now))
	}
	return
}
