package bubbletea

import (
	"image/color"

	"github.com/user/microui-go/types"
)

// TUITheme returns a high-contrast theme optimized for terminal display.
// Colors are chosen to be distinguishable even on limited color terminals.
func TUITheme() types.ThemeColors {
	return types.ThemeColors{
		Text:         color.RGBA{R: 255, G: 255, B: 255, A: 255}, // Bright white
		Border:       color.RGBA{R: 100, G: 100, B: 100, A: 255}, // Medium gray
		WindowBg:     color.RGBA{R: 40, G: 40, B: 50, A: 255},    // Dark blue-gray
		WindowTitle:  color.RGBA{R: 60, G: 60, B: 80, A: 255},    // Slightly lighter
		WindowBorder: color.RGBA{R: 80, G: 80, B: 100, A: 255},   // Visible border
		TitleText:    color.RGBA{R: 255, G: 255, B: 255, A: 255}, // Bright white
		PanelBg:      color.RGBA{R: 30, G: 30, B: 40, A: 255},    // Darker panel
		Button:       color.RGBA{R: 70, G: 70, B: 90, A: 255},    // Button base
		ButtonHover:  color.RGBA{R: 90, G: 90, B: 120, A: 255},   // Brighter on hover
		ButtonActive: color.RGBA{R: 110, G: 110, B: 150, A: 255}, // Brightest on click
		Base:         color.RGBA{R: 50, G: 50, B: 60, A: 255},    // Input bg
		BaseHover:    color.RGBA{R: 60, G: 60, B: 70, A: 255},    // Input hover
		BaseFocus:    color.RGBA{R: 70, G: 70, B: 80, A: 255},    // Input focus
		CheckBg:      color.RGBA{R: 60, G: 60, B: 70, A: 255},    // Checkbox bg
		CheckActive:  color.RGBA{R: 80, G: 180, B: 80, A: 255},   // Green check
		ScrollBase:   color.RGBA{R: 50, G: 50, B: 60, A: 255},    // Scrollbar track
		ScrollThumb:  color.RGBA{R: 100, G: 100, B: 120, A: 255}, // Scrollbar thumb
	}
}

// BorlandTheme returns a theme inspired by Borland Turbo Vision from the 90s.
// Classic blue/cyan color scheme with high contrast.
// Note: Use with custom DrawFrame (tuiDrawFrame) to draw borders only on windows.
func BorlandTheme() types.ThemeColors {
	return types.ThemeColors{
		Text:         color.RGBA{R: 255, G: 255, B: 255, A: 255}, // Bright white
		Border:       color.RGBA{R: 0, G: 0, B: 0, A: 255},       // Black (used by custom DrawFrame for windows only)
		WindowBg:     color.RGBA{R: 0, G: 170, B: 170, A: 255},   // Cyan window background
		WindowTitle:  color.RGBA{R: 0, G: 0, B: 170, A: 255},     // Blue title bar
		WindowBorder: color.RGBA{R: 0, G: 0, B: 0, A: 255},       // Black window border
		TitleText:    color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White title text
		PanelBg:      color.RGBA{R: 0, G: 170, B: 170, A: 255},   // Cyan panel
		Button:       color.RGBA{R: 0, G: 170, B: 0, A: 255},     // Green buttons
		ButtonHover:  color.RGBA{R: 0, G: 255, B: 0, A: 255},     // Bright green hover
		ButtonActive: color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White when pressed
		Base:         color.RGBA{R: 0, G: 0, B: 170, A: 255},     // Blue input bg
		BaseHover:    color.RGBA{R: 0, G: 0, B: 200, A: 255},     // Lighter blue hover
		BaseFocus:    color.RGBA{R: 0, G: 0, B: 255, A: 255},     // Bright blue focus
		CheckBg:      color.RGBA{R: 0, G: 170, B: 170, A: 255},   // Cyan checkbox bg
		CheckActive:  color.RGBA{R: 255, G: 255, B: 0, A: 255},   // Yellow checkmark
		ScrollBase:   color.RGBA{R: 0, G: 85, B: 85, A: 255},     // Dark cyan track
		ScrollThumb:  color.RGBA{R: 0, G: 255, B: 255, A: 255},   // Bright cyan thumb
	}
}

// DesktopBlue is the classic Borland desktop background color.
var DesktopBlue = color.RGBA{R: 0, G: 0, B: 168, A: 255}

// DesktopCyan is the lighter color for the dithered pattern.
var DesktopCyan = color.RGBA{R: 0, G: 170, B: 170, A: 255}

// DesktopPattern is the light shade character for the dithered background.
const DesktopPattern = '░' // U+2591 Light Shade

// Scrollbar characters for classic Turbo Vision look
const (
	ScrollTrackChar = '░' // U+2591 Light Shade - scrollbar track
	ScrollThumbChar = '█' // U+2588 Full Block - scrollbar thumb
)

// Scrollbar colors - thumb must contrast with track for visibility
var (
	// Track: subtle cyan pattern on blue
	ScrollTrackFg = color.RGBA{R: 0, G: 128, B: 128, A: 255} // Dim cyan
	ScrollTrackBg = color.RGBA{R: 0, G: 0, B: 128, A: 255}   // Dark blue

	// Thumb: bright/white block that stands out
	ScrollThumbFg = color.RGBA{R: 0, G: 0, B: 0, A: 255}     // Black (char color, not visible for █)
	ScrollThumbBg = color.RGBA{R: 0, G: 255, B: 255, A: 255} // Bright cyan background
)

// Status bar colors - classic Turbo Vision style
var (
	StatusBarFg = color.RGBA{R: 0, G: 0, B: 0, A: 255}     // Black text
	StatusBarBg = color.RGBA{R: 0, G: 170, B: 170, A: 255} // Cyan background
)
