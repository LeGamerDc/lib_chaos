package bt

func NewWait(secs BvTime) BvNode {
	return BvNode(&BvNodeWait{
		Head: BvWait,
		Secs: secs,
	})
}

func NewWaitForever() BvNode {
	return BvNode(&BvNodeWaitForever{
		Head: BvWaitForever,
	})
}

func NewAction(fn ActionFn) BvNode {
	return BvNode(&BvNodeAction{
		Head: BvAction,
		fn:   fn,
	})
}

func NewFail(inner BvNode) BvNode {
	return BvNode(&BvNodeFail{
		Head:  BvFail,
		inner: inner,
	})
}

func NewSuccess(inner BvNode) BvNode {
	return BvNode(&BvNodeAlwaysSucceed{
		Head:  BvAlwaysSucceed,
		inner: inner,
	})
}

func NewSelect(todo ...BvNode) BvNode {
	return BvNode(&BvNodeSelect{
		Head: BvSelect,
		todo: todo,
	})
}

func NewSequence(todo ...BvNode) BvNode {
	return BvNode(&BvNodeSequence{
		Head: BvSequence,
		todo: todo,
	})
}

func NewWhenAll(todo ...BvNode) BvNode {
	return BvNode(&BvNodeWhenAll{
		Head: BvWhenAll,
		todo: todo,
	})
}

func NewWhenAny(todo ...BvNode) BvNode {
	return BvNode(&BvNodeWhenAny{
		Head: BvWhenAny,
		todo: todo,
	})
}

func NewWhile(cond BvNode, todo ...BvNode) BvNode {
	return BvNode(&BvNodeWhile{
		Head: BvWhile,
		cond: cond,
		todo: todo,
	})
}

func NewIf(cond, succ, fail BvNode) BvNode {
	return BvNode(&BvNodeIf{
		Head: BvIf,
		cond: cond,
		succ: succ,
		fail: fail,
	})
}
