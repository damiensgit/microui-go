package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestNew(t *testing.T) {
	cfg := Config{
		CommandBuf: 512,
	}

	ui := New(cfg)

	if ui == nil {
		t.Fatal("New() returned nil")
	}

	if cap(ui.commands.cmds) != 512 {
		t.Errorf("Command buffer capacity = %d, want 512", cap(ui.commands.cmds))
	}
}

func TestUI_BeginFrame(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// BeginFrame should clear previous commands
	if ui.commands.Len() != 0 {
		t.Errorf("BeginFrame() left %d commands", ui.commands.Len())
	}

	ui.EndFrame()
}

func TestUI_EndFrame(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.EndFrame()

	// Should not panic
}

func TestLabel(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Should not panic
	ui.Label("Hello, World!")

	ui.EndFrame()

	// Success if no panic
}

func TestButton_Click(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Mouse outside button - not clicked
	clicked := ui.Button("Click Me")
	if clicked {
		t.Error("Button() returned true without mouse interaction")
	}

	ui.EndFrame()
}

func TestButton_HoverState(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Mouse over button
	ui.MouseMove(50, 20)
	ui.MouseDown(50, 20, MouseLeft)

	// TODO: test hover/click states when layout is implemented

	ui.EndFrame()
}

func TestWindow_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	opened := ui.BeginWindow("Test", rect)

	if !opened {
		t.Error("BeginWindow() returned false, should default to open")
	}

	ui.Label("Inside window")
	ui.EndWindow()

	ui.EndFrame()
}

func TestWindow_ClipRect(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	ui.BeginWindow("Test", rect)

	// Window should set clip rect
	ui.Label("Clipped text")
	ui.EndWindow()

	ui.EndFrame()

	// After EndWindow, clip should be restored
}
