package utils

import (
	"go.uber.org/zap"
	"lib_chaos/log"
)

func PanicWrap(f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic", zap.Any("err", err), zap.Stack("stack"))
		}
	}()
	f()
}

func PanicWrap1[A any](f func(A), a A) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic", zap.Any("err", err), zap.Stack("stack"))
		}
	}()
	f(a)
}

func PanicWrap2[A, B any](f func(A, B), a A, b B) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic", zap.Any("err", err), zap.Stack("stack"))
		}
	}()
	f(a, b)
}

func PanicWrap3[A, B, C any](f func(A, B, C), a A, b B, c C) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic", zap.Any("err", err), zap.Stack("stack"))
		}
	}()
	f(a, b, c)
}
