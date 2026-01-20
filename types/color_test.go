package types

import (
	"image/color"
	"testing"
)

func TestRGBA_FromStdLib(t *testing.T) {
	std := color.RGBA{R: 255, G: 128, B: 64, A: 255}
	c := RGBAFromColor(std)

	if c.R != 255 || c.G != 128 || c.B != 64 || c.A != 255 {
		t.Errorf("RGBAFromColor() = %v, want {255, 128, 64, 255}", c)
	}
}

func TestRGBA_ToStdLib(t *testing.T) {
	c := RGBA{R: 255, G: 128, B: 64, A: 255}
	std := c.ToColor()

	r, g, b, a := std.RGBA()
	if r != 65535 || g != 32896 || b != 16448 || a != 65535 {
		t.Errorf("ToColor().RGBA() = %d, %d, %d, %d", r, g, b, a)
	}
}

func TestRGBA_Premultiply(t *testing.T) {
	c := RGBA{R: 255, G: 128, B: 64, A: 128}
	got := c.Premultiply()

	// 255 * 128 / 255 = 128
	// 128 * 128 / 255 = 64
	// 64 * 128 / 255 = 32
	if got.R != 128 || got.G != 64 || got.B != 32 || got.A != 128 {
		t.Errorf("Premultiply() = %v, want {128, 64, 32, 128}", got)
	}
}

func TestDarkThemeColors(t *testing.T) {
	dark := DarkTheme()
	// Check that colors are not nil
	if dark.Text == nil {
		t.Error("DarkTheme() Text should not be nil")
	}
	if dark.WindowBg == nil {
		t.Error("DarkTheme() WindowBg should not be nil")
	}
}

func TestLightThemeColors(t *testing.T) {
	light := LightTheme()
	// Check that colors are not nil
	if light.Text == nil {
		t.Error("LightTheme() Text should not be nil")
	}
	if light.WindowBg == nil {
		t.Error("LightTheme() WindowBg should not be nil")
	}
}
