package retro

import (
	"bytes"
	_ "embed"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed fonts/PixeloidSans.ttf
var pixeloidSansData []byte

//go:embed fonts/PixeloidSans-Bold.ttf
var pixeloidSansBoldData []byte

//go:embed fonts/PixeloidMono.ttf
var pixeloidMonoData []byte

// Font wraps a Pixeloid font face for rendering.
type Font struct {
	face     *text.GoTextFace
	source   *text.GoTextFaceSource
	size     float64
	lineHeight int
}

// FontStyle specifies which Pixeloid variant to use.
type FontStyle int

const (
	FontSans FontStyle = iota
	FontSansBold
	FontMono
)

// NewFont creates a new font with the specified style and size.
func NewFont(style FontStyle, size float64) (*Font, error) {
	var data []byte
	switch style {
	case FontSansBold:
		data = pixeloidSansBoldData
	case FontMono:
		data = pixeloidMonoData
	default:
		data = pixeloidSansData
	}

	source, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	face := &text.GoTextFace{
		Source: source,
		Size:   size,
	}

	return &Font{
		face:       face,
		source:     source,
		size:       size,
		lineHeight: int(size * 1.2),
	}, nil
}

// Draw renders text at the specified position.
func (f *Font) Draw(target *ebiten.Image, str string, x, y int, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(target, str, f.face, op)
}

// Width returns the width of the text in pixels.
func (f *Font) Width(str string) int {
	w, _ := text.Measure(str, f.face, float64(f.lineHeight))
	return int(w)
}

// Height returns the line height in pixels.
func (f *Font) Height() int {
	return f.lineHeight
}

// SetSize changes the font size.
func (f *Font) SetSize(size float64) {
	f.size = size
	f.face.Size = size
	f.lineHeight = int(size * 1.2)
}
