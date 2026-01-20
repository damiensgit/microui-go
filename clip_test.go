package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestGetClipRect(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Push a clip rect
	expected := types.Rect{X: 10, Y: 20, W: 100, H: 80}
	ui.PushClip(expected)

	got := ui.GetClipRect()
	if got != expected {
		t.Errorf("GetClipRect() = %+v, want %+v", got, expected)
	}

	ui.PopClip()
	ui.EndFrame()
}

func TestGetClipRect_Empty(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// No clip pushed - should return unclipped rect (large rect)
	got := ui.GetClipRect()
	if got.W < 1000 {
		t.Errorf("GetClipRect() with no clip should return large rect, got %+v", got)
	}

	ui.EndFrame()
}

func TestCheckClip(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Set up clip rect
	ui.PushClip(types.Rect{X: 100, Y: 100, W: 200, H: 200})

	// Fully inside - not clipped
	result := ui.CheckClip(types.Rect{X: 150, Y: 150, W: 50, H: 50})
	if result != ClipNone {
		t.Errorf("Rect inside clip should return ClipNone, got %d", result)
	}

	// Fully outside - completely clipped
	result = ui.CheckClip(types.Rect{X: 0, Y: 0, W: 50, H: 50})
	if result != ClipAll {
		t.Errorf("Rect outside clip should return ClipAll, got %d", result)
	}

	// Partially inside - partially clipped
	result = ui.CheckClip(types.Rect{X: 50, Y: 150, W: 100, H: 50})
	if result != ClipPart {
		t.Errorf("Rect crossing clip should return ClipPart, got %d", result)
	}

	ui.PopClip()
	ui.EndFrame()
}
