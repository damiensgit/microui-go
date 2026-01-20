package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestPopup_DoesNotCloseOnOpeningClick(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Create window with button, simulate click to open popup
	// Button is positioned at (10, 29) to (110, 59) in window at (0,0)
	// Title height=24, padding Y=5, so content starts at Y=29
	ui.MouseMove(60, 45)
	ui.MouseDown(60, 45, MouseLeft)

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	if ui.Button("Open") {
		ui.OpenPopup("test_popup")
	}
	ui.EndWindow()

	// Popup should be open after button click
	if ui.BeginPopup("test_popup") {
		ui.Label("Content")
		ui.EndPopup()
	}
	ui.EndFrame()

	// Frame 2: Release mouse, popup should still be open
	ui.MouseUp(60, 45, MouseLeft)

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Button("Open")
	ui.EndWindow()

	popupOpen := ui.BeginPopup("test_popup")
	if popupOpen {
		ui.Label("Content")
		ui.EndPopup()
	}
	ui.EndFrame()

	if !popupOpen {
		t.Error("Popup should remain open on the frame after opening click")
	}
}

func TestPopup_ClosesOnOutsideClick(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Open popup
	ui.BeginFrame()
	ui.OpenPopup("test_popup")
	if ui.BeginPopup("test_popup") {
		ui.Label("Content")
		ui.EndPopup()
	}
	ui.EndFrame()

	// Frame 2: Move mouse outside popup (no click yet)
	// C microui behavior: hover_root is updated based on previous frame's mouse position.
	// So we need to move the mouse first, then click in the next frame.
	ui.MouseMove(500, 500) // Far from popup

	ui.BeginFrame()
	popupOpen := ui.BeginPopup("test_popup")
	if popupOpen {
		ui.Label("Content")
		ui.EndPopup()
	}
	ui.EndFrame()

	// Frame 3: Click outside - now hover_root should be nil (from Frame 2's mouse move)
	ui.MouseDown(500, 500, MouseLeft)

	ui.BeginFrame()
	popupStillOpen := ui.BeginPopup("test_popup")
	if popupStillOpen {
		ui.EndPopup()
	}
	ui.EndFrame()

	if popupStillOpen {
		t.Error("Popup should close after outside click")
	}
}
