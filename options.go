package microui

import "github.com/user/microui-go/types"

// Option flags for controls
const (
	OptAlignCenter    = 1 << iota // Center text alignment
	OptAlignRight                 // Right text alignment
	OptNoInteract                 // Non-interactive (display only)
	OptNoFrame                    // Don't draw control frame
	OptNoResize                   // Window: disable resize
	OptNoScroll                   // Panel: disable scrollbars
	OptNoClose                    // Window: no close button
	OptNoTitle                    // Window: no title bar
	OptHoldFocus                  // Keep focus after interaction
	OptAutoSize                   // Container: auto-size to content
	OptPopup                      // Popup behavior
	OptClosed                     // Start closed/collapsed
	OptExpanded                   // Start expanded (default for headers)
	OptAlwaysOnTop                // Window: always render on top of other windows
	OptNoControlInset             // Text: don't apply control margin/padding (for labels, titles)
)

// Response flags returned by controls
const (
	ResChange = 1 << iota // Value changed
	ResSubmit             // Enter pressed / submitted
	ResActive             // Control is active (has focus)
)

// Clip result constants
const (
	ClipNone = 0 // Rect fully visible
	ClipPart = 1 // Rect partially visible
	ClipAll  = 2 // Rect fully clipped (invisible)
)

// FrameKind identifies the type of UI component being drawn.
type FrameKind int

const (
	FrameWindow      FrameKind = iota // Window background
	FrameTitle                        // Window title bar
	FramePanel                        // Panel/container background
	FrameButton                       // Push button
	FrameInput                        // Text input field, slider track, checkbox
	FrameSliderThumb                  // Slider thumb
	FrameScrollTrack                  // Scrollbar track
	FrameScrollThumb                  // Scrollbar thumb
	FrameHeader                       // Collapsible header / tree node

	// FrameCustomBase is the starting point for user-defined frame kinds.
	// Custom controls can define their own kinds starting from this value:
	//
	//   const (
	//       FrameMyGauge    = microui.FrameCustomBase + iota
	//       FrameMyTimeline
	//       FrameMyWaveform
	//   )
	//
	// Handle these in your custom DrawFrame callback.
	FrameCustomBase FrameKind = 1000
)

// FrameState represents the interaction state of a component.
type FrameState int

const (
	StateNormal FrameState = iota
	StateHover
	StateFocus // Also means "active" or "pressed"
)

// FrameInfo contains all information needed to render a component frame.
type FrameInfo struct {
	Kind  FrameKind
	State FrameState
	Rect  types.Rect
}

// Color IDs for text and border colors.
// For component background colors, use FrameKind and FrameState with GetColor().
const (
	ColorText = iota
	ColorBorder
	ColorTitleText
)
