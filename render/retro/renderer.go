package retro

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/user/microui-go"
	"github.com/user/microui-go/types"
)

// Renderer is a skeuomorphic pixel-art style renderer for microui.
type Renderer struct {
	target    *ebiten.Image
	theme     *Theme
	font      *Font
	fontBold  *Font
	fontMono  *Font
	clipStack []types.Rect
}

// NewRenderer creates a new retro renderer with the specified theme.
func NewRenderer(theme *Theme) (*Renderer, error) {
	if theme == nil {
		theme = DarkTheme()
	}

	font, err := NewFont(FontSans, 16)
	if err != nil {
		return nil, err
	}

	fontBold, err := NewFont(FontSansBold, 16)
	if err != nil {
		return nil, err
	}

	fontMono, err := NewFont(FontMono, 14)
	if err != nil {
		return nil, err
	}

	return &Renderer{
		theme:     theme,
		font:      font,
		fontBold:  fontBold,
		fontMono:  fontMono,
		clipStack: make([]types.Rect, 0, 16),
	}, nil
}

// SetTarget sets the render target.
func (r *Renderer) SetTarget(target *ebiten.Image) {
	r.target = target
}

// SetTheme changes the current theme.
func (r *Renderer) SetTheme(theme *Theme) {
	r.theme = theme
}

// Theme returns the current theme.
func (r *Renderer) Theme() *Theme {
	return r.theme
}

// Font returns the font for layout calculations (implements types.Font).
func (r *Renderer) Font() types.Font {
	return &fontAdapter{r.font}
}

// Clear fills the screen with the background color.
func (r *Renderer) Clear() {
	if r.target != nil {
		r.target.Fill(r.theme.Background)
	}
}

// DrawRect draws a filled rectangle (implements BaseRenderer).
func (r *Renderer) DrawRect(pos, size types.Vec2, c color.Color) {
	if r.target == nil {
		return
	}
	r.fillRect(pos.X, pos.Y, size.X, size.Y, c)
}

// DrawBox draws a box outline (implements BoxRenderer).
func (r *Renderer) DrawBox(rect types.Rect, c color.Color) {
	if r.target == nil {
		return
	}
	// Draw 1px border
	r.fillRect(rect.X, rect.Y, rect.W, 1, c)                 // Top
	r.fillRect(rect.X, rect.Y+rect.H-1, rect.W, 1, c)        // Bottom
	r.fillRect(rect.X, rect.Y, 1, rect.H, c)                 // Left
	r.fillRect(rect.X+rect.W-1, rect.Y, 1, rect.H, c)        // Right
}

// DrawText draws text at the specified position with clipping.
func (r *Renderer) DrawText(text string, pos types.Vec2, font types.Font, c color.Color) {
	if r.target == nil || text == "" {
		return
	}

	// If no clip rect, draw directly
	if len(r.clipStack) == 0 {
		r.font.Draw(r.target, text, pos.X, pos.Y, c)
		return
	}

	// Get clip rect
	clip := r.clipStack[len(r.clipStack)-1]

	// Get text dimensions
	textW := r.font.Width(text)
	textH := r.font.Height()

	// Quick reject: completely outside clip rect
	if pos.X >= clip.X+clip.W || pos.Y >= clip.Y+clip.H {
		return
	}
	if pos.X+textW <= clip.X || pos.Y+textH <= clip.Y {
		return
	}

	// Clamp clip rect to target bounds
	targetBounds := r.target.Bounds()
	clipX := clip.X
	clipY := clip.Y
	clipW := clip.W
	clipH := clip.H

	if clipX < targetBounds.Min.X {
		clipW -= targetBounds.Min.X - clipX
		clipX = targetBounds.Min.X
	}
	if clipY < targetBounds.Min.Y {
		clipH -= targetBounds.Min.Y - clipY
		clipY = targetBounds.Min.Y
	}
	if clipX+clipW > targetBounds.Max.X {
		clipW = targetBounds.Max.X - clipX
	}
	if clipY+clipH > targetBounds.Max.Y {
		clipH = targetBounds.Max.Y - clipY
	}

	if clipW <= 0 || clipH <= 0 {
		return
	}

	// Use SubImage for clipping
	clipRect := image.Rect(clipX, clipY, clipX+clipW, clipY+clipH)
	subImg := r.target.SubImage(clipRect).(*ebiten.Image)

	// Draw text to SubImage (coordinates are absolute)
	r.font.Draw(subImg, text, pos.X, pos.Y, c)
}

// DrawIcon draws an icon with blocky pixel-art style.
// Icons are drawn using the theme's BevelDepth as the pixel scale,
// so each icon "pixel" is BevelDepth x BevelDepth screen pixels.
func (r *Renderer) DrawIcon(id int, rect types.Rect, c color.Color) {
	if r.target == nil {
		return
	}

	// Use BevelDepth as the pixel scale (default to 2 if not set)
	scale := r.theme.BevelDepth
	if scale < 1 {
		scale = 2
	}

	// Icon pattern is 7x7 logical pixels, scaled up
	const patternSize = 7
	iconSize := patternSize * scale

	// Center the icon in the rect
	offsetX := rect.X + (rect.W-iconSize)/2
	offsetY := rect.Y + (rect.H-iconSize)/2

	// Helper to draw a single scaled pixel (scale x scale block)
	px := func(x, y int) {
		r.fillRect(offsetX+x*scale, offsetY+y*scale, scale, scale, c)
	}

	switch id {
	case microui.IconClose:
		// X pattern (5x5 with 1px border = 7x7):
		// . . . . . . .
		// . X . . . X .
		// . . X . X . .
		// . . . X . . .
		// . . X . X . .
		// . X . . . X .
		// . . . . . . .
		px(1, 1); px(5, 1)
		px(2, 2); px(4, 2)
		px(3, 3)
		px(2, 4); px(4, 4)
		px(1, 5); px(5, 5)

	case microui.IconCheck:
		// Thicker checkmark pattern:
		// . . . . . . .
		// . . . . . X X
		// . . . . X X .
		// . X . X X . .
		// . X X X . . .
		// . . X . . . .
		// . . . . . . .
		px(5, 1); px(6, 1)
		px(4, 2); px(5, 2)
		px(1, 3); px(3, 3); px(4, 3)
		px(1, 4); px(2, 4); px(3, 4)
		px(2, 5)

	case microui.IconCollapsed:
		// Right-pointing triangle:
		// . . . . . . .
		// . X . . . . .
		// . X X . . . .
		// . X X X . . .
		// . X X . . . .
		// . X . . . . .
		// . . . . . . .
		px(1, 1)
		px(1, 2); px(2, 2)
		px(1, 3); px(2, 3); px(3, 3)
		px(1, 4); px(2, 4)
		px(1, 5)

	case microui.IconExpanded:
		// Down-pointing triangle:
		// . . . . . . .
		// . X X X X X .
		// . . X X X . .
		// . . . X . . .
		// . . . . . . .
		// . . . . . . .
		// . . . . . . .
		px(1, 1); px(2, 1); px(3, 1); px(4, 1); px(5, 1)
		px(2, 2); px(3, 2); px(4, 2)
		px(3, 3)

	case microui.IconResize:
		// Resize gripper (diagonal dots in bottom-right):
		// . . . . . . .
		// . . . . . . .
		// . . . . . X .
		// . . . . . . .
		// . . . X . X .
		// . . . . . . .
		// . X . X . X .
		px(5, 2)
		px(3, 4); px(5, 4)
		px(1, 6); px(3, 6); px(5, 6)

	default:
		// Unknown icon - draw a filled box for visibility
		for y := 1; y <= 5; y++ {
			for x := 1; x <= 5; x++ {
				px(x, y)
			}
		}
	}
}

// SetClip sets the clipping rectangle.
func (r *Renderer) SetClip(rect types.Rect) {
	// microui sends explicit clip rects, not push/pop - just replace current
	if len(r.clipStack) == 0 {
		r.clipStack = append(r.clipStack, rect)
	} else {
		r.clipStack[len(r.clipStack)-1] = rect
	}
}

// Flush completes rendering (no-op for Ebiten).
func (r *Renderer) Flush() {}

// DrawScrollTrack draws a scrollbar track (implements ScrollRenderer).
func (r *Renderer) DrawScrollTrack(rect types.Rect) {
	if r.target == nil {
		return
	}
	r.DrawBeveledRect(rect, r.theme.ScrollTrack, true)
}

// DrawScrollThumb draws a scrollbar thumb (implements ScrollRenderer).
func (r *Renderer) DrawScrollThumb(rect types.Rect) {
	if r.target == nil {
		return
	}
	r.DrawBeveledRect(rect, r.theme.ScrollThumb, false)
}

// DrawBeveledRect draws a rectangle with 3D beveled edges.
// If sunken is true, the highlight/shadow are inverted for an inset look.
func (r *Renderer) DrawBeveledRect(rect types.Rect, colors SkeuoColor, sunken bool) {
	if r.target == nil {
		return
	}

	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	depth := r.theme.BevelDepth

	highlight := colors.Highlight
	shadow := colors.Shadow
	if sunken {
		highlight, shadow = shadow, highlight
	}

	// Draw shadow edges (bottom and right)
	for i := 0; i < depth; i++ {
		// Bottom edge
		r.fillRect(x+i, y+h-1-i, w-i, 1, shadow)
		// Right edge
		r.fillRect(x+w-1-i, y+i, 1, h-i, shadow)
	}

	// Draw highlight edges (top and left)
	for i := 0; i < depth; i++ {
		// Top edge
		r.fillRect(x+i, y+i, w-i*2, 1, highlight)
		// Left edge
		r.fillRect(x+i, y+i, 1, h-i*2, highlight)
	}

	// Draw base fill
	r.fillRect(x+depth, y+depth, w-depth*2, h-depth*2, colors.Base)
}

// DrawFrame implements the microui DrawFrame callback for skeuomorphic rendering.
// Uses ui.DrawRect to add commands to the buffer for proper z-ordering.
func (r *Renderer) DrawFrame(ui *microui.UI, rect types.Rect, colorID int) {
	switch colorID {
	case microui.ColorButton, microui.ColorButtonHover:
		colors := r.theme.Button
		if colorID == microui.ColorButtonHover {
			colors = r.theme.ButtonHover
		}
		r.drawBeveledRectUI(ui, rect, colors, false)

	case microui.ColorButtonFocus:
		// Pressed button - sunken
		r.drawBeveledRectUI(ui, rect, r.theme.Button, true)

	case microui.ColorBase, microui.ColorBaseHover, microui.ColorBaseFocus:
		// Input fields - sunken
		colors := r.theme.Input
		if colorID == microui.ColorBaseFocus {
			colors = r.theme.InputFocus
		}
		r.drawBeveledRectUI(ui, rect, colors, true)

	case microui.ColorWindowBG:
		r.drawBeveledRectUI(ui, rect, r.theme.Panel, false)

	case microui.ColorTitleBG:
		r.drawBeveledRectUI(ui, rect, r.theme.TitleBar, false)

	case microui.ColorPanelBG:
		r.drawBeveledRectUI(ui, rect, r.theme.Panel, true)

	case microui.ColorScrollBase:
		r.drawBeveledRectUI(ui, rect, r.theme.ScrollTrack, true)

	case microui.ColorScrollThumb:
		r.drawBeveledRectUI(ui, rect, r.theme.ScrollThumb, false)

	default:
		// Fallback to flat rect
		ui.DrawRect(rect, r.theme.Panel.Base)
	}
}

// drawBeveledRectUI draws a beveled rectangle using ui.DrawRect for proper z-ordering.
// This adds commands to microui's buffer instead of drawing directly.
func (r *Renderer) drawBeveledRectUI(ui *microui.UI, rect types.Rect, colors SkeuoColor, sunken bool) {
	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	depth := r.theme.BevelDepth

	highlight := colors.Highlight
	shadow := colors.Shadow
	if sunken {
		highlight, shadow = shadow, highlight
	}

	// Draw base fill first
	ui.DrawRect(types.Rect{X: x, Y: y, W: w, H: h}, colors.Base)

	// Draw shadow edges (bottom and right)
	for i := 0; i < depth; i++ {
		// Bottom edge
		ui.DrawRect(types.Rect{X: x + i, Y: y + h - 1 - i, W: w - i, H: 1}, shadow)
		// Right edge
		ui.DrawRect(types.Rect{X: x + w - 1 - i, Y: y + i, W: 1, H: h - i}, shadow)
	}

	// Draw highlight edges (top and left)
	for i := 0; i < depth; i++ {
		// Top edge
		ui.DrawRect(types.Rect{X: x + i, Y: y + i, W: w - i*2, H: 1}, highlight)
		// Left edge
		ui.DrawRect(types.Rect{X: x + i, Y: y + i, W: 1, H: h - i*2}, highlight)
	}
}

// fillRect draws a filled rectangle without anti-aliasing, clipped to current clip rect.
func (r *Renderer) fillRect(x, y, w, h int, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}

	// Clip to current clip rect
	if len(r.clipStack) > 0 {
		clip := r.clipStack[len(r.clipStack)-1]
		// Intersect with clip rect
		x1, y1 := x, y
		x2, y2 := x+w, y+h
		cx1, cy1 := clip.X, clip.Y
		cx2, cy2 := clip.X+clip.W, clip.Y+clip.H

		if x1 < cx1 {
			x1 = cx1
		}
		if y1 < cy1 {
			y1 = cy1
		}
		if x2 > cx2 {
			x2 = cx2
		}
		if y2 > cy2 {
			y2 = cy2
		}

		x, y = x1, y1
		w, h = x2-x1, y2-y1

		if w <= 0 || h <= 0 {
			return
		}
	}

	vector.DrawFilledRect(r.target, float32(x), float32(y), float32(w), float32(h), c, false)
}

// drawLine draws a line.
func (r *Renderer) drawLine(x1, y1, x2, y2 float32, c color.Color) {
	vector.StrokeLine(r.target, x1, y1, x2, y2, 1, c, false)
}

// fontAdapter wraps Font to implement types.Font.
type fontAdapter struct {
	font *Font
}

func (f *fontAdapter) Width(text string) int {
	return f.font.Width(text)
}

func (f *fontAdapter) Height() int {
	return f.font.Height()
}

// Render processes all microui commands and renders them.
func (r *Renderer) Render(ui *microui.UI) {
	ui.Render(r)
}

// GetClipRect returns the current clip rectangle as an image.Rectangle.
func (r *Renderer) GetClipRect() image.Rectangle {
	if len(r.clipStack) == 0 {
		if r.target != nil {
			return r.target.Bounds()
		}
		return image.Rectangle{}
	}
	rect := r.clipStack[len(r.clipStack)-1]
	return image.Rect(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
