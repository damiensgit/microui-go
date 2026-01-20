package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestTextbox_CursorAtEndScrollsRight(t *testing.T) {
	ui := New(Config{})

	// Long text that exceeds textbox width
	buf := []byte("This is a very long text that definitely exceeds the textbox width")

	// Window at X=100, Y=20, content area starts at Y=44 (20 + 24 title bar)
	// Textbox at approximately X=100, Y=44, W=100, H=30
	// Click position should be inside the textbox: (150, 55)

	// Frame 1: Move mouse to establish hover on textbox
	ui.MouseMove(150, 55)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30) // Narrow textbox
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to focus textbox
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Press End to move cursor to end (keep mouse down to maintain focus)
	ui.KeyDown(KeyEnd)
	ui.MouseDown(150, 55, MouseLeft) // Keep focus
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Cursor should be at end
	if ui.textboxCursor != len(buf) {
		t.Errorf("cursor = %d, want %d", ui.textboxCursor, len(buf))
	}

	// Scroll offset should be positive (scrolled right)
	if ui.textboxScrollX <= 0 {
		t.Errorf("textboxScrollX = %d, should be > 0 when cursor at end of long text", ui.textboxScrollX)
	}
}

func TestTextbox_CursorAtStartScrollsLeft(t *testing.T) {
	ui := New(Config{})

	buf := []byte("This is a very long text that definitely exceeds the textbox width")

	// Frame 1: Move mouse to establish hover
	ui.MouseMove(150, 55)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to focus
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Press End to scroll right
	ui.KeyDown(KeyEnd)
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Verify we scrolled right
	if ui.textboxScrollX <= 0 {
		t.Fatalf("Expected scroll right first, got textboxScrollX = %d", ui.textboxScrollX)
	}

	// Frame 4: Press Home to move cursor to start
	ui.KeyUp(KeyEnd)
	ui.KeyDown(KeyHome)
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Cursor should be at start
	if ui.textboxCursor != 0 {
		t.Errorf("cursor = %d, want 0", ui.textboxCursor)
	}

	// Scroll offset should be 0 (scrolled back to start)
	if ui.textboxScrollX != 0 {
		t.Errorf("textboxScrollX = %d, should be 0 when cursor at start", ui.textboxScrollX)
	}
}

func TestTextbox_ScrollResetsOnFocusChange(t *testing.T) {
	ui := New(Config{})

	buf1 := []byte("Long text in first textbox that exceeds width")
	buf2 := []byte("Short")

	// Window content starts at Y=44 (20 + 24 title)
	// First textbox: Y=44 to 74
	// Second textbox: Y=79 to 109 (after 5px spacing)

	// Frame 1: Move mouse to first textbox
	ui.MouseMove(150, 55)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf1, 256)
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf2, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to focus first textbox
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf1, 256)
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf2, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Press End to scroll right
	ui.KeyDown(KeyEnd)
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf1, 256)
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf2, 256)
	ui.EndWindow()
	ui.EndFrame()

	scrollAfterEnd := ui.textboxScrollX
	if scrollAfterEnd <= 0 {
		t.Fatalf("Expected positive scroll after End key, got %d", scrollAfterEnd)
	}

	// Frame 4: Move mouse to second textbox
	ui.KeyUp(KeyEnd)
	ui.MouseUp(150, 55, MouseLeft)
	ui.MouseMove(150, 90)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf1, 256)
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf2, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 5: Click on second textbox
	ui.MouseDown(150, 90, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf1, 256)
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf2, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Scroll should reset when switching to new textbox
	if ui.textboxScrollX != 0 {
		t.Errorf("Scroll should reset to 0 when switching textbox, got %d", ui.textboxScrollX)
	}
}

func TestTextbox_ScrollFollowsCursorOnTyping(t *testing.T) {
	ui := New(Config{})

	// Start with short text
	buf := []byte("Hi")

	// Frame 1: Move mouse to textbox
	ui.MouseMove(150, 55)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{80}, 30) // Very narrow textbox
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to focus
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{80}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Type lots of text to exceed width
	ui.TextInput(" this is a lot of additional text to make it overflow")
	ui.MouseDown(150, 55, MouseLeft) // Keep focus
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{80}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Text was added, cursor moved right, scroll should follow
	textWidth := ui.style.Font.Width(string(buf))
	textboxWidth := 80 - ui.style.Padding.X*2 // Account for padding

	if textWidth > textboxWidth && ui.textboxScrollX <= 0 {
		t.Errorf("Text width (%d) exceeds textbox width (%d), but scroll is %d (should be positive)",
			textWidth, textboxWidth, ui.textboxScrollX)
	}
}

func TestTextbox_ScrollClampsToZero(t *testing.T) {
	ui := New(Config{})

	buf := []byte("Short")

	// Frame 1: Move mouse to textbox
	ui.MouseMove(200, 55)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 300, H: 100})
	ui.LayoutRow(1, []int{200}, 30) // Wide textbox, short text
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to focus
	ui.MouseDown(200, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 300, H: 100})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Press Home (cursor at 0)
	ui.KeyDown(KeyHome)
	ui.MouseDown(200, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 300, H: 100})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Scroll should never go negative
	if ui.textboxScrollX < 0 {
		t.Errorf("textboxScrollX = %d, should never be negative", ui.textboxScrollX)
	}
}

func TestTextbox_ScrollWithArrowKeys(t *testing.T) {
	ui := New(Config{})

	buf := []byte("This is a very long text that definitely exceeds the textbox width for arrow test")

	// Frame 1: Move mouse to textbox
	ui.MouseMove(150, 55)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to focus
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Press End to go to end (cursor at end, scroll right)
	ui.KeyDown(KeyEnd)
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	scrollAtEnd := ui.textboxScrollX
	cursorAtEnd := ui.textboxCursor

	// Frame 4: Press Left arrow to move cursor left
	ui.KeyUp(KeyEnd)
	ui.KeyDown(KeyLeft)
	ui.MouseDown(150, 55, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 100, Y: 20, W: 200, H: 100})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Textbox(&buf, 256)
	ui.EndWindow()
	ui.EndFrame()

	// Cursor should have moved left
	if ui.textboxCursor >= cursorAtEnd {
		t.Errorf("cursor should move left, was %d, now %d", cursorAtEnd, ui.textboxCursor)
	}

	// Scroll might stay same or decrease, but cursor should remain visible
	// (scroll should not increase when moving cursor left)
	if ui.textboxScrollX > scrollAtEnd {
		t.Errorf("scroll should not increase when moving cursor left, was %d, now %d",
			scrollAtEnd, ui.textboxScrollX)
	}
}
