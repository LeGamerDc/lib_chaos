package skill_system

type Effect interface {
	ActOn(mass *T, source, target int64)
}

type HealEffect struct {
	HealAmount  float64
	HealPercent float64
}

func (e *HealEffect) ActOn(mass *T, source, target int64) {}

type DamageEffect struct {
	DamageAmount  float64
	DamagePercent float64
}

func (e *DamageEffect) ActOn(mass *T, source, target int64) {}

type BuffEffect struct {
	BuffId int32
	Last   int32
}

func (e *BuffEffect) ActOn(mass *T, source, target int64) {}
