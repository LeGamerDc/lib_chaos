package skill_system

import (
	"testing"
)

func insert(h *HashLink, n int, k K, v V) {
	for i := 0; i < n; i++ {
		h.Insert(k, v+V(i))
	}
}
func cnt(h *HashLink, k K) int {
	var n = 0
	h.RemoveMatch(k, func(v V) bool {
		n++
		return false
	})
	return n
}
func del(h *HashLink, k K) {
	h.RemoveMatch(k, func(v V) bool {
		if v%2 == 0 {
			return true
		}
		return false
	})
}

func TestHashLink_Insert(t *testing.T) {
	var h = NewHashLink(1000)
	insert(h, 5, 1000, 1)
	if h.Remain() != 995 {
		t.Error("e4")
	}
	insert(h, 6, 1001, 1)
	if h.Remain() != 989 {
		t.Error("e5")
	}
	insert(h, 7, 1002, 1)
	if h.Remain() != 982 {
		t.Error("e6")
	}
	insert(h, 8, 1001, 1)
	if h.Remain() != 974 {
		t.Error("e6")
	}
	if cnt(h, 1000) != 5 {
		t.Error("e1")
	}
	if cnt(h, 1001) != 14 {
		t.Error("e2")
	}
	if cnt(h, 1002) != 7 {
		t.Error("e3")
	}
}

func TestHashLink_Remove(t *testing.T) {
	var h = NewHashLink(1000)
	insert(h, 5, 1000, 1)
	del(h, 1000)
	if cnt(h, 1000) != 3 {
		t.Error("e1")
	}
	if h.Remain() != 997 {
		t.Error("e6")
	}
	del(h, 1000)
	del(h, 1001)
	if cnt(h, 1000) != 3 {
		t.Error("e2")
	}
	if cnt(h, 1001) != 0 {
		t.Error("e3")
	}
	insert(h, 6, 1000, 10)
	if cnt(h, 1000) != 9 {
		t.Error("e4")
	}
	del(h, 1000)
	if cnt(h, 1000) != 6 {
		t.Error("e5")
	}
	if h.Remain() != 994 {
		t.Error("e7")
	}
}
