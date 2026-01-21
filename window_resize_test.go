package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestWindow_ResizeFromCorner(t *testing.T) {
	ui := New(Config{})
	style := ui.Style()

	// Window size = body size (content area). System adds chrome.
	// GUIStyle: BorderWidth=0, TitleHeight=24
	// Body: 200x150 → Window: 200x174 (150+24)
	bodyW, bodyH := 200, 150
	windowH := bodyH + style.TitleHeight + style.BorderWidth // 174

	// Window at (100, 50) with size 200x174
	// Bottom-right corner at (300, 224)
	cornerX := 100 + bodyW - 5 // 295
	cornerY := 50 + windowH - 5 // 219

	// Setup frame: establish initial mouse position
	ui.MouseMove(cornerX, cornerY)
	ui.BeginFrame()
	ui.BeginWindow("Resizable", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click on resize corner
	ui.MouseDown(cornerX, cornerY, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Resizable", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag to resize 50 wider, 30 taller
	ui.MouseMove(cornerX+50, cornerY+30)
	ui.BeginFrame()
	ui.BeginWindow("Resizable", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Check size changed
	cnt := ui.GetContainer("Resizable")
	if cnt.Rect().W != 250 {
		t.Errorf("Window W = %d, want 250 (resized +50)", cnt.Rect().W)
	}
	expectedH := windowH + 30 // 174 + 30 = 204
	if cnt.Rect().H != expectedH {
		t.Errorf("Window H = %d, want %d (resized +30)", cnt.Rect().H, expectedH)
	}
}

func TestWindow_ResizeMinimum(t *testing.T) {
	ui := New(Config{})

	// Setup frame: establish initial mouse position
	ui.MouseMove(295, 195)
	ui.BeginFrame()
	ui.BeginWindow("Minimum", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click on resize corner
	ui.MouseDown(295, 195, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Minimum", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag far left/up (try to make very small)
	ui.MouseMove(110, 60) // Way past origin
	ui.BeginFrame()
	ui.BeginWindow("Minimum", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Check minimum size enforced (at least 10x5 for TUI-friendly minimums)
	cnt := ui.GetContainer("Minimum")
	if cnt.Rect().W < 10 {
		t.Errorf("Window W = %d, should be at least 10 (minimum)", cnt.Rect().W)
	}
	if cnt.Rect().H < 5 {
		t.Errorf("Window H = %d, should be at least 5 (minimum)", cnt.Rect().H)
	}
}

func TestWindow_NoResizeOption(t *testing.T) {
	ui := New(Config{})

	// Setup frame: establish initial mouse position
	ui.MouseMove(295, 195)
	ui.BeginFrame()
	ui.BeginWindowOpt("NoResize", types.Rect{X: 100, Y: 50, W: 200, H: 150}, OptNoResize)
	ui.EndWindow()
	ui.EndFrame()

	// Click on resize corner of non-resizable window
	ui.MouseDown(295, 195, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindowOpt("NoResize", types.Rect{X: 100, Y: 50, W: 200, H: 150}, OptNoResize)
	ui.EndWindow()
	ui.EndFrame()

	// Move mouse
	ui.MouseMove(345, 225)
	ui.BeginFrame()
	ui.BeginWindowOpt("NoResize", types.Rect{X: 100, Y: 50, W: 200, H: 150}, OptNoResize)
	ui.EndWindow()
	ui.EndFrame()

	// Window should NOT have resized
	cnt := ui.GetContainer("NoResize")
	if cnt.Rect().W != 200 {
		t.Errorf("Window W = %d, want 200 (OptNoResize should prevent resize)", cnt.Rect().W)
	}
}

func TestWindow_ResizeClearsOnMouseRelease(t *testing.T) {
	ui := New(Config{})
	style := ui.Style()

	// Window size = body size (content area). System adds chrome.
	bodyW, bodyH := 200, 150
	windowH := bodyH + style.TitleHeight + style.BorderWidth // 174

	// Window at (100, 50), corner at (295, 219)
	cornerX := 100 + bodyW - 5
	cornerY := 50 + windowH - 5

	// Setup frame: establish initial mouse position
	ui.MouseMove(cornerX, cornerY)
	ui.BeginFrame()
	ui.BeginWindow("ResizeRelease", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click on resize corner
	ui.MouseDown(cornerX, cornerY, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("ResizeRelease", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag while mouse down (+50, +30)
	ui.MouseMove(cornerX+50, cornerY+30)
	ui.BeginFrame()
	ui.BeginWindow("ResizeRelease", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Release mouse
	ui.MouseUp(cornerX+50, cornerY+30, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("ResizeRelease", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Size should be at resized dimensions
	cnt := ui.GetContainer("ResizeRelease")
	if cnt.Rect().W != 250 {
		t.Errorf("Window W = %d, want 250", cnt.Rect().W)
	}

	// Frame 4: Move mouse further (should not affect size since resize ended)
	ui.MouseMove(400, 300)
	ui.BeginFrame()
	ui.BeginWindow("ResizeRelease", types.Rect{X: 100, Y: 50, W: bodyW, H: bodyH})
	ui.EndWindow()
	ui.EndFrame()

	// Size should remain at 250 (not continue resizing)
	if cnt.Rect().W != 250 {
		t.Errorf("Window W = %d after release, want 250 (should not continue resizing)", cnt.Rect().W)
	}
}

func TestWindow_ResizePositionUnchanged(t *testing.T) {
	ui := New(Config{})

	// Setup frame: establish initial mouse position
	ui.MouseMove(295, 195)
	ui.BeginFrame()
	ui.BeginWindow("ResizePos", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click on resize corner
	ui.MouseDown(295, 195, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("ResizePos", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag to resize
	ui.MouseMove(345, 225)
	ui.BeginFrame()
	ui.BeginWindow("ResizePos", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Position should remain unchanged (only size changes)
	cnt := ui.GetContainer("ResizePos")
	if cnt.Rect().X != 100 {
		t.Errorf("Window X = %d, want 100 (position should not change during resize)", cnt.Rect().X)
	}
	if cnt.Rect().Y != 50 {
		t.Errorf("Window Y = %d, want 50 (position should not change during resize)", cnt.Rect().Y)
	}
}
