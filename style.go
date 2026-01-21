package microui

import "github.com/user/microui-go/types"

// Style configures the visual appearance of UI controls.
type Style struct {
	// Typography
	Font types.Font

	// Colors
	Colors types.ThemeColors

	// Sizing
	Size           types.Vec2 // Default control size
	Padding        types.Vec2 // Internal padding
	Spacing        int        // Space between controls
	Indent         int        // Tree/header indent
	TitleHeight    int        // Window title bar height
	ScrollbarSize   int // Scrollbar track width
	ScrollbarMargin int // Visible margin around scrollbar track
	ScrollbarBorder int // Visual border width that scrollbar must clear
	ThumbSize       int // Slider thumb size

	// Content inset (for TUI where borders are drawn on-edge)
	BorderWidth int // Window border width - content is inset by this amount
	                // GUI: 0 (borders drawn outside/expanded, no inset needed)
	                // TUI: 1 (borders drawn on-edge, content must be inset)

	// Visual border widths for clipping (renderer-specific, GUI only)
	// These define how much space the rendered borders occupy, used for:
	// - Clipping content so it doesn't render into borders
	// - Positioning scrollbars inside the visual window area
	WindowBorder int // Window frame border width (for content clipping)

	// Control layout (buttons, inputs, etc.)
	// Layout: rect edge | margin (border) | padding | content | padding | margin | rect edge
	ControlMargin  int // Visual border width (clipping boundary)
	ControlPadding int // Additional space between border and content
}

// GUIStyle returns a style optimized for pixel-based GUI rendering.
func GUIStyle() Style {
	return Style{
		Font:          &types.MockFont{},
		Colors:        types.DarkTheme(),
		Size:          types.Vec2{X: 68, Y: 10}, // Pixel dimensions for controls
		Padding:       types.Vec2{X: 5, Y: 5},   // 5 pixels internal padding
		Spacing:       4,                        // 4 pixels between controls
		Indent:        24,                       // 24 pixels for tree indentation
		TitleHeight:   24,                       // 24 pixel title bar
		ScrollbarSize: 12,                       // 12 pixel scrollbar width
		ThumbSize:     8,                        // 8 pixel slider thumb
		// BorderWidth: 0 (default) - GUI borders are expanded outside, no content inset needed
	}
}

// TUIStyle returns a style optimized for cell-based terminal rendering.
// Use this for TUI renderers like Bubble Tea, tcell, termbox, etc.
func TUIStyle() Style {
	return Style{
		Font:          &types.MockFont{},
		Colors:        types.DarkTheme(),
		Size:          types.Vec2{X: 20, Y: 1}, // Cell dimensions for controls
		Padding:       types.Vec2{X: 1, Y: 0},  // 1 cell horizontal padding
		Spacing:       1,                       // 1 cell between controls
		Indent:        2,                       // 2 cells for tree indentation
		TitleHeight:   1,                       // 1 cell title bar
		ScrollbarSize: 1,                       // 1 cell scrollbar width
		ThumbSize:     1,                       // 1 cell slider thumb
		BorderWidth:   1,                       // 1 cell border - content inset for on-edge borders
	}
}

// DefaultStyle returns the GUI style for backwards compatibility.
// Prefer GUIStyle() or TUIStyle() for explicit intent.
func DefaultStyle() Style {
	return GUIStyle()
}
