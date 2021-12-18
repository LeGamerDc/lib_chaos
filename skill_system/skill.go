package skill_system

type Skill interface {
	Cast(mass *T, ctx SkillManager, source, target int64)
}

type DelayedSkill struct {
	Delay int32
	Skill Skill
}

func (s *DelayedSkill) Cast(mass *T, ctx SkillManager, source, target int64) {
	ctx.Insert(s.Delay, func(skillCtx SkillManager, mass *T) {
		s.Skill.Cast(mass, skillCtx, source, target)
	}, source, 0)
}

type TargetSkill struct {
	Selector TargetSelector
	Effect   Effect
}

func (s *TargetSkill) Cast(mass *T, ctx SkillManager, source, target int64) {
	var ts = s.Selector.Select(mass, source, target)
	for _, t := range ts {
		s.Effect.ActOn(mass, source, t)
	}
}

type ArrowRainLikeSkill struct {
	Delay         int32
	PosSelector   PosSelector
	RangeSelector LandSelector
	Effect        Effect
}

func (s *ArrowRainLikeSkill) Cast(mass *T, ctx SkillManager, source, target int64) {
	var ts = s.PosSelector.Select(mass, source, target)
	if len(ts) == 0 {
		return
	}
	var pos = ts[0]
	ctx.Insert(s.Delay, func(skillCtx SkillManager, mass *T) {
		var ts = s.RangeSelector.Select(mass, pos)
		for _, t := range ts {
			s.Effect.ActOn(mass, source, t)
		}
	}, source, 0)
}

type ChainLikeSkill struct {
	Max      int32
	Gap      int32
	Decay    float64
	Selector ExcludeSelector
	Effect   Effect
}

func (s *ChainLikeSkill) Cast(mass *T, ctx SkillManager, source, target int64) {
	var f CastFunc
	var remain = s.Max
	var visited = make([]int64, 0, s.Max)
	f = func(skillCtx SkillManager, mass *T) {
		if remain <= 0 {
			return
		}
		remain--
		var ts = s.Selector.Select(mass, source, target, visited)
		if len(ts) == 0 {
			return
		}
		var t = ts[0]
		s.Effect.ActOn(mass, source, t)
		ctx.Insert(s.Gap, f, source, 0)
	}

	ctx.Insert(s.Gap, f, source, 0)
}
