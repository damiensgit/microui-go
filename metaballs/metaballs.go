// Package metaballs provides a metaball field simulation that can be rendered
// by different backends (GUI pixels or TUI half-block characters).
package metaballs

import "math"

// Config configures the metaballs effect
type Config struct {
	BallCount int     // Number of metaballs
	Threshold float64 // Metaball threshold (typically 1.0)
	Speed     float64 // Speed multiplier for ball movement
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		BallCount: 5,
		Threshold: 1.0,
		Speed:     1.0,
	}
}

// Field represents the metaball simulation state
type Field struct {
	config Config

	// Ball positions and velocities (normalized 0-1 coordinates)
	ballX, ballY   []float64
	ballVX, ballVY []float64
	ballRadius     []float64

	// Animation time
	time float64

	// Field dimensions (for bouncing)
	width, height float64
}

// New creates a new metaball field
func New(config Config) *Field {
	f := &Field{
		config:     config,
		ballX:      make([]float64, config.BallCount),
		ballY:      make([]float64, config.BallCount),
		ballVX:     make([]float64, config.BallCount),
		ballVY:     make([]float64, config.BallCount),
		ballRadius: make([]float64, config.BallCount),
		width:      1.0,
		height:     1.0,
	}

	// Initialize balls with varied properties
	for i := 0; i < config.BallCount; i++ {
		// Stagger initial positions in a circle
		angle := float64(i) * 2.0 * math.Pi / float64(config.BallCount)
		f.ballX[i] = 0.5 + 0.3*math.Cos(angle)
		f.ballY[i] = 0.5 + 0.3*math.Sin(angle)

		// Varied velocities (normalized coordinates per second)
		speed := 0.12 + 0.08*float64(i%3)
		velAngle := angle + math.Pi/4
		f.ballVX[i] = speed * math.Cos(velAngle)
		f.ballVY[i] = speed * math.Sin(velAngle)

		// Varied radii for visual interest
		f.ballRadius[i] = 0.06 + 0.03*float64(i%3)
	}

	return f
}

// Update advances the animation by dt seconds
func (f *Field) Update(dt float64) {
	// Apply speed multiplier
	dt *= f.config.Speed
	f.time += dt

	// Update ball positions with bouncing
	for i := 0; i < f.config.BallCount; i++ {
		f.ballX[i] += f.ballVX[i] * dt
		f.ballY[i] += f.ballVY[i] * dt

		// Bounce at edges
		if f.ballX[i] < 0 {
			f.ballX[i] = 0
			f.ballVX[i] = -f.ballVX[i]
		}
		if f.ballX[i] > f.width {
			f.ballX[i] = f.width
			f.ballVX[i] = -f.ballVX[i]
		}
		if f.ballY[i] < 0 {
			f.ballY[i] = 0
			f.ballVY[i] = -f.ballVY[i]
		}
		if f.ballY[i] > f.height {
			f.ballY[i] = f.height
			f.ballVY[i] = -f.ballVY[i]
		}
	}
}

// SetBounds sets the field bounds (for bouncing). Default is 1.0 x 1.0.
func (f *Field) SetBounds(width, height float64) {
	f.width = width
	f.height = height
}

// Time returns the current animation time
func (f *Field) Time() float64 {
	return f.time
}

// Sample calculates the metaball field value at the given normalized coordinates.
// Returns a value >= threshold if inside a blob, < threshold if outside.
// The value represents the "intensity" - higher values are deeper inside blobs.
func (f *Field) Sample(x, y float64) float64 {
	var field float64
	for i := 0; i < f.config.BallCount; i++ {
		dx := x - f.ballX[i]
		dy := y - f.ballY[i]
		distSq := dx*dx + dy*dy
		if distSq < 0.0001 {
			distSq = 0.0001 // Prevent division by zero
		}
		// Field contribution: r^2 / d^2
		r := f.ballRadius[i]
		field += (r * r) / distSq
	}
	return field
}

// IsInside returns true if the point is inside the metaball surface
func (f *Field) IsInside(x, y float64) bool {
	return f.Sample(x, y) >= f.config.Threshold
}

// Threshold returns the configured threshold
func (f *Field) Threshold() float64 {
	return f.config.Threshold
}

// SetSpeed sets the speed multiplier
func (f *Field) SetSpeed(speed float64) {
	f.config.Speed = speed
}

// SetThreshold sets the metaball threshold
func (f *Field) SetThreshold(threshold float64) {
	f.config.Threshold = threshold
}
