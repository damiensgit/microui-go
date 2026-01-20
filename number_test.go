package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestNumber_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)

	value := 50.0
	changed := ui.Number(&value, 1.0)

	if changed {
		t.Error("Number should not change without input")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestNumber_Drag(t *testing.T) {
	ui := New(Config{})

	value := 50.0

	// Control is at Y ~= 29 (title 24 + padding 5), height ~= 20 (size 10 + padding*2)
	// So Y=35 is inside the control
	const clickY = 35

	// Setup: Establish initial mouse position
	ui.MouseMove(50, clickY)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	ui.Number(&value, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Press mouse to gain focus (hover is set from previous frame)
	ui.MouseDown(50, clickY, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	ui.Number(&value, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag 10 pixels right - MouseMove BEFORE BeginFrame
	ui.MouseMove(60, clickY)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	changed := ui.Number(&value, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Value should have increased (delta.X > 0)
	if !changed || value <= 50.0 {
		t.Errorf("Value should have increased from drag, got %v changed=%v", value, changed)
	}
}

func TestNumber_InstantClickDrag(t *testing.T) {
	// Test clicking and dragging without a separate hover frame first
	// This simulates a user who clicks instantly without hovering
	ui := New(Config{})

	value := 50.0

	// Control is at Y ~= 29, height ~= 20, so Y=35 is inside
	const clickY = 35

	// Setup: establish initial mouse position
	ui.MouseMove(50, clickY)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	ui.Number(&value, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click on control
	ui.MouseDown(50, clickY, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	ui.Number(&value, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag while holding mouse - MouseMove BEFORE BeginFrame
	ui.MouseMove(60, clickY)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	changed := ui.Number(&value, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Value should have increased from drag
	if !changed || value <= 50.0 {
		t.Errorf("Instant click+drag should work, got value=%v changed=%v", value, changed)
	}
}

func TestNumberOpt_Drag(t *testing.T) {
	// Test NumberOpt with alignment option
	ui := New(Config{})

	value := 50.0

	// Control is at Y ~= 29, height ~= 20, so Y=35 is inside
	const clickY = 35

	// Setup: establish initial mouse position
	ui.MouseMove(50, clickY)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	ui.NumberOpt(&value, 1.0, "%.0f", OptAlignRight)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click
	ui.MouseDown(50, clickY, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	ui.NumberOpt(&value, 1.0, "%.0f", OptAlignRight)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag - MouseMove BEFORE BeginFrame
	ui.MouseMove(60, clickY)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)
	changed := ui.NumberOpt(&value, 1.0, "%.0f", OptAlignRight)
	ui.EndWindow()
	ui.EndFrame()

	if !changed || value <= 50.0 {
		t.Errorf("NumberOpt drag should work, got value=%v changed=%v", value, changed)
	}
}
