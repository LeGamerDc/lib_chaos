package skill_system

import "github.com/fogleman/ln/ln"

type TargetSelector interface {
	Select(mass *T, source, target int64) []int64
}

// ----------------------------------------------------------------------------------------------

type PosSelector interface {
	Select(mass *T, source, target int64) []ln.Vector
}

// ----------------------------------------------------------------------------------------------

type LandSelector interface {
	Select(mass *T, pos ln.Vector) []int64
}

// ----------------------------------------------------------------------------------------------

type ExcludeSelector interface {
	Select(mass *T, source, target int64, excludes []int64) []int64
}
