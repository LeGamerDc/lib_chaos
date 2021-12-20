package main

import (
	"fmt"
	"lib_chaos/ecs"
)

type Char struct {
	Name string
	Id   int64
}

func main() {
	var s = ecs.MakeArray[Char](1024, 0)
	s.Set(1, Char{Name: "john", Id: 1234})
	s.Set(2, Char{Name: "lily", Id: 1324})
	s.Set(3, Char{Name: "joe", Id: 1423})
	fmt.Println(s.Size())

}

//func print[T fmt.Stringer](x T) {
//	fmt.Println(x)
//}
