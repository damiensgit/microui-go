package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestDrawFrame_CustomCallback(t *testing.T) {
	callCount := 0
	customDrawFrame := func(ui *UI, rect types.Rect, colorID int) {
		callCount++
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Click") // Should trigger draw_frame
	ui.EndWindow()
	ui.EndFrame()

	if callCount == 0 {
		t.Error("Custom DrawFrame callback was never called")
	}
}

func TestDrawFrame_DefaultProducesCommands(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Click")
	ui.EndWindow()
	ui.EndFrame()

	// Should have rect commands from default draw_frame
	hasRect := false
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdRect {
			hasRect = true
		}
	})

	if !hasRect {
		t.Error("Default DrawFrame should produce CmdRect commands")
	}
}

func TestDrawFrame_ColorIDsProvided(t *testing.T) {
	colorIDs := make(map[int]bool)
	customDrawFrame := func(ui *UI, rect types.Rect, colorID int) {
		colorIDs[colorID] = true
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Normal") // Button color
	ui.EndWindow()
	ui.EndFrame()

	// Should have received button color ID
	if !colorIDs[ColorButton] && !colorIDs[ColorBase] {
		t.Error("DrawFrame should receive control color IDs")
	}
}

func TestDrawFrame_ColorConstants(t *testing.T) {
	// Verify color constants are defined correctly (matching C microui ordering)
	expectedValues := map[int]string{
		ColorText:        "ColorText",
		ColorBorder:      "ColorBorder",
		ColorWindowBG:    "ColorWindowBG",
		ColorTitleBG:     "ColorTitleBG",
		ColorTitleText:   "ColorTitleText",
		ColorPanelBG:     "ColorPanelBG",
		ColorButton:      "ColorButton",
		ColorButtonHover: "ColorButtonHover",
		ColorButtonFocus: "ColorButtonFocus",
		ColorBase:        "ColorBase",
		ColorBaseHover:   "ColorBaseHover",
		ColorBaseFocus:   "ColorBaseFocus",
		ColorScrollBase:  "ColorScrollBase",
		ColorScrollThumb: "ColorScrollThumb",
	}

	// Check that each constant has a unique value
	seen := make(map[int]string)
	for value, name := range expectedValues {
		if existing, ok := seen[value]; ok {
			t.Errorf("Color constant %s has same value as %s (%d)", name, existing, value)
		}
		seen[value] = name
	}

	// Check constants are in expected order (iota)
	if ColorText != 0 {
		t.Errorf("ColorText should be 0, got %d", ColorText)
	}
}

func TestDrawFrame_GetColorByID(t *testing.T) {
	ui := New(Config{})

	// Test that getColorByID returns appropriate colors
	tests := []struct {
		colorID int
		name    string
	}{
		{ColorButton, "ColorButton"},
		{ColorButtonHover, "ColorButtonHover"},
		{ColorBase, "ColorBase"},
		{ColorBaseHover, "ColorBaseHover"},
		{ColorBaseFocus, "ColorBaseFocus"},
		{ColorScrollBase, "ColorScrollBase"},
		{ColorScrollThumb, "ColorScrollThumb"},
		{ColorTitleBG, "ColorTitleBG"},
		{ColorWindowBG, "ColorWindowBG"},
		{ColorPanelBG, "ColorPanelBG"},
		{ColorText, "ColorText"},
	}

	for _, tt := range tests {
		c := ui.GetColorByID(tt.colorID)
		if c == nil {
			t.Errorf("GetColorByID(%s) returned nil", tt.name)
		}
	}
}

func TestDrawFrame_NilCallbackUsesDefault(t *testing.T) {
	ui := New(Config{
		DrawFrame: nil, // Explicitly nil
	})

	// Should not panic and should use default
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Click")
	ui.EndWindow()
	ui.EndFrame()

	// Should have rect commands from default draw_frame
	hasRect := false
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdRect {
			hasRect = true
		}
	})

	if !hasRect {
		t.Error("Default DrawFrame (from nil config) should produce CmdRect commands")
	}
}

func TestDrawFrame_CalledWithCorrectRect(t *testing.T) {
	var capturedRect types.Rect
	customDrawFrame := func(ui *UI, rect types.Rect, colorID int) {
		if colorID == ColorButton {
			capturedRect = rect
		}
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Button("Click")
	ui.EndWindow()
	ui.EndFrame()

	// Button should have been rendered with non-zero dimensions
	if capturedRect.W == 0 || capturedRect.H == 0 {
		t.Errorf("DrawFrame should be called with valid rect, got %+v", capturedRect)
	}
}

func TestDrawFrame_PublicMethod(t *testing.T) {
	callCount := 0
	customDrawFrame := func(ui *UI, rect types.Rect, colorID int) {
		callCount++
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	// Test that DrawFrame public method works
	testRect := types.Rect{X: 10, Y: 20, W: 100, H: 50}
	ui.DrawFrame(testRect, ColorButton)

	if callCount != 1 {
		t.Errorf("DrawFrame method should call callback once, got %d calls", callCount)
	}
}
