package retro

import "image/color"

// SkeuoColor represents a color with highlight and shadow variants for 3D beveling.
type SkeuoColor struct {
	Base      color.Color // Main fill color
	Highlight color.Color // Light edge (top/left) - raised effect
	Shadow    color.Color // Dark edge (bottom/right) - raised effect
}

// Theme defines the complete color scheme and layout for the retro renderer.
// All dimensional values are in LOGICAL PIXELS. Use Px() to convert to screen pixels.
type Theme struct {
	// === COLORS ===

	// Window/panel backgrounds
	Panel SkeuoColor

	// Buttons
	Button      SkeuoColor
	ButtonHover SkeuoColor

	// Input fields (sunken appearance)
	Input      SkeuoColor
	InputFocus SkeuoColor

	// Title bar
	TitleBar     SkeuoColor
	TitleBarText color.Color

	// General text
	Text    color.Color
	TextDim color.Color // Disabled/secondary text

	// Canvas/drawing area (sunken)
	Canvas SkeuoColor

	// Scrollbar
	ScrollTrack SkeuoColor
	ScrollThumb SkeuoColor

	// Background behind windows
	Background color.Color

	// === LAYOUT (all in LOGICAL pixels) ===

	// PixelScale: screen pixels per logical pixel.
	// Set to 2 for chunky retro look, 1 for crisp, 3 for extra chunky.
	PixelScale int

	// BorderSize: thickness of border on each edge (2 = outer + inner)
	BorderSize int

	// CornerRadius: corner notch size (0 = square, 1 = subtle notch)
	CornerRadius int

	// ControlHeight: height of buttons, inputs, sliders
	ControlHeight int

	// TitleHeight: height of window title bar content (excluding border inset)
	TitleHeight int

	// Padding: internal spacing inside controls {X: horizontal, Y: vertical}
	Padding struct{ X, Y int }

	// ScrollbarWidth: width of scrollbar track (logical pixels)
	ScrollbarWidth int

	// ScrollbarMargin: margin around scrollbar on all sides (logical pixels)
	// Total scrollbar area = margin + width + margin
	// e.g., 2 + 4 + 2 = 8 logical pixels horizontal space for vertical scrollbar
	ScrollbarMargin int

	// ThumbWidth: width of slider thumb (logical pixels, including borders)
	ThumbWidth int

	// Spacing: gap between controls
	Spacing int

	// === EFFECTS ===

	// ShadowAlpha: opacity for drop shadows (0-255, 0 = no shadow)
	ShadowAlpha uint8

	// UseFlat: use flat double-border style instead of 3D bevels
	UseFlat bool
}

// Px converts logical pixels to screen pixels.
// Example: theme.Px(2) returns 4 when PixelScale is 2.
func (t *Theme) Px(logical int) int {
	scale := t.PixelScale
	if scale < 1 {
		scale = 1
	}
	return logical * scale
}

// ShadowColor returns the drop shadow color using Panel.Shadow with ShadowAlpha.
func (t *Theme) ShadowColor() color.Color {
	if t.ShadowAlpha == 0 {
		return nil
	}
	r, g, b, _ := t.Panel.Shadow.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: t.ShadowAlpha,
	}
}

// StyleTitleHeight returns the total title height for microui style
// (content height + border inset at top).
func (t *Theme) StyleTitleHeight() int {
	return t.Px(t.TitleHeight + t.BorderSize)
}

// StyleControlHeight returns the control height for microui style.
func (t *Theme) StyleControlHeight() int {
	return t.Px(t.ControlHeight)
}

// StylePadding returns padding for microui style.
func (t *Theme) StylePadding() (x, y int) {
	return t.Px(t.Padding.X), t.Px(t.Padding.Y)
}

// StyleScrollbarWidth returns scrollbar width for microui style.
func (t *Theme) StyleScrollbarWidth() int {
	return t.Px(t.ScrollbarWidth)
}

// StyleScrollbarMargin returns scrollbar margin for microui style.
func (t *Theme) StyleScrollbarMargin() int {
	return t.Px(t.ScrollbarMargin)
}

// StyleScrollbarTotalWidth returns total scrollbar area width (margin + track + margin).
func (t *Theme) StyleScrollbarTotalWidth() int {
	return t.Px(t.ScrollbarMargin*2 + t.ScrollbarWidth)
}

// StyleSpacing returns spacing for microui style.
func (t *Theme) StyleSpacing() int {
	return t.Px(t.Spacing)
}

// StyleWindowBorder returns the window border width for content clipping.
// Content should be clipped to not render into the visual window border.
func (t *Theme) StyleWindowBorder() int {
	return t.Px(t.BorderSize)
}

// StyleControlMargin returns the control margin (visual border width).
// This is the clipping boundary - text is clipped inside this.
func (t *Theme) StyleControlMargin() int {
	return t.Px(t.BorderSize)
}

// StyleControlPadding returns additional padding inside the control border.
// Content is positioned inside margin + padding.
func (t *Theme) StyleControlPadding() int {
	// For pixel theme: no additional padding, text goes right up to border interior
	return 0
}

// StyleThumbSize returns the slider thumb width for microui style.
func (t *Theme) StyleThumbSize() int {
	return t.Px(t.ThumbWidth)
}

// StyleScrollbarBorder returns the border width for scrollbar positioning.
// Scrollbars must clear this distance from the window edge.
func (t *Theme) StyleScrollbarBorder() int {
	return t.Px(t.BorderSize)
}

// DarkTheme returns a dark skeuomorphic theme inspired by game editors.
func DarkTheme() *Theme {
	return &Theme{
		Panel: SkeuoColor{
			Base:      color.RGBA{R: 45, G: 45, B: 48, A: 255},
			Highlight: color.RGBA{R: 65, G: 65, B: 70, A: 255},
			Shadow:    color.RGBA{R: 25, G: 25, B: 28, A: 255},
		},
		Button: SkeuoColor{
			Base:      color.RGBA{R: 60, G: 60, B: 65, A: 255},
			Highlight: color.RGBA{R: 85, G: 85, B: 90, A: 255},
			Shadow:    color.RGBA{R: 35, G: 35, B: 38, A: 255},
		},
		ButtonHover: SkeuoColor{
			Base:      color.RGBA{R: 70, G: 70, B: 75, A: 255},
			Highlight: color.RGBA{R: 95, G: 95, B: 100, A: 255},
			Shadow:    color.RGBA{R: 45, G: 45, B: 48, A: 255},
		},
		Input: SkeuoColor{
			Base:      color.RGBA{R: 30, G: 30, B: 33, A: 255},
			Highlight: color.RGBA{R: 20, G: 20, B: 22, A: 255}, // Inverted for sunken
			Shadow:    color.RGBA{R: 50, G: 50, B: 55, A: 255}, // Inverted for sunken
		},
		InputFocus: SkeuoColor{
			Base:      color.RGBA{R: 35, G: 35, B: 40, A: 255},
			Highlight: color.RGBA{R: 20, G: 20, B: 22, A: 255},
			Shadow:    color.RGBA{R: 60, G: 60, B: 65, A: 255},
		},
		TitleBar: SkeuoColor{
			Base:      color.RGBA{R: 50, G: 50, B: 55, A: 255},
			Highlight: color.RGBA{R: 70, G: 70, B: 75, A: 255},
			Shadow:    color.RGBA{R: 30, G: 30, B: 33, A: 255},
		},
		TitleBarText: color.RGBA{R: 220, G: 220, B: 220, A: 255},
		Text:         color.RGBA{R: 200, G: 200, B: 200, A: 255},
		TextDim:      color.RGBA{R: 120, G: 120, B: 120, A: 255},
		Canvas: SkeuoColor{
			Base:      color.RGBA{R: 20, G: 20, B: 22, A: 255},
			Highlight: color.RGBA{R: 10, G: 10, B: 12, A: 255},
			Shadow:    color.RGBA{R: 40, G: 40, B: 45, A: 255},
		},
		ScrollTrack: SkeuoColor{
			Base:      color.RGBA{R: 35, G: 35, B: 38, A: 255},
			Highlight: color.RGBA{R: 25, G: 25, B: 28, A: 255},
			Shadow:    color.RGBA{R: 45, G: 45, B: 48, A: 255},
		},
		ScrollThumb: SkeuoColor{
			Base:      color.RGBA{R: 70, G: 70, B: 75, A: 255},
			Highlight: color.RGBA{R: 90, G: 90, B: 95, A: 255},
			Shadow:    color.RGBA{R: 50, G: 50, B: 55, A: 255},
		},
		Background: color.RGBA{R: 30, G: 32, B: 34, A: 255},
		PixelScale: 2, // 1 logical px = 2 screen px
	}
}

// LightTheme returns a light skeuomorphic theme with a softer feel.
func LightTheme() *Theme {
	return &Theme{
		Panel: SkeuoColor{
			Base:      color.RGBA{R: 200, G: 200, B: 200, A: 255},
			Highlight: color.RGBA{R: 240, G: 240, B: 240, A: 255},
			Shadow:    color.RGBA{R: 140, G: 140, B: 140, A: 255},
		},
		Button: SkeuoColor{
			Base:      color.RGBA{R: 180, G: 180, B: 180, A: 255},
			Highlight: color.RGBA{R: 230, G: 230, B: 230, A: 255},
			Shadow:    color.RGBA{R: 120, G: 120, B: 120, A: 255},
		},
		ButtonHover: SkeuoColor{
			Base:      color.RGBA{R: 190, G: 190, B: 190, A: 255},
			Highlight: color.RGBA{R: 240, G: 240, B: 240, A: 255},
			Shadow:    color.RGBA{R: 130, G: 130, B: 130, A: 255},
		},
		Input: SkeuoColor{
			Base:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
			Highlight: color.RGBA{R: 160, G: 160, B: 160, A: 255}, // Inverted for sunken
			Shadow:    color.RGBA{R: 240, G: 240, B: 240, A: 255}, // Inverted for sunken
		},
		InputFocus: SkeuoColor{
			Base:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
			Highlight: color.RGBA{R: 100, G: 150, B: 200, A: 255},
			Shadow:    color.RGBA{R: 200, G: 220, B: 240, A: 255},
		},
		TitleBar: SkeuoColor{
			Base:      color.RGBA{R: 160, G: 160, B: 165, A: 255},
			Highlight: color.RGBA{R: 200, G: 200, B: 205, A: 255},
			Shadow:    color.RGBA{R: 110, G: 110, B: 115, A: 255},
		},
		TitleBarText: color.RGBA{R: 40, G: 40, B: 40, A: 255},
		Text:         color.RGBA{R: 30, G: 30, B: 30, A: 255},
		TextDim:      color.RGBA{R: 120, G: 120, B: 120, A: 255},
		Canvas: SkeuoColor{
			Base:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
			Highlight: color.RGBA{R: 180, G: 180, B: 180, A: 255},
			Shadow:    color.RGBA{R: 230, G: 230, B: 230, A: 255},
		},
		ScrollTrack: SkeuoColor{
			Base:      color.RGBA{R: 220, G: 220, B: 220, A: 255},
			Highlight: color.RGBA{R: 180, G: 180, B: 180, A: 255},
			Shadow:    color.RGBA{R: 240, G: 240, B: 240, A: 255},
		},
		ScrollThumb: SkeuoColor{
			Base:      color.RGBA{R: 160, G: 160, B: 160, A: 255},
			Highlight: color.RGBA{R: 200, G: 200, B: 200, A: 255},
			Shadow:    color.RGBA{R: 120, G: 120, B: 120, A: 255},
		},
		Background: color.RGBA{R: 140, G: 180, B: 140, A: 255}, // Soft green like reference
		PixelScale: 2,                                          // 1 logical px = 2 screen px
	}
}

// MintTheme returns a theme matching the first reference image (mint green background).
func MintTheme() *Theme {
	return &Theme{
		Panel: SkeuoColor{
			Base:      color.RGBA{R: 68, G: 68, B: 68, A: 255},
			Highlight: color.RGBA{R: 98, G: 98, B: 98, A: 255},
			Shadow:    color.RGBA{R: 38, G: 38, B: 38, A: 255},
		},
		Button: SkeuoColor{
			Base:      color.RGBA{R: 85, G: 85, B: 85, A: 255},
			Highlight: color.RGBA{R: 115, G: 115, B: 115, A: 255},
			Shadow:    color.RGBA{R: 55, G: 55, B: 55, A: 255},
		},
		ButtonHover: SkeuoColor{
			Base:      color.RGBA{R: 95, G: 95, B: 95, A: 255},
			Highlight: color.RGBA{R: 125, G: 125, B: 125, A: 255},
			Shadow:    color.RGBA{R: 65, G: 65, B: 65, A: 255},
		},
		Input: SkeuoColor{
			Base:      color.RGBA{R: 50, G: 50, B: 50, A: 255},
			Highlight: color.RGBA{R: 30, G: 30, B: 30, A: 255},
			Shadow:    color.RGBA{R: 70, G: 70, B: 70, A: 255},
		},
		InputFocus: SkeuoColor{
			Base:      color.RGBA{R: 55, G: 55, B: 55, A: 255},
			Highlight: color.RGBA{R: 35, G: 35, B: 35, A: 255},
			Shadow:    color.RGBA{R: 75, G: 75, B: 75, A: 255},
		},
		TitleBar: SkeuoColor{
			Base:      color.RGBA{R: 58, G: 58, B: 58, A: 255},
			Highlight: color.RGBA{R: 88, G: 88, B: 88, A: 255},
			Shadow:    color.RGBA{R: 28, G: 28, B: 28, A: 255},
		},
		TitleBarText: color.RGBA{R: 220, G: 220, B: 220, A: 255},
		Text:         color.RGBA{R: 220, G: 220, B: 220, A: 255},
		TextDim:      color.RGBA{R: 140, G: 140, B: 140, A: 255},
		Canvas: SkeuoColor{
			Base:      color.RGBA{R: 40, G: 40, B: 40, A: 255},
			Highlight: color.RGBA{R: 20, G: 20, B: 20, A: 255},
			Shadow:    color.RGBA{R: 60, G: 60, B: 60, A: 255},
		},
		ScrollTrack: SkeuoColor{
			Base:      color.RGBA{R: 50, G: 50, B: 50, A: 255},
			Highlight: color.RGBA{R: 30, G: 30, B: 30, A: 255},
			Shadow:    color.RGBA{R: 70, G: 70, B: 70, A: 255},
		},
		ScrollThumb: SkeuoColor{
			Base:      color.RGBA{R: 90, G: 90, B: 90, A: 255},
			Highlight: color.RGBA{R: 120, G: 120, B: 120, A: 255},
			Shadow:    color.RGBA{R: 60, G: 60, B: 60, A: 255},
		},
		Background: color.RGBA{R: 156, G: 203, B: 161, A: 255}, // Mint green
		PixelScale: 2,                                          // 1 logical px = 2 screen px
	}
}

// PixelTheme returns a dark theme matching modern pixel art editors.
// Features double borders, corner notches, and drop shadows.
// All layout values are in LOGICAL pixels - change PixelScale to resize everything.
func PixelTheme() *Theme {
	// Colors extracted from reference screenshot
	windowBody := color.RGBA{R: 50, G: 50, B: 50, A: 255}  // Dark gray window body
	titleBarBg := color.RGBA{R: 70, G: 70, B: 70, A: 255}  // Lighter gray title bar
	outerBorder := color.RGBA{R: 20, G: 20, B: 20, A: 255} // Black outer border
	innerBorder := color.RGBA{R: 90, G: 90, B: 90, A: 255} // Gray inner border
	buttonBg := color.RGBA{R: 55, G: 55, B: 55, A: 255}    // Button fill
	inputBg := color.RGBA{R: 40, G: 40, B: 40, A: 255}     // Input field

	return &Theme{
		// === COLORS ===
		Panel: SkeuoColor{
			Base:      windowBody,
			Highlight: innerBorder,
			Shadow:    outerBorder,
		},
		Button: SkeuoColor{
			Base:      buttonBg,
			Highlight: innerBorder,
			Shadow:    outerBorder,
		},
		ButtonHover: SkeuoColor{
			Base:      color.RGBA{R: 65, G: 65, B: 65, A: 255},
			Highlight: color.RGBA{R: 100, G: 100, B: 100, A: 255},
			Shadow:    outerBorder,
		},
		Input: SkeuoColor{
			Base:      inputBg,
			Highlight: innerBorder,
			Shadow:    outerBorder,
		},
		InputFocus: SkeuoColor{
			Base:      color.RGBA{R: 45, G: 45, B: 45, A: 255},
			Highlight: color.RGBA{R: 100, G: 150, B: 200, A: 255},
			Shadow:    outerBorder,
		},
		TitleBar: SkeuoColor{
			Base:      titleBarBg,
			Highlight: innerBorder,
			Shadow:    outerBorder,
		},
		TitleBarText: color.RGBA{R: 220, G: 220, B: 220, A: 255},
		Text:         color.RGBA{R: 220, G: 220, B: 220, A: 255},
		TextDim:      color.RGBA{R: 140, G: 140, B: 140, A: 255},
		Canvas: SkeuoColor{
			Base:      color.RGBA{R: 30, G: 30, B: 30, A: 255},
			Highlight: innerBorder,
			Shadow:    outerBorder,
		},
		ScrollTrack: SkeuoColor{
			Base:      color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 80}, // Dark with alpha
			Highlight: color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 80}, // Same (flat)
			Shadow:    color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 80}, // Same (flat)
		},
		ScrollThumb: SkeuoColor{
			Base:      color.RGBA{R: 30, G: 30, B: 30, A: 255}, // Dark thumb
			Highlight: color.RGBA{R: 30, G: 30, B: 30, A: 255}, // Same (flat)
			Shadow:    color.RGBA{R: 30, G: 30, B: 30, A: 255}, // Same (flat)
		},
		Background: color.RGBA{R: 35, G: 35, B: 35, A: 255},

		// === LAYOUT (all in logical pixels) ===
		PixelScale:      2,  // 1 logical px = 2 screen px (chunky retro)
		BorderSize:      2,  // outer + inner border
		CornerRadius:    1,  // corner notch
		ControlHeight:   14, // buttons, inputs height
		TitleHeight:     14, // title bar content height
		Padding:         struct{ X, Y int }{X: 4, Y: 2},
		ScrollbarWidth:  4, // scrollbar track width
		ScrollbarMargin: 2, // margin around scrollbar (total area = 2+4+2 = 8 logical)
		ThumbWidth:      8, // slider thumb width (logical pixels)
		Spacing:         2,

		// === EFFECTS ===
		ShadowAlpha: 60, // subtle drop shadow
		UseFlat:     true,
	}
}
