package bubbletea

import (
	"image/color"
	"strings"
	"sync"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/user/microui-go/types"
)

// ColorMode represents the terminal color depth.
type ColorMode int

const (
	ColorAuto      ColorMode = iota // Auto-detect (default, assumes true color)
	Color16                         // 16 ANSI colors
	Color256                        // 256 color palette
	ColorTrueColor                  // 24-bit true color
)

// Cell represents a single terminal cell with character and colors.
type Cell struct {
	Char rune        // Character to display (0 = empty/space)
	Fg   color.Color // Foreground color
	Bg   color.Color // Background color
}

// Renderer implements render.Renderer for terminal output.
// It maintains a double-buffered cell buffer for thread-safe rendering.
// The back buffer is updated by Draw operations, then swapped to front
// for the Draw() method to read from.
type Renderer struct {
	mu        sync.RWMutex
	front     [][]Cell   // Buffer for Draw() to read from (ticker goroutine)
	back      [][]Cell   // Buffer for updates (main goroutine)
	width     int        // Terminal width in cells
	height    int        // Terminal height in cells
	clipRect  types.Rect // Current clipping rectangle
	colorMode ColorMode  // Terminal color depth for shadow style
}

// NewRenderer creates a new TUI renderer with the given dimensions.
func NewRenderer(width, height int) *Renderer {
	r := &Renderer{
		width:  width,
		height: height,
	}
	r.front = make([][]Cell, height)
	r.back = make([][]Cell, height)
	for y := 0; y < height; y++ {
		r.front[y] = make([]Cell, width)
		r.back[y] = make([]Cell, width)
	}
	r.clipRect = types.Rect{X: 0, Y: 0, W: width, H: height}
	return r
}

// SetColorMode sets the terminal color depth.
// This affects shadow rendering: 16-color uses classic TV style (black/gray),
// while 256+ colors use gradient darkening for a smoother look.
func (r *Renderer) SetColorMode(mode ColorMode) {
	r.colorMode = mode
}

// Resize updates the renderer dimensions.
func (r *Renderer) Resize(width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if width == r.width && height == r.height {
		return
	}
	r.width = width
	r.height = height
	r.front = make([][]Cell, height)
	r.back = make([][]Cell, height)
	for y := 0; y < height; y++ {
		r.front[y] = make([]Cell, width)
		r.back[y] = make([]Cell, width)
	}
	r.clipRect = types.Rect{X: 0, Y: 0, W: width, H: height}
}

// Clear resets the back buffer for a new frame.
func (r *Renderer) Clear() {
	for y := range r.back {
		for x := range r.back[y] {
			r.back[y][x] = Cell{}
		}
	}
}

// FillBackground fills the entire buffer with a character and colors.
// Used for drawing patterned backgrounds like the classic Borland desktop.
func (r *Renderer) FillBackground(ch rune, fg, bg color.Color) {
	for y := range r.back {
		for x := range r.back[y] {
			r.back[y][x] = Cell{
				Char: ch,
				Fg:   fg,
				Bg:   bg,
			}
		}
	}
}

// Swap atomically swaps the front and back buffers.
// Call this after rendering a complete frame to make it visible to Draw().
func (r *Renderer) Swap() {
	r.mu.Lock()
	r.front, r.back = r.back, r.front
	r.mu.Unlock()
}

// Width returns the renderer width.
func (r *Renderer) Width() int {
	return r.width
}

// Height returns the renderer height.
func (r *Renderer) Height() int {
	return r.height
}

// GetCell returns the cell at the given position.
func (r *Renderer) GetCell(x, y int) Cell {
	if x < 0 || x >= r.width || y < 0 || y >= r.height {
		return Cell{}
	}
	return r.back[y][x]
}

// inClip checks if a position is within the current clip rectangle.
func (r *Renderer) inClip(x, y int) bool {
	return x >= r.clipRect.X && x < r.clipRect.X+r.clipRect.W &&
		y >= r.clipRect.Y && y < r.clipRect.Y+r.clipRect.H
}

// inBounds checks if a position is within the buffer bounds.
func (r *Renderer) inBounds(x, y int) bool {
	return x >= 0 && x < r.width && y >= 0 && y < r.height
}

// DrawRect fills a rectangle with the given color.
// In TUI mode, this fills cells with a space and background color.
// Special case: 1x1 rects are treated as cursors and invert the existing cell colors.
func (r *Renderer) DrawRect(pos, size types.Vec2, c color.Color) {
	// Special case for cursor: 1x1 rect inverts colors instead of overwriting
	if size.X == 1 && size.Y == 1 {
		x, y := pos.X, pos.Y
		if r.inClip(x, y) && r.inBounds(x, y) {
			existing := r.back[y][x]
			// Swap fg/bg colors for inverted cursor effect
			// If cell is empty, show a block cursor with the given color
			if existing.Char == 0 || existing.Char == ' ' {
				r.back[y][x] = Cell{
					Char: ' ',
					Fg:   existing.Bg,
					Bg:   c,
				}
			} else {
				// Invert: old bg becomes fg, given color becomes bg
				r.back[y][x] = Cell{
					Char: existing.Char,
					Fg:   existing.Bg,
					Bg:   c,
				}
			}
		}
		return
	}

	// Calculate clipped bounds
	x1 := pos.X
	y1 := pos.Y
	x2 := pos.X + size.X
	y2 := pos.Y + size.Y

	// Clip to clip rectangle
	if x1 < r.clipRect.X {
		x1 = r.clipRect.X
	}
	if y1 < r.clipRect.Y {
		y1 = r.clipRect.Y
	}
	if x2 > r.clipRect.X+r.clipRect.W {
		x2 = r.clipRect.X + r.clipRect.W
	}
	if y2 > r.clipRect.Y+r.clipRect.H {
		y2 = r.clipRect.Y + r.clipRect.H
	}

	// Fill cells
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if r.inBounds(x, y) {
				r.back[y][x] = Cell{
					Char: ' ',
					Bg:   c,
				}
			}
		}
	}
}

// Shadow colors - classic Turbo Vision style for 16-color mode
// These are ANSI colors 0 and 8, which exist in all color modes
var (
	ShadowBg = color.RGBA{R: 0, G: 0, B: 0, A: 255}    // Black (ANSI 0)
	ShadowFg = color.RGBA{R: 85, G: 85, B: 85, A: 255} // Dark gray (ANSI 8)
)

// darkenColor reduces brightness of a color by the given factor (0.0-1.0).
// A factor of 0.4 makes the color 40% as bright.
func darkenColor(c color.Color, factor float64) color.Color {
	if c == nil {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(float64(r>>8) * factor),
		G: uint8(float64(g>>8) * factor),
		B: uint8(float64(b>>8) * factor),
		A: uint8(a >> 8),
	}
}

// DrawShadow renders a shadow over existing cells in the given rectangle.
// For 16-color mode: uses classic Turbo Vision style (black bg, dark gray fg).
// For 256+ colors: uses gradient darkening by the given factor for smooth shadows.
func (r *Renderer) DrawShadow(rect types.Rect, factor float64) {
	x1, y1 := rect.X, rect.Y
	x2, y2 := rect.X+rect.W, rect.Y+rect.H

	// Clip to clip rectangle
	if x1 < r.clipRect.X {
		x1 = r.clipRect.X
	}
	if y1 < r.clipRect.Y {
		y1 = r.clipRect.Y
	}
	if x2 > r.clipRect.X+r.clipRect.W {
		x2 = r.clipRect.X + r.clipRect.W
	}
	if y2 > r.clipRect.Y+r.clipRect.H {
		y2 = r.clipRect.Y + r.clipRect.H
	}

	// Use classic TV style for 16 colors, gradient for 256+
	use16ColorStyle := r.colorMode == Color16

	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if r.inBounds(x, y) {
				existing := r.back[y][x]
				if use16ColorStyle {
					// Classic Turbo Vision: black bg, dark gray fg
					r.back[y][x] = Cell{
						Char: existing.Char,
						Fg:   ShadowFg,
						Bg:   ShadowBg,
					}
				} else {
					// Gradient darkening: darken existing colors
					r.back[y][x] = Cell{
						Char: existing.Char,
						Fg:   darkenColor(existing.Fg, factor),
						Bg:   darkenColor(existing.Bg, factor),
					}
				}
			}
		}
	}
}

// Box-drawing characters for TUI borders
const (
	boxTopLeft     = '┌'
	boxTopRight    = '┐'
	boxBottomLeft  = '└'
	boxBottomRight = '┘'
	boxHorizontal  = '─'
	boxVertical    = '│'
)

// DrawBox draws an outlined rectangle using box-drawing characters.
// This is the TUI equivalent of drawing a border - uses ┌─┐│└─┘ characters.
func (r *Renderer) DrawBox(rect types.Rect, c color.Color) {
	x1, y1 := rect.X, rect.Y
	x2, y2 := rect.X+rect.W-1, rect.Y+rect.H-1

	// Draw corners
	r.setCell(x1, y1, boxTopLeft, c)
	r.setCell(x2, y1, boxTopRight, c)
	r.setCell(x1, y2, boxBottomLeft, c)
	r.setCell(x2, y2, boxBottomRight, c)

	// Draw horizontal edges
	for x := x1 + 1; x < x2; x++ {
		r.setCell(x, y1, boxHorizontal, c)
		r.setCell(x, y2, boxHorizontal, c)
	}

	// Draw vertical edges
	for y := y1 + 1; y < y2; y++ {
		r.setCell(x1, y, boxVertical, c)
		r.setCell(x2, y, boxVertical, c)
	}
}

// setCell sets a single cell with clipping, preserving background color.
func (r *Renderer) setCell(x, y int, ch rune, fg color.Color) {
	if !r.inClip(x, y) || !r.inBounds(x, y) {
		return
	}
	bg := r.back[y][x].Bg
	r.back[y][x] = Cell{
		Char: ch,
		Fg:   fg,
		Bg:   bg,
	}
}

// SetCellFull sets a single cell with character and both fg/bg colors.
// Used for custom rendering like metaballs half-block characters.
func (r *Renderer) SetCellFull(x, y int, ch rune, fg, bg color.Color) {
	if !r.inBounds(x, y) {
		return
	}
	r.back[y][x] = Cell{
		Char: ch,
		Fg:   fg,
		Bg:   bg,
	}
}

// FillRectChar fills a rectangle with a specific character and colors.
// Used for TUI elements like scrollbars that need character-based rendering.
func (r *Renderer) FillRectChar(rect types.Rect, ch rune, fg, bg color.Color) {
	// Calculate clipped bounds
	x1 := rect.X
	y1 := rect.Y
	x2 := rect.X + rect.W
	y2 := rect.Y + rect.H

	// Clip to clip rectangle
	if x1 < r.clipRect.X {
		x1 = r.clipRect.X
	}
	if y1 < r.clipRect.Y {
		y1 = r.clipRect.Y
	}
	if x2 > r.clipRect.X+r.clipRect.W {
		x2 = r.clipRect.X + r.clipRect.W
	}
	if y2 > r.clipRect.Y+r.clipRect.H {
		y2 = r.clipRect.Y + r.clipRect.H
	}

	// Fill cells with character
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if r.inBounds(x, y) {
				r.back[y][x] = Cell{
					Char: ch,
					Fg:   fg,
					Bg:   bg,
				}
			}
		}
	}
}

// DrawScrollTrack draws a scrollbar track (background).
// Uses the light shade character (░) for classic TUI look.
func (r *Renderer) DrawScrollTrack(rect types.Rect) {
	r.FillRectChar(rect, ScrollTrackChar, ScrollTrackFg, ScrollTrackBg)
}

// DrawScrollThumb draws a scrollbar thumb (draggable part).
// Uses the full block character (█) for visibility.
func (r *Renderer) DrawScrollThumb(rect types.Rect) {
	r.FillRectChar(rect, ScrollThumbChar, ScrollThumbFg, ScrollThumbBg)
}

// DrawText renders text at the specified position.
func (r *Renderer) DrawText(text string, pos types.Vec2, font types.Font, c color.Color) {
	x := pos.X
	y := pos.Y

	// Skip if completely outside clip rect vertically
	if y < r.clipRect.Y || y >= r.clipRect.Y+r.clipRect.H {
		return
	}

	for _, ch := range text {
		// Skip if outside clip rect horizontally
		if x >= r.clipRect.X && x < r.clipRect.X+r.clipRect.W {
			if r.inBounds(x, y) {
				// Preserve existing background color
				bg := r.back[y][x].Bg
				r.back[y][x] = Cell{
					Char: ch,
					Fg:   c,
					Bg:   bg,
				}
			}
		}
		x++
	}
}

// DrawIcon renders an icon using Unicode symbols.
func (r *Renderer) DrawIcon(id int, rect types.Rect, c color.Color) {
	icon := IconToRune(id)

	// Center the icon in the rect (for single-char icons)
	x := rect.X + rect.W/2
	y := rect.Y + rect.H/2

	if r.inClip(x, y) && r.inBounds(x, y) {
		bg := r.back[y][x].Bg
		r.back[y][x] = Cell{
			Char: icon,
			Fg:   c,
			Bg:   bg,
		}
	}
}

// SetClip sets the clipping rectangle for subsequent drawing operations.
func (r *Renderer) SetClip(rect types.Rect) {
	r.clipRect = rect
}

// RenderToString converts the cell buffer to a string for debugging.
// Each line is separated by newline. Useful for testing.
func (r *Renderer) RenderToString() string {
	var sb strings.Builder
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			ch := r.back[y][x].Char
			if ch == 0 {
				ch = ' '
			}
			sb.WriteRune(ch)
		}
		if y < r.height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// colorKey extracts a comparable key for a color (or 0 if nil)
func colorKey(c color.Color) uint32 {
	if c == nil {
		return 0
	}
	r, g, b, _ := c.RGBA()
	return ((r >> 8) << 16) | ((g >> 8) << 8) | (b >> 8)
}

// RenderToANSI converts the cell buffer to an ANSI-colored string.
// Optimized to batch colors and minimize escape codes.
func (r *Renderer) RenderToANSI() string {
	// Pre-allocate for better performance
	var sb strings.Builder
	sb.Grow(r.width * r.height * 4) // Rough estimate

	var curFg, curBg uint32 = 0, 0
	needsReset := false

	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			cell := r.back[y][x]
			ch := cell.Char
			if ch == 0 {
				ch = ' '
			}

			// Get color keys for this cell
			newFg := colorKey(cell.Fg)
			newBg := colorKey(cell.Bg)

			// Check if colors changed from current state
			fgChanged := newFg != curFg
			bgChanged := newBg != curBg

			if fgChanged || bgChanged {
				// Need to change colors
				if newFg == 0 && newBg == 0 {
					// Reset to default
					if needsReset {
						sb.WriteString("\x1b[0m")
						needsReset = false
					}
				} else {
					// Emit color codes
					sb.WriteString("\x1b[")
					first := true

					if fgChanged && newFg != 0 {
						sb.WriteString("38;2;")
						sb.WriteString(itoa(int((newFg >> 16) & 0xFF)))
						sb.WriteRune(';')
						sb.WriteString(itoa(int((newFg >> 8) & 0xFF)))
						sb.WriteRune(';')
						sb.WriteString(itoa(int(newFg & 0xFF)))
						first = false
					} else if fgChanged && newFg == 0 {
						sb.WriteString("39") // Default fg
						first = false
					}

					if bgChanged && newBg != 0 {
						if !first {
							sb.WriteRune(';')
						}
						sb.WriteString("48;2;")
						sb.WriteString(itoa(int((newBg >> 16) & 0xFF)))
						sb.WriteRune(';')
						sb.WriteString(itoa(int((newBg >> 8) & 0xFF)))
						sb.WriteRune(';')
						sb.WriteString(itoa(int(newBg & 0xFF)))
					} else if bgChanged && newBg == 0 {
						if !first {
							sb.WriteRune(';')
						}
						sb.WriteString("49") // Default bg
					}

					sb.WriteRune('m')
					needsReset = true
				}
				curFg = newFg
				curBg = newBg
			}

			sb.WriteRune(ch)
		}

		// Reset colors at end of line for cleaner output
		if needsReset {
			sb.WriteString("\x1b[0m")
			curFg, curBg = 0, 0
			needsReset = false
		}

		if y < r.height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// itoa converts int to string without importing strconv
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// DebugLog is a callback for debug logging (set externally).
var DebugLog func(format string, args ...any)

// ContentHash returns a hash of the current cell buffer content.
// Used for detecting if content actually changed between frames.
func (r *Renderer) ContentHash() uint64 {
	var hash uint64 = 14695981039346656037 // FNV-1a offset basis
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			cell := r.back[y][x]
			// Hash the rune
			hash ^= uint64(cell.Char)
			hash *= 1099511628211 // FNV-1a prime
			// Hash colors
			hash ^= uint64(colorKey(cell.Fg))
			hash *= 1099511628211
			hash ^= uint64(colorKey(cell.Bg))
			hash *= 1099511628211
		}
	}
	return hash
}

// Draw implements tea.Layer interface for Bubble Tea v2 rendering.
// Reads from the front buffer (swapped after View completes).
// This is called from the ticker goroutine, while updates happen on main goroutine.
func (r *Renderer) Draw(s uv.Screen, rect uv.Rectangle) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for y := rect.Min.Y; y < rect.Max.Y && y < r.height; y++ {
		if y < 0 {
			continue
		}
		for x := rect.Min.X; x < rect.Max.X && x < r.width; x++ {
			if x < 0 {
				continue
			}

			cell := r.front[y][x]

			// Only set cells that have content (char, fg, or bg)
			// Empty cells are left for ultraviolet to handle via its Clear()
			hasContent := cell.Char != 0 || cell.Fg != nil || cell.Bg != nil
			if !hasContent {
				continue
			}

			ch := cell.Char
			if ch == 0 {
				ch = ' '
			}

			s.SetCell(x, y, &uv.Cell{
				Content: string(ch),
				Style: uv.Style{
					Fg: cell.Fg,
					Bg: cell.Bg,
				},
				Width: 1,
			})
		}
	}
}
