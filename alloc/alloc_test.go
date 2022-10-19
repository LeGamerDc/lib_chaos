package alloc

import (
	"testing"
)

type Data struct {
	I1, I2, I3, I4 int
	F1, F2, F3, F4 float64
	S1, S2, S3, S4 []int
}

func feed(s *[]int, n int) {
	for i := 0; i < n; i++ {
		*s = append(*s, i+1)
	}
}

func newData() *Data {
	d := &Data{
		I1: 1,
		I2: 2,
		I3: 3,
		I4: 4,
		F1: 1,
		F2: 2,
		F3: 3,
		F4: 4,
		S1: make([]int, 0, 4),
		S2: make([]int, 0, 8),
		S3: make([]int, 0, 12),
		S4: make([]int, 0, 16),
	}
	feed(&d.S1, 4)
	feed(&d.S2, 8)
	feed(&d.S3, 12)
	feed(&d.S4, 16)
	return d
}

func newData2(buf *Buf) *Data {
	d := Malloc[Data](buf)
	d.I1, d.I2, d.I3, d.I4 = 1, 2, 3, 4
	d.F1, d.F2, d.F3, d.F4 = 1, 2, 3, 4
	d.S1 = MallocSlice[int](buf, 0, 4)
	d.S2 = MallocSlice[int](buf, 0, 8)
	d.S3 = MallocSlice[int](buf, 0, 12)
	d.S4 = MallocSlice[int](buf, 0, 16)
	feed(&d.S1, 4)
	feed(&d.S2, 8)
	feed(&d.S3, 12)
	feed(&d.S4, 16)
	return d
}

type House struct {
	Addr string
	Area int
}

type Person struct {
	Name string
	Age  int
	Home *House
	Data *Data
}

func (p *Person) call() {}

type Escape interface {
	call()
}

// use noinline to escape x
//go:noinline
func Use(x Escape) {
	x.call()
}

func TestAlloc(t *testing.T) {
	var a = new(Allocator)
	a.Init()
	for i := 0; i < 10; i++ {
		m := a.CreateMsg(func(buf *Buf) interface{} {
			p := Malloc[Person](buf)
			p.Name = "john"
			p.Age = 18
			p.Home = Malloc[House](buf)
			p.Home.Addr = "aaeaeaesfe"
			p.Home.Area = 169
			p.Data = newData2(buf)
			return p
		})
		Use(m.msg.(Escape))
		//fmt.Println(a.cp.off, a.cp.cnt)
		_ = m.Close()
	}
}

func BenchmarkAlloc(b *testing.B) {
	var a = new(Allocator)
	a.Init()
	for i := 0; i < b.N; i++ {
		m := a.CreateMsg(func(buf *Buf) interface{} {
			p := Malloc[Person](buf)
			p.Name = "john"
			p.Age = 18
			p.Home = Malloc[House](buf)
			p.Home.Addr = "aaeaeaesfe"
			p.Home.Area = 169
			p.Data = newData2(buf)
			return p
		})
		Use(m.msg.(Escape))
		_ = m.Close()
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := &Person{
			Name: "john",
			Age:  18,
			Home: &House{
				Addr: "aaeaeaesfe",
				Area: 169,
			},
			Data: newData(),
		}
		Use(m)
	}
}
