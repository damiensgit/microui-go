package atlas

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Font renders text using the microui bitmap atlas
type Font struct {
	atlas *ebiten.Image
}

// NewFont creates a new atlas-based font
func NewFont() *Font {
	// Create NRGBA image with white color and grayscale as alpha
	// This matches C microui's GL_ALPHA texture approach
	// The grayscale value becomes the alpha, color is applied when drawing
	img := image.NewNRGBA(image.Rect(0, 0, AtlasWidth, AtlasHeight))
	for y := 0; y < AtlasHeight; y++ {
		for x := 0; x < AtlasWidth; x++ {
			alpha := AtlasTexture[y*AtlasWidth+x]
			// White color with variable alpha - color tinting happens via ColorScale
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: alpha})
		}
	}

	// Convert to Ebiten image
	ebitenImg := ebiten.NewImageFromImage(img)

	return &Font{
		atlas: ebitenImg,
	}
}

// Draw renders text at the specified position with the given color
func (f *Font) Draw(target *ebiten.Image, text string, x, y int, c color.Color) {
	if f.atlas == nil {
		return
	}

	r, g, b, a := c.RGBA()
	cr := float64(r) / 0xffff
	cg := float64(g) / 0xffff
	cb := float64(b) / 0xffff
	ca := float64(a) / 0xffff

	curX := x
	for _, ch := range text {
		if ch == '\n' {
			curX = x
			y += 17 // Font height
			continue
		}

		charIdx := AtlasFont + int(ch)
		rect, ok := AtlasRects[charIdx]
		if !ok {
			// Unknown character, use space width
			curX += 6
			continue
		}

		// Get character from atlas
		srcRect := image.Rect(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H)
		charImg := f.atlas.SubImage(srcRect).(*ebiten.Image)

		// Draw with color (nearest-neighbor for pixel-perfect text)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(curX), float64(y))
		op.ColorScale.Scale(float32(cr), float32(cg), float32(cb), float32(ca))
		op.Filter = ebiten.FilterNearest
		target.DrawImage(charImg, op)

		curX += rect.W
	}
}

// Width returns the pixel width of the given text
func (f *Font) Width(text string) int {
	width := 0
	for _, ch := range text {
		if ch == '\n' {
			continue
		}
		charIdx := AtlasFont + int(ch)
		rect, ok := AtlasRects[charIdx]
		if !ok {
			width += 6 // Default width for unknown chars
			continue
		}
		width += rect.W
	}
	return width
}

// Height returns the font height in pixels
func (f *Font) Height() int {
	return 17
}

// GetIconRect returns the atlas rect for an icon
func GetIconRect(iconID int) (Rect, bool) {
	rect, ok := AtlasRects[iconID]
	return rect, ok
}

// GetAtlasImage returns the atlas image for icon rendering
func (f *Font) GetAtlasImage() *ebiten.Image {
	return f.atlas
}

// HasIcon returns true if the atlas has the specified icon
func (f *Font) HasIcon(iconID int) bool {
	_, ok := AtlasRects[iconID]
	return ok
}

// DrawIcon renders an icon from the atlas centered within the destination rect
func (f *Font) DrawIcon(target *ebiten.Image, iconID int, destRect image.Rectangle, c color.Color) {
	if f.atlas == nil {
		return
	}

	atlasRect, ok := AtlasRects[iconID]
	if !ok {
		return
	}

	r, g, b, a := c.RGBA()
	cr := float64(r) / 0xffff
	cg := float64(g) / 0xffff
	cb := float64(b) / 0xffff
	ca := float64(a) / 0xffff

	// Get icon from atlas
	srcRect := image.Rect(atlasRect.X, atlasRect.Y, atlasRect.X+atlasRect.W, atlasRect.Y+atlasRect.H)
	iconImg := f.atlas.SubImage(srcRect).(*ebiten.Image)

	// Center icon within destination rect
	destW := destRect.Dx()
	destH := destRect.Dy()
	offsetX := (destW - atlasRect.W) / 2
	offsetY := (destH - atlasRect.H) / 2

	// Draw with color (nearest-neighbor for pixel-perfect icons)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(destRect.Min.X+offsetX), float64(destRect.Min.Y+offsetY))
	op.ColorScale.Scale(float32(cr), float32(cg), float32(cb), float32(ca))
	op.Filter = ebiten.FilterNearest
	target.DrawImage(iconImg, op)
}
