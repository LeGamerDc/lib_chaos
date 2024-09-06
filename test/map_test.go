package main

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"testing"
)

const (
	N = 10000
	M = 3000
	X = M * N
)

type Unit struct {
	Name string
	Age  int
	Next *Unit
}

func get() []*Unit {
	us := make([]*Unit, X)
	for i := 0; i < X; i++ {
		us[i] = &Unit{
			Name: "name_" + strconv.Itoa(i),
			Age:  rand.Int(),
		}
	}
	for i := 0; i < X-M; i++ {
		us[i].Next = us[i+M]
	}
	return us
}

func getMap() map[int]*Unit {
	us := get()
	m := make(map[int]*Unit, M)
	for i := 0; i < M; i++ {
		m[i] = us[i]
	}
	return m
}

func getSlice() []*Unit {
	us := get()
	s := make([]*Unit, M)
	for i := 0; i < M; i++ {
		s[i] = us[i]
	}
	return s
}

func BenchmarkGcMap(b *testing.B) {
	b.StopTimer()
	m := getMap()
	runtime.GC()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		runtime.GC()
	}
	runtime.KeepAlive(m)
}

func BenchmarkGcSlice(b *testing.B) {
	b.StopTimer()
	m := getSlice()
	runtime.GC()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		runtime.GC()
	}
	runtime.KeepAlive(m)
}

func BenchmarkMapNoGc(b *testing.B) {
	b.StopTimer()
	m := getMap()
	runtime.GC()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		x := m[i%M]
		x.Age = i
	}
	runtime.KeepAlive(m)
}

func BenchmarkMap(b *testing.B) {
	b.StopTimer()
	m := getMap()
	runtime.GC()
	var (
		wg sync.WaitGroup
		ch = make(chan struct{})
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ch:
				return
			default:
				runtime.GC()
			}
		}
	}()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		x := m[i%M]
		x.Age = i
	}
	b.StopTimer()
	runtime.KeepAlive(m)
	close(ch)
	wg.Wait()
}

func BenchmarkSlice(b *testing.B) {
	b.StopTimer()
	m := getSlice()
	runtime.GC()
	var (
		wg sync.WaitGroup
		ch = make(chan struct{})
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ch:
				return
			default:
				runtime.GC()
			}
		}
	}()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = m[i%M]
	}
	b.StopTimer()
	runtime.KeepAlive(m)
	close(ch)
	wg.Wait()
}
