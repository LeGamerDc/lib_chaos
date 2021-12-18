package skill_system

import "github.com/fogleman/ln/ln"

type SkillAttr uint32

const (
	SkillAttrInterrupt = SkillAttr(1 << iota)
	SkillAttrChannel
	SkillAttrBuff
	SkillAttrDeBuff
	SkillAttrFromAlly
	SkillAttrFromEnemy
)

type T struct{}

func (t *T) Pos(charId int64) ln.Vector { return ln.Vector{} }

type CastFunc func(skillCtx SkillManager, mass *T)

type SkillManager interface {
	Insert(when int32, f CastFunc, charId int64, attr SkillAttr)
}
type node struct {
	charId  int64
	attr    SkillAttr
	next    int32
	cast    CastFunc
	deleted bool // 该技能已经被删除，不会再触发
}

type SkillWheel struct {
	filter *HashLink
	slots  []node
	free   []int32

	cycle            [60]int32
	curTick, maxTick int32
}

func NewSkillWheel(n int, m int32) *SkillWheel {
	var s = &SkillWheel{
		filter:  NewHashLink(n),
		slots:   make([]node, n),
		free:    make([]int32, n),
		maxTick: m,
	}
	for i := 0; i < n; i++ {
		s.free[i] = int32(n - 1 - i)
	}
	return s
}

// 根据charId和技能属性过滤掉wheel中的技能并删除
func (s *SkillWheel) Filter(charId int64, bitMask SkillAttr) {
	s.filter.RemoveMatch(charId, func(v V) bool {
		var node = &s.slots[v]
		if node.charId == charId && (node.attr&bitMask) == bitMask {
			node.deleted = true
			node.cast = nil
			return true
		}
		return false
	})
}

func (s *SkillWheel) Tick(mass *T) {
	s.tick(mass, s.curTick)
	s.curTick = (s.curTick + 1) % s.maxTick
}
func (s *SkillWheel) tick(mass *T, t int32) {
	var n = s.cycle[t]
	for n != -1 {
		var node = &s.slots[n]
		if !node.deleted {
			node.cast(s, mass)
			if s.filter != nil {
				s.filter.Remove(node.charId, n)
			}
		}
		node.cast = nil
		s.free = append(s.free, n)
		n = node.next
	}
	s.cycle[t] = -1
}
func (s *SkillWheel) getNode() int32 {
	var l = len(s.free) - 1
	if l < 0 {
		return -1
	}
	var n = s.free[l]
	s.free = s.free[:l]
	return n
}
func (s *SkillWheel) Insert(when int32, f CastFunc, charId int64, attr SkillAttr) {
	if when == 0 {
		return
	}
	var (
		idx = (s.curTick + when) % s.maxTick
		n   = s.getNode()
	)
	if n == -1 {
		return
	}
	s.slots[n] = node{
		charId: charId,
		attr:   attr,
		next:   s.cycle[idx],
		cast:   f,
	}
	if s.filter != nil {
		s.filter.Insert(charId, n)
	}
	s.cycle[idx] = n
}
