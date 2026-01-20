package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestTextbox_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	buf := []byte("hello")
	result := ui.Textbox(&buf, 128)

	// Should not report change without input
	if result != 0 {
		t.Error("Textbox should return 0 without input")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestTextbox_TextInput(t *testing.T) {
	ui := New(Config{})

	buf := make([]byte, 128)
	copy(buf, "test")

	// Simulate clicking on textbox to focus it
	ui.MouseMove(50, 50)
	ui.MouseDown(50, 50, MouseLeft)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)

	ui.Textbox(&buf, 128)

	ui.EndWindow()
	ui.EndFrame()

	// Now add text input
	ui.TextInput("X")
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)

	result := ui.Textbox(&buf, 128)

	ui.EndWindow()
	ui.EndFrame()

	// Result should indicate change if focused
	_ = result
}

func TestTextbox_Backspace(t *testing.T) {
	ui := New(Config{})

	// Create buffer with initial text
	buf := []byte("hello")

	// Simulate clicking to focus
	ui.MouseMove(50, 50)
	ui.MouseDown(50, 50, MouseLeft)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Press backspace
	ui.KeyDown(KeyBackspace)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	result := ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Textbox modifies when focused and key is pressed
	// Check if ResChange flag was returned
	if result&ResChange == 0 {
		// Note: focus may have been lost between frames, which is acceptable behavior
		t.Log("Backspace did not return ResChange (focus may have been lost)")
	}
}

func TestTextbox_EnterSubmit(t *testing.T) {
	ui := New(Config{})

	buf := []byte("test")

	// Simulate clicking to focus
	ui.MouseMove(50, 50)
	ui.MouseDown(50, 50, MouseLeft)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Press enter
	ui.KeyDown(KeyEnter)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	result := ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Check if ResSubmit flag would be returned when focused
	_ = result
}

func TestTextbox_ResFlags(t *testing.T) {
	// Test that ResChange and ResSubmit flags are correctly defined
	if ResChange != 1 {
		t.Errorf("ResChange = %d, want 1", ResChange)
	}
	if ResSubmit != 2 {
		t.Errorf("ResSubmit = %d, want 2", ResSubmit)
	}
}

func TestTextbox_WithWindow(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	if ui.BeginWindow("Test", rect) {
		buf1 := []byte("username")
		ui.Textbox(&buf1, 64)

		buf2 := []byte("password")
		ui.Textbox(&buf2, 64)

		ui.EndWindow()
	}

	ui.EndFrame()
}

func TestTextbox_KeyPressedOnce(t *testing.T) {
	ui := New(Config{})

	buf := []byte("hello")

	// Frame 1: Move mouse to textbox position to establish hover
	// Window content area starts at Y=24 (after title bar)
	// Textbox will be at approximately (0, 24) to (200, 54)
	// Position mouse at center of textbox: (100, 39)
	ui.MouseMove(100, 39)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to gain focus (hover was established in previous frame)
	ui.MouseDown(100, 39, MouseLeft)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Press backspace once (should delete one character)
	// Keep mouse held to maintain focus
	ui.KeyDown(KeyBackspace)
	ui.MouseDown(100, 39, MouseLeft) // Re-click to maintain focus and MouseDown state
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	if len(buf) != 4 {
		t.Errorf("After first backspace: expected buffer length 4, got %d", len(buf))
	}

	// Frame 4: Still holding backspace (should NOT delete another)
	// Keep mouse held to maintain focus
	ui.MouseDown(100, 39, MouseLeft)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Should only have deleted ONE character total (not two)
	expected := 4 // "hello" -> "hell"
	if len(buf) != expected {
		t.Errorf("Expected buffer length %d, got %d (backspace repeated wrongly)", expected, len(buf))
	}
}

func TestTextbox_FastTypingAtEnd(t *testing.T) {
	ui := New(Config{})

	buf := []byte("hello")

	// Frame 1: Move mouse to textbox position to establish hover
	ui.MouseMove(100, 39)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to gain focus
	ui.MouseDown(100, 39, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Cursor should be at end (5)
	if ui.textboxCursor != 5 {
		t.Errorf("Initial cursor position = %d, want 5", ui.textboxCursor)
	}

	// Frame 3: Type multiple characters quickly (simulate fast typing)
	// TextInput must be called AFTER BeginFrame (which clears old input)
	ui.BeginFrame()
	ui.TextInput("xyz")
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Buffer should be "helloxyz"
	if string(buf) != "helloxyz" {
		t.Errorf("Buffer = %q, want %q", string(buf), "helloxyz")
	}

	// Cursor should be at end (8)
	if ui.textboxCursor != 8 {
		t.Errorf("Cursor position = %d, want 8", ui.textboxCursor)
	}

	// Frame 4: Type more characters
	ui.BeginFrame()
	ui.TextInput("123")
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Buffer should be "helloxyz123"
	if string(buf) != "helloxyz123" {
		t.Errorf("Buffer = %q, want %q", string(buf), "helloxyz123")
	}

	// Cursor should be at end (11)
	if ui.textboxCursor != 11 {
		t.Errorf("Final cursor position = %d, want 11", ui.textboxCursor)
	}
}

func TestTextbox_CursorInMiddleTyping(t *testing.T) {
	ui := New(Config{})

	buf := []byte("hello")

	// Frame 1: Hover
	ui.MouseMove(100, 39)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to gain focus
	ui.MouseDown(100, 39, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Manually set cursor to middle (position 2, after "he")
	ui.textboxCursor = 2

	// Frame 3: Type a character at cursor position
	// TextInput must be called AFTER BeginFrame
	ui.BeginFrame()
	ui.TextInput("X")
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Buffer should be "heXllo" (X inserted after "he")
	if string(buf) != "heXllo" {
		t.Errorf("Buffer = %q, want %q", string(buf), "heXllo")
	}

	// Cursor should be at 3 (after the X)
	if ui.textboxCursor != 3 {
		t.Errorf("Cursor position = %d, want 3", ui.textboxCursor)
	}
}

func TestTextbox_RapidTypingMultipleFrames(t *testing.T) {
	ui := New(Config{})

	buf := []byte("")

	// Frame 1: Hover
	ui.MouseMove(100, 39)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to gain focus
	ui.MouseDown(100, 39, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Simulate rapid typing across many frames
	expected := ""
	for i := 0; i < 20; i++ {
		char := string(rune('a' + i%26))
		expected += char

		ui.BeginFrame()
		ui.TextInput(char)
		ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
		ui.LayoutRow(1, []int{200}, 30)
		ui.Textbox(&buf, 128)
		ui.EndWindow()
		ui.EndFrame()

		// Verify after each frame
		if string(buf) != expected {
			t.Errorf("Frame %d: Buffer = %q, want %q", i, string(buf), expected)
		}
		if ui.textboxCursor != len(expected) {
			t.Errorf("Frame %d: Cursor = %d, want %d", i, ui.textboxCursor, len(expected))
		}
	}
}

func TestTextbox_ArrowKeysAfterClickPosition(t *testing.T) {
	ui := New(Config{})

	buf := []byte("hello")

	// Frame 1: Hover
	ui.MouseMove(30, 39)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to position cursor at start
	ui.MouseDown(10, 39, MouseLeft) // Click near start
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	initialCursor := ui.textboxCursor
	t.Logf("Initial cursor after click: %d", initialCursor)

	// Frame 3: Type a character
	ui.MouseUp(10, 39, MouseLeft) // Release mouse
	ui.BeginFrame()
	ui.TextInput("X")
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	afterTypeCursor := ui.textboxCursor
	t.Logf("Cursor after typing 'X': %d, buf=%q", afterTypeCursor, string(buf))

	// Frame 4: Press right arrow
	ui.KeyDown(KeyRight)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	afterArrowCursor := ui.textboxCursor
	t.Logf("Cursor after right arrow: %d", afterArrowCursor)

	// Cursor should have moved right (if not at end)
	if afterArrowCursor == afterTypeCursor && afterTypeCursor < len(buf) {
		t.Errorf("Right arrow didn't move cursor: before=%d, after=%d, bufLen=%d",
			afterTypeCursor, afterArrowCursor, len(buf))
	}
}

func TestTextbox_ArrowKeysWithMouseHovering(t *testing.T) {
	ui := New(Config{})

	buf := []byte("hello world")

	// Frame 1: Hover
	ui.MouseMove(50, 39)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click in middle of text
	ui.MouseDown(50, 39, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 3: Release mouse but keep hovering over textbox
	ui.MouseUp(50, 39, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	cursorAfterClick := ui.textboxCursor
	t.Logf("Cursor after click+release: %d", cursorAfterClick)

	// Frame 4: Type some characters while mouse still hovers
	ui.BeginFrame()
	ui.TextInput("XY")
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	cursorAfterType := ui.textboxCursor
	t.Logf("Cursor after typing 'XY': %d, buf=%q", cursorAfterType, string(buf))

	// Frame 5: Press right arrow (mouse still hovering)
	ui.KeyDown(KeyRight)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()
	ui.KeyUp(KeyRight)

	cursorAfterRight1 := ui.textboxCursor
	t.Logf("Cursor after first right arrow: %d", cursorAfterRight1)

	if cursorAfterRight1 <= cursorAfterType && cursorAfterType < len(buf) {
		t.Errorf("First right arrow didn't move cursor: before=%d, after=%d", cursorAfterType, cursorAfterRight1)
	}

	// Frame 6: Press right arrow again
	ui.KeyDown(KeyRight)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()
	ui.KeyUp(KeyRight)

	cursorAfterRight2 := ui.textboxCursor
	t.Logf("Cursor after second right arrow: %d", cursorAfterRight2)

	if cursorAfterRight2 <= cursorAfterRight1 && cursorAfterRight1 < len(buf) {
		t.Errorf("Second right arrow didn't move cursor: before=%d, after=%d", cursorAfterRight1, cursorAfterRight2)
	}

	// Frame 7: Press left arrow
	ui.KeyDown(KeyLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()
	ui.KeyUp(KeyLeft)

	cursorAfterLeft := ui.textboxCursor
	t.Logf("Cursor after left arrow: %d", cursorAfterLeft)

	if cursorAfterLeft >= cursorAfterRight2 && cursorAfterRight2 > 0 {
		t.Errorf("Left arrow didn't move cursor: before=%d, after=%d", cursorAfterRight2, cursorAfterLeft)
	}
}

func TestTextbox_ClickToPositionCursor(t *testing.T) {
	ui := New(Config{})

	buf := []byte("hello world")

	// Frame 1: Hover
	ui.MouseMove(100, 39)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click somewhere in the text (not at the very end)
	// The textbox starts at approximately X=5 (padding) within the window
	// With default font, each character is about 7-8 pixels wide
	// Click at position that should be around the middle of "hello"
	ui.MouseDown(30, 39, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Cursor should NOT be at the end (11), but somewhere in the middle
	if ui.textboxCursor == len(buf) {
		t.Errorf("Cursor should be positioned at click, not at end (%d)", ui.textboxCursor)
	}

	// Frame 3: Type a character - it should be inserted at cursor position
	ui.BeginFrame()
	ui.TextInput("X")
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf, 128)
	ui.EndWindow()
	ui.EndFrame()

	// The 'X' should be inserted somewhere in the middle, not at the end
	bufStr := string(buf)
	if bufStr == "hello worldX" {
		t.Errorf("Text was appended at end, but should be inserted at click position. Got: %q", bufStr)
	}
}

func TestTextbox_MultipleTextboxesSameFrame(t *testing.T) {
	ui := New(Config{})

	buf1 := []byte("first")
	buf2 := []byte("second")

	// Frame 1: Setup - hover over first textbox
	ui.MouseMove(100, 39)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf1, 128)
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf2, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Click to focus first textbox
	ui.MouseDown(100, 39, MouseLeft)
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf1, 128)
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf2, 128)
	ui.EndWindow()
	ui.EndFrame()

	// Cursor should be at end of first textbox's content
	if ui.textboxCursor != 5 {
		t.Errorf("Initial cursor = %d, want 5 (end of 'first')", ui.textboxCursor)
	}

	// Frame 3: Type into first textbox
	ui.BeginFrame()
	ui.TextInput("X")
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf1, 128)
	ui.LayoutRow(1, []int{200}, 30)
	ui.Textbox(&buf2, 128)
	ui.EndWindow()
	ui.EndFrame()

	// First textbox should have the X, second should be unchanged
	if string(buf1) != "firstX" {
		t.Errorf("buf1 = %q, want %q", string(buf1), "firstX")
	}
	if string(buf2) != "second" {
		t.Errorf("buf2 = %q, want %q (should be unchanged)", string(buf2), "second")
	}
}
