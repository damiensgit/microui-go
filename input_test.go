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
