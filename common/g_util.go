package common

import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func MaxN[T constraints.Ordered](xs ...T) T {
	return GetMax(xs)
}

func MinN[T constraints.Ordered](xs ...T) T {
	return GetMin(xs)
}

func GetMax[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		panic("empty param")
	}
	var m = xs[0]
	for _, x := range xs[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

func GetMin[T constraints.Ordered](xs []T) T {
	if len(xs) == 0 {
		panic("empty param")
	}
	var m = xs[0]
	for _, x := range xs[1:] {
		if x < m {
			m = x
		}
	}
	return m
}

func Reverse[T any](xs []T) {
	var l = len(xs)
	for i := 0; i < l/2; i++ {
		j := l - 1 - i
		xs[i], xs[j] = xs[j], xs[i]
	}
}

func EraseOnce[T comparable](xs []T, x T) []T {
	for i, v := range xs {
		if v == x {
			return append(xs[:i], xs[i+1:]...)
		}
	}
	return xs
}

// Index return array index if found, else return -1
func Index[T comparable](xs []T, y T) int {
	for i, x := range xs {
		if x == y {
			return i
		}
	}
	return -1
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	var a = make([]K, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	return a
}

func ToSet[K comparable](a []K) map[K]struct{} {
	var m = make(map[K]struct{}, len(a))
	for _, k := range a {
		m[k] = struct{}{}
	}
	return m
}

func Contains[T comparable](a, b []T) bool {
	var m = ToSet(a)
	for _, x := range b {
		if _, ok := m[x]; !ok {
			return false
		}
	}
	return true
}

func Union[K comparable](a, b map[K]struct{}) map[K]struct{} {
	var c = make(map[K]struct{}, len(a)+len(b))
	for k := range a {
		c[k] = struct{}{}
	}
	for k := range b {
		c[k] = struct{}{}
	}
	return c
}

func Intersection[K comparable](a, b map[K]struct{}) map[K]struct{} {
	var c = make(map[K]struct{})
	for k := range a {
		if _, ok := b[k]; ok {
			c[k] = struct{}{}
		}
	}
	return c
}

func ToSlice[K comparable](a map[K]struct{}) []K {
	var s = make([]K, 0, len(a))
	for k := range a {
		s = append(s, k)
	}
	return s
}
