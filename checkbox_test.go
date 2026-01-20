package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestCheckbox(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	checked := false

	// Not clicked initially
	if ui.Checkbox("Check me", &checked) {
		t.Error("Checkbox() returned true without click")
	}

	if checked {
		t.Error("Checkbox modified value without click")
	}

	ui.EndFrame()
}

func TestCheckbox_Toggle(t *testing.T) {
	ui := New(Config{})

	// Simulate mouse click at checkbox position
	ui.MouseMove(15, 15) // Inside default checkbox area
	ui.MouseDown(15, 15, MouseLeft)
	ui.BeginFrame()

	checked := false
	changed := ui.Checkbox("Check me", &checked)

	if changed && !checked {
		t.Error("Checkbox returned true but didn't set checked=true")
	}

	ui.EndFrame()
}

func TestCheckbox_WithWindow(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	if ui.BeginWindow("Test", rect) {
		checked := true
		ui.Checkbox("Option 1", &checked)

		checked2 := false
		ui.Checkbox("Option 2", &checked2)

		ui.EndWindow()
	}

	ui.EndFrame()
}
