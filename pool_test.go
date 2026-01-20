package microui

import (
	"testing"
)

func TestGrowPool_Fixed(t *testing.T) {
	p := growPool[int]{}
	p.Init(4, 100)

	// Allocate from fixed pool
	items := []*int{}
	for i := 0; i < 4; i++ {
		item := p.Alloc()
		items = append(items, item)
	}

	if len(p.grow) != 0 {
		t.Errorf("Alloc() used grow pool, want fixed pool only")
	}
}

func TestGrowPool_Growth(t *testing.T) {
	p := growPool[struct{}]{}
	p.Init(2, 10)

	// Exhaust fixed pool
	for i := 0; i < 2; i++ {
		p.Alloc()
	}

	// Should grow
	p.Alloc()

	if len(p.grow) < 1 {
		t.Errorf("After growth, len(p.grow) = %d, want >= 1", len(p.grow))
	}
}

func TestGrowPool_Exhausted(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Alloc() should panic when pool exhausted")
		}
	}()

	p := growPool[struct{}]{}
	p.Init(2, 2)

	for i := 0; i < 3; i++ {
		p.Alloc()
	}
}

func TestGrowStack(t *testing.T) {
	s := growStack[int]{}
	s.Init(4)

	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Len() != 3 {
		t.Errorf("Len() = %d, want 3", s.Len())
	}

	if s.Peek() != 3 {
		t.Errorf("Peek() = %d, want 3", s.Peek())
	}

	v := s.Pop()
	if v != 3 {
		t.Errorf("Pop() = %d, want 3", v)
	}

	if s.Len() != 2 {
		t.Errorf("After Pop, Len() = %d, want 2", s.Len())
	}
}
