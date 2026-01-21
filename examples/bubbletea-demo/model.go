package main

import (
	"fmt"
	"image/color"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	microui "github.com/user/microui-go"
	"github.com/user/microui-go/metaballs"
	"github.com/user/microui-go/render/bubbletea"
	"github.com/user/microui-go/types"
)

// layerWrapper wraps the renderer to force View change detection each frame.
// Bubble Tea v2 compares View.Content by pointer - same pointer = no redraw.
// By creating a new wrapper each frame, we ensure the pointer is different.
type layerWrapper struct {
	r *bubbletea.Renderer
}

func (l *layerWrapper) Draw(s tea.Screen, r tea.Rectangle) {
	l.r.Draw(s, r)
}

// Debug logging to file
var debugFile *os.File

func init() {
	var err error
	debugFile, err = os.Create("debug.log")
	if err != nil {
		panic(err)
	}
}

func debugLog(format string, args ...any) {
	fmt.Fprintf(debugFile, format+"\n", args...)
	debugFile.Sync()
}

// Model implements tea.Model for the microui TUI demo.
type Model struct {
	ui       *microui.UI
	renderer *bubbletea.Renderer
	font     *bubbletea.MonospaceFont

	// Demo state
	checks    [3]bool
	sliderVal float64
	textBuf   []byte

	// Additional demo state (matching ebiten demo)
	logBuf       string    // Event log buffer
	numberVal    float64   // Number input value
	sliderStep   float64   // Slider with step value
	readOnlyBuf  []byte    // Read-only textbox buffer
	treeChecks   [2]bool   // Separate checkboxes for tree window

	// Metaballs viewport
	metaField    *metaballs.Field
	metaRenderer *metaballs.TUIRenderer
	lastTime     time.Time

	// Metaballs controls
	metaSpeed      float64
	metaThreshold  float64
	metaHue        float64
	metaSaturation float64
	metaViewport   types.Rect // Viewport rect from layout

	// Window open state (for close button support)
	demoWindowOpen      bool
	inputWindowOpen     bool
	scrollWindowOpen    bool
	paletteWindowOpen   bool
	boxTestWindowOpen   bool
	treeWindowOpen      bool
	popupWindowOpen      bool
	columnWindowOpen     bool
	enhancedWindowOpen   bool
	logWindowOpen        bool
	metaballsWindowOpen  bool
	showWindowsMenu     bool // Toggle for windows restore menu
	wantsQuit           bool // Signal to quit application

	// Window dimensions
	width  int
	height int

	// Frame state - tracks if BeginFrame was called in Update
	frameStarted bool

	// FPS tracking
	frameCount    int
	lastFPSUpdate time.Time
	currentFPS    float64

	// Content hash for change detection
	lastContentHash uint64
	layer           *layerWrapper

	// Pending mouse position - coalesces rapid motion events
	pendingMouseX int
	pendingMouseY int
	hasMouseMove  bool

	// Skip rendering for motion-only events
	skipRender bool
	cachedView tea.View
}

// windowPositions defines the ideal grid layout for large terminals
var windowPositions = []struct {
	name string
	grid types.Rect // Position for large terminals (90x48+)
	w, h int        // Fixed size
}{
	{"Demo", types.Rect{X: 1, Y: 1, W: 30, H: 14}, 30, 14},
	{"Input", types.Rect{X: 32, Y: 1, W: 30, H: 5}, 30, 5},
	{"Box Test", types.Rect{X: 63, Y: 1, W: 26, H: 6}, 26, 6},
	{"Scroll Test", types.Rect{X: 1, Y: 16, W: 30, H: 7}, 30, 7},
	{"Color Palette", types.Rect{X: 32, Y: 7, W: 30, H: 16}, 30, 16},
	{"Column Layout", types.Rect{X: 63, Y: 8, W: 26, H: 8}, 26, 8},
	{"Tree & Text", types.Rect{X: 1, Y: 24, W: 30, H: 10}, 30, 10},
	{"Popup Demo", types.Rect{X: 32, Y: 24, W: 30, H: 5}, 30, 5},
	{"Enhanced Controls", types.Rect{X: 63, Y: 17, W: 26, H: 8}, 26, 8},
	{"Event Log", types.Rect{X: 63, Y: 26, W: 26, H: 8}, 26, 8},
	{"Metaballs", types.Rect{X: 1, Y: 35, W: 40, H: 12}, 40, 12},
}

// getWindowRect returns the position for a window based on terminal size.
// Large terminals (90x48+) use grid layout, smaller ones use cascade.
func (m *Model) getWindowRect(name string) types.Rect {
	// Find window in positions list
	idx := -1
	for i, wp := range windowPositions {
		if wp.name == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return types.Rect{X: 1, Y: 1, W: 30, H: 10} // fallback
	}

	wp := windowPositions[idx]

	// Check if terminal is large enough for grid layout
	// Grid needs ~90 cols and ~48 rows to fit nicely
	if m.width >= 90 && m.height >= 48 {
		return wp.grid
	}

	// Cascade layout: stack windows diagonally so all title bars visible
	// Each window offset by (+2, +1) from previous - simple linear cascade
	cascadeX := 1 + idx*2
	cascadeY := 1 + idx*1

	// Clamp to screen bounds but don't wrap (user can drag off-screen windows)
	maxX := m.width - wp.w
	maxY := m.height - wp.h - 1 // Leave room for status bar
	if cascadeX > maxX && maxX > 0 {
		cascadeX = maxX
	}
	if cascadeY > maxY && maxY > 0 {
		cascadeY = maxY
	}
	if cascadeX < 1 {
		cascadeX = 1
	}
	if cascadeY < 1 {
		cascadeY = 1
	}

	return types.Rect{X: cascadeX, Y: cascadeY, W: wp.w, H: wp.h}
}

// cascadeWindows resets all window positions to cascade layout
func (m *Model) cascadeWindows() {
	for i, wp := range windowPositions {
		cnt := m.ui.GetContainer(wp.name)
		if cnt != nil {
			// Calculate cascade position
			cascadeX := 1 + i*2
			cascadeY := 1 + i*1

			// Clamp to screen bounds
			maxX := m.width - wp.w
			maxY := m.height - wp.h - 1
			if cascadeX > maxX && maxX > 0 {
				cascadeX = maxX
			}
			if cascadeY > maxY && maxY > 0 {
				cascadeY = maxY
			}
			if cascadeX < 1 {
				cascadeX = 1
			}
			if cascadeY < 1 {
				cascadeY = 1
			}

			cnt.SetRect(types.Rect{X: cascadeX, Y: cascadeY, W: wp.w, H: wp.h})
		}
	}
}

// tileWindows arranges windows in a grid layout to fit the screen
func (m *Model) tileWindows() {
	if m.width == 0 || m.height == 0 {
		return
	}

	// Available space (leave 1 row for status bar)
	availW := m.width
	availH := m.height - 1

	// Try to fit windows using their defined sizes
	// Use a simple row-based packing algorithm
	x := 1
	y := 1
	rowHeight := 0

	for _, wp := range windowPositions {
		cnt := m.ui.GetContainer(wp.name)
		if cnt == nil {
			continue
		}

		// Check if window fits in current row
		if x+wp.w > availW && x > 1 {
			// Move to next row
			x = 1
			y += rowHeight + 1
			rowHeight = 0
		}

		// Check if we've run out of vertical space - wrap to top with offset
		if y+wp.h > availH {
			y = 1
			x += 2 // Offset so windows don't completely overlap
		}

		cnt.SetRect(types.Rect{X: x, Y: y, W: wp.w, H: wp.h})

		// Track tallest window in this row
		if wp.h > rowHeight {
			rowHeight = wp.h
		}

		// Move to next position in row
		x += wp.w + 1
	}
}

// writeLog adds a message to the event log
func (m *Model) writeLog(text string) {
	if len(m.logBuf) > 0 {
		m.logBuf += "\n"
	}
	m.logBuf += text
	// Limit log size to prevent memory growth
	if len(m.logBuf) > 1000 {
		m.logBuf = m.logBuf[len(m.logBuf)-800:]
	}
}

// tuiDrawFrame is a custom DrawFrame for TUI that draws backgrounds and window borders.
// Content area is already inset by BorderWidth in core, so border is drawn ON the rect edge.
func tuiDrawFrame(ui *microui.UI, info microui.FrameInfo) {
	// Draw the filled background
	c := ui.GetColor(info.Kind, info.State)
	ui.DrawRect(info.Rect, c)

	// Only draw border for window backgrounds
	if info.Kind == microui.FrameWindow {
		borderColor := ui.GetColorByID(microui.ColorBorder)
		if borderColor != nil {
			_, _, _, a := borderColor.RGBA()
			if a > 0 {
				// Draw border ON the rect edge (content is inset by BorderWidth)
				ui.DrawBox(info.Rect, borderColor)
			}
		}
	}
}

// newMetaRenderer creates a metaballs renderer with the appropriate color mode
func newMetaRenderer(field *metaballs.Field, colorMode bubbletea.ColorMode) *metaballs.TUIRenderer {
	r := metaballs.NewTUIRenderer(field, 80, 24)
	switch colorMode {
	case bubbletea.Color16:
		r.SetColorMode(metaballs.ColorMode16)
	case bubbletea.Color256:
		r.SetColorMode(metaballs.ColorMode256)
	default:
		r.SetColorMode(metaballs.ColorModeTrueColor)
	}
	return r
}

// NewModel creates a new demo model.
func NewModel(colorMode bubbletea.ColorMode) *Model {
	font := &bubbletea.MonospaceFont{}
	theme := bubbletea.BorlandTheme() // Classic 90s Borland Turbo Vision look
	renderer := bubbletea.NewRenderer(80, 24)
	renderer.SetColorMode(colorMode)

	// Use TUIStyle() for cell-based TUI defaults, override colors and font
	style := microui.TUIStyle()
	style.Colors = theme
	style.Font = font

	ui := microui.New(microui.Config{
		Style:     style,
		DrawFrame: tuiDrawFrame, // Custom DrawFrame for TUI window borders
	})

	// Enable microui debug logging to diagnose close button
	ui.SetDebug(debugLog)

	// Disable renderer debug logging
	// bubbletea.DebugLog = debugLog

	// Explicitly open windows (needed when using OptClosed)
	ui.OpenWindow("Demo")
	ui.OpenWindow("Input")
	ui.OpenWindow("Scroll Test")
	ui.OpenWindow("Color Palette")
	ui.OpenWindow("Box Test")
	ui.OpenWindow("Tree & Text")
	ui.OpenWindow("Popup Demo")
	ui.OpenWindow("Column Layout")
	ui.OpenWindow("Enhanced Controls")
	ui.OpenWindow("Event Log")
	ui.OpenWindow("Metaballs")

	// Initialize metaballs field and TUI renderer
	metaCfg := metaballs.DefaultConfig()
	metaCfg.BallCount = 4 // Fewer balls for TUI clarity
	metaField := metaballs.New(metaCfg)

	return &Model{
		ui:                  ui,
		renderer:            renderer,
		font:                font,
		sliderVal:           0.5,
		textBuf:             make([]byte, 0, 128), // length 0, capacity 128
		checks:              [3]bool{true, false, true},
		numberVal:           42.0,
		sliderStep:          50.0,
		readOnlyBuf:         []byte("Read-only text"),
		demoWindowOpen:      true,
		inputWindowOpen:     true,
		scrollWindowOpen:    true,
		paletteWindowOpen:   true,
		boxTestWindowOpen:  true,
		treeWindowOpen:     true,
		popupWindowOpen:     true,
		columnWindowOpen:    true,
		enhancedWindowOpen:  true,
		logWindowOpen:       true,
		metaballsWindowOpen: true,
		width:               0, // Set by WindowSizeMsg before first render
		height:              0,
		lastFPSUpdate:       time.Now(),
		metaField:           metaField,
		metaRenderer:        newMetaRenderer(metaField, colorMode),
		lastTime:            time.Now(),
		// Metaballs controls defaults
		metaSpeed:      1.0,
		metaThreshold:  1.0,
		metaHue:        0.0,
		metaSaturation: 0.8,
	}
}

// frameTickMsg triggers a frame render
type frameTickMsg time.Time

// frameTick returns a command that ticks at ~60 FPS
func frameTick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return frameTickMsg(t)
	})
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return frameTick() // Start the frame ticker
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for quit request (from Quit button)
	if m.wantsQuit {
		return m, tea.Quit
	}

	// Store text input to add after BeginFrame (which clears it)
	var textToInput string

	// Process input events FIRST to update mouse position
	// This ensures MouseDelta is calculated correctly in BeginFrame
	switch msg := msg.(type) {
	case frameTickMsg:
		// Frame tick - apply any pending mouse position
		if m.hasMouseMove {
			m.ui.MouseMove(m.pendingMouseX, m.pendingMouseY)
			m.hasMouseMove = false
		}
		// Update metaballs animation
		now := time.Now()
		dt := now.Sub(m.lastTime).Seconds()
		m.lastTime = now
		m.metaField.Update(dt)
		// Continue to BeginFrame/View, and schedule next tick
		m.ui.BeginFrame()
		m.frameStarted = true
		return m, frameTick()

	case tea.WindowSizeMsg:
		debugLog("WindowSize: %dx%d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		m.renderer.Resize(msg.Width, msg.Height)

	case tea.KeyPressMsg:
		debugLog("KeyPress: %s", msg.String())
		// Handle quit (only Ctrl+C, not 'q' which conflicts with text input)
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Handle Escape to toggle windows menu
		if msg.String() == "esc" {
			m.showWindowsMenu = !m.showWindowsMenu
			return m, nil
		}
		// Store text for later (after BeginFrame)
		key := msg.Key()
		if key.Text != "" {
			textToInput = key.Text
		}
		// Handle special keys (these set flags that persist)
		handleKeyPress(m.ui, msg)

	case tea.MouseClickMsg:
		debugLog("MouseClick: x=%d y=%d button=%v (calling MouseDown)", msg.X, msg.Y, msg.Button)
		// Update position first for correct delta
		m.ui.MouseMove(msg.X, msg.Y)
		m.ui.MouseDown(msg.X, msg.Y, microui.MouseLeft)
		debugLog("  After MouseDown: input set")

	case tea.MouseReleaseMsg:
		debugLog("MouseRelease: x=%d y=%d button=%v", msg.X, msg.Y, msg.Button)
		m.ui.MouseMove(msg.X, msg.Y)
		m.ui.MouseUp(msg.X, msg.Y, microui.MouseLeft)

	case tea.MouseMotionMsg:
		// Just store the position - don't process yet
		// This coalesces all motion events into the latest position
		m.pendingMouseX = msg.X
		m.pendingMouseY = msg.Y
		m.hasMouseMove = true
		m.skipRender = true // Tell View() to return cached
		return m, nil

	case tea.MouseWheelMsg:
		debugLog("MouseWheel: x=%d y=%d button=%v", msg.X, msg.Y, msg.Button)
		// Check button for direction
		switch msg.Button {
		case tea.MouseWheelUp:
			m.ui.Scroll(0, -3)
		case tea.MouseWheelDown:
			m.ui.Scroll(0, 3)
		case tea.MouseWheelLeft:
			m.ui.Scroll(-3, 0)
		case tea.MouseWheelRight:
			m.ui.Scroll(3, 0)
		}

	default:
		// Log unknown message types to see what we're getting
		debugLog("Unknown msg type: %T", msg)
	}

	// NOW call BeginFrame - MousePos is updated, so delta will be correct
	m.ui.BeginFrame()
	m.frameStarted = true

	// Add text input AFTER BeginFrame (which clears TextInput)
	if textToInput != "" {
		debugLog("  -> TextInput: %q", textToInput)
		m.ui.TextInput(textToInput)
	}

	return m, nil
}

// View implements tea.Model for v2 - returns tea.View.
func (m *Model) View() tea.View {
	// Skip rendering for motion-only events - return cached view
	if m.skipRender && m.cachedView.Content != nil {
		m.skipRender = false
		return m.cachedView
	}
	m.skipRender = false

	// Update FPS counter
	m.frameCount++
	now := time.Now()
	elapsed := now.Sub(m.lastFPSUpdate)
	if elapsed >= time.Second {
		m.currentFPS = float64(m.frameCount) / elapsed.Seconds()
		m.frameCount = 0
		m.lastFPSUpdate = now
	}

	// Draw Borland-style desktop background (dithered blue pattern)
	// Light cyan pattern on dark blue creates the classic dithered look
	m.renderer.FillBackground(
		bubbletea.DesktopPattern,  // ░ light shade character
		bubbletea.DesktopCyan,     // Cyan foreground for the pattern dots
		bubbletea.DesktopBlue,     // Dark blue background
	)

	// If frame wasn't started in Update (initial View call), start it now
	if !m.frameStarted {
		m.ui.BeginFrame()
	}
	m.frameStarted = false // Reset for next Update/View cycle

	// Build demo UI
	m.buildDemoUI()

	// End frame to finalize container command ranges
	m.ui.EndFrame()

	// Render with per-container shadows (correct z-order)
	// For each container in z-order: draw shadow, then render container
	m.renderWithShadows()

	// Draw status bar at bottom (overwrites anything beneath)
	m.drawStatusBar()

	// Check if content actually changed using hash
	contentHash := m.renderer.ContentHash()
	if contentHash != m.lastContentHash || m.layer == nil {
		// Content changed - create new wrapper to force redraw
		m.layer = &layerWrapper{m.renderer}
		m.lastContentHash = contentHash
	}

	// Swap buffers - makes the completed frame visible to Draw()
	m.renderer.Swap()

	// Use Layer interface - ultraviolet handles dirty tracking
	v := tea.NewView(m.layer)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion

	// Cache for motion-only events
	m.cachedView = v
	return v
}

// buildDemoUI creates the demo windows and controls.
func (m *Model) buildDemoUI() {
	// Wait for WindowSizeMsg before creating windows (need dimensions for layout)
	if m.width == 0 || m.height == 0 {
		return
	}

	// Demo Window - use OptClosed to prevent auto-reopen after close button click
	if m.demoWindowOpen {
		if m.ui.BeginWindowOpt("Demo", m.getWindowRect("Demo"), microui.OptClosed) {
			// Header: Test Buttons (expanded by default)
			m.ui.LayoutRow(1, []int{-1}, 0)
			if m.ui.HeaderEx("Test Buttons", microui.OptExpanded) {
				m.ui.LayoutRow(2, []int{14, 14}, 1)
				if m.ui.Button("Button 1") {
					m.writeLog("Button 1")
					debugLog("!!! Button 1 CLICKED !!!")
				}
				if m.ui.Button("Button 2") {
					m.writeLog("Button 2")
					debugLog("!!! Button 2 CLICKED !!!")
				}
			}

			// Header: Checkboxes (collapsed by default)
			m.ui.LayoutRow(1, []int{-1}, 0)
			if m.ui.Header("Checkboxes") {
				m.ui.LayoutRow(1, []int{-1}, 1)
				m.ui.Checkbox("Check 1", &m.checks[0])
				m.ui.Checkbox("Check 2", &m.checks[1])
				m.ui.Checkbox("Check 3", &m.checks[2])
			}

			// Header: Slider (expanded by default)
			m.ui.LayoutRow(1, []int{-1}, 0)
			if m.ui.HeaderEx("Slider", microui.OptExpanded) {
				m.ui.LayoutRow(2, []int{8, -1}, 1)
				m.ui.Label("Value:")
				oldVal := m.sliderVal
				if m.ui.Slider(&m.sliderVal, 0, 1) {
					m.writeLog(fmt.Sprintf("Slider: %.2f", m.sliderVal))
					debugLog("Slider changed: %.2f -> %.2f", oldVal, m.sliderVal)
				}
			}

			// Metaballs window toggle
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Checkbox("Metaballs", &m.metaballsWindowOpen)

			m.ui.EndWindow()
		} else {
			// Close button was clicked
			m.demoWindowOpen = false
			debugLog("Demo window closed by user")
		}
	}

	// Input Window - use OptClosed to prevent auto-reopen after close button click
	if m.inputWindowOpen {
		if m.ui.BeginWindowOpt("Input", m.getWindowRect("Input"), microui.OptClosed) {
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Label("Type here:")
			oldLen := len(m.textBuf)
			result := m.ui.Textbox(&m.textBuf, 128)
			if result != 0 || len(m.textBuf) != oldLen {
				debugLog("Textbox result=%d bufLen=%d content=%q", result, len(m.textBuf), string(m.textBuf))
			}
			m.ui.EndWindow()
		} else {
			// Close button was clicked
			m.inputWindowOpen = false
			debugLog("Input window closed by user")
		}
	}

	// Scroll Test Window - demonstrates both scrollbars
	if m.scrollWindowOpen {
		if m.ui.BeginWindowOpt("Scroll Test", m.getWindowRect("Scroll Test"), microui.OptClosed) {
			// Set panel size to fill remaining window space (-1 = fill)
			m.ui.LayoutRow(1, []int{-1}, -1)

			// Create a panel with lots of content to trigger scrollbars
			m.ui.BeginPanel("scrollpanel")

			// Wide content for horizontal scroll (wider than window)
			m.ui.LayoutRow(1, []int{80}, 1)

			// Many rows for vertical scroll
			for i := 0; i < 15; i++ {
				m.ui.Label(fmt.Sprintf("Row %02d: Wide scrollable content here..............end", i+1))
			}

			m.ui.EndPanel()
			m.ui.EndWindow()
		} else {
			m.scrollWindowOpen = false
			debugLog("Scroll Test window closed by user")
		}
	}

	// Color Palette Window - shows how colors render in different modes
	if m.paletteWindowOpen {
		if m.ui.BeginWindowOpt("Color Palette", m.getWindowRect("Color Palette"), microui.OptClosed) {
			m.buildColorPalette()
			m.ui.EndWindow()
		} else {
			m.paletteWindowOpen = false
			debugLog("Color Palette window closed by user")
		}
	}

	// Box Test Window - demonstrates box-drawing characters with layout-relative positions
	if m.boxTestWindowOpen {
		if m.ui.BeginWindowOpt("Box Test", m.getWindowRect("Box Test"), microui.OptClosed) {
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Label("Box drawing:")

			// Use LayoutNext to get rects that move with the window
			boxColor := color.RGBA{255, 255, 0, 255} // Yellow
			m.ui.LayoutRow(2, []int{10, 10}, 3)      // Two columns, 3 cells tall
			rect1 := m.ui.LayoutNext()
			m.ui.DrawBox(rect1, boxColor)
			rect2 := m.ui.LayoutNext()
			m.ui.DrawBox(rect2, boxColor)

			m.ui.EndWindow()
		} else {
			m.boxTestWindowOpen = false
			debugLog("Box Test window closed by user")
		}
	}

	// Tree & Text Window - demonstrates tree nodes and text wrapping
	if m.treeWindowOpen {
		if m.ui.BeginWindowOpt("Tree & Text", m.getWindowRect("Tree & Text"), microui.OptClosed) {
			m.ui.LayoutRow(2, []int{18, -1}, -1)

			// Left column - tree
			m.ui.LayoutBeginColumn()
			if m.ui.BeginTreeNode("Test 1") {
				if m.ui.BeginTreeNode("Test 1a") {
					m.ui.Label("Hello")
					m.ui.Label("world")
					m.ui.EndTreeNode()
				}
				if m.ui.BeginTreeNode("Test 1b") {
					if m.ui.Button("Btn 1") {
						m.writeLog("Tree btn 1")
					}
					if m.ui.Button("Btn 2") {
						m.writeLog("Tree btn 2")
					}
					m.ui.EndTreeNode()
				}
				m.ui.EndTreeNode()
			}
			if m.ui.BeginTreeNode("Test 2") {
				m.ui.Checkbox("Check A", &m.treeChecks[0])
				m.ui.Checkbox("Check B", &m.treeChecks[1])
				m.ui.EndTreeNode()
			}
			m.ui.LayoutEndColumn()

			// Right column - wrapped text
			m.ui.LayoutBeginColumn()
			m.ui.LayoutRow(1, []int{-1}, 0)
			m.ui.Text("This is a longer text that should wrap across multiple lines in the TUI.")
			m.ui.LayoutEndColumn()

			m.ui.EndWindow()
		} else {
			m.treeWindowOpen = false
			debugLog("Tree & Text window closed by user")
		}
	}

	// Popup Demo Window
	if m.popupWindowOpen {
		if m.ui.BeginWindowOpt("Popup Demo", m.getWindowRect("Popup Demo"), microui.OptClosed) {
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Label("Click for popup:")
			m.ui.LayoutRow(1, []int{-1}, 1)
			if m.ui.Button("Open Popup") {
				m.ui.OpenPopup("demo_popup")
			}
			m.ui.EndWindow()
		} else {
			m.popupWindowOpen = false
			debugLog("Popup Demo window closed by user")
		}
	}

	// Column Layout Window - demonstrates multi-column layouts
	if m.columnWindowOpen {
		if m.ui.BeginWindowOpt("Column Layout", m.getWindowRect("Column Layout"), microui.OptClosed) {
			// Two-column layout
			m.ui.LayoutRow(2, []int{12, -1}, 6)

			// Left column
			m.ui.LayoutBeginColumn()
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Label("Left Col")
			if m.ui.Button("L1") {
				m.writeLog("Left 1")
			}
			if m.ui.Button("L2") {
				m.writeLog("Left 2")
			}
			m.ui.LayoutEndColumn()

			// Right column
			m.ui.LayoutBeginColumn()
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Label("Right Col")
			if m.ui.Button("R1") {
				m.writeLog("Right 1")
			}
			m.ui.LayoutEndColumn()

			m.ui.EndWindow()
		} else {
			m.columnWindowOpen = false
			debugLog("Column Layout window closed by user")
		}
	}

	// Enhanced Controls Window - slider with step, number, read-only textbox
	if m.enhancedWindowOpen {
		if m.ui.BeginWindowOpt("Enhanced Controls", m.getWindowRect("Enhanced Controls"), microui.OptClosed) {
			// Slider with step
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Label("Slider (step=10):")
			m.ui.SliderOpt(&m.sliderStep, 0, 100, 10, "%.0f", 0)

			// Number input
			m.ui.LayoutRow(2, []int{10, -1}, 1)
			m.ui.Label("Number:")
			m.ui.Number(&m.numberVal, 1.0)

			// Read-only textbox
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Label("Read-only:")
			m.ui.TextboxOpt(&m.readOnlyBuf, 64, microui.OptNoInteract)

			m.ui.EndWindow()
		} else {
			m.enhancedWindowOpen = false
			debugLog("Enhanced Controls window closed by user")
		}
	}

	// Event Log Window - shows logged events
	if m.logWindowOpen {
		if m.ui.BeginWindowOpt("Event Log", m.getWindowRect("Event Log"), microui.OptClosed) {
			// Panel for scrollable log
			m.ui.LayoutRow(1, []int{-1}, -1)
			m.ui.BeginPanel("LogPanel")
			// Dense log: use Text() which now has tight line spacing
			if len(m.logBuf) > 0 {
				m.ui.LayoutRow(1, []int{-1}, 0)
				m.ui.Text(m.logBuf)
			} else {
				m.ui.LayoutRow(1, []int{-1}, 1)
				m.ui.Label("(interact...)")
			}
			m.ui.EndPanel()
			m.ui.EndWindow()
		} else {
			m.logWindowOpen = false
			debugLog("Event Log window closed by user")
		}
	}

	// Metaballs Viewport Window - shows metaball animation through half-block characters
	// Acts as a "porthole" into the animation - coordinates are screen-relative, not window-relative
	// Content is rendered in renderWithShadows() after container background
	if m.metaballsWindowOpen {
		if m.ui.BeginWindowOpt("Metaballs", m.getWindowRect("Metaballs"), microui.OptClosed) {
			// 1 cell gap below title
			m.ui.Space(1)

			// Row 1: Speed + Mix (50/50)
			m.ui.LayoutRow(4, []int{5, 11, 5, 11}, 1)
			m.ui.Label("Spd:")
			m.ui.SliderOpt(&m.metaSpeed, 0.1, 3.0, 0.1, "%.1f", 0)
			m.ui.Label("Mix:")
			m.ui.SliderOpt(&m.metaThreshold, 0.3, 2.0, 0.1, "%.1f", 0)

			// Row 2: Hue + Sat (50/50)
			m.ui.LayoutRow(4, []int{5, 11, 5, 11}, 1)
			m.ui.Label("Hue:")
			m.ui.SliderOpt(&m.metaHue, 0.0, 1.0, 0.05, "%.2f", 0)
			m.ui.Label("Sat:")
			m.ui.SliderOpt(&m.metaSaturation, 0.0, 1.0, 0.1, "%.1f", 0)

			// Apply settings to field and renderer
			m.metaField.SetSpeed(m.metaSpeed)
			m.metaField.SetThreshold(m.metaThreshold)
			m.metaRenderer.SetHSVParams(m.metaHue, m.metaSaturation, 0.5)

			// Get viewport rect - fill remaining space
			m.ui.LayoutRow(1, []int{-1}, -1)
			m.metaViewport = m.ui.LayoutNext()

			m.ui.EndWindow()
		} else {
			m.metaballsWindowOpen = false
			debugLog("Metaballs window closed by user")
		}
	}

	// Popup content (render last so it draws on top)
	if m.ui.BeginPopup("demo_popup") {
		m.ui.Label("Popup Menu!")
		if m.ui.Button("Action 1") {
			m.writeLog("Popup action 1")
		}
		if m.ui.Button("Action 2") {
			m.writeLog("Popup action 2")
		}
		m.ui.EndPopup()
	}

	// Windows menu (Esc to toggle)
	if m.showWindowsMenu {
		// Center the menu
		menuW, menuH := 22, 22
		menuX := (m.width - menuW) / 2
		menuY := (m.height - menuH) / 2
		if menuX < 0 {
			menuX = 0
		}
		if menuY < 0 {
			menuY = 0
		}

		if m.ui.BeginWindowOpt("Windows", types.Rect{X: menuX, Y: menuY, W: menuW, H: menuH}, 0) {
			m.ui.LayoutRow(1, []int{-1}, 1)
			m.ui.Checkbox("Demo", &m.demoWindowOpen)
			m.ui.Checkbox("Input", &m.inputWindowOpen)
			m.ui.Checkbox("Scroll", &m.scrollWindowOpen)
			m.ui.Checkbox("Palette", &m.paletteWindowOpen)
			m.ui.Checkbox("Box Test", &m.boxTestWindowOpen)
			m.ui.Checkbox("Tree", &m.treeWindowOpen)
			m.ui.Checkbox("Popup", &m.popupWindowOpen)
			m.ui.Checkbox("Columns", &m.columnWindowOpen)
			m.ui.Checkbox("Enhanced", &m.enhancedWindowOpen)
			m.ui.Checkbox("Log", &m.logWindowOpen)
			m.ui.Checkbox("Metaballs", &m.metaballsWindowOpen)

			m.ui.Space(1)
			m.ui.LayoutRow(1, []int{-1}, 1)
			if m.ui.Button("Show All") {
				m.demoWindowOpen = true
				m.inputWindowOpen = true
				m.scrollWindowOpen = true
				m.paletteWindowOpen = true
				m.boxTestWindowOpen = true
				m.treeWindowOpen = true
				m.popupWindowOpen = true
				m.columnWindowOpen = true
				m.enhancedWindowOpen = true
				m.logWindowOpen = true
				m.metaballsWindowOpen = true
			}
			if m.ui.Button("Hide All") {
				m.demoWindowOpen = false
				m.inputWindowOpen = false
				m.scrollWindowOpen = false
				m.paletteWindowOpen = false
				m.boxTestWindowOpen = false
				m.treeWindowOpen = false
				m.popupWindowOpen = false
				m.columnWindowOpen = false
				m.enhancedWindowOpen = false
				m.logWindowOpen = false
				m.metaballsWindowOpen = false
			}
			if m.ui.Button("Cascade") {
				// Show all windows and reset to cascade positions
				m.demoWindowOpen = true
				m.inputWindowOpen = true
				m.scrollWindowOpen = true
				m.paletteWindowOpen = true
				m.boxTestWindowOpen = true
				m.treeWindowOpen = true
				m.popupWindowOpen = true
				m.columnWindowOpen = true
				m.enhancedWindowOpen = true
				m.logWindowOpen = true
				m.metaballsWindowOpen = true
				m.cascadeWindows()
			}
			if m.ui.Button("Tile") {
				// Show all windows and arrange in grid
				m.demoWindowOpen = true
				m.inputWindowOpen = true
				m.scrollWindowOpen = true
				m.paletteWindowOpen = true
				m.boxTestWindowOpen = true
				m.treeWindowOpen = true
				m.popupWindowOpen = true
				m.columnWindowOpen = true
				m.enhancedWindowOpen = true
				m.logWindowOpen = true
				m.metaballsWindowOpen = true
				m.tileWindows()
			}
			m.ui.Space(1)
			if m.ui.Button("Close") {
				m.showWindowsMenu = false
			}
			if m.ui.Button("Quit") {
				m.wantsQuit = true
			}
			m.ui.EndWindow()
		}
	}
}

// buildColorPalette draws color swatches to visualize color mode differences.
func (m *Model) buildColorPalette() {
	// 16 ANSI colors - these should look identical in all modes
	m.ui.LayoutRow(1, []int{-1}, 1)
	m.ui.Label("16 ANSI Colors:")

	// Draw 8 dark colors
	m.ui.LayoutRow(8, []int{3, 3, 3, 3, 3, 3, 3, 3}, 1)
	ansiDark := []color.Color{
		color.RGBA{0, 0, 0, 255},       // Black
		color.RGBA{170, 0, 0, 255},     // Red
		color.RGBA{0, 170, 0, 255},     // Green
		color.RGBA{170, 170, 0, 255},   // Yellow/Brown
		color.RGBA{0, 0, 170, 255},     // Blue
		color.RGBA{170, 0, 170, 255},   // Magenta
		color.RGBA{0, 170, 170, 255},   // Cyan
		color.RGBA{170, 170, 170, 255}, // White/Gray
	}
	for _, c := range ansiDark {
		rect := m.ui.LayoutNext()
		m.ui.DrawRect(rect, c)
	}

	// Draw 8 bright colors
	m.ui.LayoutRow(8, []int{3, 3, 3, 3, 3, 3, 3, 3}, 1)
	ansiBright := []color.Color{
		color.RGBA{85, 85, 85, 255},    // Bright Black (Gray)
		color.RGBA{255, 85, 85, 255},   // Bright Red
		color.RGBA{85, 255, 85, 255},   // Bright Green
		color.RGBA{255, 255, 85, 255},  // Bright Yellow
		color.RGBA{85, 85, 255, 255},   // Bright Blue
		color.RGBA{255, 85, 255, 255},  // Bright Magenta
		color.RGBA{85, 255, 255, 255},  // Bright Cyan
		color.RGBA{255, 255, 255, 255}, // Bright White
	}
	for _, c := range ansiBright {
		rect := m.ui.LayoutNext()
		m.ui.DrawRect(rect, c)
	}

	// RGB gradient - shows true color vs palette mapping
	m.ui.LayoutRow(1, []int{-1}, 1)
	m.ui.Label("Red Gradient (256/true):")
	m.ui.LayoutRow(8, []int{3, 3, 3, 3, 3, 3, 3, 3}, 1)
	for i := 0; i < 8; i++ {
		rect := m.ui.LayoutNext()
		c := color.RGBA{uint8(i * 36), 0, 0, 255}
		m.ui.DrawRect(rect, c)
	}

	// Cyan gradient - matches our theme
	m.ui.LayoutRow(1, []int{-1}, 1)
	m.ui.Label("Cyan Gradient:")
	m.ui.LayoutRow(8, []int{3, 3, 3, 3, 3, 3, 3, 3}, 1)
	for i := 0; i < 8; i++ {
		rect := m.ui.LayoutNext()
		c := color.RGBA{0, uint8(i * 36), uint8(i * 36), 255}
		m.ui.DrawRect(rect, c)
	}

	// Shadow style - classic Turbo Vision (ANSI compatible)
	m.ui.LayoutRow(1, []int{-1}, 1)
	m.ui.Label("Shadow: Black bg, Gray fg")
	m.ui.LayoutRow(2, []int{12, 12}, 1)
	// Original cyan
	rect := m.ui.LayoutNext()
	m.ui.DrawRect(rect, color.RGBA{0, 170, 170, 255})
	// Shadow colors (ANSI 0 and 8)
	rect = m.ui.LayoutNext()
	m.ui.DrawRect(rect, bubbletea.ShadowBg) // Black

	// Desktop pattern colors
	m.ui.LayoutRow(1, []int{-1}, 1)
	m.ui.Label("Desktop: Blue|Cyan|ShadFg")
	m.ui.LayoutRow(3, []int{8, 8, 8}, 1)
	rect = m.ui.LayoutNext()
	m.ui.DrawRect(rect, bubbletea.DesktopBlue)
	rect = m.ui.LayoutNext()
	m.ui.DrawRect(rect, bubbletea.DesktopCyan)
	rect = m.ui.LayoutNext()
	m.ui.DrawRect(rect, bubbletea.ShadowFg) // Dark gray
}

// renderWithShadows renders all containers with proper z-ordered shadows.
// For each container (back to front): draw shadow, then render container.
// This ensures shadows appear behind windows but in front of windows further back.
func (m *Model) renderWithShadows() {
	// Shadow factor: 0.4 = 40% brightness (60% darker)
	const shadowFactor = 0.4

	// Get containers sorted by z-index (back to front)
	containers := m.ui.RootContainersSorted()

	for _, cnt := range containers {
		if !cnt.Open() {
			continue
		}

		rect := cnt.Rect()

		// Draw shadow BEFORE this container's content
		// Right shadow: 2 cells wide, starts 1 cell down from top
		m.renderer.DrawShadow(types.Rect{
			X: rect.X + rect.W,
			Y: rect.Y + 1,
			W: 2,
			H: rect.H,
		}, shadowFactor)
		// Bottom shadow: stops before right shadow to avoid double-darkening corner
		m.renderer.DrawShadow(types.Rect{
			X: rect.X + 2,
			Y: rect.Y + rect.H,
			W: rect.W - 2, // Don't overlap with right shadow
			H: 1,
		}, shadowFactor)

		// Render this container's commands
		m.ui.RenderContainer(cnt, m.renderer)

		// Draw metaballs content into the viewport area
		if cnt.Name() == "Metaballs" && m.metaballsWindowOpen {
			vp := m.metaViewport
			body := cnt.Body()
			if vp.W > 0 && vp.H > 0 {
				m.metaRenderer.SetScreenSize(m.width, m.height)
				cells := m.metaRenderer.RenderWindow(vp.X, vp.Y, vp.W, vp.H)
				for y, row := range cells {
					for x, cell := range row {
						screenX := vp.X + x
						screenY := vp.Y + y
						// Clip to container body (respects scrolling)
						if screenX >= body.X && screenX < body.X+body.W &&
							screenY >= body.Y && screenY < body.Y+body.H {
							m.renderer.SetCellFull(screenX, screenY, cell.Char, cell.Fg, cell.Bg)
						}
					}
				}
			}
		}
	}
}

// drawStatusBar draws a status/hint bar at the bottom of the screen.
func (m *Model) drawStatusBar() {
	if m.height < 1 {
		return
	}
	y := m.height - 1

	// Left: key hints
	left := " Ctrl+C Quit │ Esc Windows │ Drag titles"
	// Right: FPS counter
	right := fmt.Sprintf("FPS:%.0f ", m.currentFPS)

	// Fill entire bottom row with background first
	for x := 0; x < m.width; x++ {
		m.renderer.SetCellFull(x, y, ' ', bubbletea.StatusBarFg, bubbletea.StatusBarBg)
	}

	// Draw left-aligned text
	for i, r := range left {
		if i < m.width {
			m.renderer.SetCellFull(i, y, r, bubbletea.StatusBarFg, bubbletea.StatusBarBg)
		}
	}

	// Draw right-aligned FPS
	rightStart := m.width - len(right)
	if rightStart > len(left) { // Don't overlap with left text
		for i, r := range right {
			m.renderer.SetCellFull(rightStart+i, y, r, bubbletea.StatusBarFg, bubbletea.StatusBarBg)
		}
	}
}

// handleKeyPress bridges Bubble Tea key events to microui.
// Note: Text input is handled separately in Update() after BeginFrame.
func handleKeyPress(ui *microui.UI, msg tea.KeyPressMsg) {
	// Get the underlying Key
	key := msg.Key()

	debugLog("KeyPress: code=%d text=%q", key.Code, key.Text)

	switch key.Code {
	case tea.KeyBackspace:
		debugLog("  -> Backspace")
		ui.KeyDown(microui.KeyBackspace)
		ui.KeyUp(microui.KeyBackspace)
	case tea.KeyLeft:
		debugLog("  -> Left")
		ui.KeyDown(microui.KeyLeft)
		ui.KeyUp(microui.KeyLeft)
	case tea.KeyRight:
		debugLog("  -> Right")
		ui.KeyDown(microui.KeyRight)
		ui.KeyUp(microui.KeyRight)
	case tea.KeyEnter:
		debugLog("  -> Enter")
		ui.KeyDown(microui.KeyEnter)
		ui.KeyUp(microui.KeyEnter)
	}
	// Text input is handled in Update() after BeginFrame
}
