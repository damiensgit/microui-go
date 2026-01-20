package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// MetaballsConfig configures the metaballs effect
type MetaballsConfig struct {
	GridResolution int     // Resolution divisor (higher = faster but blockier, e.g., 4 = 1/4 resolution)
	BallCount      int     // Number of metaballs
	Threshold      float64 // Metaball threshold (typically 1.0)
	ColorCycle     float64 // Speed of color cycling
	Speed          float64 // Speed multiplier for ball movement
}

// DefaultMetaballsConfig returns sensible defaults
func DefaultMetaballsConfig() MetaballsConfig {
	return MetaballsConfig{
		GridResolution: 6,   // Higher default for better perf on large screens
		BallCount:      6,   // 6 bouncing balls
		Threshold:      1.0, // Standard threshold
		ColorCycle:     0.5, // Moderate color cycling speed
		Speed:          1.0, // Normal speed
	}
}

// Metaballs handles the animated metaballs background effect
type Metaballs struct {
	config MetaballsConfig

	// Ball positions and velocities
	ballX, ballY   []float64
	ballVX, ballVY []float64
	ballRadius     []float64

	// Mouse tracking ball
	mouseX, mouseY float64 // Normalized 0-1 coordinates
	mouseRadius    float64

	// Cached computation grid
	gridW, gridH int
	gridImage    *ebiten.Image

	// Preallocated buffers (avoid allocs in Draw)
	pixels  []byte
	ballGX  []float64
	ballGY  []float64
	ballGR2 []float64

	// Animation time
	time float64

	// Screen dimensions (for detecting resize)
	screenW, screenH int
}

// NewMetaballs creates a new metaballs effect
func NewMetaballs(config MetaballsConfig) *Metaballs {
	m := &Metaballs{
		config:      config,
		ballX:       make([]float64, config.BallCount),
		ballY:       make([]float64, config.BallCount),
		ballVX:      make([]float64, config.BallCount),
		ballVY:      make([]float64, config.BallCount),
		ballRadius:  make([]float64, config.BallCount),
		ballGX:      make([]float64, config.BallCount),
		ballGY:      make([]float64, config.BallCount),
		ballGR2:     make([]float64, config.BallCount),
		mouseX:      0.5,
		mouseY:      0.5,
		mouseRadius: 0.1,
	}

	// Initialize balls with varied properties
	for i := 0; i < config.BallCount; i++ {
		// Stagger initial positions using simple ratios (avoid trig at init)
		t := float64(i) / float64(config.BallCount)
		// Simple circular distribution without sin/cos
		m.ballX[i] = 0.5 + 0.3*(2*t-1)
		m.ballY[i] = 0.5 + 0.3*(1-2*((t*2)-float64(int(t*2))))

		// Varied velocities
		speed := 0.15 + 0.1*float64(i%3)
		m.ballVX[i] = speed * (0.7 - t)
		m.ballVY[i] = speed * (t - 0.3)

		// Varied radii
		m.ballRadius[i] = 0.08 + 0.04*float64(i%3)
	}

	return m
}

// Update advances the animation by dt seconds
func (m *Metaballs) Update(dt float64, screenW, screenH int) {
	dt *= m.config.Speed
	m.time += dt

	// Update ball positions with bouncing
	for i := 0; i < m.config.BallCount; i++ {
		m.ballX[i] += m.ballVX[i] * dt
		m.ballY[i] += m.ballVY[i] * dt

		if m.ballX[i] < 0 {
			m.ballX[i] = 0
			m.ballVX[i] = -m.ballVX[i]
		} else if m.ballX[i] > 1.0 {
			m.ballX[i] = 1.0
			m.ballVX[i] = -m.ballVX[i]
		}
		if m.ballY[i] < 0 {
			m.ballY[i] = 0
			m.ballVY[i] = -m.ballVY[i]
		} else if m.ballY[i] > 1.0 {
			m.ballY[i] = 1.0
			m.ballVY[i] = -m.ballVY[i]
		}
	}

	// Check if we need to recreate the grid image
	gridW := screenW / m.config.GridResolution
	gridH := screenH / m.config.GridResolution
	if gridW < 1 {
		gridW = 1
	}
	if gridH < 1 {
		gridH = 1
	}

	if m.gridImage == nil || m.gridW != gridW || m.gridH != gridH {
		m.gridW = gridW
		m.gridH = gridH
		m.gridImage = ebiten.NewImage(gridW, gridH)
		m.pixels = make([]byte, gridW*gridH*4)
	}

	m.screenW = screenW
	m.screenH = screenH
}

// SetMousePosition updates the mouse tracking ball position
func (m *Metaballs) SetMousePosition(mx, my int) {
	if m.screenW > 0 && m.screenH > 0 {
		m.mouseX = float64(mx) / float64(m.screenW)
		m.mouseY = float64(my) / float64(m.screenH)
	}
}

// Draw renders the metaballs to the screen
func (m *Metaballs) Draw(screen *ebiten.Image, baseColor color.RGBA) {
	if m.gridImage == nil || m.pixels == nil {
		return
	}

	gridW := m.gridW
	gridH := m.gridH
	ballCount := m.config.BallCount
	threshold := m.config.Threshold
	invThreshold := 1.0 / threshold

	// Precompute ball positions in grid coordinates (reuse slices)
	for i := 0; i < ballCount; i++ {
		m.ballGX[i] = m.ballX[i] * float64(gridW)
		m.ballGY[i] = m.ballY[i] * float64(gridH)
		r := m.ballRadius[i] * float64(gridW)
		m.ballGR2[i] = r * r
	}

	// Mouse ball in grid coordinates
	mouseGX := m.mouseX * float64(gridW)
	mouseGY := m.mouseY * float64(gridH)
	mouseR := m.mouseRadius * float64(gridW)
	mouseGR2 := mouseR * mouseR

	// Precompute color cycle base (avoid per-pixel math.Floor)
	hueBase := m.time * m.config.ColorCycle
	// Keep hue in 0-6 range for faster HSV
	hueBase = hueBase - float64(int(hueBase))
	hueBase *= 6.0

	invGridSum := 0.5 / float64(gridW+gridH)

	// Process each pixel in the grid
	pixels := m.pixels
	idx := 0
	for y := 0; y < gridH; y++ {
		fy := float64(y) + 0.5
		for x := 0; x < gridW; x++ {
			fx := float64(x) + 0.5

			// Calculate metaball field value: sum(r^2 / d^2)
			var field float64
			for i := 0; i < ballCount; i++ {
				dx := fx - m.ballGX[i]
				dy := fy - m.ballGY[i]
				distSq := dx*dx + dy*dy
				if distSq < 0.0001 {
					distSq = 0.0001
				}
				field += m.ballGR2[i] / distSq
			}

			// Mouse ball contribution
			dx := fx - mouseGX
			dy := fy - mouseGY
			distSq := dx*dx + dy*dy
			if distSq < 0.0001 {
				distSq = 0.0001
			}
			field += mouseGR2 / distSq

			// Fast color calculation
			if field >= threshold {
				// Inside metaball
				intensity := (field - threshold) * invThreshold
				if intensity > 1.0 {
					intensity = 1.0
				}

				// Simplified HSV: hue varies by position, sat=0.7, val=0.4+0.5*intensity
				hue := hueBase + float64(x+y)*invGridSum*3.0 // *6*0.5
				r, g, b := fastHSV(hue, 0.7, 0.4+0.5*intensity)

				alpha := 128 + uint8(intensity*127)

				pixels[idx] = r
				pixels[idx+1] = g
				pixels[idx+2] = b
				pixels[idx+3] = alpha
			} else {
				// Outside - check for glow
				norm := field * invThreshold
				glow := norm * norm

				if glow > 0.05 {
					hue := hueBase + float64(x+y)*invGridSum*3.0
					v := 0.3 * glow
					r, g, b := fastHSV(hue, 0.6*glow, v)
					alpha := uint8(glow * 153) // 0.6 * 255

					pixels[idx] = r
					pixels[idx+1] = g
					pixels[idx+2] = b
					pixels[idx+3] = alpha
				} else {
					// Transparent
					pixels[idx] = 0
					pixels[idx+1] = 0
					pixels[idx+2] = 0
					pixels[idx+3] = 0
				}
			}
			idx += 4
		}
	}

	m.gridImage.WritePixels(pixels)

	// Fill background then draw metaballs
	screen.Fill(baseColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(m.config.GridResolution), float64(m.config.GridResolution))
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(m.gridImage, op)
}

// fastHSV converts HSV to RGB without math.Floor (hue already in 0-6 range)
func fastHSV(h, s, v float64) (r, g, b uint8) {
	// Wrap hue to 0-6
	for h >= 6 {
		h -= 6
	}
	for h < 0 {
		h += 6
	}

	hi := int(h)
	f := h - float64(hi)
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	var rf, gf, bf float64
	switch hi {
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
	default: // 5
		rf, gf, bf = v, p, q
	}

	return uint8(rf * 255), uint8(gf * 255), uint8(bf * 255)
}
