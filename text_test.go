package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestText_WordWrap(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)

	// Long text should word-wrap
	ui.Text("This is a long text that should wrap to multiple lines within the available width.")

	ui.EndWindow()
	ui.EndFrame()

	// Should generate multiple text commands
	textCmdCount := 0
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdText {
			textCmdCount++
		}
	})

	// At minimum we need more than one text command for wrapped text
	// (title bar + at least 2 lines of wrapped text)
	if textCmdCount < 3 {
		t.Errorf("Expected at least 3 text commands, got %d (text not wrapping?)", textCmdCount)
	}
}

func TestText_SingleLine(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 800, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)

	// Short text should fit in one line
	ui.Text("Short text")

	ui.EndWindow()
	ui.EndFrame()

	// Count text commands (1 for window title, 1 for text)
	textCmdCount := 0
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdText {
			textCmdCount++
		}
	})

	// Should have exactly 2 text commands (window title + text)
	if textCmdCount != 2 {
		t.Errorf("Expected 2 text commands, got %d", textCmdCount)
	}
}
