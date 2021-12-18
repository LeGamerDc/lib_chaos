package bt

import (
	"fmt"
	"testing"
)

func TestWhile(t *testing.T) {
	var value = 0
	var fn ActionFn = func() int32 {
		value++
		return Success
	}
	var bv = NewWhile(NewWait(50), NewWait(1), NewAction(fn))
	var st = StateNew(bv, 0)
	fmt.Println(Event(st, 1))
	fmt.Println(Event(st, 3))
	fmt.Println(value)
}
