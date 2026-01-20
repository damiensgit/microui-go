package types

// Font represents a font that can measure text dimensions.
type Font interface {
	// Width returns the width of the text in pixels.
	Width(text string) int
	// Height returns the font height in pixels.
	Height() int
}

// MockFont is a test implementation of Font.
type MockFont struct {
	Widths map[rune]int
	H      int
}

// Width returns the sum of character widths.
func (m *MockFont) Width(text string) int {
	w := 0
	for _, r := range text {
		if rw, ok := m.Widths[r]; ok {
			w += rw
		} else {
			w += 8 // default
		}
	}
	return w
}

// Height returns the mock font height.
func (m *MockFont) Height() int {
	if m.H > 0 {
		return m.H
	}
	return 16 // default
}
