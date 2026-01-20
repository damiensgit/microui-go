package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestWindow_DragByTitleBar(t *testing.T) {
	ui := New(Config{})

	// Setup frame: establish initial mouse position (no delta on first real frame)
	ui.MouseMove(150, 10)
	ui.BeginFrame()
	ui.BeginWindow("Draggable", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click on title bar (mouse position already at 150, 10)
	ui.MouseDown(150, 10, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Draggable", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag mouse 50 pixels right, 30 down
	ui.MouseMove(200, 40)
	ui.BeginFrame()
	ui.BeginWindow("Draggable", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Get container and check position changed
	cnt := ui.GetContainer("Draggable")
	if cnt.Rect().X != 150 {
		t.Errorf("Window X = %d, want 150 (dragged 50px)", cnt.Rect().X)
	}
	if cnt.Rect().Y != 30 {
		t.Errorf("Window Y = %d, want 30 (dragged 30px)", cnt.Rect().Y)
	}
}

func TestWindow_DragOnlyFromTitleBar(t *testing.T) {
	ui := New(Config{})

	// Click in body area (not title bar)
	// Input BEFORE BeginFrame
	ui.MouseMove(150, 100) // Inside window body
	ui.MouseDown(150, 100, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("NoDrag", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Move mouse
	ui.MouseMove(200, 150)
	ui.BeginFrame()
	ui.BeginWindow("NoDrag", types.Rect{X: 100, Y: 50, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Window should NOT have moved
	cnt := ui.GetContainer("NoDrag")
	if cnt.Rect().X != 100 {
		t.Errorf("Window X = %d, want 100 (should not drag from body)", cnt.Rect().X)
	}
}

func TestWindow_DragClearsOnMouseRelease(t *testing.T) {
	ui := New(Config{})

	// Setup frame: establish initial mouse position
	ui.MouseMove(150, 10)
	ui.BeginFrame()
	ui.BeginWindow("DragRelease", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 1: Click on title bar
	ui.MouseDown(150, 10, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("DragRelease", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Drag while mouse down
	ui.MouseMove(200, 40)
	ui.BeginFrame()
	ui.BeginWindow("DragRelease", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Release mouse
	ui.MouseUp(200, 40, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("DragRelease", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Position should be at dragged location
	cnt := ui.GetContainer("DragRelease")
	if cnt.Rect().X != 150 {
		t.Errorf("Window X = %d, want 150", cnt.Rect().X)
	}

	// Frame 4: Move mouse further (should not affect position since drag ended)
	ui.MouseMove(300, 100)
	ui.BeginFrame()
	ui.BeginWindow("DragRelease", types.Rect{X: 100, Y: 0, W: 200, H: 150})
	ui.EndWindow()
	ui.EndFrame()

	// Position should remain at 150 (not continue dragging)
	if cnt.Rect().X != 150 {
		t.Errorf("Window X = %d after release, want 150 (should not continue dragging)", cnt.Rect().X)
	}
}

func TestWindow_NoTitleNoDrag(t *testing.T) {
	ui := New(Config{})

	// Click on what would be title bar area of a titleless window
	ui.MouseMove(150, 10)
	ui.MouseDown(150, 10, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindowOpt("NoTitleWindow", types.Rect{X: 100, Y: 0, W: 200, H: 150}, OptNoTitle)
	ui.EndWindow()
	ui.EndFrame()

	// Move mouse
	ui.MouseMove(200, 40)
	ui.BeginFrame()
	ui.BeginWindowOpt("NoTitleWindow", types.Rect{X: 100, Y: 0, W: 200, H: 150}, OptNoTitle)
	ui.EndWindow()
	ui.EndFrame()

	// Window should NOT have moved (no title bar to drag)
	cnt := ui.GetContainer("NoTitleWindow")
	if cnt.Rect().X != 100 {
		t.Errorf("Window X = %d, want 100 (titleless windows should not be draggable by title)", cnt.Rect().X)
	}
}
