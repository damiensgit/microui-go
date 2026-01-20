package bubbletea

// Icon IDs (must match microui.IconClose, IconCheck, etc.)
const (
	iconClose     = 1
	iconCheck     = 2
	iconCollapsed = 3
	iconExpanded  = 4
	iconResize    = 5
)

// Icon rune mappings for terminal display.
// Classic Turbo Vision style characters.
const (
	IconRuneClose     = '\u25A0' // ■ (black square - classic TV close button)
	IconRuneCheck     = '\u2713' // ✓ (check mark)
	IconRuneCollapsed = '\u25BA' // ► (black right-pointing pointer)
	IconRuneExpanded  = '\u25BC' // ▼ (black down-pointing triangle)
	IconRuneFallback  = '\u25A1' // □ (white square, fallback)
	IconRuneResize    = '\u2518' // ┘ (box drawings light up and left - resize gripper)
)

// IconToRune converts a microui icon ID to a Unicode rune for terminal display.
func IconToRune(id int) rune {
	switch id {
	case iconClose:
		return IconRuneClose
	case iconCheck:
		return IconRuneCheck
	case iconCollapsed:
		return IconRuneCollapsed
	case iconExpanded:
		return IconRuneExpanded
	case iconResize:
		return IconRuneResize
	default:
		return IconRuneFallback
	}
}
