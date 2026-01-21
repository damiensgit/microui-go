package types

import (
	"testing"
)

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
