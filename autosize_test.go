package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestWindow_AutoSizeGrowsToFitContent(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Create autosize window with content
	// C microui: autosize uses PREVIOUS frame's contentSize, so first frame calculates size
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 50, H: 50}, OptAutoSize)
	ui.LayoutRow(1, []int{200}, 0) // Content wants 200 width
	ui.Label("Wide content here")
	ui.Label("More content")
	ui.Label("Even more content")
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Autosize takes effect using frame 1's contentSize
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 50, H: 50}, OptAutoSize)
	ui.LayoutRow(1, []int{200}, 0)
	ui.Label("Wide content here")
	ui.Label("More content")
	ui.Label("Even more content")
	ui.EndWindow()
	ui.EndFrame()

	// Window should have grown to fit content
	cnt := ui.GetContainer("AutoSize")
	if cnt.Rect().W < 200 {
		t.Errorf("AutoSize window W = %d, want >= 200", cnt.Rect().W)
	}
}

func TestWindow_AutoSizeGrowsHeight(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Calculate content size
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 200, H: 50}, OptAutoSize)
	for i := 0; i < 10; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content line")
	}
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Autosize takes effect
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 200, H: 50}, OptAutoSize)
	for i := 0; i < 10; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content line")
	}
	ui.EndWindow()
	ui.EndFrame()

	cnt := ui.GetContainer("AutoSize")
	// 10 lines at ~30px each (default size.Y) + title bar (~24px) + padding = ~350px
	if cnt.Rect().H < 100 {
		t.Errorf("AutoSize window H = %d, want > 100 (should grow for content)", cnt.Rect().H)
	}
}

func TestWindow_NormalWindowDoesNotAutoSize(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	// Normal window without OptAutoSize
	ui.BeginWindow("Normal", types.Rect{X: 100, Y: 100, W: 200, H: 100})
	ui.LayoutRow(1, []int{300}, 0) // Content wider than window
	ui.Label("Wide content")
	ui.EndWindow()
	ui.EndFrame()

	cnt := ui.GetContainer("Normal")
	// Window should NOT have grown
	if cnt.Rect().W != 200 {
		t.Errorf("Normal window W = %d, want 200 (should not auto-size)", cnt.Rect().W)
	}
}

func TestWindow_AutoSizeWithMinimumSize(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Calculate content size
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 10, H: 10}, OptAutoSize)
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Label("x") // Very small content
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Autosize takes effect
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 10, H: 10}, OptAutoSize)
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Label("x")
	ui.EndWindow()
	ui.EndFrame()

	cnt := ui.GetContainer("AutoSize")
	// Should have at least a reasonable minimum size
	if cnt.Rect().W < 50 || cnt.Rect().H < 50 {
		t.Errorf("AutoSize window should have minimum size, got %dx%d", cnt.Rect().W, cnt.Rect().H)
	}
}

func TestWindow_AutoSizeAdjustsToContent(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Create autosize window with large content
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 50, H: 50}, OptAutoSize)
	ui.LayoutRow(1, []int{300}, 0)
	ui.Label("Wide content")
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Autosize takes effect (grows to fit Frame 1's content)
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 50, H: 50}, OptAutoSize)
	ui.LayoutRow(1, []int{300}, 0)
	ui.Label("Wide content")
	ui.EndWindow()
	ui.EndFrame()

	cnt := ui.GetContainer("AutoSize")
	// C microui: autosize sets window to fit content (can grow or shrink)
	// After 2 frames with 300-width content, window should have grown
	if cnt.Rect().W < 300 {
		t.Errorf("AutoSize window W = %d, want >= 300 to fit content", cnt.Rect().W)
	}
}

func TestWindow_AutoSizeWithNoTitle(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Calculate content size
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 50, H: 50}, OptAutoSize|OptNoTitle)
	ui.LayoutRow(1, []int{200}, 0)
	ui.Label("Content")
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Autosize takes effect
	ui.BeginFrame()
	ui.BeginWindowOpt("AutoSize", types.Rect{X: 100, Y: 100, W: 50, H: 50}, OptAutoSize|OptNoTitle)
	ui.LayoutRow(1, []int{200}, 0)
	ui.Label("Content")
	ui.EndWindow()
	ui.EndFrame()

	cnt := ui.GetContainer("AutoSize")
	// Should still grow to fit content
	if cnt.Rect().W < 200 {
		t.Errorf("AutoSize window (no title) W = %d, want >= 200", cnt.Rect().W)
	}
}
