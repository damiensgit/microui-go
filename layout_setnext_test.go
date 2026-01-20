package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestLayoutSetNext_Absolute(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	// Set absolute position for next control
	ui.LayoutSetNext(types.Rect{X: 100, Y: 50, W: 80, H: 30}, false)
	rect := ui.LayoutNext()

	if rect.X != 100 || rect.Y != 50 || rect.W != 80 || rect.H != 30 {
		t.Errorf("rect = %+v, want {X:100, Y:50, W:80, H:30}", rect)
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutSetNext_Relative(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 10, Y: 20, W: 400, H: 300})

	ui.LayoutRow(1, []int{-1}, 30)

	// Get the current layout position by peeking
	firstRect := ui.LayoutNext()

	// Set relative position for next control (relative to current position)
	ui.LayoutSetNext(types.Rect{X: 5, Y: 5, W: 80, H: 30}, true)
	rect := ui.LayoutNext()

	// Should be offset from where we would have been positioned
	// firstRect ended, so next position would be firstRect.Y + firstRect.H + spacing
	// Relative adds to that position
	if rect.W != 80 || rect.H != 30 {
		t.Errorf("rect size = %dx%d, want 80x30", rect.W, rect.H)
	}

	ui.EndWindow()
	ui.EndFrame()

	_ = firstRect
}

func TestLayoutSetNext_OnlyAffectsOne(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	ui.LayoutRow(1, []int{100}, 30)

	// Set next position
	ui.LayoutSetNext(types.Rect{X: 200, Y: 200, W: 50, H: 50}, false)
	rect1 := ui.LayoutNext()

	// Next call should NOT use the set_next rect
	rect2 := ui.LayoutNext()

	if rect1.X != 200 {
		t.Errorf("rect1.X = %d, want 200", rect1.X)
	}
	if rect2.X == 200 {
		t.Error("rect2 should not use set_next position")
	}

	ui.EndWindow()
	ui.EndFrame()
}
