package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestLayoutRow(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Create a window first to have a layout context
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(2, []int{100, -1}, 30)

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutNext(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(2, []int{100, 200}, 30)

	rect1 := ui.LayoutNext()
	rect2 := ui.LayoutNext()

	// Second control should be to the right of first
	if rect2.X <= rect1.X {
		t.Errorf("Second layout rect X=%d should be > first X=%d", rect2.X, rect1.X)
	}

	// Both should have same Y
	if rect2.Y != rect1.Y {
		t.Errorf("Same row: rect2.Y=%d should equal rect1.Y=%d", rect2.Y, rect1.Y)
	}

	// Widths should match what we specified
	if rect1.W != 100 {
		t.Errorf("rect1.W=%d, want 100", rect1.W)
	}
	if rect2.W != 200 {
		t.Errorf("rect2.W=%d, want 200", rect2.W)
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutRowWrapping(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(2, []int{100, 100}, 30)

	rect1 := ui.LayoutNext()
	rect2 := ui.LayoutNext()
	// After 2 items, should wrap to next row
	rect3 := ui.LayoutNext()

	// rect3 should be on a new row (Y > rect1.Y)
	if rect3.Y <= rect1.Y {
		t.Errorf("rect3.Y=%d should be > rect1.Y=%d (new row)", rect3.Y, rect1.Y)
	}

	// rect3 should be back at left edge
	if rect3.X != rect1.X {
		t.Errorf("rect3.X=%d should equal rect1.X=%d (new row)", rect3.X, rect1.X)
	}

	_ = rect2

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutFillWidth(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(2, []int{100, -1}, 30)

	rect1 := ui.LayoutNext()
	rect2 := ui.LayoutNext()

	// rect2 should fill remaining space
	// Available = window width - padding - spacing
	if rect2.W <= 0 {
		t.Errorf("Fill width rect2.W=%d should be > 0", rect2.W)
	}

	// rect2 width + rect1 width + spacing should use available space
	_ = rect1
	_ = rect2

	ui.EndWindow()
	ui.EndFrame()
}
