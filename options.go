package microui

// Option flags for controls
const (
	OptAlignCenter = 1 << iota // Center text alignment
	OptAlignRight              // Right text alignment
	OptNoInteract              // Non-interactive (display only)
	OptNoFrame                 // Don't draw control frame
	OptNoResize                // Window: disable resize
	OptNoScroll                // Panel: disable scrollbars
	OptNoClose                 // Window: no close button
	OptNoTitle                 // Window: no title bar
	OptHoldFocus               // Keep focus after interaction
	OptAutoSize                // Container: auto-size to content
	OptPopup                   // Popup behavior
	OptClosed                  // Start closed/collapsed
	OptExpanded                // Start expanded (default for headers)
	OptAlwaysOnTop             // Window: always render on top of other windows
	OptSnapToEdge              // Window: snap to screen edges and other windows when dragging
	OptSnapTarget              // Window: can be snapped TO by other windows
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

// Color IDs for DrawFrame callback
const (
	ColorText = iota
	ColorBorder
	ColorWindowBG
	ColorTitleBG
	ColorTitleText
	ColorPanelBG
	ColorButton
	ColorButtonHover
	ColorButtonFocus
	ColorBase
	ColorBaseHover
	ColorBaseFocus
	ColorScrollBase
	ColorScrollThumb
)
