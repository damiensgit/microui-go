package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestLayoutWidth(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	ui.LayoutRow(1, []int{-1}, 0)

	// Set custom width for next control
	ui.LayoutWidth(150)
	rect := ui.LayoutNext()

	if rect.W != 150 {
		t.Errorf("rect.W = %d, want 150", rect.W)
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutHeight(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	ui.LayoutRow(1, []int{-1}, 0)

	// Set custom height for next control
	ui.LayoutHeight(50)
	rect := ui.LayoutNext()

	if rect.H != 50 {
		t.Errorf("rect.H = %d, want 50", rect.H)
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutWidth_OnlyAffectsOne(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	ui.LayoutRow(1, []int{100}, 30)

	// Set custom width
	ui.LayoutWidth(150)
	rect1 := ui.LayoutNext()
	rect2 := ui.LayoutNext() // Should NOT use 150

	if rect1.W != 150 {
		t.Errorf("rect1.W = %d, want 150", rect1.W)
	}
	if rect2.W == 150 {
		t.Errorf("rect2.W should not be 150, LayoutWidth should only affect one control")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutHeight_OnlyAffectsOne(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	ui.LayoutRow(1, []int{100}, 30)

	// Set custom height
	ui.LayoutHeight(50)
	rect1 := ui.LayoutNext()
	rect2 := ui.LayoutNext() // Should NOT use 50

	if rect1.H != 50 {
		t.Errorf("rect1.H = %d, want 50", rect1.H)
	}
	if rect2.H == 50 {
		t.Errorf("rect2.H should not be 50, LayoutHeight should only affect one control")
	}

	ui.EndWindow()
	ui.EndFrame()
}
