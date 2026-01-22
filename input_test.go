package microui

import (
	"testing"
)

func TestUI_MouseMove(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.MouseMove(42, 99)

	if ui.input.MousePos.X != 42 || ui.input.MousePos.Y != 99 {
		t.Errorf("MouseMove() set pos to %v, want (42, 99)", ui.input.MousePos)
	}
}

func TestUI_MouseDown(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.MouseDown(10, 20, MouseLeft)

	if !ui.input.MouseDown[0] {
		t.Error("MouseDown() did not set MouseDown[0]")
	}
}

func TestUI_MouseUp(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.MouseDown(10, 20, MouseLeft)
	ui.MouseUp(10, 20, MouseLeft)

	if ui.input.MouseDown[0] {
		t.Error("MouseUp() did not clear MouseDown[0]")
	}
}

func TestUI_KeyDown(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.KeyDown(KeyEnter)

	if !ui.input.KeyDown[KeyEnter] {
		t.Error("KeyDown() did not set KeyDown[KeyEnter]")
	}
}

func TestUI_KeyUp(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.KeyDown(KeyEnter)
	ui.KeyUp(KeyEnter)

	if ui.input.KeyDown[KeyEnter] {
		t.Error("KeyUp() did not clear KeyDown[KeyEnter]")
	}
}

func TestUI_InputChan(t *testing.T) {
	ui := New(Config{})

	ch := ui.InputChan()
	if ch == nil {
		t.Fatal("InputChan() returned nil")
	}

	// Send event
	ch <- KeyEvent{Key: KeyEnter, Down: true}
	ui.processInput()

	if !ui.input.KeyDown[KeyEnter] {
		t.Error("InputChan event was not processed")
	}
}

func TestUI_MousePressedCleared(t *testing.T) {
	ui := New(Config{})

	// MousePressed should persist during frame, cleared at EndFrame
	ui.MouseDown(10, 20, MouseLeft)
	ui.BeginFrame()

	// Should still be true during the frame
	if !ui.input.MousePressed[0] {
		t.Error("MousePressed should persist during frame")
	}

	ui.EndFrame()

	// Should be cleared after EndFrame
	if ui.input.MousePressed[0] {
		t.Error("MousePressed should be cleared after EndFrame")
	}
}

// TestUI_MouseDeltaFromChannel tests that channel-driven mouse events are processed
// BEFORE MouseDelta is computed, so drag math is correct on the same frame.
// This tests the fix for: "BeginFrame computes MouseDelta before processInput"
func TestUI_MouseDeltaFromChannel(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Set initial position
	ui.MouseMove(100, 100)
	ui.BeginFrame()
	ui.EndFrame()

	// Frame 2: Send mouse button event via channel BEFORE BeginFrame
	// MouseEvent updates position when processing button events
	ch := ui.InputChan()
	ch <- MouseEvent{X: 150, Y: 120, Btn: MouseLeft, Down: true} // Move +50, +20

	// BeginFrame should process channel events BEFORE computing delta
	ui.BeginFrame()

	// Delta should reflect the channel event's position update
	if ui.input.MouseDelta.X != 50 {
		t.Errorf("MouseDelta.X = %d, want 50 (channel event should be processed before delta)", ui.input.MouseDelta.X)
	}
	if ui.input.MouseDelta.Y != 20 {
		t.Errorf("MouseDelta.Y = %d, want 20 (channel event should be processed before delta)", ui.input.MouseDelta.Y)
	}

	// Position should be updated
	if ui.input.MousePos.X != 150 || ui.input.MousePos.Y != 120 {
		t.Errorf("MousePos = %v, want (150, 120)", ui.input.MousePos)
	}

	ui.EndFrame()
}

// TestUI_MouseDeltaMultipleChannelEvents tests that multiple channel events
// are all processed and the final delta reflects the total movement.
func TestUI_MouseDeltaMultipleChannelEvents(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Set initial position
	ui.MouseMove(0, 0)
	ui.BeginFrame()
	ui.EndFrame()

	// Frame 2: Send multiple mouse button events via channel
	// Each MouseEvent updates the position
	ch := ui.InputChan()
	ch <- MouseEvent{X: 10, Y: 5, Btn: MouseLeft, Down: true}
	ch <- MouseEvent{X: 25, Y: 15, Btn: MouseLeft, Down: false}
	ch <- MouseEvent{X: 30, Y: 30, Btn: MouseLeft, Down: true} // Final position

	ui.BeginFrame()

	// Delta should be from last known pos (0,0) to final pos (30,30)
	if ui.input.MouseDelta.X != 30 {
		t.Errorf("MouseDelta.X = %d, want 30", ui.input.MouseDelta.X)
	}
	if ui.input.MouseDelta.Y != 30 {
		t.Errorf("MouseDelta.Y = %d, want 30", ui.input.MouseDelta.Y)
	}

	ui.EndFrame()
}

func TestIsKeyDown_TracksKeyState(t *testing.T) {
	ui := New(Config{})

	if ui.IsKeyDown(KeyShift) {
		t.Error("shift should not be down initially")
	}

	ui.KeyDown(KeyShift)
	if !ui.IsKeyDown(KeyShift) {
		t.Error("shift should be down after KeyDown")
	}

	ui.KeyUp(KeyShift)
	if ui.IsKeyDown(KeyShift) {
		t.Error("shift should be up after KeyUp")
	}
}

func TestIsKeyDown_MultipleKeys(t *testing.T) {
	ui := New(Config{})

	ui.KeyDown(KeyCtrl)
	ui.KeyDown(KeyShift)

	if !ui.IsKeyDown(KeyCtrl) || !ui.IsKeyDown(KeyShift) {
		t.Error("both ctrl and shift should be down")
	}

	ui.KeyUp(KeyCtrl)

	if ui.IsKeyDown(KeyCtrl) {
		t.Error("ctrl should be up")
	}
	if !ui.IsKeyDown(KeyShift) {
		t.Error("shift should still be down")
	}
}

func TestTextChar_AccumulatesInput(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.TextChar('H')
	ui.TextChar('i')
	ui.TextChar('!')

	if ui.input.TextInput != "Hi!" {
		t.Errorf("TextInput = %q, want %q", ui.input.TextInput, "Hi!")
	}

	ui.EndFrame()
}

func TestTextChar_ClearsEachFrame(t *testing.T) {
	ui := New(Config{})

	// Frame 1
	ui.BeginFrame()
	ui.TextChar('A')
	ui.EndFrame()

	// Frame 2 - input should be cleared
	ui.BeginFrame()
	if ui.input.TextInput != "" {
		t.Errorf("TextInput should be cleared between frames, got %q", ui.input.TextInput)
	}
	ui.EndFrame()
}

func TestTextEvent_ViaChannel(t *testing.T) {
	ui := New(Config{})

	ch := ui.InputChan()
	ch <- TextEvent{Rune: 'X'}
	ch <- TextEvent{Rune: 'Y'}

	ui.BeginFrame()

	if ui.input.TextInput != "XY" {
		t.Errorf("TextInput = %q, want %q", ui.input.TextInput, "XY")
	}

	ui.EndFrame()
}
