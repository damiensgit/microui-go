package microui

import (
	"testing"
)

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
