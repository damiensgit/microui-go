package types

// Vec2 represents a 2D vector or point.
type Vec2 struct {
	X, Y int
}

// Add returns the sum of two vectors.
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the difference of two vectors.
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Rect represents a rectangle.
type Rect struct {
	X, Y, W, H int
}

// Contains returns true if the point is inside the rectangle.
func (r Rect) Contains(p Vec2) bool {
	return p.X >= r.X && p.X < r.X+r.W &&
		p.Y >= r.Y && p.Y < r.Y+r.H
}

// Empty returns true if the rectangle has zero or negative area.
func (r Rect) Empty() bool {
	return r.W <= 0 || r.H <= 0
}
