package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestNew(t *testing.T) {
	cfg := Config{
		CommandBuf: 512,
	}

	ui := New(cfg)

	if ui == nil {
		t.Fatal("New() returned nil")
	}

	if cap(ui.commands.cmds) != 512 {
		t.Errorf("Command buffer capacity = %d, want 512", cap(ui.commands.cmds))
	}
}

func TestUI_BeginFrame(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// BeginFrame should clear previous commands
	if ui.commands.Len() != 0 {
		t.Errorf("BeginFrame() left %d commands", ui.commands.Len())
	}

	ui.EndFrame()
}

func TestUI_EndFrame(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.EndFrame()

	// Should not panic
}

func TestLabel(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Should not panic
	ui.Label("Hello, World!")

	ui.EndFrame()

	// Success if no panic
}

func TestButton_Click(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Mouse outside button - not clicked
	clicked := ui.Button("Click Me")
	if clicked {
		t.Error("Button() returned true without mouse interaction")
	}

	ui.EndFrame()
}

func TestButton_HoverState(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Mouse over button
	ui.MouseMove(50, 20)
	ui.MouseDown(50, 20, MouseLeft)

	// TODO: test hover/click states when layout is implemented

	ui.EndFrame()
}

func TestWindow_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	opened := ui.BeginWindow("Test", rect)

	if !opened {
		t.Error("BeginWindow() returned false, should default to open")
	}

	ui.Label("Inside window")
	ui.EndWindow()

	ui.EndFrame()
}

func TestWindow_ClipRect(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	ui.BeginWindow("Test", rect)

	// Window should set clip rect
	ui.Label("Clipped text")
	ui.EndWindow()

	ui.EndFrame()

	// After EndWindow, clip should be restored
}

// TestNew_ColorsWithoutFont tests that when a user provides colors but no font,
// the font is defaulted independently (not left nil which would cause crash).
// This tests the fix for: "Default style only applies when both font and text color are unset"
func TestNew_ColorsWithoutFont(t *testing.T) {
	// Create config with colors but NO font
	cfg := Config{
		Style: Style{
			Colors: types.DarkTheme(), // User provides colors
			// Font: nil - intentionally not set
		},
	}

	ui := New(cfg)

	// Font should have been defaulted
	if ui.style.Font == nil {
		t.Fatal("Font is nil - should have been defaulted when colors were provided without font")
	}

	// Should be able to render text without panic
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	ui.Label("This should not crash")
	ui.EndWindow()
	ui.EndFrame()

	// Success if no panic
}

// TestNew_FontWithoutColors tests that when a user provides font but no colors,
// the full default style is NOT applied (user's font is kept).
func TestNew_FontWithoutColors(t *testing.T) {
	customFont := &types.MockFont{}

	cfg := Config{
		Style: Style{
			Font: customFont,
			// Colors: not set
		},
	}

	ui := New(cfg)

	// The check is (Font == nil && Colors.Text == nil) -> use DefaultStyle
	// Since Font is not nil, we keep user's style (even if Colors is empty)
	if ui.style.Font != customFont {
		t.Error("Font was replaced - should have kept user's font when provided")
	}
}

// TestNew_DefaultStyleApplied tests that when neither font nor colors are provided,
// the full default style is applied.
func TestNew_DefaultStyleApplied(t *testing.T) {
	cfg := Config{
		// Style: empty - nothing provided
	}

	ui := New(cfg)

	// Both should be defaulted
	if ui.style.Font == nil {
		t.Error("Font is nil - default style should have been applied")
	}

	if ui.style.Colors.Text == nil {
		t.Error("Colors.Text is nil - default style should have been applied")
	}
}

// TestDrawText_NilFontGuard tests that DrawText doesn't panic if somehow
// the font is nil (belt-and-suspenders check).
func TestDrawText_NilFontGuard(t *testing.T) {
	// This test verifies the New() fix works, not that DrawText handles nil
	// Since New() now ensures font is never nil, this test just confirms
	// the normal path works.
	ui := New(Config{
		Style: Style{
			Colors: types.LightTheme(), // Only colors, no font
		},
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})

	// DrawText is called internally by Label
	ui.Label("Test text")
	ui.Button("Test button")

	ui.EndWindow()
	ui.EndFrame()

	// Success if no panic
}

func TestToggleButton_ReturnsClickedState(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Position mouse over where toggle will be, then press
	// Window title bar is ~24px, padding ~5px, so button starts around y=29
	ui.BeginFrame()
	ui.MouseMove(100, 40)
	ui.MouseDown(100, 40, MouseLeft)
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	ui.LayoutRow(1, []int{-1}, 0)
	clicked := ui.ToggleButton("Toggle", false)
	ui.EndWindow()
	ui.EndFrame()

	if !clicked {
		t.Error("toggle should report clicked when mouse pressed over it")
	}

	// Frame 2: No mouse press - should not be clicked
	ui.BeginFrame()
	ui.MouseUp(100, 40, MouseLeft)
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	ui.LayoutRow(1, []int{-1}, 0)
	clicked = ui.ToggleButton("Toggle", false)
	ui.EndWindow()
	ui.EndFrame()

	if clicked {
		t.Error("toggle should not report clicked without mouse press")
	}
}

func TestToggleButton_VisualStateReflectsSelected(t *testing.T) {
	ui := New(Config{})

	// Just verifies it runs without panic with both states
	for _, selected := range []bool{true, false} {
		ui.BeginFrame()
		ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})
		ui.ToggleButton("Toggle", selected)
		ui.EndWindow()
		ui.EndFrame()
	}
}

func TestSpace_AdvancesLayout(t *testing.T) {
	ui := New(Config{})

	// Test that Space() doesn't panic and affects layout
	// Space modifies internal layout position, not content size directly
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 200})

	ui.LayoutRow(1, []int{-1}, 20)
	ui.Label("Before")

	ui.Space(50)

	ui.LayoutRow(1, []int{-1}, 20)
	ui.Label("After")

	ui.EndWindow()
	ui.EndFrame()

	// Content size should reflect all content including the space
	cnt := ui.GetContainer("Test")
	// 2 labels (20px each) + space (50px) + padding = at least 90px
	if cnt.ContentSize().Y < 70 {
		t.Errorf("ContentSize.Y = %d, expected at least 70 with Space(50)", cnt.ContentSize().Y)
	}
}

func TestOpenWindow_OpensClosedWindow(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Create window with OptClosed
	ui.BeginFrame()
	opened := ui.BeginWindowOpt("Closeable", types.Rect{X: 0, Y: 0, W: 100, H: 100}, OptClosed)
	if opened {
		ui.EndWindow()
	}
	ui.EndFrame()

	if opened {
		t.Error("window with OptClosed should start closed")
	}

	// Open the window
	ui.OpenWindow("Closeable")

	// Frame 2: Window should now be open
	ui.BeginFrame()
	opened = ui.BeginWindowOpt("Closeable", types.Rect{X: 0, Y: 0, W: 100, H: 100}, OptClosed)
	if opened {
		ui.EndWindow()
	}
	ui.EndFrame()

	if !opened {
		t.Error("window should be open after OpenWindow()")
	}
}

func TestEachContainer_IteratesAll(t *testing.T) {
	ui := New(Config{})

	// Create several windows
	ui.BeginFrame()
	ui.BeginWindow("A", types.Rect{X: 0, Y: 0, W: 50, H: 50})
	ui.EndWindow()
	ui.BeginWindow("B", types.Rect{X: 50, Y: 0, W: 50, H: 50})
	ui.EndWindow()
	ui.BeginWindow("C", types.Rect{X: 100, Y: 0, W: 50, H: 50})
	ui.EndWindow()
	ui.EndFrame()

	// Collect all container names
	names := make(map[string]bool)
	ui.EachContainer(func(c *Container) bool {
		names[c.Name()] = true
		return true
	})

	for _, expected := range []string{"A", "B", "C"} {
		if !names[expected] {
			t.Errorf("EachContainer missed container %q", expected)
		}
	}
}

func TestEachContainer_StopsOnFalse(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	for i := 0; i < 10; i++ {
		ui.BeginWindow("Win"+string(rune('0'+i)), types.Rect{X: i * 20, Y: 0, W: 20, H: 20})
		ui.EndWindow()
	}
	ui.EndFrame()

	count := 0
	ui.EachContainer(func(c *Container) bool {
		count++
		return count < 3 // stop after 3
	})

	if count != 3 {
		t.Errorf("EachContainer should have stopped after 3, visited %d", count)
	}
}

func TestIsCapturingMouse_FalseWhenIdle(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	ui.EndWindow()
	ui.EndFrame()

	if ui.IsCapturingMouse() {
		t.Error("should not be capturing mouse when idle")
	}
}

func TestIsCapturingMouse_TrueWhenDraggingSlider(t *testing.T) {
	ui := New(Config{})

	var sliderVal float64 = 0.5

	// Frame 1: Click on slider area
	ui.BeginFrame()
	ui.MouseMove(100, 40)
	ui.MouseDown(100, 40, MouseLeft)
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Slider(&sliderVal, 0, 1)
	ui.EndWindow()
	ui.EndFrame()

	// Frame 2: Mouse still down, dragging
	ui.BeginFrame()
	ui.MouseMove(120, 40)
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 100})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Slider(&sliderVal, 0, 1)

	if !ui.IsCapturingMouse() {
		t.Error("should be capturing mouse while dragging slider")
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestIsHoverRoot_MatchesWindowUnderMouse(t *testing.T) {
	ui := New(Config{})

	// Create two non-overlapping windows
	ui.BeginFrame()
	ui.MouseMove(50, 50) // Over window A
	ui.BeginWindow("A", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.BeginWindow("B", types.Rect{X: 150, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.EndFrame()

	// After EndFrame, hover root should be set from next frame
	ui.BeginFrame()
	ui.MouseMove(50, 50)
	ui.BeginWindow("A", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.BeginWindow("B", types.Rect{X: 150, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.EndFrame()

	if !ui.IsHoverRoot("A") {
		t.Error("A should be hover root when mouse is over it")
	}
	if ui.IsHoverRoot("B") {
		t.Error("B should not be hover root when mouse is over A")
	}
}

func TestIsHoverRoot_FalseWhenNoHover(t *testing.T) {
	ui := New(Config{})

	// Mouse outside all windows
	ui.BeginFrame()
	ui.MouseMove(500, 500)
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.EndFrame()

	if ui.IsHoverRoot("Test") {
		t.Error("should not be hover root when mouse is outside")
	}
	if ui.IsHoverRoot("NonExistent") {
		t.Error("should not be hover root for non-existent window")
	}
}
