package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestHeader_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)

	// Header should return true when expanded (default)
	if ui.Header("Section 1") {
		ui.Label("Content 1")
	}

	// Another header
	if ui.Header("Section 2") {
		ui.Label("Content 2")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestHeader_Toggle(t *testing.T) {
	ui := New(Config{})

	// First frame - render header
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 30)

	expanded1 := ui.Header("Section")

	ui.EndWindow()
	ui.EndFrame()

	// Default should be expanded
	if !expanded1 {
		t.Error("Header should be expanded by default")
	}

	// Click on header to toggle
	// Header is at Y ~= 29 (after title 24 + padding 5), height ~= 20 (size.Y 10 + padding*2)
	// So click at Y=35 to be inside the control
	ui.MouseMove(50, 35)
	ui.MouseDown(50, 35, MouseLeft)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 30)

	expanded2 := ui.Header("Section")

	ui.EndWindow()
	ui.EndFrame()

	// Should now be collapsed after click
	if expanded2 {
		t.Error("Header should be collapsed after click")
	}
}
