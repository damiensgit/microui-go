package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestGetCurrentContainer(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("TestWindow", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	cnt := ui.GetCurrentContainer()
	if cnt == nil {
		t.Fatal("GetCurrentContainer returned nil inside window")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestGetContainer(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("MyWindow", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.EndWindow()

	// Get container by name
	cnt := ui.GetContainer("MyWindow")
	if cnt == nil {
		t.Fatal("GetContainer returned nil for existing window")
	}

	ui.EndFrame()
}

func TestContainerMethods(t *testing.T) {
	ui := New(Config{})
	style := ui.Style()
	ui.BeginFrame()

	// Window size = body size (content area). System adds chrome.
	bodyRect := types.Rect{X: 10, Y: 20, W: 400, H: 300}
	ui.BeginWindow("TestWindow", bodyRect)

	cnt := ui.GetCurrentContainer()
	if cnt == nil {
		t.Fatal("GetCurrentContainer returned nil")
	}

	// Test ID
	if cnt.ID() == 0 {
		t.Error("Container ID should not be zero")
	}

	// Test Rect - container rect includes chrome (title, borders)
	// GUIStyle: BorderWidth=0, TitleHeight=24
	expectedRect := types.Rect{
		X: bodyRect.X,
		Y: bodyRect.Y,
		W: bodyRect.W + style.BorderWidth*2,
		H: bodyRect.H + style.TitleHeight + style.BorderWidth,
	}
	if cnt.Rect() != expectedRect {
		t.Errorf("Container Rect mismatch: got %v, want %v", cnt.Rect(), expectedRect)
	}

	// Test SetRect - sets the total window rect (including chrome)
	newRect := types.Rect{X: 50, Y: 60, W: 200, H: 150}
	cnt.SetRect(newRect)
	if cnt.Rect() != newRect {
		t.Errorf("SetRect failed: got %v, want %v", cnt.Rect(), newRect)
	}

	// Test Scroll
	scroll := types.Vec2{X: 0, Y: 0}
	if cnt.Scroll() != scroll {
		t.Errorf("Initial scroll should be zero: got %v", cnt.Scroll())
	}

	// Test SetScroll
	newScroll := types.Vec2{X: 10, Y: 20}
	cnt.SetScroll(newScroll)
	if cnt.Scroll() != newScroll {
		t.Errorf("SetScroll failed: got %v, want %v", cnt.Scroll(), newScroll)
	}

	// Test ZIndex
	if cnt.ZIndex() <= 0 {
		t.Error("Container ZIndex should be positive")
	}

	// Test Open
	if !cnt.Open() {
		t.Error("Container should be open")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestGetCurrentContainerOutsideWindow(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Before any window, should return nil
	cnt := ui.GetCurrentContainer()
	if cnt != nil {
		t.Error("GetCurrentContainer should return nil outside any window")
	}

	ui.EndFrame()
}

func TestContainerPersistence(t *testing.T) {
	ui := New(Config{})

	// First frame - create container with overflow content so scroll is valid
	ui.BeginFrame()
	ui.BeginWindow("PersistentWindow", types.Rect{X: 0, Y: 0, W: 400, H: 100})
	// Add content that overflows to make scroll valid
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 30)
		ui.Label("Content line")
	}
	cnt1 := ui.GetContainer("PersistentWindow")
	cnt1.SetScroll(types.Vec2{X: 0, Y: 100}) // Set valid scroll within content range
	ui.EndWindow()
	ui.EndFrame()

	// Second frame - container should persist
	ui.BeginFrame()
	ui.BeginWindow("PersistentWindow", types.Rect{X: 0, Y: 0, W: 400, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 30)
		ui.Label("Content line")
	}
	cnt2 := ui.GetContainer("PersistentWindow")
	if cnt2 == nil {
		t.Fatal("Container should persist between frames")
	}
	// Scroll should persist (may be clamped to max valid scroll)
	if cnt2.Scroll().Y == 0 {
		t.Errorf("Container scroll should persist (non-zero): got %v", cnt2.Scroll())
	}
	ui.EndWindow()
	ui.EndFrame()
}

func TestBringToFront(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Window1", types.Rect{X: 0, Y: 0, W: 200, H: 200})
	ui.EndWindow()

	ui.BeginWindow("Window2", types.Rect{X: 50, Y: 50, W: 200, H: 200})
	ui.EndWindow()

	cnt1 := ui.GetContainer("Window1")
	cnt2 := ui.GetContainer("Window2")

	// Window2 should be on top initially (opened later)
	if cnt1.ZIndex() >= cnt2.ZIndex() {
		t.Error("Window2 should have higher z-index initially")
	}

	// Bring Window1 to front
	ui.BringToFront(cnt1)

	if cnt1.ZIndex() <= cnt2.ZIndex() {
		t.Error("After BringToFront, Window1 should have higher z-index")
	}

	ui.EndFrame()
}
