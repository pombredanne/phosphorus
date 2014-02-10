package id

import (
	"reflect"
	"syscall"
	"testing"
)

func BenchmarkSystime(b *testing.B) {
	var tv syscall.Timeval
	for i := 0; i < b.N; i++ {
		timeSys(&tv)
	}
}

func BenchmarkLibtime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		timeLib()
	}
}

func BenchmarkIdGen(b *testing.B) {
	g := NewGenerator(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.SafeId()
	}
}

func TestIds(t *testing.T) {
	g0 := NewGenerator(0)
	i0 := make([]int64, 0, 10)
	for i := 0; i < 10; i++ {
		j, _ := g0.Id()
		i0 = append(i0, j)
	}

	g1 := NewGenerator(255)
	i1 := make([]int64, 0, 10)
	for i := 0; i < 10; i++ {
		j, _ := g1.Id()
		i1 = append(i1, j)
	}

	if reflect.DeepEqual(i0, i1) {
		t.Fail()
	}
}

func TestGet(t *testing.T) { // will roll over
	g := NewGenerator(0)
	for i := 0; i <= 0x4000; i++ {
		_, err := g.Id()
		if err != nil {
			return
		}
	}
	t.Fail()
}

func TestSafeGet(t *testing.T) {
	g := NewGenerator(0)
	for i := 0; i <= 0x4000; i++ {
		_ = g.SafeId()
	}
}

func TestMono(t *testing.T) {
	g := NewGenerator(0)
	g.lastRollover = 1 << 62
	for i := 0; i <= 0x4000; i++ {
		_, err := g.Id()
		if err != nil {
			return
		}
	}
	t.Fail()
}
