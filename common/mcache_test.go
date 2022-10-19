package common

import (
	"fmt"
	"testing"
)

func TestTable(t *testing.T) {
	for i := 0; i <= 4096; i++ {
		var c = sizeClass(i)
		if class[c] < i {
			t.Errorf("error for %d, class: %d, size: %d", i, c, class[c])
		}
	}
	fmt.Println("test table finish")
}

func BenchmarkSwitch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := 0; j < largeSizeMax; j++ {
			_ = support(j)
		}
	}
}

func BenchmarkMap(b *testing.B) {
	var m = make(map[int]int, len(class))
	for i, v := range class {
		m[v] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < largeSizeMax; j++ {
			_ = m[j]
		}
	}
}
