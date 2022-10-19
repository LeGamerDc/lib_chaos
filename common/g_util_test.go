package common

import (
    "testing"
)

func TestMax(t *testing.T) {
    var s = []int{3, 5, 7, 9, 2, 4, 6, 8}
    if GetMax(s) != 9 {
        t.Error("not equal")
    }
    if MaxN(1, 3, 2) != 3 {
        t.Error("not equal")
    }
    if Max(1, 2) != 2 {
        t.Error("not equal")
    }
}

func TestMin(t *testing.T) {
    var s = []int{3, 5, 7, 9, 2, 4, 6, 8}
    if GetMin(s) != 2 {
        t.Error("not equal")
    }
    if MinN(1, 3, 2) != 1 {
        t.Error("not equal")
    }
    if Min(1, 2) != 1 {
        t.Error("not equal")
    }
}

func TestMap(t *testing.T) {
    var s = []int{3, 5, 7, 9, 2, 4, 6, 8, 7}
    var m = ToSet(s)
    var s2 = MapKeys(m)
    if !Contains(s, s2) || !Contains(s2, s) {
        t.Error("not equal")
    }
}
