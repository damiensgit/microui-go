package metaballs

import (
	"image/color"
	"math"
)

// Half-block Unicode characters for 2x vertical resolution
const (
	UpperHalf = '▀' // U+2580 - top pixel filled
	LowerHalf = '▄' // U+2584 - bottom pixel filled
	FullBlock = '█' // U+2588 - both pixels filled
	// Space = ' ' - neither pixel filled (use background)
)

// ColorMode for TUI rendering
type ColorMode int

const (
	ColorMode16    ColorMode = iota // 16 ANSI colors - simple cyan/blue
	ColorMode256                    // 256 colors - HSV with limited palette
	ColorModeTrueColor              // 24-bit true color - full HSV
)

// Cell represents a single TUI cell with half-block encoding
type Cell struct {
	Char rune
	Fg   color.Color
	Bg   color.Color
}

// TUIRenderer renders metaballs to a grid of half-block characters.
// Each cell represents 2 vertical "pixels", doubling vertical resolution.
type TUIRenderer struct {
	field     *Field
	screenW   int // Total screen width for coordinate mapping
	screenH   int // Total screen height (in half-pixels) for coordinate mapping
	colorMode ColorMode

	// Simple mode colors (16-color)
	blobColor color.Color // Color for blob interiors
	glowColor color.Color // Color for blob edges/glow
	bgColor   color.Color // Background color

	// HSV mode settings (256/true color)
	hueOffset   float64 // Base hue offset (0-1)
	saturation  float64 // Color saturation (0-1)
	colorCycle  float64 // How much hue varies across screen (0-1)
}

// NewTUIRenderer creates a renderer for the given field
func NewTUIRenderer(field *Field, screenW, screenH int) *TUIRenderer {
	return &TUIRenderer{
		field:     field,
		screenW:   screenW,
		screenH:   screenH * 2, // Double for half-block pixels
		colorMode: ColorModeTrueColor,
		// 16-color defaults
		blobColor: color.RGBA{0, 255, 255, 255}, // Cyan
		glowColor: color.RGBA{0, 170, 170, 255}, // Darker cyan
		bgColor:   color.RGBA{0, 0, 128, 255},   // Dark blue
		// HSV defaults
		hueOffset:   0.0,
		saturation:  0.8,
		colorCycle:  0.5,
	}
}

// SetColorMode sets the color rendering mode
func (r *TUIRenderer) SetColorMode(mode ColorMode) {
	r.colorMode = mode
}

// SetColors configures the rendering colors (for 16-color mode)
func (r *TUIRenderer) SetColors(blob, glow, bg color.Color) {
	r.blobColor = blob
	r.glowColor = glow
	r.bgColor = bg
}

// SetHSVParams configures HSV rendering (for 256/true color modes)
func (r *TUIRenderer) SetHSVParams(hueOffset, saturation, colorCycle float64) {
	r.hueOffset = hueOffset
	r.saturation = saturation
	r.colorCycle = colorCycle
}

// SetScreenSize updates the screen dimensions for coordinate mapping
func (r *TUIRenderer) SetScreenSize(w, h int) {
	r.screenW = w
	r.screenH = h * 2 // Double for half-block pixels
}

// hsvToRGB converts HSV (hue 0-1, saturation 0-1, value 0-1) to RGB
func hsvToRGB(h, s, v float64) color.RGBA {
	// Normalize hue to 0-1 range
	h = h - math.Floor(h)

	hi := int(h * 6)
	f := h*6 - float64(hi)
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	var rf, gf, bf float64
	switch hi % 6 {
	case 0:
		rf, gf, bf = v, t, p
	case 1:
		rf, gf, bf = q, v, p
	case 2:
		rf, gf, bf = p, v, t
	case 3:
		rf, gf, bf = p, q, v
	case 4:
		rf, gf, bf = t, p, v
	case 5:
		rf, gf, bf = v, p, q
	}

	return color.RGBA{
		R: uint8(rf * 255),
		G: uint8(gf * 255),
		B: uint8(bf * 255),
		A: 255,
	}
}

// colorForField returns the color for a given field value and position
func (r *TUIRenderer) colorForField(field float64, x, y int, threshold float64) color.Color {
	if r.colorMode == ColorMode16 {
		// Simple mode - use preset colors
		if field >= threshold {
			return r.blobColor
		} else if field >= threshold*0.5 {
			return r.glowColor
		}
		return r.bgColor
	}

	// HSV mode for 256/true color
	// Use animation time for color cycling
	time := r.field.Time()
	hue := r.hueOffset + time*0.3 + float64(x+y)/float64(r.screenW+r.screenH)*r.colorCycle

	if field >= threshold {
		// Inside blob - bright, saturated color
		intensity := math.Min((field-threshold)/threshold, 1.0)
		value := 0.5 + 0.5*intensity
		return hsvToRGB(hue, r.saturation, value)
	} else if field >= threshold*0.3 {
		// Glow region - dimmer
		glow := (field - threshold*0.3) / (threshold * 0.7)
		value := 0.2 + 0.3*glow
		sat := r.saturation * glow
		return hsvToRGB(hue, sat, value)
	}

	// Background - dark blue/purple
	return hsvToRGB(0.7, 0.5, 0.1)
}

// RenderCell renders a single cell at the given screen position.
// screenX, screenY are in cell coordinates (not half-pixels).
// The cell acts as a viewport - its screen position determines what part
// of the infinite metaball field is visible.
func (r *TUIRenderer) RenderCell(screenX, screenY int) Cell {
	// Convert screen position to normalized field coordinates
	// Top pixel of this cell
	topX := float64(screenX) / float64(r.screenW)
	topY := float64(screenY*2) / float64(r.screenH)
	// Bottom pixel of this cell
	botX := topX
	botY := float64(screenY*2+1) / float64(r.screenH)

	// Sample the field at both positions
	topField := r.field.Sample(topX, topY)
	botField := r.field.Sample(botX, botY)

	threshold := r.field.Threshold()
	topInside := topField >= threshold
	botInside := botField >= threshold

	// Get colors for each half-pixel
	topColor := r.colorForField(topField, screenX, screenY*2, threshold)
	botColor := r.colorForField(botField, screenX, screenY*2+1, threshold)

	// Determine character and colors based on which pixels are filled
	// For half-blocks: Fg = foreground char color, Bg = background
	switch {
	case topInside && botInside:
		// Both inside - full block, use top color for both
		return Cell{Char: FullBlock, Fg: topColor, Bg: topColor}
	case topInside && !botInside:
		// Only top pixel - upper half block
		// ▀ draws top half in Fg, bottom half shows Bg
		return Cell{Char: UpperHalf, Fg: topColor, Bg: botColor}
	case !topInside && botInside:
		// Only bottom pixel - lower half block
		// ▄ draws bottom half in Fg, top half shows Bg
		return Cell{Char: LowerHalf, Fg: botColor, Bg: topColor}
	default:
		// Neither inside - check for glow
		topGlow := topField >= threshold*0.3
		botGlow := botField >= threshold*0.3
		if topGlow && botGlow {
			return Cell{Char: FullBlock, Fg: topColor, Bg: topColor}
		} else if topGlow {
			return Cell{Char: UpperHalf, Fg: topColor, Bg: botColor}
		} else if botGlow {
			return Cell{Char: LowerHalf, Fg: botColor, Bg: topColor}
		}
		// Far from blobs - background
		bgColor := r.colorForField(0, screenX, screenY, threshold)
		return Cell{Char: ' ', Fg: bgColor, Bg: bgColor}
	}
}

// RenderWindow renders a rectangular window of cells.
// windowX, windowY is the top-left screen position of the window.
// width, height are the window dimensions in cells.
// Returns a 2D grid of cells [y][x].
func (r *TUIRenderer) RenderWindow(windowX, windowY, width, height int) [][]Cell {
	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			// Screen position = window position + cell offset
			cells[y][x] = r.RenderCell(windowX+x, windowY+y)
		}
	}
	return cells
}
