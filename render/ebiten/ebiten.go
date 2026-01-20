package ebiten

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/user/microui-go/types"
)

// emptyImage is a 1x1 white pixel used as source for DrawTriangles with solid colors
var emptyImage = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
}()

// IconProvider can render icons from an atlas
type IconProvider interface {
	DrawIcon(target *ebiten.Image, iconID int, rect image.Rectangle, c color.Color)
	HasIcon(iconID int) bool
}

// Renderer implements microui.Renderer using Ebiten v2.
type Renderer struct {
	target       *ebiten.Image
	font         Font
	iconProvider IconProvider
	clipRect     types.Rect
	mu           sync.Mutex
}

// NewRenderer creates a new Ebiten renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		clipRect: types.Rect{X: 0, Y: 0, W: 10000, H: 10000},
		font:     &defaultFont{},
	}
}

// SetFont sets the font used for text rendering.
// Pass nil to use the default debug font.
func (r *Renderer) SetFont(font Font) {
	r.mu.Lock()
	if font == nil {
		r.font = &defaultFont{}
	} else {
		r.font = font
	}
	r.mu.Unlock()
}

// SetIconProvider sets the icon provider for atlas-based icon rendering.
// Pass nil to use default geometric icons.
func (r *Renderer) SetIconProvider(provider IconProvider) {
	r.mu.Lock()
	r.iconProvider = provider
	r.mu.Unlock()
}

// SetTarget sets the render target.
func (r *Renderer) SetTarget(target *ebiten.Image) {
	r.mu.Lock()
	r.target = target
	r.mu.Unlock()
}

// DrawRect fills a rectangle with the given color.
func (r *Renderer) DrawRect(pos, size types.Vec2, c color.Color) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.target == nil {
		return
	}

	// Apply clipping
	x, y, w, h := r.applyClip(pos.X, pos.Y, size.X, size.Y)
	if w <= 0 || h <= 0 {
		return
	}

	rgba := color.NRGBAModel.Convert(c).(color.NRGBA)
	vector.DrawFilledRect(
		r.target,
		float32(x), float32(y),
		float32(w), float32(h),
		rgba,
		false,
	)
}

// DrawBox draws an unfilled rectangle outline (border only).
func (r *Renderer) DrawBox(rect types.Rect, c color.Color) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.target == nil {
		return
	}

	rgba := color.NRGBAModel.Convert(c).(color.NRGBA)

	// Draw box using 4 filled rects

	// Top edge
	r.drawClippedRect(rect.X+1, rect.Y, rect.W-2, 1, rgba)
	// Bottom edge
	r.drawClippedRect(rect.X+1, rect.Y+rect.H-1, rect.W-2, 1, rgba)
	// Left edge
	r.drawClippedRect(rect.X, rect.Y, 1, rect.H, rgba)
	// Right edge
	r.drawClippedRect(rect.X+rect.W-1, rect.Y, 1, rect.H, rgba)
}

// drawClippedRect draws a filled rect with clipping applied
func (r *Renderer) drawClippedRect(x, y, w, h int, rgba color.NRGBA) {
	// Apply clipping
	x, y, w, h = r.applyClip(x, y, w, h)
	if w <= 0 || h <= 0 {
		return
	}

	vector.DrawFilledRect(
		r.target,
		float32(x), float32(y),
		float32(w), float32(h),
		rgba,
		false,
	)
}

// DrawText renders text at the specified position with proper clipping.
func (r *Renderer) DrawText(text string, pos types.Vec2, font types.Font, c color.Color) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.target == nil || text == "" {
		return
	}

	// Get text dimensions
	textW := r.font.Width(text)
	textH := r.font.Height()

	// Quick reject: completely outside clip rect
	if pos.X >= r.clipRect.X+r.clipRect.W || pos.Y >= r.clipRect.Y+r.clipRect.H {
		return
	}
	if pos.X+textW <= r.clipRect.X || pos.Y+textH <= r.clipRect.Y {
		return
	}

	// Use SubImage clipping for proper text clipping
	// Get the intersection of the text bounds and clip rect
	clipX := r.clipRect.X
	clipY := r.clipRect.Y
	clipW := r.clipRect.W
	clipH := r.clipRect.H

	// Clamp clip rect to target bounds
	targetBounds := r.target.Bounds()
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

	// Get SubImage for clipping
	clipRect := image.Rect(clipX, clipY, clipX+clipW, clipY+clipH)
	subImg := r.target.SubImage(clipRect).(*ebiten.Image)

	// Draw text to SubImage with adjusted coordinates
	// SubImage coordinates are relative to the original image, so we use absolute coords
	r.font.Draw(subImg, text, pos.X, pos.Y, c)
}

// Icon IDs (must match microui constants)
const (
	iconClose     = 1
	iconCheck     = 2
	iconCollapsed = 3
	iconExpanded  = 4
	iconResize    = 5
)

// DrawIcon renders an icon with proper clipping.
// Uses atlas icons if an IconProvider is set, otherwise falls back to geometric shapes.
func (r *Renderer) DrawIcon(id int, rect types.Rect, c color.Color) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.target == nil {
		return
	}

	// Try atlas-based icon first
	if r.iconProvider != nil && r.iconProvider.HasIcon(id) {
		// Get clipped subimage
		clipX := r.clipRect.X
		clipY := r.clipRect.Y
		clipW := r.clipRect.W
		clipH := r.clipRect.H

		targetBounds := r.target.Bounds()
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

		if clipW > 0 && clipH > 0 {
			clipImgRect := image.Rect(clipX, clipY, clipX+clipW, clipY+clipH)
			subImg := r.target.SubImage(clipImgRect).(*ebiten.Image)
			iconRect := image.Rect(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H)
			r.iconProvider.DrawIcon(subImg, id, iconRect, c)
		}
		return
	}

	// Fall back to geometric shapes

	// Quick reject: completely outside clip rect
	if rect.X >= r.clipRect.X+r.clipRect.W || rect.Y >= r.clipRect.Y+r.clipRect.H {
		return
	}
	if rect.X+rect.W <= r.clipRect.X || rect.Y+rect.H <= r.clipRect.Y {
		return
	}

	// Get SubImage for clipping
	clipX := r.clipRect.X
	clipY := r.clipRect.Y
	clipW := r.clipRect.W
	clipH := r.clipRect.H

	// Clamp clip rect to target bounds
	targetBounds := r.target.Bounds()
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

	clipImgRect := image.Rect(clipX, clipY, clipX+clipW, clipY+clipH)
	subImg := r.target.SubImage(clipImgRect).(*ebiten.Image)

	// Calculate center and size
	cx := float32(rect.X + rect.W/2)
	cy := float32(rect.Y + rect.H/2)
	size := float32(rect.W)
	if float32(rect.H) < size {
		size = float32(rect.H)
	}
	size *= 0.6 // Icon is 60% of rect size

	rgba := color.NRGBAModel.Convert(c).(color.NRGBA)

	switch id {
	case iconClose: // X shape
		half := size / 2
		vector.StrokeLine(subImg, cx-half, cy-half, cx+half, cy+half, 2, rgba, false)
		vector.StrokeLine(subImg, cx+half, cy-half, cx-half, cy+half, 2, rgba, false)

	case iconCheck: // Checkmark shape
		// Classic checkmark: short leg down-left, long leg up-right
		vector.StrokeLine(subImg, cx-size*0.3, cy-size*0.05, cx-size*0.05, cy+size*0.2, 1.5, rgba, false)
		vector.StrokeLine(subImg, cx-size*0.05, cy+size*0.2, cx+size*0.35, cy-size*0.3, 1.5, rgba, false)

	case iconCollapsed: // Right-pointing triangle (>) - filled
		x1, y1 := cx-size*0.2, cy-size*0.35
		x2, y2 := cx+size*0.3, cy
		x3, y3 := cx-size*0.2, cy+size*0.35
		// Use Path for filled triangle
		var path vector.Path
		path.MoveTo(x1, y1)
		path.LineTo(x2, y2)
		path.LineTo(x3, y3)
		path.Close()
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vs {
			vs[i].SrcX = 1
			vs[i].SrcY = 1
			vs[i].ColorR = float32(rgba.R) / 255
			vs[i].ColorG = float32(rgba.G) / 255
			vs[i].ColorB = float32(rgba.B) / 255
			vs[i].ColorA = float32(rgba.A) / 255
		}
		subImg.DrawTriangles(vs, is, emptyImage, nil)

	case iconExpanded: // Down-pointing triangle (v) - filled
		x1, y1 := cx-size*0.35, cy-size*0.2
		x2, y2 := cx+size*0.35, cy-size*0.2
		x3, y3 := cx, cy+size*0.3
		// Use Path for filled triangle
		var path vector.Path
		path.MoveTo(x1, y1)
		path.LineTo(x2, y2)
		path.LineTo(x3, y3)
		path.Close()
		vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vs {
			vs[i].SrcX = 1
			vs[i].SrcY = 1
			vs[i].ColorR = float32(rgba.R) / 255
			vs[i].ColorG = float32(rgba.G) / 255
			vs[i].ColorB = float32(rgba.B) / 255
			vs[i].ColorA = float32(rgba.A) / 255
		}
		subImg.DrawTriangles(vs, is, emptyImage, nil)

	case iconResize:
		// GUI: no visual for resize gripper - the area still works for dragging
	}
}

// SetClip sets the clipping rectangle.
func (r *Renderer) SetClip(rect types.Rect) {
	r.mu.Lock()
	r.clipRect = rect
	r.mu.Unlock()
}

// Scrollbar colors for GUI rendering
var (
	scrollTrackColor = color.RGBA{R: 50, G: 50, B: 60, A: 255}
	scrollThumbColor = color.RGBA{R: 100, G: 100, B: 120, A: 255}
)

// DrawScrollTrack draws a scrollbar track (background).
func (r *Renderer) DrawScrollTrack(rect types.Rect) {
	r.DrawRect(types.Vec2{X: rect.X, Y: rect.Y}, types.Vec2{X: rect.W, Y: rect.H}, scrollTrackColor)
}

// DrawScrollThumb draws a scrollbar thumb (draggable part).
func (r *Renderer) DrawScrollThumb(rect types.Rect) {
	r.DrawRect(types.Vec2{X: rect.X, Y: rect.Y}, types.Vec2{X: rect.W, Y: rect.H}, scrollThumbColor)
}

func (r *Renderer) applyClip(x, y, w, h int) (int, int, int, int) {
	// Simple rectangle intersection
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
	return x, y, w, h
}

// Font is the interface for text rendering in Ebiten.
type Font interface {
	Draw(target *ebiten.Image, text string, x, y int, c color.Color)
	Width(text string) int
	Height() int
}

// defaultFont is a simple placeholder font.
type defaultFont struct{}

func (d *defaultFont) Draw(target *ebiten.Image, text string, x, y int, c color.Color) {
	// Simple text drawing using ebitenutil
	// Note: DebugPrintAt only draws white text, but it's enough for the demo
	// In production, use a proper font like ebiten/text with font faces
	ebitenutil.DebugPrintAt(target, text, x, y)
}

func (d *defaultFont) Width(text string) int {
	return len(text) * 8 // Approximate monospace width
}

func (d *defaultFont) Height() int {
	return 16 // Default debug font height
}
