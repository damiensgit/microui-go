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

//go:embed fonts/m6x11.ttf
var m6x11Data []byte

//go:embed fonts/m5x7.ttf
var m5x7Data []byte

//go:embed fonts/QuinqueFive.ttf
var quinqueFiveData []byte

// FontFace is the interface for font rendering.
type FontFace interface {
	Draw(target *ebiten.Image, str string, x, y int, c color.Color)
	Width(str string) int
	Height() int
}

// Font wraps a Pixeloid font face for rendering.
type Font struct {
	face       *text.GoTextFace
	source     *text.GoTextFaceSource
	size       float64
	lineHeight int
	yOffset    int // Vertical adjustment for baseline alignment
}

// FontStyle specifies which Pixeloid variant to use.
type FontStyle int

const (
	FontSans FontStyle = iota
	FontSansBold
	FontMono
	FontM6x11      // Pixel-perfect at 16, 32, 48px
	FontM5x7       // Pixel-perfect at 16, 32, 48px
	FontQuinque    // Pixel-perfect at 5, 10, 15, 20px
)

// NewFont creates a new font with the specified style and size.
func NewFont(style FontStyle, size float64) (*Font, error) {
	var data []byte
	var yOffset int
	switch style {
	case FontSansBold:
		data = pixeloidSansBoldData
	case FontMono:
		data = pixeloidMonoData
	case FontM6x11:
		data = m6x11Data
		yOffset = int(size / 8) // Adjust baseline down
	case FontM5x7:
		data = m5x7Data
		yOffset = int(size / 8) // Adjust baseline down (4px at 32px)
	case FontQuinque:
		data = quinqueFiveData
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
		yOffset:    yOffset,
	}, nil
}

// Draw renders text at the specified position.
func (f *Font) Draw(target *ebiten.Image, str string, x, y int, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y+f.yOffset))
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
