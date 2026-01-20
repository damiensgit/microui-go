package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestNumber_ShiftClickEntersTextbox(t *testing.T) {
	ui := New(Config{})
	val := 42.0

	// Frame 1: Hover first (hover root needs to be set)
	ui.MouseMove(50, 35)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Shift+click to enter textbox edit mode
	ui.KeyDown(KeyShift)
	ui.MouseDown(50, 35, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Verify we're in textbox edit mode
	if ui.numberTextboxID == 0 {
		t.Error("Shift+click should enter textbox edit mode")
	}
}

func TestNumber_TextboxEditChangesValue(t *testing.T) {
	ui := New(Config{})
	val := 42.0

	// Frame 1: Hover first
	ui.MouseMove(50, 35)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Shift+click to enter edit mode
	ui.KeyDown(KeyShift)
	ui.MouseDown(50, 35, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Clear buffer and type new value
	ui.KeyUp(KeyShift)
	// Simulate selecting all and typing new value
	ui.numberTextboxBuf = []byte("100")
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 4: Press Enter to confirm
	ui.KeyDown(KeyEnter)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	if val != 100.0 {
		t.Errorf("val = %f, want 100.0 after textbox edit", val)
	}
}

func TestNumber_NormalClickStillDrags(t *testing.T) {
	ui := New(Config{})
	val := 50.0

	// Frame 1: Hover over control first (mouse NOT down)
	ui.MouseMove(50, 35)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Normal click (no shift) to gain focus
	ui.MouseDown(50, 35, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Should NOT be in textbox mode
	if ui.numberTextboxID != 0 {
		t.Error("Normal click should not enter textbox mode")
	}

	initialVal := val

	// Frame 3: Drag right to increase value
	ui.MouseMove(100, 35) // Move 50 pixels right
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Value should have changed via drag
	if val == initialVal {
		t.Error("Dragging should change value")
	}
}

func TestNumber_ExitTextboxOnFocusLoss(t *testing.T) {
	ui := New(Config{})
	val := 42.0

	// Frame 1: Hover first
	ui.MouseMove(50, 35)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Shift+click to enter edit mode
	ui.KeyDown(KeyShift)
	ui.MouseDown(50, 35, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Verify we're in textbox edit mode
	if ui.numberTextboxID == 0 {
		t.Fatal("Expected to be in textbox edit mode")
	}

	// Frame 3: Release mouse to clear mouse pressed state
	ui.KeyUp(KeyShift)
	ui.MouseUp(50, 35, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 4: Click elsewhere to lose focus (outside the control area)
	ui.MouseMove(300, 200) // Different location
	ui.MouseDown(300, 200, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Number(&val, 1.0)
	ui.EndWindow()
	ui.EndFrame()

	// Should have exited textbox mode
	if ui.numberTextboxID != 0 {
		t.Error("Should exit textbox mode when focus is lost")
	}
}

func TestNumber_TextboxPreservesFormat(t *testing.T) {
	ui := New(Config{})
	val := 3.14159

	// Frame 1: Hover first
	ui.MouseMove(50, 35)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.NumberOpt(&val, 0.01, "%.4f", 0)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Shift+click to enter edit mode with custom format
	ui.KeyDown(KeyShift)
	ui.MouseDown(50, 35, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.NumberOpt(&val, 0.01, "%.4f", 0)
	ui.EndWindow()
	ui.EndFrame()

	// Verify buffer was initialized with formatted value
	expected := "3.1416" // "%.4f" rounds to 4 decimal places
	got := string(ui.numberTextboxBuf)
	if got != expected {
		t.Errorf("Buffer = %q, want %q", got, expected)
	}
}
