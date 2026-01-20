package main

import "image/color"

// Tool represents a drawing tool.
type Tool int

const (
	ToolPencil Tool = iota
	ToolEraser
	ToolFill
)

// Editor holds the pixel editor state.
type Editor struct {
	width        int
	height       int
	pixels       []color.Color
	currentColor color.Color
	tool         Tool
	brushSize    float64
	zoom         int
	palette      []color.Color
	// For line interpolation
	lastX, lastY int
	drawing      bool
}

// NewEditor creates a new pixel editor with the given canvas size.
func NewEditor(width, height int) *Editor {
	pixels := make([]color.Color, width*height)
	// Initialize with transparent/white
	bgColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for i := range pixels {
		pixels[i] = bgColor
	}

	return &Editor{
		width:        width,
		height:       height,
		pixels:       pixels,
		currentColor: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		tool:         ToolPencil,
		brushSize:    1,
		zoom:         12,
		palette:      defaultPalette(),
	}
}

// defaultPalette returns a nice pixel art color palette.
func defaultPalette() []color.Color {
	return []color.Color{
		// Row 1: Grayscale
		color.RGBA{R: 0, G: 0, B: 0, A: 255},
		color.RGBA{R: 85, G: 85, B: 85, A: 255},
		color.RGBA{R: 170, G: 170, B: 170, A: 255},
		color.RGBA{R: 255, G: 255, B: 255, A: 255},

		// Row 2: Reds/Browns
		color.RGBA{R: 127, G: 36, B: 36, A: 255},
		color.RGBA{R: 191, G: 64, B: 64, A: 255},
		color.RGBA{R: 255, G: 100, B: 100, A: 255},
		color.RGBA{R: 139, G: 90, B: 43, A: 255},

		// Row 3: Oranges/Yellows
		color.RGBA{R: 255, G: 127, B: 39, A: 255},
		color.RGBA{R: 255, G: 180, B: 80, A: 255},
		color.RGBA{R: 255, G: 220, B: 100, A: 255},
		color.RGBA{R: 255, G: 255, B: 100, A: 255},

		// Row 4: Greens
		color.RGBA{R: 34, G: 85, B: 34, A: 255},
		color.RGBA{R: 50, G: 150, B: 50, A: 255},
		color.RGBA{R: 100, G: 200, B: 100, A: 255},
		color.RGBA{R: 180, G: 230, B: 130, A: 255},

		// Row 5: Blues
		color.RGBA{R: 30, G: 60, B: 114, A: 255},
		color.RGBA{R: 50, G: 100, B: 180, A: 255},
		color.RGBA{R: 100, G: 150, B: 230, A: 255},
		color.RGBA{R: 150, G: 200, B: 255, A: 255},

		// Row 6: Purples/Pinks
		color.RGBA{R: 85, G: 34, B: 102, A: 255},
		color.RGBA{R: 140, G: 70, B: 160, A: 255},
		color.RGBA{R: 200, G: 120, B: 200, A: 255},
		color.RGBA{R: 255, G: 180, B: 200, A: 255},

		// Row 7: Skin tones
		color.RGBA{R: 255, G: 224, B: 189, A: 255},
		color.RGBA{R: 228, G: 185, B: 145, A: 255},
		color.RGBA{R: 198, G: 145, B: 108, A: 255},
		color.RGBA{R: 141, G: 95, B: 68, A: 255},

		// Row 8: Extras
		color.RGBA{R: 0, G: 170, B: 170, A: 255},
		color.RGBA{R: 255, G: 100, B: 150, A: 255},
		color.RGBA{R: 100, G: 80, B: 60, A: 255},
		color.RGBA{R: 60, G: 60, B: 80, A: 255},
	}
}

// GetPixel returns the color at the given position.
func (e *Editor) GetPixel(x, y int) color.Color {
	if x < 0 || x >= e.width || y < 0 || y >= e.height {
		return color.RGBA{}
	}
	return e.pixels[y*e.width+x]
}

// SetPixel sets the color at the given position.
func (e *Editor) SetPixel(x, y int, c color.Color) {
	if x < 0 || x >= e.width || y < 0 || y >= e.height {
		return
	}
	e.pixels[y*e.width+x] = c
}

// SetTool changes the current tool.
func (e *Editor) SetTool(t Tool) {
	e.tool = t
}

// SetColor changes the current drawing color.
func (e *Editor) SetColor(c color.Color) {
	e.currentColor = c
}

// ToolName returns the name of the current tool.
func (e *Editor) ToolName() string {
	switch e.tool {
	case ToolPencil:
		return "Pencil"
	case ToolEraser:
		return "Eraser"
	case ToolFill:
		return "Fill"
	default:
		return "Unknown"
	}
}

// StartStroke begins a drawing stroke at the given position.
func (e *Editor) StartStroke(x, y int) {
	e.drawing = true
	e.lastX = x
	e.lastY = y
	e.ApplyToolAt(x, y)
}

// ContinueStroke continues the stroke to the given position with line interpolation.
func (e *Editor) ContinueStroke(x, y int) {
	if !e.drawing {
		e.StartStroke(x, y)
		return
	}
	// Draw line from last position to current position
	e.drawLine(e.lastX, e.lastY, x, y)
	e.lastX = x
	e.lastY = y
}

// EndStroke ends the current drawing stroke.
func (e *Editor) EndStroke() {
	e.drawing = false
}

// ApplyTool applies the current tool at the given canvas position (legacy single-point).
func (e *Editor) ApplyTool(x, y int) {
	e.ApplyToolAt(x, y)
}

// ApplyToolAt applies the current tool at the given canvas position.
func (e *Editor) ApplyToolAt(x, y int) {
	switch e.tool {
	case ToolPencil:
		e.drawBrush(x, y, e.currentColor)
	case ToolEraser:
		e.drawBrush(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	case ToolFill:
		e.floodFill(x, y, e.currentColor)
	}
}

// drawLine draws from (x0,y0) to (x1,y1) using Bresenham's algorithm.
func (e *Editor) drawLine(x0, y0, x1, y1 int) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx + dy

	for {
		e.ApplyToolAt(x0, y0)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			if x0 == x1 {
				break
			}
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			if y0 == y1 {
				break
			}
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawBrush draws with the current brush size.
func (e *Editor) drawBrush(cx, cy int, c color.Color) {
	size := int(e.brushSize)
	offset := size / 2

	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			e.SetPixel(cx-offset+dx, cy-offset+dy, c)
		}
	}
}

// floodFill performs a flood fill from the given position.
func (e *Editor) floodFill(x, y int, newColor color.Color) {
	if x < 0 || x >= e.width || y < 0 || y >= e.height {
		return
	}

	targetColor := e.GetPixel(x, y)
	if colorsEqual(targetColor, newColor) {
		return
	}

	// Simple stack-based flood fill
	type point struct{ x, y int }
	stack := []point{{x, y}}

	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if p.x < 0 || p.x >= e.width || p.y < 0 || p.y >= e.height {
			continue
		}

		current := e.GetPixel(p.x, p.y)
		if !colorsEqual(current, targetColor) {
			continue
		}

		e.SetPixel(p.x, p.y, newColor)

		stack = append(stack, point{p.x + 1, p.y})
		stack = append(stack, point{p.x - 1, p.y})
		stack = append(stack, point{p.x, p.y + 1})
		stack = append(stack, point{p.x, p.y - 1})
	}
}

// ZoomIn increases the zoom level.
func (e *Editor) ZoomIn() {
	if e.zoom < 24 {
		e.zoom += 2
	}
}

// ZoomOut decreases the zoom level.
func (e *Editor) ZoomOut() {
	if e.zoom > 4 {
		e.zoom -= 2
	}
}

// Clear resets the canvas to white.
func (e *Editor) Clear() {
	bgColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for i := range e.pixels {
		e.pixels[i] = bgColor
	}
}
