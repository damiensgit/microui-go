package types

import (
	"image/color"
)

// RGBA represents a color in RGBA format.
// Values are 0-255.
type RGBA struct {
	R, G, B, A uint8
}

// RGBAFromColor creates a types.RGBA from standard color.Color.
func RGBAFromColor(c color.Color) RGBA {
	if c == nil {
		return RGBA{}
	}
	r, g, b, a := c.RGBA()
	return RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

// ToColor converts to standard color.Color.
func (c RGBA) ToColor() color.Color {
	return color.RGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: c.A,
	}
}

// Premultiply returns alpha-premultiplied color.
func (c RGBA) Premultiply() RGBA {
	if c.A == 255 {
		return c
	}
	a := uint16(c.A)
	return RGBA{
		R: uint8((uint16(c.R) * a) / 255),
		G: uint8((uint16(c.G) * a) / 255),
		B: uint8((uint16(c.B) * a) / 255),
		A: c.A,
	}
}

// Common colors
var (
	ColorTransparent = RGBA{A: 0}
	ColorBlack       = RGBA{R: 0, G: 0, B: 0, A: 255}
	ColorWhite       = RGBA{R: 255, G: 255, B: 255, A: 255}
)

// DarkTheme returns the default dark theme colors.
func DarkTheme() ThemeColors {
	return ThemeColors{
		Text:         color.RGBA{R: 230, G: 230, B: 230, A: 255},
		Border:       color.RGBA{R: 25, G: 25, B: 25, A: 255},
		WindowBg:     color.RGBA{R: 50, G: 50, B: 50, A: 255},
		WindowTitle:  color.RGBA{R: 25, G: 25, B: 25, A: 255},
		WindowBorder: color.RGBA{R: 25, G: 25, B: 25, A: 255},
		TitleText:    color.RGBA{R: 240, G: 240, B: 240, A: 255},
		PanelBg:      color.RGBA{R: 0, G: 0, B: 0, A: 0},
		Button:       color.RGBA{R: 75, G: 75, B: 75, A: 255},
		ButtonHover:  color.RGBA{R: 95, G: 95, B: 95, A: 255},
		ButtonActive: color.RGBA{R: 115, G: 115, B: 115, A: 255},
		Base:         color.RGBA{R: 30, G: 30, B: 30, A: 255},
		BaseHover:    color.RGBA{R: 35, G: 35, B: 35, A: 255},
		BaseFocus:    color.RGBA{R: 40, G: 40, B: 40, A: 255},
		CheckBg:      color.RGBA{R: 60, G: 60, B: 60, A: 255},
		CheckActive:  color.RGBA{R: 100, G: 180, B: 100, A: 255},
		ScrollBase:   color.RGBA{R: 43, G: 43, B: 43, A: 255},
		ScrollThumb:  color.RGBA{R: 30, G: 30, B: 30, A: 255},
	}
}

// LightTheme returns the default light theme colors.
func LightTheme() ThemeColors {
	return ThemeColors{
		Text:         color.RGBA{R: 30, G: 30, B: 30, A: 255},
		Border:       color.RGBA{R: 180, G: 180, B: 180, A: 255},
		WindowBg:     color.RGBA{R: 240, G: 240, B: 240, A: 255},
		WindowTitle:  color.RGBA{R: 220, G: 220, B: 220, A: 255},
		WindowBorder: color.RGBA{R: 180, G: 180, B: 180, A: 255},
		TitleText:    color.RGBA{R: 30, G: 30, B: 30, A: 255},
		PanelBg:      color.RGBA{R: 235, G: 235, B: 235, A: 255},
		Button:       color.RGBA{R: 200, G: 200, B: 200, A: 255},
		ButtonHover:  color.RGBA{R: 180, G: 180, B: 180, A: 255},
		ButtonActive: color.RGBA{R: 160, G: 160, B: 160, A: 255},
		Base:         color.RGBA{R: 230, G: 230, B: 230, A: 255},
		BaseHover:    color.RGBA{R: 220, G: 220, B: 220, A: 255},
		BaseFocus:    color.RGBA{R: 210, G: 210, B: 210, A: 255},
		CheckBg:      color.RGBA{R: 210, G: 210, B: 210, A: 255},
		CheckActive:  color.RGBA{R: 60, G: 140, B: 60, A: 255},
		ScrollBase:   color.RGBA{R: 220, G: 220, B: 220, A: 255},
		ScrollThumb:  color.RGBA{R: 140, G: 140, B: 140, A: 255},
	}
}

// ThemeColors contains all color values for theming.
type ThemeColors struct {
	Text         color.Color
	Border       color.Color
	WindowBg     color.Color
	WindowTitle  color.Color
	WindowBorder color.Color
	TitleText    color.Color // Title bar text
	PanelBg      color.Color
	Button       color.Color
	ButtonHover  color.Color
	ButtonActive color.Color
	Base         color.Color // Generic control bg
	BaseHover    color.Color // Generic control hover
	BaseFocus    color.Color // Generic control focus
	CheckBg      color.Color
	CheckActive  color.Color
	ScrollBase   color.Color // Scrollbar track
	ScrollThumb  color.Color // Scrollbar thumb
}
