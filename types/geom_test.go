package types

import "testing"

func TestVec2_Add(t *testing.T) {
	a := Vec2{X: 1, Y: 2}
	b := Vec2{X: 3, Y: 4}
	got := a.Add(b)
	want := Vec2{X: 4, Y: 6}
	if got != want {
		t.Errorf("Add() = %v, want %v", got, want)
	}
}

func TestVec2_Sub(t *testing.T) {
	a := Vec2{X: 5, Y: 7}
	b := Vec2{X: 2, Y: 3}
	got := a.Sub(b)
	want := Vec2{X: 3, Y: 4}
	if got != want {
		t.Errorf("Sub() = %v, want %v", got, want)
	}
}

func TestRect_Contains(t *testing.T) {
	r := Rect{X: 10, Y: 10, W: 100, H: 50}

	tests := []struct {
		name string
		p    Vec2
		want bool
	}{
		{"inside", Vec2{X: 50, Y: 30}, true},
		{"at corner", Vec2{X: 10, Y: 10}, true},
		{"outside left", Vec2{X: 5, Y: 30}, false},
		{"outside right", Vec2{X: 115, Y: 30}, false},
		{"outside top", Vec2{X: 50, Y: 5}, false},
		{"outside bottom", Vec2{X: 50, Y: 65}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.Contains(tt.p)
			if got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Empty(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want bool
	}{
		{"zero width", Rect{X: 10, Y: 10, W: 0, H: 50}, true},
		{"zero height", Rect{X: 10, Y: 10, W: 100, H: 0}, true},
		{"negative width", Rect{X: 10, Y: 10, W: -10, H: 50}, true},
		{"negative height", Rect{X: 10, Y: 10, W: 100, H: -5}, true},
		{"valid rect", Rect{X: 10, Y: 10, W: 100, H: 50}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Empty()
			if got != tt.want {
				t.Errorf("Empty() = %v, want %v", got, tt.want)
			}
		})
	}
}
