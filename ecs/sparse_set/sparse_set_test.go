package sparse_set

import (
    "fmt"
    "sync/atomic"
    "testing"
    "time"
)

type T struct {
    V uint64
}

func TestArray(t *testing.T) {
    var a = MakeArray[T](1024, 0)
    var i1 = a.Set(101, T{V: 3})
    var i2 = a.Set(102, T{V: 4})
    fmt.Println(a.Free())
    fmt.Println(a.first)
    fmt.Println(a.Get(i1))
    a.Remove(i1)
    fmt.Println(a.first)
    a.Remove(i2)
    fmt.Println(a.first)
    fmt.Println(a.Get(i1))
    fmt.Println(a.Free())
    a.Set(101, T{V: 3})
    a.Set(101, T{V: 3})
    a.Set(101, T{V: 3})
    fmt.Println(a.first)
    fmt.Println(a.Get(i1))
    fmt.Println(a.Free())
}

func TestSparse(t *testing.T) {
    var a = MakeSparseArray[T](8)
    a.grow()
    a.grow()
    //fmt.Printf("%d %v\n", len(a.chucks), a.freeBits)
    var ids []int64
    var s uint64
    for i := 0; i < 10000; i++ {
        id := a.Set(23, 3, T{V: uint64(i)})
        s += uint64(i)
        if IdTag(id) != 3 {
            t.Error("id tag is wrong")
        }
        ids = append(ids, id)
    }

    var sum uint64
    a.ParForeach(func(t *T) {
        atomic.AddUint64(&sum, t.V)
    }, 24)
    fmt.Println(sum, s)
    for _, id := range ids {
        a.Remove(id)
    }
    fmt.Printf("%d %v\n", len(a.chucks), a.freeBits)
}

func TestParallel(t *testing.T) {
    var a = MakeSparseArray[T](1024)
    for i := 0; i < 10000000; i++ {
        a.Set(23, 3, T{V: uint64(i)})
    }
    var s1 uint64
    timer(func() {
        a.Foreach(func(t *T) {
            //s1 += t.V
            t.V++
        })
    })
    var s2 uint64
    timer(func() {
        a.ParForeach(func(t *T) {
            t.V++
            //s2 += t.V
            //atomic.AddUint64(&s2, t.V)
        }, 2)
    })
    fmt.Println(s1, s2)

}

func timer(f func()) {
    var s = time.Now()
    f()
    fmt.Println(time.Since(s))
}
