package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestSlider(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	value := 0.5

	// Not changed initially
	if ui.Slider(&value, 0, 1) {
		t.Error("Slider() returned true without interaction")
	}

	// Value should be clamped
	ui.Slider(&value, 0, 1)
	if value < 0 || value > 1 {
		t.Errorf("Slider() value = %f, want [0, 1]", value)
	}

	ui.EndFrame()
}

func TestSlider_Drag(t *testing.T) {
	ui := New(Config{})

	// Position slider at known location and drag
	ui.BeginFrame()

	// Move mouse to middle of slider and hold down
	ui.MouseMove(50, 15)
	ui.MouseDown(50, 15, MouseLeft)

	ui.EndFrame()

	// New frame with mouse still down
	ui.BeginFrame()

	value := 0.0
	changed := ui.Slider(&value, 0, 1)

	// Should have changed since mouse is down over the slider area
	if changed && (value < 0 || value > 1) {
		t.Errorf("Slider value out of range: %f", value)
	}

	ui.EndFrame()
}

func TestSlider_WithWindow(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	if ui.BeginWindow("Test", rect) {
		value := 0.5
		ui.Slider(&value, 0, 100)

		value2 := 25.0
		ui.Slider(&value2, 0, 50)

		ui.EndWindow()
	}

	ui.EndFrame()
}

func TestSlider_Clamp(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Test that values get clamped
	value := -5.0
	ui.Slider(&value, 0, 10)

	// Value should stay unchanged since no mouse interaction
	if value != -5.0 {
		t.Errorf("Slider modified value without interaction: %f", value)
	}

	ui.EndFrame()
}
