package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestPopup_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 0)

	if ui.Button("Open Popup") {
		ui.OpenPopup("my_popup")
	}

	if ui.BeginPopup("my_popup") {
		ui.Label("Popup content")
		ui.Button("Close") // Popup closes when clicking outside
		ui.EndPopup()
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestPopup_OpenClose(t *testing.T) {
	ui := New(Config{})

	// First frame - popup should not be open
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 30)

	popupOpen := ui.BeginPopup("test_popup")
	if popupOpen {
		ui.Label("Content")
		ui.EndPopup()
	}

	ui.EndWindow()
	ui.EndFrame()

	if popupOpen {
		t.Error("Popup should not be open initially")
	}

	// Open the popup
	ui.OpenPopup("test_popup")

	// Next frame - popup should be open
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{100}, 30)

	popupOpen = ui.BeginPopup("test_popup")
	if popupOpen {
		ui.Label("Content")
		ui.EndPopup()
	}

	ui.EndWindow()
	ui.EndFrame()

	if !popupOpen {
		t.Error("Popup should be open after OpenPopup")
	}
}
