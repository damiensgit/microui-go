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
	target   *ebiten.Image
	theme    *Theme
	font     *Font
	fontBold *Font
	fontMono *Font
	clipRect types.Rect
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
		theme:    theme,
		font:     font,
		fontBold: fontBold,
		fontMono: fontMono,
		clipRect: types.Rect{X: 0, Y: 0, W: 10000, H: 10000}, // Default to "unclipped"
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
// Uses logical pixels (PixelScale) for line thickness.
func (r *Renderer) DrawBox(rect types.Rect, c color.Color) {
	if r.target == nil {
		return
	}
	px := r.theme.PixelScale
	if px < 1 {
		px = 2
	}
	// Draw border using logical pixel thickness
	r.fillRect(rect.X, rect.Y, rect.W, px, c)                   // Top
	r.fillRect(rect.X, rect.Y+rect.H-px, rect.W, px, c)         // Bottom
	r.fillRect(rect.X, rect.Y, px, rect.H, c)                   // Left
	r.fillRect(rect.X+rect.W-px, rect.Y, px, rect.H, c)         // Right
}

// DrawText draws text at the specified position with clipping.
func (r *Renderer) DrawText(text string, pos types.Vec2, font types.Font, c color.Color) {
	if r.target == nil || text == "" {
		return
	}

	// Get clip rect (always apply clipping)
	clip := r.clipRect

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
// Icons are drawn using the theme's PixelScale,
// so each icon "pixel" is PixelScale x PixelScale screen pixels.
func (r *Renderer) DrawIcon(id int, rect types.Rect, c color.Color) {
	if r.target == nil {
		return
	}

	// Use PixelScale as the pixel scale (default to 2 if not set)
	scale := r.theme.PixelScale
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
	r.clipRect = rect
}

// Flush completes rendering (no-op for Ebiten).
func (r *Renderer) Flush() {}

// DrawScrollTrack draws a scrollbar track (implements ScrollRenderer).
func (r *Renderer) DrawScrollTrack(rect types.Rect) {
	if r.target == nil {
		return
	}
	if r.theme.UseFlat {
		// Flat floating scrollbar - just the base color
		r.fillRect(rect.X, rect.Y, rect.W, rect.H, r.theme.ScrollTrack.Base)
		return
	}
	r.DrawBeveledRect(rect, r.theme.ScrollTrack, true)
}

// DrawScrollThumb draws a scrollbar thumb (implements ScrollRenderer).
func (r *Renderer) DrawScrollThumb(rect types.Rect) {
	if r.target == nil {
		return
	}
	if r.theme.UseFlat {
		// Flat floating scrollbar thumb - just the base color
		r.fillRect(rect.X, rect.Y, rect.W, rect.H, r.theme.ScrollThumb.Base)
		return
	}
	r.DrawBeveledRect(rect, r.theme.ScrollThumb, false)
}

// DrawBeveledRect draws a rectangle with 3D beveled edges or flat style.
// If sunken is true, the highlight/shadow are inverted for an inset look.
func (r *Renderer) DrawBeveledRect(rect types.Rect, colors SkeuoColor, sunken bool) {
	if r.target == nil {
		return
	}

	// Use flat style if theme specifies it
	if r.theme.UseFlat {
		r.DrawFlatRect(rect, colors, sunken)
		return
	}

	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	depth := r.theme.PixelScale

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

// DrawFlatRect draws a pixel-art style rectangle with double borders and drop shadow.
// Structure: drop shadow -> outer border (Shadow) -> inner border (Highlight) -> fill (Base)
func (r *Renderer) DrawFlatRect(rect types.Rect, colors SkeuoColor, sunken bool) {
	if r.target == nil {
		return
	}

	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	radius := r.theme.CornerRadius

	// Colors: Shadow = outer border (dark), Highlight = inner border (lighter), Base = fill
	outerBorder := colors.Shadow
	innerBorder := colors.Highlight
	fill := colors.Base

	if sunken {
		innerBorder = colors.Shadow // Same as outer for sunken
	}

	// 1. Draw drop shadow
	shadowColor := r.theme.ShadowColor()
	if shadowColor != nil {
		// Shadow on right edge
		r.fillRect(x+w, y+radius+1, 1, h-radius, shadowColor)
		// Shadow on bottom edge
		r.fillRect(x+radius+1, y+h, w-radius, 1, shadowColor)
		// Shadow corner (bottom-right)
		if radius >= 1 {
			r.fillRect(x+w-1, y+h, 1, 1, shadowColor)
			r.fillRect(x+w, y+h-1, 1, 1, shadowColor)
		}
	}

	// 2. Draw outer border with 1px corner radius
	r.drawRoundedBorder(x, y, w, h, radius, outerBorder)

	// 3. Draw inner border inset by 1px
	r.drawRoundedBorder(x+1, y+1, w-2, h-2, radius, innerBorder)

	// 4. Draw fill (inset by 2px for both borders)
	r.fillRect(x+2, y+2, w-4, h-4, fill)
}

// fillRoundedRect draws a filled rectangle with corner notches (1px corner radius).
func (r *Renderer) fillRoundedRect(x, y, w, h, radius int, c color.Color) {
	if radius <= 0 {
		r.fillRect(x, y, w, h, c)
		return
	}

	// Fill main body (excluding corners)
	// Top row (excluding corner pixels)
	r.fillRect(x+radius, y, w-radius*2, radius, c)
	// Middle rows (full width)
	r.fillRect(x, y+radius, w, h-radius*2, c)
	// Bottom row (excluding corner pixels)
	r.fillRect(x+radius, y+h-radius, w-radius*2, radius, c)
}

// drawRoundedBorder draws a 1px border with corner notches.
func (r *Renderer) drawRoundedBorder(x, y, w, h, radius int, c color.Color) {
	if radius <= 0 {
		// Simple box border
		r.fillRect(x, y, w, 1, c)           // Top
		r.fillRect(x, y+h-1, w, 1, c)       // Bottom
		r.fillRect(x, y, 1, h, c)           // Left
		r.fillRect(x+w-1, y, 1, h, c)       // Right
		return
	}

	// Top edge (excluding corners)
	r.fillRect(x+radius, y, w-radius*2, 1, c)
	// Bottom edge (excluding corners)
	r.fillRect(x+radius, y+h-1, w-radius*2, 1, c)
	// Left edge (excluding corners)
	r.fillRect(x, y+radius, 1, h-radius*2, c)
	// Right edge (excluding corners)
	r.fillRect(x+w-1, y+radius, 1, h-radius*2, c)

	// Draw corner pixels (for 1px radius, just the diagonal pixel)
	if radius == 1 {
		// Top-left corner
		r.fillRect(x, y+1, 1, 1, c)
		r.fillRect(x+1, y, 1, 1, c)
		// Top-right corner
		r.fillRect(x+w-2, y, 1, 1, c)
		r.fillRect(x+w-1, y+1, 1, 1, c)
		// Bottom-left corner
		r.fillRect(x, y+h-2, 1, 1, c)
		r.fillRect(x+1, y+h-1, 1, 1, c)
		// Bottom-right corner
		r.fillRect(x+w-2, y+h-1, 1, 1, c)
		r.fillRect(x+w-1, y+h-2, 1, 1, c)
	}
}

// DrawFrame implements the microui DrawFrame callback for pixel art style rendering.
// Uses ui.DrawRect to add commands to the buffer for proper z-ordering.
func (r *Renderer) DrawFrame(ui *microui.UI, info microui.FrameInfo) {
	// Use flat pixel art style if enabled
	if r.theme.UseFlat {
		r.drawFrameFlat(ui, info)
		return
	}

	// Beveled style - dispatch based on FrameKind and FrameState
	colors, sunken := r.getStyleForFrame(info.Kind, info.State)
	r.drawBeveledRectUI(ui, info.Rect, colors, sunken)
}

// getStyleForFrame returns theme colors and sunken state for a frame.
func (r *Renderer) getStyleForFrame(kind microui.FrameKind, state microui.FrameState) (SkeuoColor, bool) {
	switch kind {
	case microui.FrameWindow:
		return r.theme.Panel, false
	case microui.FrameTitle:
		return r.theme.TitleBar, false
	case microui.FramePanel:
		return r.theme.Panel, true
	case microui.FrameButton, microui.FrameHeader:
		if state == microui.StateHover {
			return r.theme.ButtonHover, false
		}
		return r.theme.Button, state == microui.StateFocus
	case microui.FrameInput:
		if state == microui.StateFocus {
			return r.theme.InputFocus, true
		}
		return r.theme.Input, true
	case microui.FrameSliderThumb:
		if state == microui.StateHover {
			return r.theme.ButtonHover, false
		}
		return r.theme.Button, state == microui.StateFocus
	case microui.FrameScrollTrack:
		return r.theme.ScrollTrack, true
	case microui.FrameScrollThumb:
		return r.theme.ScrollThumb, false
	default:
		return r.theme.Panel, false
	}
}

// drawFrameFlat draws frames in pixel art style with double borders.
// All measurements use logical pixels (1 logical px = PixelScale screen px).
func (r *Renderer) drawFrameFlat(ui *microui.UI, info microui.FrameInfo) {
	rect := info.Rect
	x, y, w, h := rect.X, rect.Y, rect.W, rect.H

	// Convert logical pixels to screen pixels
	scale := r.theme.PixelScale
	if scale < 1 {
		scale = 2
	}
	border := r.theme.BorderSize * scale // Total border in screen px (e.g., 2 logical * 2 scale = 4)
	onePx := scale                       // 1 logical pixel in screen px

	// Shadow starts 3 logical pixels from top
	shadowInset := onePx * 3

	switch info.Kind {
	case microui.FrameWindow:
		// Window: shadow + outer border + fill + inner border (on top)
		outerBorder := r.theme.Panel.Shadow    // Black
		innerBorder := r.theme.Panel.Highlight // Gray
		fill := r.theme.Panel.Base             // Dark gray

		// 1. Drop shadow - subtle, starts 3 logical px from top
		if shadow := r.theme.ShadowColor(); shadow != nil {
			// Right shadow (starts 3 logical px down)
			ui.DrawRect(types.Rect{X: x + w, Y: y + shadowInset, W: onePx, H: h - shadowInset + onePx}, shadow)
			// Bottom shadow
			ui.DrawRect(types.Rect{X: x + shadowInset, Y: y + h, W: w - shadowInset + onePx, H: onePx}, shadow)
		}

		// 2. Outer border (1 logical px thick, with corner notch)
		r.drawPixelBorderUI(ui, x, y, w, h, onePx, outerBorder)

		// 3. Fill (inside outer border)
		ui.DrawRect(types.Rect{X: x + onePx, Y: y + onePx, W: w - onePx*2, H: h - onePx*2}, fill)

		// 4. Inner border ON TOP of fill
		r.drawPixelBorderUI(ui, x+onePx, y+onePx, w-onePx*2, h-onePx*2, onePx, innerBorder)

	case microui.FrameTitle:
		// Title bar - lighter fill inside the window's double border
		// Inset by BorderSize on left, top, right (not bottom - touches content)
		ui.DrawRect(types.Rect{
			X: x + border,
			Y: y + border,
			W: w - border*2,
			H: h - border, // Only inset top, not bottom
		}, r.theme.TitleBar.Base)

	case microui.FramePanel:
		// Panels inside windows - just fill
		ui.DrawRect(rect, r.theme.Panel.Base)

	case microui.FrameButton, microui.FrameHeader, microui.FrameSliderThumb:
		// Buttons/headers/slider thumb get double border
		colors := r.theme.Button
		if info.State == microui.StateHover {
			colors = r.theme.ButtonHover
		}

		// Outer border
		r.drawPixelBorderUI(ui, x, y, w, h, onePx, colors.Shadow)
		// Fill (inside outer border)
		ui.DrawRect(types.Rect{X: x + onePx, Y: y + onePx, W: w - onePx*2, H: h - onePx*2}, colors.Base)
		// Inner border on top
		r.drawPixelBorderUI(ui, x+onePx, y+onePx, w-onePx*2, h-onePx*2, onePx, colors.Highlight)

	case microui.FrameInput:
		// Input fields - style varies by state
		switch info.State {
		case microui.StateNormal:
			// Normal: single border dark box
			colors := r.theme.Input
			r.drawPixelBorderUI(ui, x, y, w, h, onePx, colors.Shadow)
			ui.DrawRect(types.Rect{X: x + onePx, Y: y + onePx, W: w - onePx*2, H: h - onePx*2}, colors.Base)

		case microui.StateHover:
			// Hover: 2 pixel border, lighter interior
			colors := r.theme.Input
			r.drawPixelBorderUI(ui, x, y, w, h, onePx, colors.Shadow)
			r.drawPixelBorderUI(ui, x+onePx, y+onePx, w-onePx*2, h-onePx*2, onePx, colors.Highlight)
			ui.DrawRect(types.Rect{X: x + onePx*2, Y: y + onePx*2, W: w - onePx*4, H: h - onePx*4}, colors.Highlight)

		case microui.StateFocus:
			// Focus: single border, light interior
			colors := r.theme.InputFocus
			r.drawPixelBorderUI(ui, x, y, w, h, onePx, colors.Shadow)
			ui.DrawRect(types.Rect{X: x + onePx, Y: y + onePx, W: w - onePx*2, H: h - onePx*2}, colors.Highlight)
		}

	case microui.FrameScrollTrack:
		// Floating scrollbar track - dark with alpha, no borders
		ui.DrawRect(rect, r.theme.ScrollTrack.Base)

	case microui.FrameScrollThumb:
		// Floating scrollbar thumb - solid color, no borders
		ui.DrawRect(rect, r.theme.ScrollThumb.Base)

	default:
		ui.DrawRect(rect, r.theme.Panel.Base)
	}
}

// drawPixelBorderUI draws a border using logical pixels.
// px is the logical pixel size (PixelScale). Border is px screen pixels thick.
// Corner is notched by px screen pixels.
func (r *Renderer) drawPixelBorderUI(ui *microui.UI, x, y, w, h, px int, c color.Color) {
	if w <= 0 || h <= 0 || w <= px*2 || h <= px*2 {
		// Too small for corner notch, draw simple rect border
		ui.DrawRect(types.Rect{X: x, Y: y, W: w, H: px}, c)             // Top
		ui.DrawRect(types.Rect{X: x, Y: y + h - px, W: w, H: px}, c)    // Bottom
		ui.DrawRect(types.Rect{X: x, Y: y + px, W: px, H: h - px*2}, c) // Left
		ui.DrawRect(types.Rect{X: x + w - px, Y: y + px, W: px, H: h - px*2}, c) // Right
		return
	}

	// Border with corner notch (corners cut by px screen pixels)
	// Top edge (skip corner notch area)
	ui.DrawRect(types.Rect{X: x + px, Y: y, W: w - px*2, H: px}, c)
	// Bottom edge (skip corner notch area)
	ui.DrawRect(types.Rect{X: x + px, Y: y + h - px, W: w - px*2, H: px}, c)
	// Left edge (skip corner notch area)
	ui.DrawRect(types.Rect{X: x, Y: y + px, W: px, H: h - px*2}, c)
	// Right edge (skip corner notch area)
	ui.DrawRect(types.Rect{X: x + w - px, Y: y + px, W: px, H: h - px*2}, c)
}

// drawBeveledRectUI draws a beveled or flat rectangle using ui.DrawRect for proper z-ordering.
// This adds commands to microui's buffer instead of drawing directly.
func (r *Renderer) drawBeveledRectUI(ui *microui.UI, rect types.Rect, colors SkeuoColor, sunken bool) {
	// Use flat style if theme specifies it
	if r.theme.UseFlat {
		r.drawFlatRectUI(ui, rect, colors, sunken)
		return
	}

	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	depth := r.theme.PixelScale

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

// drawFlatRectUI draws a pixel-art style rectangle with double borders and drop shadow.
// Structure: drop shadow -> outer border (Shadow) -> inner border (Highlight) -> fill (Base)
func (r *Renderer) drawFlatRectUI(ui *microui.UI, rect types.Rect, colors SkeuoColor, sunken bool) {
	x, y, w, h := rect.X, rect.Y, rect.W, rect.H
	radius := r.theme.CornerRadius

	// Colors: Shadow = outer border (dark), Highlight = inner border (lighter), Base = fill
	outerBorder := colors.Shadow
	innerBorder := colors.Highlight
	fill := colors.Base

	// For sunken elements, we might want different styling
	if sunken {
		// Sunken elements are typically darker, less prominent
		outerBorder = colors.Shadow
		innerBorder = colors.Shadow // Same as outer for sunken
	}

	// 1. Draw drop shadow (offset 1px right and down, on left/right/bottom edges)
	shadowColor := r.theme.ShadowColor()
	if shadowColor != nil {
		// Shadow on right edge (1px to the right of window)
		ui.DrawRect(types.Rect{X: x + w, Y: y + radius + 1, W: 1, H: h - radius}, shadowColor)
		// Shadow on bottom edge (1px below window)
		ui.DrawRect(types.Rect{X: x + radius + 1, Y: y + h, W: w - radius, H: 1}, shadowColor)
		// Shadow corner (bottom-right)
		if radius >= 1 {
			ui.DrawRect(types.Rect{X: x + w - 1, Y: y + h, W: 1, H: 1}, shadowColor)
			ui.DrawRect(types.Rect{X: x + w, Y: y + h - 1, W: 1, H: 1}, shadowColor)
		}
	}

	// 2. Draw outer border (black/dark) with 1px corner radius
	r.drawRoundedBorderUI(ui, x, y, w, h, radius, outerBorder)

	// 3. Draw inner border (gray/lighter) inset by 1px
	r.drawRoundedBorderUI(ui, x+1, y+1, w-2, h-2, radius, innerBorder)

	// 4. Draw fill (inset by 2px for both borders)
	r.fillRoundedRectUI(ui, x+2, y+2, w-4, h-4, 0, fill) // No radius needed for inner fill
}

// drawRoundedBorderUI draws a 1px border with corner notches using ui.DrawRect.
func (r *Renderer) drawRoundedBorderUI(ui *microui.UI, x, y, w, h, radius int, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}

	if radius <= 0 {
		// Simple box border
		ui.DrawRect(types.Rect{X: x, Y: y, W: w, H: 1}, c)         // Top
		ui.DrawRect(types.Rect{X: x, Y: y + h - 1, W: w, H: 1}, c) // Bottom
		ui.DrawRect(types.Rect{X: x, Y: y + 1, W: 1, H: h - 2}, c) // Left (excluding corners)
		ui.DrawRect(types.Rect{X: x + w - 1, Y: y + 1, W: 1, H: h - 2}, c) // Right (excluding corners)
		return
	}

	// Top edge (excluding corner pixels)
	ui.DrawRect(types.Rect{X: x + radius, Y: y, W: w - radius*2, H: 1}, c)
	// Bottom edge (excluding corner pixels)
	ui.DrawRect(types.Rect{X: x + radius, Y: y + h - 1, W: w - radius*2, H: 1}, c)
	// Left edge (excluding corner pixels)
	ui.DrawRect(types.Rect{X: x, Y: y + radius, W: 1, H: h - radius*2}, c)
	// Right edge (excluding corner pixels)
	ui.DrawRect(types.Rect{X: x + w - 1, Y: y + radius, W: 1, H: h - radius*2}, c)

	// Corner pixels for 1px radius (diagonal notch)
	if radius == 1 {
		// Top-left corner - draw the two pixels that form the diagonal
		ui.DrawRect(types.Rect{X: x + 1, Y: y, W: 1, H: 1}, c)
		ui.DrawRect(types.Rect{X: x, Y: y + 1, W: 1, H: 1}, c)
		// Top-right corner
		ui.DrawRect(types.Rect{X: x + w - 2, Y: y, W: 1, H: 1}, c)
		ui.DrawRect(types.Rect{X: x + w - 1, Y: y + 1, W: 1, H: 1}, c)
		// Bottom-left corner
		ui.DrawRect(types.Rect{X: x + 1, Y: y + h - 1, W: 1, H: 1}, c)
		ui.DrawRect(types.Rect{X: x, Y: y + h - 2, W: 1, H: 1}, c)
		// Bottom-right corner
		ui.DrawRect(types.Rect{X: x + w - 2, Y: y + h - 1, W: 1, H: 1}, c)
		ui.DrawRect(types.Rect{X: x + w - 1, Y: y + h - 2, W: 1, H: 1}, c)
	}
}

// fillRoundedRectUI fills a rectangle with corner notches using ui.DrawRect.
func (r *Renderer) fillRoundedRectUI(ui *microui.UI, x, y, w, h, radius int, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}

	if radius <= 0 {
		ui.DrawRect(types.Rect{X: x, Y: y, W: w, H: h}, c)
		return
	}

	// Fill excluding corner pixels
	// Top row (excluding corner pixels)
	ui.DrawRect(types.Rect{X: x + radius, Y: y, W: w - radius*2, H: radius}, c)
	// Middle rows (full width)
	ui.DrawRect(types.Rect{X: x, Y: y + radius, W: w, H: h - radius*2}, c)
	// Bottom row (excluding corner pixels)
	ui.DrawRect(types.Rect{X: x + radius, Y: y + h - radius, W: w - radius*2, H: radius}, c)
}

// fillRect draws a filled rectangle without anti-aliasing, clipped to current clip rect.
func (r *Renderer) fillRect(x, y, w, h int, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}

	// Apply clipping (always, matching ebiten renderer behavior)
	if x < r.clipRect.X {
		w -= r.clipRect.X - x
		x = r.clipRect.X
	}
	if y < r.clipRect.Y {
		h -= r.clipRect.Y - y
		y = r.clipRect.Y
	}
	if x+w > r.clipRect.X+r.clipRect.W {
		w = r.clipRect.X + r.clipRect.W - x
	}
	if y+h > r.clipRect.Y+r.clipRect.H {
		h = r.clipRect.Y + r.clipRect.H - y
	}

	if w <= 0 || h <= 0 {
		return
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
	return image.Rect(r.clipRect.X, r.clipRect.Y, r.clipRect.X+r.clipRect.W, r.clipRect.Y+r.clipRect.H)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
