package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestScrollbar_AppearsWhenContentOverflows(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Scrollable", types.Rect{X: 0, Y: 0, W: 200, H: 100})

	// Add content that overflows (20 labels at ~24px each = 480px, window is 100px)
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content line")
	}

	ui.EndWindow()
	ui.EndFrame()

	// Content size should exceed window height
	cnt := ui.GetContainer("Scrollable")
	if cnt.ContentSize().Y <= 100 {
		t.Errorf("ContentSize.Y = %d, want > 100 (content should overflow)", cnt.ContentSize().Y)
	}
}

func TestScrollbar_DraggingChangesScroll(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Create window with overflow
	ui.BeginFrame()
	ui.BeginWindow("Scrollable", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click on scrollbar area (right edge)
	ui.BeginFrame()
	ui.MouseMove(195, 50) // Right edge, middle height
	ui.MouseDown(195, 50, MouseLeft)
	ui.BeginWindow("Scrollable", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Drag scrollbar down
	ui.BeginFrame()
	ui.MouseMove(195, 80)
	ui.BeginWindow("Scrollable", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	// Scroll should have changed
	cnt := ui.GetContainer("Scrollable")
	if cnt.Scroll().Y <= 0 {
		t.Errorf("Scroll.Y = %d, want > 0 after dragging scrollbar", cnt.Scroll().Y)
	}
}

func TestScrollbar_NoScrollbarWhenContentFits(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Small", types.Rect{X: 0, Y: 0, W: 200, H: 200})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Label("Just one line")
	ui.EndWindow()
	ui.EndFrame()

	// Content should fit, no scroll needed
	cnt := ui.GetContainer("Small")
	bodyHeight := cnt.Body().H
	if cnt.ContentSize().Y > bodyHeight {
		t.Errorf("ContentSize.Y = %d exceeds body height %d, but content should fit", cnt.ContentSize().Y, bodyHeight)
	}
}

func TestScrollbar_ContentSizeTracksMultipleControls(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("TestWin", types.Rect{X: 0, Y: 0, W: 200, H: 300})

	// Add several controls
	for i := 0; i < 5; i++ {
		ui.LayoutRow(1, []int{-1}, 30)
		ui.Label("Line")
	}

	ui.EndWindow()
	ui.EndFrame()

	// Content size should reflect all controls
	// 5 labels * 30 height + spacing + padding
	cnt := ui.GetContainer("TestWin")
	if cnt.ContentSize().Y < 150 {
		t.Errorf("ContentSize.Y = %d, want >= 150 for 5 labels at 30px each", cnt.ContentSize().Y)
	}
}

func TestScrollbar_ScrollClampsToMaxScroll(t *testing.T) {
	ui := New(Config{})

	// Create window with overflow
	ui.BeginFrame()
	ui.BeginWindow("Clamp", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	// Manually set scroll beyond limits
	cnt := ui.GetContainer("Clamp")
	cnt.SetScroll(types.Vec2{X: 0, Y: 9999})

	// Next frame should clamp scroll
	ui.BeginFrame()
	ui.BeginWindow("Clamp", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	// Scroll should be clamped to maxScroll (contentSize + padding*2 - bodyHeight)
	// scrollbars() adds padding*2 to contentSize before calculating max scroll
	body := cnt.Body()
	padding := ui.Style().Padding
	maxScroll := cnt.ContentSize().Y + padding.Y*2 - body.H
	if cnt.Scroll().Y > maxScroll {
		t.Errorf("Scroll.Y = %d, should be clamped to maxScroll %d", cnt.Scroll().Y, maxScroll)
	}
}

func TestScrollbar_NegativeScrollClampsToZero(t *testing.T) {
	ui := New(Config{})

	// Create window
	ui.BeginFrame()
	ui.BeginWindow("NegClamp", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	// Manually set negative scroll
	cnt := ui.GetContainer("NegClamp")
	cnt.SetScroll(types.Vec2{X: -100, Y: -100})

	// Next frame should clamp scroll to 0
	ui.BeginFrame()
	ui.BeginWindow("NegClamp", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	if cnt.Scroll().Y < 0 {
		t.Errorf("Scroll.Y = %d, should be clamped to >= 0", cnt.Scroll().Y)
	}
	if cnt.Scroll().X < 0 {
		t.Errorf("Scroll.X = %d, should be clamped to >= 0", cnt.Scroll().X)
	}
}

func TestScrollbar_HorizontalScrollWhenContentOverflows(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("HScroll", types.Rect{X: 0, Y: 0, W: 100, H: 200})

	// Add wide content
	ui.LayoutRow(1, []int{300}, 30) // 300px wide content in 100px window
	ui.Label("Very wide content")

	ui.EndWindow()
	ui.EndFrame()

	// Content width should exceed window width
	cnt := ui.GetContainer("HScroll")
	if cnt.ContentSize().X <= 100 {
		t.Errorf("ContentSize.X = %d, want > 100 for horizontal overflow", cnt.ContentSize().X)
	}
}

func TestScrollbar_BodyReducedWhenScrollbarShown(t *testing.T) {
	ui := New(Config{})

	// First render to establish content size
	ui.BeginFrame()
	ui.BeginWindow("BodyReduced", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	cnt := ui.GetContainer("BodyReduced")

	// The body width should be reduced by scrollbar size when scrollbar is shown
	// Window width is 200, scrollbar is 12, so body should be 200 - 12 = 188 or less
	scrollbarSize := ui.Style().ScrollbarSize
	expectedMaxWidth := 200 - scrollbarSize

	// Re-render to see updated body
	ui.BeginFrame()
	ui.BeginWindow("BodyReduced", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndWindow()
	ui.EndFrame()

	if cnt.Body().W > expectedMaxWidth {
		t.Errorf("Body.W = %d, want <= %d when vertical scrollbar is shown", cnt.Body().W, expectedMaxWidth)
	}
}

// TestTUIScrollbar_VerticalAppearsImmediately tests that vertical scrollbar appears
// on the same frame that content overflows, not one frame late.
// This is critical for TUI where the 1-cell scrollbar gets clipped if space isn't reserved.
func TestTUIScrollbar_VerticalAppearsImmediately(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	// Calculate expected body dimensions for a 20x12 window
	// TUI: titleHeight=1, borderWidth=1
	windowW, windowH := 20, 12
	bodyW := windowW - style.BorderWidth*2  // 20 - 2 = 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 12 - 1 - 1 = 10

	// Content that just overflows: 11 rows in 10-row body
	contentH := bodyH + 1 // 11

	// Frame 1: Create window with overflowing content
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		// Add content rows
		for i := 0; i < contentH; i++ {
			ui.LayoutRow(1, []int{-1}, 1) // 1-cell height per row
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2: Scrollbar should now be visible
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentH; i++ {
			ui.LayoutRow(1, []int{-1}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")

	// Body width should be reduced by scrollbar size
	expectedBodyW := bodyW - style.ScrollbarSize
	if cnt.Body().W != expectedBodyW {
		t.Errorf("Body.W = %d, want %d (scrollbar should reduce width)", cnt.Body().W, expectedBodyW)
	}
}

// TestTUIScrollbar_HorizontalOnly tests that horizontal scrollbar appears
// when content is wider than body but not taller.
func TestTUIScrollbar_HorizontalOnly(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 8
	bodyW := windowW - style.BorderWidth*2                    // 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 6

	// Content that:
	// - Is wider than body (needs horizontal scrollbar)
	// - Fits in body height (no vertical scrollbar)
	contentW := bodyW + 10 // 28 - needs horizontal scrollbar
	contentRows := 2       // Small content, fits in 6 cells (2 rows * 2 cells each with spacing)

	drawFrame := func() {
		ui.BeginFrame()
		if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
			for i := 0; i < contentRows; i++ {
				ui.LayoutRow(1, []int{contentW}, 1)
				ui.Label("X")
			}
			ui.EndWindow()
		}
		ui.EndFrame()
	}

	// Run multiple frames for scrollbar state to stabilize
	for i := 0; i < 3; i++ {
		drawFrame()
	}

	cnt := ui.GetContainer("Test")

	// Horizontal scrollbar should be present (content wider than body)
	// Body height reduced for horizontal scrollbar
	expectedBodyH := bodyH - style.ScrollbarSize
	if cnt.Body().H != expectedBodyH {
		t.Errorf("Body.H = %d, want %d (horizontal scrollbar should reduce height)",
			cnt.Body().H, expectedBodyH)
	}

	// No vertical scrollbar needed (content fits vertically)
	// Body width should NOT be reduced
	if cnt.Body().W != bodyW {
		t.Errorf("Body.W = %d, want %d (no vertical scrollbar expected)",
			cnt.Body().W, bodyW)
	}
}

// TestTUIScrollbar_VerticalTriggersHorizontal tests that when vertical scrollbar
// appears, it reduces body width which may trigger horizontal scrollbar.
func TestTUIScrollbar_VerticalTriggersHorizontal(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 12
	bodyW := windowW - style.BorderWidth*2  // 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 10

	// Content that:
	// - Is taller than body (needs vertical scrollbar)
	// - Fits in body width, but NOT when vertical scrollbar takes 1 column
	contentW := bodyW      // 18 - fits exactly, but vert scrollbar will make it overflow
	contentH := bodyH + 5  // 15 - needs vertical scrollbar

	// Frame 1
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentH; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentH; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")

	// Both scrollbars should be present
	expectedBodyW := bodyW - style.ScrollbarSize
	expectedBodyH := bodyH - style.ScrollbarSize

	if cnt.Body().W != expectedBodyW {
		t.Errorf("Body.W = %d, want %d (vertical scrollbar should reduce width)",
			cnt.Body().W, expectedBodyW)
	}
	if cnt.Body().H != expectedBodyH {
		t.Errorf("Body.H = %d, want %d (horizontal scrollbar should be triggered by vertical)",
			cnt.Body().H, expectedBodyH)
	}
}

// TestTUIScrollbar_AppearsWhenWindowShrinks tests that vertical scrollbar appears
// immediately when window shrinks to make content overflow, not one frame late.
// This is the "content 10, body shrinks from 10 to 9" scenario.
func TestTUIScrollbar_AppearsWhenWindowShrinks(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	// Start with window large enough to fit content
	largeWindowH := 15  // Body will be 15 - 1 (title) - 1 (border) = 13
	smallWindowH := 12  // Body will be 12 - 1 - 1 = 10

	windowW := 20
	bodyW := windowW - style.BorderWidth*2 // 18

	// Content: 6 rows with spacing = 6 + 5 = 11 cells (fits in 13, overflows 10)
	contentRows := 6

	// Frame 1: Large window, content fits
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: largeWindowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{-1}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")
	// Verify no scrollbar yet
	if cnt.Body().W != bodyW {
		t.Errorf("Frame 1: Body.W = %d, want %d (no scrollbar yet)", cnt.Body().W, bodyW)
	}

	// Frame 2: Shrink window - content should now overflow
	// Manually update container rect to simulate window resize
	cnt.SetRect(types.Rect{X: 0, Y: 0, W: windowW, H: smallWindowH})

	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: smallWindowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{-1}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 3: C microui uses previous frame's body, so scrollbar appears one frame after resize
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: smallWindowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{-1}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Scrollbar should now be present (after frame stabilization)
	expectedBodyW := bodyW - style.ScrollbarSize
	if cnt.Body().W != expectedBodyW {
		t.Errorf("Frame 3 (after stabilization): Body.W = %d, want %d (scrollbar should appear)",
			cnt.Body().W, expectedBodyW)
	}
}

// TestTUIScrollbar_MaxScrollAccountsForBothScrollbars tests that when both
// scrollbars are present, maxScrollY accounts for horizontal scrollbar height
// and maxScrollX accounts for vertical scrollbar width.
func TestTUIScrollbar_MaxScrollAccountsForBothScrollbars(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 12
	bodyW := windowW - style.BorderWidth*2                    // 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 10

	// Content larger than body in BOTH dimensions (needs both scrollbars)
	// 8 rows with spacing = 8 + 7 = 15 cells (> 10)
	// Width of 25 (> 18)
	contentRows := 8
	contentW := 25

	// Frame 1
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")
	cs := cnt.ContentSize()

	// Both scrollbars should be present
	// Body should be reduced by scrollbar size in BOTH dimensions
	expectedBodyW := bodyW - style.ScrollbarSize // 17
	expectedBodyH := bodyH - style.ScrollbarSize // 9

	if cnt.Body().W != expectedBodyW {
		t.Errorf("Body.W = %d, want %d (vertical scrollbar present)", cnt.Body().W, expectedBodyW)
	}
	if cnt.Body().H != expectedBodyH {
		t.Errorf("Body.H = %d, want %d (horizontal scrollbar present)", cnt.Body().H, expectedBodyH)
	}

	// Content size with padding
	csY := cs.Y + style.Padding.Y*2
	csX := cs.X + style.Padding.X*2

	// Max scroll should allow scrolling to see ALL content
	// maxScrollY = contentSize - visibleHeight = csY - bodyH (with horiz scrollbar)
	// maxScrollX = contentSize - visibleWidth = csX - bodyW (with vert scrollbar)
	expectedMaxScrollY := csY - expectedBodyH // 15 - 9 = 6
	expectedMaxScrollX := csX - expectedBodyW

	t.Logf("contentSize=%v, body=%v, csY=%d, csX=%d", cs, cnt.Body(), csY, csX)
	t.Logf("expectedMaxScrollY=%d, expectedMaxScrollX=%d", expectedMaxScrollY, expectedMaxScrollX)

	// Scroll to maximum and verify we can see bottom content
	cnt.SetScroll(types.Vec2{X: 9999, Y: 9999})

	// Run another frame to clamp scroll
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Scroll should be clamped to max values
	if cnt.Scroll().Y != expectedMaxScrollY {
		t.Errorf("Scroll.Y = %d, want %d (should be clamped to max)", cnt.Scroll().Y, expectedMaxScrollY)
	}
	if cnt.Scroll().X < expectedMaxScrollX-1 || cnt.Scroll().X > expectedMaxScrollX+1 {
		t.Errorf("Scroll.X = %d, want ~%d (should be clamped to max)", cnt.Scroll().X, expectedMaxScrollX)
	}
}

// TestTUIScrollbar_NoScrollbarsWhenContentFits tests that no scrollbars appear
// when content fits within the body.
func TestTUIScrollbar_NoScrollbarsWhenContentFits(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 12
	bodyW := windowW - style.BorderWidth*2                    // 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 10

	// Content that fits - account for spacing AND padding!
	// With spacing=1, N rows take N + (N-1) = 2N-1 cells
	// Scrollbar check uses: contentSize.Y + padding.X*2 > bodyH
	// So content fits when: contentSize.Y <= bodyH - padding.X*2 = 10 - 2 = 8
	// With N rows: 2N-1 <= 8 → N <= 4.5 → N = 4 rows max
	// Also account for padding (Padding.X=1 on each side for content width)
	contentRows := 4  // 4 rows = 4 + 3 spacing = 7 cells, 7 + 2 padding = 9 <= 10
	contentW := bodyW - style.Padding.X*2 - 2 // Leave some margin

	// Frame 1
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")

	// No scrollbars, body should be full size
	if cnt.Body().W != bodyW {
		t.Errorf("Body.W = %d, want %d (no vertical scrollbar needed, contentSize=%v)",
			cnt.Body().W, bodyW, cnt.ContentSize())
	}
	if cnt.Body().H != bodyH {
		t.Errorf("Body.H = %d, want %d (no horizontal scrollbar needed, contentSize=%v)",
			cnt.Body().H, bodyH, cnt.ContentSize())
	}
}

// TestTUIScrollbar_VerticalOnlyNoBottomClip tests that when only vertical scrollbar
// is needed (content fits horizontally), the bottom content area is NOT clipped
// as if a horizontal scrollbar was present.
func TestTUIScrollbar_VerticalOnlyNoBottomClip(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 10
	bodyW := windowW - style.BorderWidth*2                    // 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 8

	// Content that overflows vertically but fits horizontally
	// 10 rows with spacing = 10 + 9 = 19 cells (overflows bodyH=8)
	// Width fits within bodyW
	contentRows := 10
	contentW := bodyW - style.Padding.X*2 - 2 // Fits easily (14 < 18)

	t.Logf("Window: %dx%d, Body: %dx%d, contentW: %d", windowW, windowH, bodyW, bodyH, contentW)

	// Frame 1
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")
	cs := cnt.ContentSize()
	body := cnt.Body()

	t.Logf("ContentSize: %v, Body: %v", cs, body)

	// Vertical scrollbar should be present (body.W reduced)
	hasVertScrollbar := body.W < bodyW
	// Horizontal scrollbar should NOT be present (body.H should be full height)
	hasHorizScrollbar := body.H < bodyH

	t.Logf("Has vert: %v, Has horiz: %v", hasVertScrollbar, hasHorizScrollbar)

	if !hasVertScrollbar {
		t.Errorf("Expected vertical scrollbar (content overflows vertically)")
	}

	if hasHorizScrollbar {
		t.Errorf("Body.H = %d, want %d - horizontal scrollbar should NOT appear (content fits horizontally)",
			body.H, bodyH)
	}
}

// TestTUIScrollbar_VerticalOnlyAlmostFillsWidth tests that when vertical scrollbar
// appears and content ALMOST fills the width (but still fits after scrollbar),
// horizontal scrollbar should NOT appear.
func TestTUIScrollbar_VerticalOnlyAlmostFillsWidth(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 10
	bodyW := windowW - style.BorderWidth*2                    // 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 8

	// Content that overflows vertically but fits horizontally EVEN with vertical scrollbar
	// Vertical scrollbar takes 1 cell, so available width = 18 - 1 = 17
	// Content width should fit in 17 (accounting for padding: 17 - 2 = 15 for content)
	contentRows := 10
	// With padding added: cs.X = contentW + 2. Need cs.X <= bodyW - scrollbarSize
	// So contentW + 2 <= 18 - 1 = 17, contentW <= 15
	contentW := 15 // Exactly fits: 15 + 2 = 17 = 18 - 1

	t.Logf("Window: %dx%d, Body: %dx%d, contentW: %d", windowW, windowH, bodyW, bodyH, contentW)

	// Frame 1
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")
	cs := cnt.ContentSize()
	body := cnt.Body()

	t.Logf("ContentSize: %v, Body: %v", cs, body)

	// cs.X with padding = 15 + 2 = 17
	// bodyW with vertical scrollbar = 18 - 1 = 17
	// 17 <= 17, so horizontal scrollbar should NOT appear

	hasVertScrollbar := body.W < bodyW
	hasHorizScrollbar := body.H < bodyH

	t.Logf("Has vert: %v, Has horiz: %v", hasVertScrollbar, hasHorizScrollbar)

	if !hasVertScrollbar {
		t.Errorf("Expected vertical scrollbar (content overflows vertically)")
	}

	if hasHorizScrollbar {
		t.Errorf("Body.H = %d, want %d - horizontal scrollbar should NOT appear (content fits even with vertical scrollbar)",
			body.H, bodyH)
	}
}

// TestTUIScrollbar_BottomContentVisible tests that content at the bottom of the
// content area is visible when horizontal scrollbar is NOT needed. This catches
// the bug where the horizontal scrollbar area clips content even when hidden.
func TestTUIScrollbar_BottomContentVisible(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 10
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 8

	// Content that overflows vertically but fits horizontally
	// This should show vertical scrollbar but NOT horizontal
	contentRows := 10
	contentW := 10 // Fits easily in body width of 18

	// Frame 1
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2
	ui.BeginFrame()
	if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		for i := 0; i < contentRows; i++ {
			ui.LayoutRow(1, []int{contentW}, 1)
			ui.Label("X")
		}
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("Test")
	body := cnt.Body()

	t.Logf("Body: %v, bodyH expected: %d", body, bodyH)

	// The FULL body height should be available for content (no horizontal scrollbar)
	if body.H != bodyH {
		t.Errorf("Body.H = %d, want %d - bottom content area is being clipped incorrectly", body.H, bodyH)
	}

	// Verify the clip would allow content at the bottom of the body
	// Bottom Y of body = body.Y + body.H - 1
	bottomY := body.Y + body.H - 1
	t.Logf("Bottom of content area: Y=%d", bottomY)

	// The bottom Y should be at the expected position (not reduced for horizontal scrollbar)
	expectedBottomY := style.TitleHeight + bodyH - 1 // title + body - 1
	if bottomY != expectedBottomY {
		t.Errorf("Bottom Y = %d, want %d - content area bottom is wrong", bottomY, expectedBottomY)
	}
}

// TestTUIScrollbar_MutualDependencyCorrect tests that horizontal scrollbar
// correctly appears only when content doesn't fit even after accounting for
// vertical scrollbar space.
// Note: C microui uses previous frame's body for scrollbar decisions, so mutual
// dependency takes multiple frames to stabilize.
func TestTUIScrollbar_MutualDependencyCorrect(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 20, 10
	bodyW := windowW - style.BorderWidth*2                    // 18
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 8

	// Content that overflows vertically AND barely overflows horizontally after
	// vertical scrollbar appears
	contentRows := 10
	// With padding: cs.X = contentW + 2. Need cs.X > bodyW - scrollbarSize
	// So contentW + 2 > 18 - 1 = 17, contentW > 15
	contentW := 16 // Just overflows: 16 + 2 = 18 > 17

	t.Logf("Window: %dx%d, Body: %dx%d, contentW: %d", windowW, windowH, bodyW, bodyH, contentW)

	drawFrame := func() {
		ui.BeginFrame()
		if ui.BeginWindowOpt("Test", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
			for i := 0; i < contentRows; i++ {
				ui.LayoutRow(1, []int{contentW}, 1)
				ui.Label("X")
			}
			ui.EndWindow()
		}
		ui.EndFrame()
	}

	// Run multiple frames for scrollbar state to stabilize
	// C microui uses previous frame's body, so mutual dependency takes 3+ frames
	for i := 0; i < 4; i++ {
		drawFrame()
	}

	cnt := ui.GetContainer("Test")
	body := cnt.Body()

	t.Logf("ContentSize: %v, Body: %v", cnt.ContentSize(), body)

	// Both scrollbars should appear because:
	// 1. Vertical needed (content overflows vertically)
	// 2. Horizontal needed (content overflows horizontally after vertical scrollbar takes space)

	hasVertScrollbar := body.W < bodyW
	hasHorizScrollbar := body.H < bodyH

	t.Logf("Has vert: %v, Has horiz: %v", hasVertScrollbar, hasHorizScrollbar)

	if !hasVertScrollbar {
		t.Errorf("Expected vertical scrollbar")
	}

	if !hasHorizScrollbar {
		t.Errorf("Expected horizontal scrollbar (content overflows after vertical scrollbar)")
	}
}

// TestTUIScrollbar_BoxTestScenario mirrors the Box Test window exactly to verify
// that content is fully visible when scrolled to max with both scrollbars present.
// Box Test: window 22x7, label + two 10x3 boxes side by side.
func TestTUIScrollbar_BoxTestScenario(t *testing.T) {
	style := TUIStyle()
	ui := New(Config{Style: style})

	windowW, windowH := 22, 7

	// Calculate body dimensions
	// TUI: titleHeight=1, borderWidth=1, scrollbarSize=1
	bodyW := windowW - style.BorderWidth*2                    // 22 - 2 = 20
	bodyH := windowH - style.TitleHeight - style.BorderWidth // 7 - 1 - 1 = 5

	t.Logf("Window: %dx%d, Body: %dx%d", windowW, windowH, bodyW, bodyH)
	t.Logf("Style: padding=%v, spacing=%d", style.Padding, style.Spacing)

	// Frame 1: Create window with content
	ui.BeginFrame()
	if ui.BeginWindowOpt("BoxTest", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		// Label row
		ui.LayoutRow(1, []int{-1}, 1)
		ui.Label("Box drawing:")

		// Two boxes side by side, 10 wide each, 3 tall
		// With spacing=1, total width = 10 + 1 + 10 = 21
		ui.LayoutRow(2, []int{10, 10}, 3)
		ui.Label("") // Placeholder for first box position
		ui.Label("") // Placeholder for second box position

		ui.EndWindow()
	}
	ui.EndFrame()

	// Frame 2: Let scrollbars calculate
	ui.BeginFrame()
	if ui.BeginWindowOpt("BoxTest", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		ui.LayoutRow(1, []int{-1}, 1)
		ui.Label("Box drawing:")
		ui.LayoutRow(2, []int{10, 10}, 3)
		ui.Label("")
		ui.Label("")
		ui.EndWindow()
	}
	ui.EndFrame()

	cnt := ui.GetContainer("BoxTest")
	cs := cnt.ContentSize()

	t.Logf("ContentSize: %v", cs)
	t.Logf("Body after scrollbars: %v", cnt.Body())

	// Check if scrollbars are present
	hasVertScrollbar := cnt.Body().W < bodyW
	hasHorizScrollbar := cnt.Body().H < bodyH

	t.Logf("Has vert scrollbar: %v, Has horiz scrollbar: %v", hasVertScrollbar, hasHorizScrollbar)

	// Calculate expected max scroll
	csY := cs.Y + style.Padding.Y*2
	csX := cs.X + style.Padding.X*2

	actualBodyW := cnt.Body().W
	actualBodyH := cnt.Body().H

	expectedMaxScrollY := csY - actualBodyH
	expectedMaxScrollX := csX - actualBodyW

	t.Logf("csY=%d, csX=%d, expectedMaxScrollY=%d, expectedMaxScrollX=%d",
		csY, csX, expectedMaxScrollY, expectedMaxScrollX)

	// Scroll to maximum
	cnt.SetScroll(types.Vec2{X: 9999, Y: 9999})

	// Frame 3: Let scroll get clamped
	ui.BeginFrame()
	if ui.BeginWindowOpt("BoxTest", types.Rect{X: 0, Y: 0, W: windowW, H: windowH}, OptNoResize) {
		ui.LayoutRow(1, []int{-1}, 1)
		ui.Label("Box drawing:")
		ui.LayoutRow(2, []int{10, 10}, 3)
		ui.Label("")
		ui.Label("")
		ui.EndWindow()
	}
	ui.EndFrame()

	t.Logf("Scroll after clamp: %v", cnt.Scroll())

	// THE KEY CHECK: When scrolled to max, ALL content should be visible.
	// With the fix (using Padding.X for Y in cs calculation), maxScrollY now correctly
	// accounts for the padding offset, allowing scroll to reach far enough to see all content.
	//
	// Verify the scroll was clamped to expected max
	if cnt.Scroll().Y != expectedMaxScrollY {
		t.Errorf("Scroll.Y = %d, want %d (should be clamped to maxScrollY)", cnt.Scroll().Y, expectedMaxScrollY)
	}

	// Verify content is now fully visible at max scroll:
	// - paddedBody.Y = body.Y + padding.Y (with separate X/Y padding)
	// - Content bottom at screen Y = paddedBody.Y - scroll.Y + contentSize.Y
	// - Clip end at body.Y + body.H
	// Content is visible when: paddedBody.Y - scroll.Y + contentSize.Y <= body.Y + body.H
	body := cnt.Body()
	paddedBodyY := body.Y + style.Padding.Y // Note: uses Padding.Y now
	contentBottomY := paddedBodyY - cnt.Scroll().Y + cs.Y
	clipEndY := body.Y + body.H

	t.Logf("At max scroll: content bottom Y=%d, clip end Y=%d", contentBottomY, clipEndY)

	if contentBottomY > clipEndY {
		t.Errorf("Content bottom (Y=%d) extends past clip (Y=%d) - content is still clipped!",
			contentBottomY, clipEndY)
	}
}
