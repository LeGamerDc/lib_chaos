package f14

type K32 interface {
    ~uint32 | ~int32 | ~float32
}

type K64 interface {
    ~uint64 | ~int64 | ~float64
}

type Map32[K K32, V any] struct {
}

type Map64[K K64, V any] struct {
}
