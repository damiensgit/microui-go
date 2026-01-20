package bubbletea

import "unicode/utf8"

// MonospaceFont implements types.Font for terminal text rendering.
// Each character occupies exactly 1 cell width and 1 cell height.
type MonospaceFont struct{}

// Width returns the width of text in terminal cells.
// For a monospace terminal, each rune is 1 cell wide.
// Note: CJK wide characters would need special handling for proper display.
func (f *MonospaceFont) Width(text string) int {
	return utf8.RuneCountInString(text)
}

// Height returns the font height in terminal rows (always 1).
func (f *MonospaceFont) Height() int {
	return 1
}
