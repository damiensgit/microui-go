package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

// Note: Window title bar is 24 pixels high by default.
// A window at {X:100, Y:50} has content area starting at Y=74 (50+24).

func TestMouseOver_TrueWhenInRectAndHoverRoot(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	// Mouse at (150, 100) - inside window content area (below title bar at Y=74)
	ui.MouseMove(150, 100)

	ui.BeginWindow("Test", types.Rect{X: 100, Y: 50, W: 200, H: 150})

	// Rect in content area (Y >= 74)
	rect := types.Rect{X: 110, Y: 80, W: 100, H: 50}
	if !ui.MouseOver(rect) {
		t.Error("MouseOver should return true when mouse is in rect and in hover root")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestMouseOver_FalseWhenOutsideRect(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.MouseMove(50, 50) // Outside window entirely

	ui.BeginWindow("Test", types.Rect{X: 100, Y: 50, W: 200, H: 150})

	rect := types.Rect{X: 110, Y: 80, W: 100, H: 50}
	if ui.MouseOver(rect) {
		t.Error("MouseOver should return false when mouse is outside rect")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestMouseOver_FalseWhenNotHoverRoot(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Establish hover root on front window
	ui.BeginFrame()
	// Mouse at (200, 120) - in overlap area of both windows
	ui.MouseMove(200, 120)

	// Back window at Y=50, content starts at Y=74
	ui.BeginWindow("Back", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()

	// Front window at Y=70, content starts at Y=94
	ui.BeginWindow("Front", types.Rect{X: 150, Y: 70, W: 200, H: 150})
	ui.EndWindow()

	ui.EndFrame()

	// Frame 2: Check MouseOver in back window
	ui.BeginFrame()
	ui.MouseMove(200, 120) // Still in overlap area

	// Back window - should NOT be hover root
	ui.BeginWindow("Back", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	// This rect is in the back window's content area
	rect := types.Rect{X: 110, Y: 80, W: 150, H: 80}
	// Mouse is in rect, but back window is not hover root
	if ui.MouseOver(rect) {
		t.Error("MouseOver should return false for non-hover-root window")
	}
	ui.EndWindow()

	// Front window - should be hover root
	ui.BeginWindow("Front", types.Rect{X: 150, Y: 70, W: 200, H: 150})
	// This rect is in the front window's content area (Y >= 94)
	rect2 := types.Rect{X: 160, Y: 100, W: 100, H: 50}
	if !ui.MouseOver(rect2) {
		t.Error("MouseOver should return true for hover root window")
	}
	ui.EndWindow()

	ui.EndFrame()
}

func TestMouseOver_RespectsClipRect(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.MouseMove(250, 120) // Outside clip but inside rect

	ui.BeginWindow("Test", types.Rect{X: 100, Y: 50, W: 200, H: 150})

	// Push a smaller clip rect (within the content area)
	ui.PushClip(types.Rect{X: 100, Y: 74, W: 100, H: 100})

	// Rect extends beyond clip
	rect := types.Rect{X: 200, Y: 100, W: 100, H: 50}
	// Mouse at (250, 120) is in rect but outside clip (clip X ends at 200)
	if ui.MouseOver(rect) {
		t.Error("MouseOver should return false when mouse is outside clip rect")
	}

	ui.PopClip()
	ui.EndWindow()
	ui.EndFrame()
}
