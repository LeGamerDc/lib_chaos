package detour

var (
	empty = node{}
)

type node struct {
	a, b, x int32
	next    int32
}

type JumpCache struct {
	cache              []node
	first              []int32
	maxNodes, hashSize int32
	cur                int32
}

func (j *JumpCache) findPtr(a, b int32) (ptr *int32, ok bool) {
	var p = &j.first[hash64(int64(a)<<32+int64(b))&(j.hashSize-1)]
	for *p != -1 {
		if j.cache[*p].a == a && j.cache[*p].b == b {
			return p, true
		}
		p = &j.cache[*p].next
	}
	return nil, false
}

func (j *JumpCache) delete(i int32) {
	if j.cache[i] == empty {
		return
	}
	ptr, ok := j.findPtr(j.cache[i].a, j.cache[i].b)
	if !ok {
		return
	}
	*ptr = j.cache[*ptr].next
}

func (j *JumpCache) Insert(a, b, x int32) {
	var key = hash64(int64(a)<<32+int64(b)) & (j.hashSize - 1)
	j.delete(j.cur)
	j.cache[j.cur] = node{
		a:    a,
		b:    b,
		x:    x,
		next: j.first[key],
	}
	j.first[key] = j.cur
	j.cur = (j.cur + 1) % j.maxNodes
}

func (j *JumpCache) Find(a, b int32) (x int32, ok bool) {
	var p = j.first[hash64(int64(a)<<32+int64(b))&(j.hashSize-1)]
	for p != -1 {
		if j.cache[p].a == a && j.cache[p].b == b {
			return j.cache[p].x, true
		}
		p = j.cache[p].next
	}
	return 0, false
}

func newJumpCache(maxNodes int32) (j *JumpCache) {
	j = new(JumpCache)
	j.maxNodes = maxNodes
	j.hashSize = int32(nextPow2(uint32(maxNodes / 4)))
	j.cache = make([]node, j.maxNodes)
	j.first = make([]int32, j.hashSize)
	memset(j.first, -1)
	return
}
