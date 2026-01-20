package retro

import "image/color"

// SkeuoColor represents a color with highlight and shadow variants for 3D beveling.
type SkeuoColor struct {
	Base      color.Color // Main fill color
	Highlight color.Color // Light edge (top/left) - raised effect
	Shadow    color.Color // Dark edge (bottom/right) - raised effect
}

// Theme defines the complete color scheme for the retro renderer.
type Theme struct {
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
	Text     color.Color
	TextDim  color.Color // Disabled/secondary text

	// Canvas/drawing area (sunken)
	Canvas SkeuoColor

	// Scrollbar
	ScrollTrack SkeuoColor
	ScrollThumb SkeuoColor

	// Background behind windows
	Background color.Color

	// Bevel depth in pixels (1-3 recommended)
	BevelDepth int
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
		BevelDepth: 2,
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
		BevelDepth: 2,
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
		BevelDepth: 2,
	}
}
