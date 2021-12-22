package common

import "constraints"

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
