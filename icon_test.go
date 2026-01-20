package microui

import (
	"image/color"
	"testing"

	"github.com/user/microui-go/types"
)

func TestDrawIcon(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 16, H: 16}
	ui.DrawIcon(IconClose, rect, color.White)

	ui.EndFrame()

	// Should have generated an icon command
	found := false
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdIcon {
			found = true
			if cmd.Icon != IconClose {
				t.Errorf("cmd.Icon = %d, want %d", cmd.Icon, IconClose)
			}
		}
	})

	if !found {
		t.Error("DrawIcon should add CmdIcon command")
	}
}

func TestDrawIcon_Properties(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 20, Y: 30, W: 24, H: 24}
	testColor := color.RGBA{R: 255, G: 128, B: 0, A: 255}
	ui.DrawIcon(IconCheck, rect, testColor)

	ui.EndFrame()

	// Verify the command properties
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdIcon {
			if cmd.Rect.X != rect.X || cmd.Rect.Y != rect.Y || cmd.Rect.W != rect.W || cmd.Rect.H != rect.H {
				t.Errorf("DrawIcon rect mismatch: got %v, want %v", cmd.Rect, rect)
			}
			if cmd.Color != testColor {
				t.Errorf("DrawIcon color mismatch: got %v, want %v", cmd.Color, testColor)
			}
			if cmd.Icon != IconCheck {
				t.Errorf("DrawIcon icon mismatch: got %d, want %d", cmd.Icon, IconCheck)
			}
		}
	})
}

func TestDrawIcon_FullyClipped(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Set small clip rect
	ui.PushClip(types.Rect{X: 100, Y: 100, W: 50, H: 50})

	// Draw icon outside clip rect
	rect := types.Rect{X: 0, Y: 0, W: 16, H: 16}
	ui.DrawIcon(IconCheck, rect, color.White)

	ui.PopClip()
	ui.EndFrame()

	// Should NOT have generated an icon command (fully clipped)
	found := false
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdIcon {
			found = true
		}
	})

	if found {
		t.Error("DrawIcon should not add command when fully clipped")
	}
}

func TestDrawIcon_PartiallyClipped(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Set clip rect that partially overlaps with icon
	clipRect := types.Rect{X: 50, Y: 50, W: 100, H: 100}
	ui.PushClip(clipRect)

	// Draw icon that partially overlaps with clip rect
	rect := types.Rect{X: 40, Y: 60, W: 32, H: 32}
	ui.DrawIcon(IconExpanded, rect, color.White)

	ui.PopClip()
	ui.EndFrame()

	// Should have generated clip and icon commands
	iconFound := false
	clipBeforeIcon := false
	clipAfterIcon := false
	lastWasIcon := false

	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdClip && !iconFound {
			clipBeforeIcon = true
		}
		if cmd.Kind == CmdIcon {
			iconFound = true
			lastWasIcon = true
		} else if lastWasIcon && cmd.Kind == CmdClip {
			clipAfterIcon = true
			lastWasIcon = false
		}
	})

	if !iconFound {
		t.Error("DrawIcon should add CmdIcon command when partially clipped")
	}
	if !clipBeforeIcon {
		t.Error("DrawIcon should add CmdClip before icon when partially clipped")
	}
	if !clipAfterIcon {
		t.Error("DrawIcon should restore clip after icon when partially clipped")
	}
}

func TestIconConstants(t *testing.T) {
	// Verify icon constants are defined and have expected values
	if IconClose != 1 {
		t.Errorf("IconClose = %d, want 1", IconClose)
	}
	if IconCheck != 2 {
		t.Errorf("IconCheck = %d, want 2", IconCheck)
	}
	if IconCollapsed != 3 {
		t.Errorf("IconCollapsed = %d, want 3", IconCollapsed)
	}
	if IconExpanded != 4 {
		t.Errorf("IconExpanded = %d, want 4", IconExpanded)
	}
	if IconMax != 6 {
		t.Errorf("IconMax = %d, want 6", IconMax)
	}
}
