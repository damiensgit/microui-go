package microui

import (
	"image/color"
	"testing"

	"github.com/user/microui-go/types"
)

func TestDrawBox(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.DrawBox(types.Rect{X: 10, Y: 10, W: 100, H: 50}, color.White)

	ui.EndFrame()

	// Should have generated a box command
	found := false
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdBox {
			found = true
		}
	})

	if !found {
		t.Error("DrawBox should add CmdBox command")
	}
}

func TestDrawBox_Properties(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 20, Y: 30, W: 150, H: 80}
	testColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	ui.DrawBox(rect, testColor)

	ui.EndFrame()

	// Verify the command properties
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdBox {
			if cmd.Rect.X != rect.X || cmd.Rect.Y != rect.Y || cmd.Rect.W != rect.W || cmd.Rect.H != rect.H {
				t.Errorf("DrawBox rect mismatch: got %v, want %v", cmd.Rect, rect)
			}
			if cmd.Color != testColor {
				t.Errorf("DrawBox color mismatch: got %v, want %v", cmd.Color, testColor)
			}
		}
	})
}
