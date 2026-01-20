package ebiten

import (
	"testing"
)

func TestNewRenderer(t *testing.T) {
	r := NewRenderer()

	if r == nil {
		t.Fatal("NewRenderer() returned nil")
	}
}

func TestRenderer_SetTarget(t *testing.T) {
	r := NewRenderer()

	// nil target should not panic
	r.SetTarget(nil)
}
