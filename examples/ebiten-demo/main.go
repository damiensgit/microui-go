package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	microui "github.com/user/microui-go"
	uirenderer "github.com/user/microui-go/render/ebiten"
	"github.com/user/microui-go/render/ebiten/atlas"
	"github.com/user/microui-go/types"
)

func main() {
	game := NewGame()

	ebiten.SetWindowSize(900, 700)
	ebiten.SetWindowTitle("MicroUI Go Demo - All Controls")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	ui       *microui.UI
	renderer *uirenderer.Renderer

	// Demo state
	checks     [3]bool
	bgColor    [3]float64
	sliderVal  float64
	clickCount int
	logBuf     string
	lastMouse  bool

	// New control state
	textboxBuf  []byte
	numberVal   float64
	numberVal2  float64 // Separate value for Enhanced Controls window
	sliderStep  float64
	alignDemo   int
	readOnlyBuf []byte
	showNoTitle bool
	showNoClose bool

	// Key repeat state
	heldKeys       map[ebiten.Key]time.Time // When each key was first pressed
	lastRepeatTime map[ebiten.Key]time.Time // When each key last repeated

	// Metaballs background
	metaballs       *Metaballs
	enableMetaballs bool
	metaResolution  float64 // Grid resolution (1-8, lower = higher quality)
	metaBallCount   float64 // Number of balls (2-12)
	metaSpeed       float64 // Speed multiplier (0.1-3.0)
	metaThreshold   float64 // Mix threshold (0.5-2.0, lower = more blobby)
	lastTime        time.Time

	// Window visibility state (ESC menu toggles these)
	showWindowsMenu       bool
	demoWindowOpen        bool
	inputWindowOpen       bool
	collapsibleWindowOpen bool
	popupWindowOpen       bool
	columnWindowOpen      bool
	featuresWindowOpen    bool
	logWindowOpen         bool
	enhancedWindowOpen    bool

	// Screen dimensions (from Layout, works in WASM)
	screenW, screenH int
}

// atlasLayoutFont wraps atlas.Font to implement types.Font for layout calculations
type atlasLayoutFont struct {
	font *atlas.Font
}

func (f *atlasLayoutFont) Width(text string) int { return f.font.Width(text) }
func (f *atlasLayoutFont) Height() int           { return f.font.Height() }

func NewGame() *Game {
	// Create atlas font for proper text rendering
	atlasFont := atlas.NewFont()

	// Use atlas font for layout too (ensures text measurement matches rendering)
	layoutFont := &atlasLayoutFont{font: atlasFont}

	// Use GUIStyle() for pixel-based GUI defaults, override font for atlas rendering
	style := microui.GUIStyle()
	style.Font = layoutFont

	ui := microui.New(microui.Config{
		Style: style,
	})

	// Create renderer with atlas font and icon provider
	renderer := uirenderer.NewRenderer()
	renderer.SetFont(atlasFont)
	renderer.SetIconProvider(atlasFont)

	// Initialize metaballs with default config
	metaConfig := DefaultMetaballsConfig()
	metaConfig.GridResolution = 4

	// Explicitly open windows (needed when using OptClosed to prevent auto-reopen)
	ui.OpenWindow("Demo Window")
	ui.OpenWindow("Input Controls")
	ui.OpenWindow("Collapsible Controls")
	ui.OpenWindow("Popup Demo")
	ui.OpenWindow("Column Layout")
	ui.OpenWindow("Window Features")
	ui.OpenWindow("Event Log")
	ui.OpenWindow("Enhanced Controls")
	ui.OpenWindow("Fixed Size")
	ui.OpenWindow("NoTitle Window")

	return &Game{
		ui:              ui,
		renderer:        renderer,
		checks:          [3]bool{true, false, true},
		bgColor:         [3]float64{50, 50, 60},
		sliderVal:       0.5,
		textboxBuf:      []byte("Edit me!"),
		numberVal:       42.0,
		numberVal2:      100.0,
		sliderStep:      5.0,
		alignDemo:       0,
		readOnlyBuf:     []byte("Read-only text"),
		showNoTitle:     true,
		showNoClose:     true,
		heldKeys:        make(map[ebiten.Key]time.Time),
		lastRepeatTime:  make(map[ebiten.Key]time.Time),
		metaballs:       NewMetaballs(metaConfig),
		enableMetaballs: true,
		metaResolution:  4.0,
		metaBallCount:   6.0,
		metaSpeed:       1.0,
		metaThreshold:   1.0,
		lastTime:        time.Now(),
		// All windows open by default
		demoWindowOpen:        true,
		inputWindowOpen:       true,
		collapsibleWindowOpen: true,
		popupWindowOpen:       true,
		columnWindowOpen:      true,
		featuresWindowOpen:    true,
		logWindowOpen:         true,
		enhancedWindowOpen:    true,
	}
}

func (g *Game) writeLog(text string) {
	if len(g.logBuf) > 0 {
		g.logBuf += "\n"
	}
	g.logBuf += text
}

func (g *Game) Update() error {
	// Update metaballs animation
	now := time.Now()
	dt := now.Sub(g.lastTime).Seconds()
	g.lastTime = now

	if g.metaballs != nil {
		g.metaballs.Update(dt, g.screenW, g.screenH)
	}

	mx, my := ebiten.CursorPosition()

	// Update mouse ball position for metaballs
	if g.metaballs != nil {
		g.metaballs.SetMousePosition(mx, my)
	}

	g.ui.MouseMove(mx, my)

	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if pressed && !g.lastMouse {
		g.ui.MouseDown(mx, my, microui.MouseLeft)
	} else if !pressed && g.lastMouse {
		g.ui.MouseUp(mx, my, microui.MouseLeft)
	}
	g.lastMouse = pressed

	// Handle scroll wheel
	_, scrollY := ebiten.Wheel()
	if scrollY != 0 {
		g.ui.Scroll(0, int(-scrollY*30)) // Negative because scroll down = positive delta
	}

	g.ui.BeginFrame()

	// Handle keyboard input AFTER BeginFrame (which clears old input)
	g.handleKeyboard()

	// === Demo Window (matches C microui demo) ===
	// Column 1, Row 1 - Main demo window
	if g.demoWindowOpen {
		if g.ui.BeginWindowOpt("Demo Window", types.Rect{X: 10, Y: 10, W: 280, H: 420}, microui.OptClosed) {
			// Enforce minimum size like C demo
		win := g.ui.GetCurrentContainer()
		if win != nil {
			r := win.Rect()
			if r.W < 240 {
				r.W = 240
			}
			if r.H < 300 {
				r.H = 300
			}
			win.SetRect(r)
		}

		// Window Info header (collapsed by default)
		g.ui.LayoutRow(1, []int{-1}, 0)
		if g.ui.Header("Window Info") {
			win := g.ui.GetCurrentContainer()
			g.ui.LayoutRow(2, []int{54, -1}, 0)
			g.ui.Label("Position:")
			if win != nil {
				r := win.Rect()
				g.ui.Label(fmt.Sprintf("%d, %d", r.X, r.Y))
				g.ui.Label("Size:")
				g.ui.Label(fmt.Sprintf("%d, %d", r.W, r.H))
			}
		}

		// Test Buttons header (expanded by default)
		g.ui.LayoutRow(1, []int{-1}, 0)
		if g.ui.HeaderEx("Test Buttons", microui.OptExpanded) {
			g.ui.LayoutRow(3, []int{86, -110, -1}, 0)
			g.ui.Label("Test buttons 1:")
			if g.ui.Button("Button 1") {
				g.writeLog("Pressed button 1")
			}
			if g.ui.Button("Button 2") {
				g.writeLog("Pressed button 2")
			}
			g.ui.Label("Test buttons 2:")
			if g.ui.Button("Button 3") {
				g.writeLog("Pressed button 3")
			}
			if g.ui.Button("Popup") {
				g.ui.OpenPopup("Test Popup")
			}
			if g.ui.BeginPopup("Test Popup") {
				g.ui.Button("Hello")
				g.ui.Button("World")
				g.ui.EndPopup()
			}
		}

		// Tree and Text header (expanded by default)
		g.ui.LayoutRow(1, []int{-1}, 0)
		if g.ui.HeaderEx("Tree and Text", microui.OptExpanded) {
			g.ui.LayoutRow(2, []int{140, -1}, 0)

			// Left column - tree
			g.ui.LayoutBeginColumn()
			if g.ui.BeginTreeNode("Test 1") {
				if g.ui.BeginTreeNode("Test 1a") {
					g.ui.Label("Hello")
					g.ui.Label("world")
					g.ui.EndTreeNode()
				}
				if g.ui.BeginTreeNode("Test 1b") {
					if g.ui.Button("Button 1") {
						g.writeLog("Pressed button 1")
					}
					if g.ui.Button("Button 2") {
						g.writeLog("Pressed button 2")
					}
					g.ui.EndTreeNode()
				}
				g.ui.EndTreeNode()
			}
			if g.ui.BeginTreeNode("Test 2") {
				g.ui.LayoutRow(2, []int{54, 54}, 0)
				if g.ui.Button("Button 3") {
					g.writeLog("Pressed button 3")
				}
				if g.ui.Button("Button 4") {
					g.writeLog("Pressed button 4")
				}
				if g.ui.Button("Button 5") {
					g.writeLog("Pressed button 5")
				}
				if g.ui.Button("Button 6") {
					g.writeLog("Pressed button 6")
				}
				g.ui.EndTreeNode()
			}
			if g.ui.BeginTreeNode("Test 3") {
				g.ui.Checkbox("Checkbox 1", &g.checks[0])
				g.ui.Checkbox("Checkbox 2", &g.checks[1])
				g.ui.Checkbox("Checkbox 3", &g.checks[2])
				g.ui.EndTreeNode()
			}
			g.ui.LayoutEndColumn()

			// Right column - text
			g.ui.LayoutBeginColumn()
			g.ui.LayoutRow(1, []int{-1}, 0)
			g.ui.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas lacinia, sem eu lacinia molestie, mi risus faucibus ipsum, eu varius magna felis a nulla.")
			g.ui.LayoutEndColumn()
		}

		// Background Color header (expanded by default)
		g.ui.LayoutRow(1, []int{-1}, 0)
		if g.ui.HeaderEx("Background Color", microui.OptExpanded) {
			g.ui.LayoutRow(2, []int{-78, -1}, 74)

			// Left column - sliders
			g.ui.LayoutBeginColumn()
			g.ui.LayoutRow(2, []int{46, -1}, 0)
			g.ui.Label("Red:")
			g.ui.Slider(&g.bgColor[0], 0, 255)
			g.ui.Label("Green:")
			g.ui.Slider(&g.bgColor[1], 0, 255)
			g.ui.Label("Blue:")
			g.ui.Slider(&g.bgColor[2], 0, 255)
			g.ui.LayoutEndColumn()

			// Right column - color preview
			rect := g.ui.LayoutNext()
			g.ui.DrawRect(rect, color.RGBA{
				R: uint8(g.bgColor[0]),
				G: uint8(g.bgColor[1]),
				B: uint8(g.bgColor[2]),
				A: 255,
			})
			hexStr := fmt.Sprintf("#%02X%02X%02X", int(g.bgColor[0]), int(g.bgColor[1]), int(g.bgColor[2]))
			g.ui.DrawControlText(hexStr, rect, microui.ColorText, microui.OptAlignCenter)

			// Metaballs controls
			g.ui.LayoutRow(2, []int{120, -1}, 0)
			g.ui.Checkbox("Metaballs", &g.enableMetaballs)
			g.ui.Label("")

			g.ui.LayoutRow(2, []int{70, -1}, 0)
			g.ui.Label("Resolution:")
			oldRes := g.metaResolution
			g.ui.SliderOpt(&g.metaResolution, 1, 8, 1, "%.0f", 0)

			g.ui.Label("Balls:")
			oldBalls := g.metaBallCount
			g.ui.SliderOpt(&g.metaBallCount, 2, 12, 1, "%.0f", 0)

			g.ui.Label("Speed:")
			g.ui.SliderOpt(&g.metaSpeed, 0.1, 3.0, 0.1, "%.1f", 0)

			g.ui.Label("Mix:")
			g.ui.SliderOpt(&g.metaThreshold, 0.3, 2.0, 0.1, "%.1f", 0)

			// Update speed and threshold in real-time (no need to recreate)
			if g.metaballs != nil {
				g.metaballs.config.Speed = g.metaSpeed
				g.metaballs.config.Threshold = g.metaThreshold
			}

			// Update metaballs config if resolution or ball count changed
			if (int(g.metaResolution) != int(oldRes) || int(g.metaBallCount) != int(oldBalls)) && g.metaballs != nil {
				newConfig := g.metaballs.config
				newConfig.GridResolution = int(g.metaResolution)
				newConfig.BallCount = int(g.metaBallCount)
				g.metaballs = NewMetaballs(newConfig)
			}
		}

			g.ui.EndWindow()
		} else {
			g.demoWindowOpen = false
		}
	}

	// === Input Controls Window (textbox, number) ===
	// Column 1, Row 2
	if g.inputWindowOpen {
		if g.ui.BeginWindowOpt("Input Controls", types.Rect{X: 10, Y: 440, W: 280, H: 130}, microui.OptClosed) {
			// Textbox
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Textbox (type to edit):")
		g.ui.LayoutRow(1, []int{-1}, 0) // Use default height (same as other controls)
		result := g.ui.Textbox(&g.textboxBuf, 128)
		if result&microui.ResSubmit != 0 {
			g.writeLog("Textbox submitted: " + string(g.textboxBuf))
		}

		// Number input
		g.ui.LayoutRow(2, []int{100, -1}, 0)
		g.ui.Label("Number (drag):")
		g.ui.Number(&g.numberVal, 0.5)

		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label(fmt.Sprintf("Value: %.2f", g.numberVal))

			g.ui.EndWindow()
		} else {
			g.inputWindowOpen = false
		}
	}

	// === Collapsible Controls Window (header, treenode) ===
	// Column 2, Row 1
	if g.collapsibleWindowOpen {
		if g.ui.BeginWindowOpt("Collapsible Controls", types.Rect{X: 300, Y: 10, W: 280, H: 260}, microui.OptClosed) {
		g.ui.LayoutRow(1, []int{-1}, 0)

		// Header sections
		if g.ui.Header("Header Section 1") {
			g.ui.Label("Content inside header 1")
			g.ui.Label("More content here")
		}

		if g.ui.Header("Header Section 2") {
			g.ui.Label("This is header 2 content")
			if g.ui.Button("Nested Button") {
				g.writeLog("Nested button clicked!")
			}
		}

		// TreeNode (hierarchical)
		if g.ui.BeginTreeNode("Tree Root") {
			g.ui.Label("Child item 1")
			g.ui.Label("Child item 2")

			if g.ui.BeginTreeNode("Nested Node") {
				g.ui.Label("Nested child A")
				g.ui.Label("Nested child B")
				g.ui.EndTreeNode()
			}

			g.ui.EndTreeNode()
		}

			g.ui.EndWindow()
		} else {
			g.collapsibleWindowOpen = false
		}
	}

	// === Popup Demo Window ===
	// Column 2, Row 2
	if g.popupWindowOpen {
		if g.ui.BeginWindowOpt("Popup Demo", types.Rect{X: 300, Y: 280, W: 280, H: 100}, microui.OptClosed) {
			g.ui.LayoutRow(1, []int{-1}, 0)
			g.ui.Label("Click button to open popup:")

			g.ui.LayoutRow(1, []int{120}, 0)
			if g.ui.Button("Open Popup") {
				g.ui.OpenPopup("demo_popup")
			}

			g.ui.EndWindow()
		} else {
			g.popupWindowOpen = false
		}
	}

	// === Column Layout Demo ===
	// Column 3, Row 1
	if g.columnWindowOpen {
		if g.ui.BeginWindowOpt("Column Layout", types.Rect{X: 590, Y: 10, W: 280, H: 180}, microui.OptClosed) {
		// Two-column layout
		g.ui.LayoutRow(2, []int{130, -1}, 100)

		// Left column
		g.ui.LayoutBeginColumn()
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Left Column")
		if g.ui.Button("L1") {
			g.writeLog("Left 1")
		}
		if g.ui.Button("L2") {
			g.writeLog("Left 2")
		}
		g.ui.LayoutEndColumn()

		// Right column
		g.ui.LayoutBeginColumn()
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Right Column")
		if g.ui.Button("R1") {
			g.writeLog("Right 1")
		}
		g.ui.LayoutEndColumn()

			g.ui.EndWindow()
		} else {
			g.columnWindowOpen = false
		}
	}

	// === Window Features Demo (drag & resize) ===
	// Column 3, Row 2
	if g.featuresWindowOpen {
		if g.ui.BeginWindowOpt("Window Features", types.Rect{X: 590, Y: 200, W: 280, H: 180}, microui.OptClosed) {
			g.ui.LayoutRow(1, []int{-1}, 0)

			g.ui.Text("Drag this window by the title bar. Resize it by dragging the bottom-right corner.")

			g.ui.Label("")
			g.ui.Label("Try these features:")
			g.ui.Label("- Drag title bar to move")
			g.ui.Label("- Drag corner to resize")
			g.ui.Label("- Click to bring to front")
			g.ui.Label("- Scroll wheel to scroll")

			g.ui.EndWindow()
		} else {
			g.featuresWindowOpen = false
		}
	}

	// === Fixed Size Window (no resize) ===
	// Column 3, Row 3
	if g.showNoClose {
		opt := microui.OptNoResize | microui.OptClosed
		if g.ui.BeginWindowOpt("Fixed Size", types.Rect{X: 590, Y: 390, W: 140, H: 80}, opt) {
			g.ui.LayoutRow(1, []int{-1}, 0)
			g.ui.Label("Can't resize me!")
			if g.ui.Button("Hide") {
				g.showNoClose = false
			}
			g.ui.EndWindow()
		} else {
			g.showNoClose = false
		}
	}

	// === Log Window ===
	// Column 3, Row 4
	if g.logWindowOpen {
		if g.ui.BeginWindowOpt("Event Log", types.Rect{X: 590, Y: 480, W: 280, H: 170}, microui.OptClosed) {
		// Use a panel for scrollable log content - explicit height to fill window body
		g.ui.LayoutRow(1, []int{-1}, 120) // Panel fills most of window body (150 - title - padding)
		g.ui.BeginPanel("LogPanel")
		g.ui.LayoutRow(1, []int{-1}, 0)
		if len(g.logBuf) > 0 {
			// Show log lines - Text handles word wrap
			g.ui.Text(g.logBuf)
		} else {
			g.ui.Label("(interact with controls)")
		}
			g.ui.EndPanel()

			g.ui.EndWindow()
		} else {
			g.logWindowOpen = false
		}
	}

	// === Enhanced Options Demo ===
	// Column 2, Row 3
	if g.enhancedWindowOpen {
		if g.ui.BeginWindowOpt("Enhanced Controls", types.Rect{X: 300, Y: 390, W: 280, H: 180}, microui.OptClosed) {
			// Slider with step - use SliderOpt
			g.ui.LayoutRow(1, []int{-1}, 0)
			g.ui.Label("Slider with step (5):")
			g.ui.SliderOpt(&g.sliderStep, 0, 100, 5, "%.0f", 0)

			// Number with format - use NumberOpt (uses separate variable to avoid ID conflict)
			g.ui.LayoutRow(2, []int{100, -1}, 0)
			g.ui.Label("Number (int):")
			g.ui.NumberOpt(&g.numberVal2, 1.0, "%.0f", microui.OptAlignRight)

			// Read-only textbox
			g.ui.LayoutRow(1, []int{-1}, 0)
			g.ui.Label("Read-only textbox:")
			g.ui.TextboxOpt(&g.readOnlyBuf, 64, microui.OptNoInteract)

			g.ui.EndWindow()
		} else {
			g.enhancedWindowOpen = false
		}
	}

	// === Window Options Demo (No Title) ===
	// Column 1, Row 3
	if g.showNoTitle {
		opt := microui.OptNoTitle | microui.OptClosed
		if g.ui.BeginWindowOpt("NoTitle Window", types.Rect{X: 10, Y: 580, W: 280, H: 70}, opt) {
			g.ui.LayoutRow(1, []int{-1}, 0)
			g.ui.Label("This window has no title bar")
			if g.ui.Button("Hide This Window") {
				g.showNoTitle = false
			}
			g.ui.EndWindow()
		} else {
			g.showNoTitle = false
		}
	}

	// === Render popups LAST so they draw on top of all windows ===
	// C microui: popup content uses default width=0 (not -1 fill) for proper autosize
	if g.ui.BeginPopup("demo_popup") {
		g.ui.Label("Popup Content!")
		if g.ui.Button("Action 1") {
			g.writeLog("Popup action 1")
		}
		if g.ui.Button("Action 2") {
			g.writeLog("Popup action 2")
		}
		g.ui.EndPopup()
	}

	// === Windows Menu (ESC to toggle) ===
	if g.showWindowsMenu {
		// Center the menu
		menuW, menuH := 200, 380
		menuX := (g.screenW - menuW) / 2
		menuY := (g.screenH - menuH) / 2

		if g.ui.BeginWindowOpt("Windows", types.Rect{X: menuX, Y: menuY, W: menuW, H: menuH}, 0) {
			g.ui.LayoutRow(1, []int{-1}, 0)

			// Helper to handle checkbox + OpenWindow
			windowCheckbox := func(label, name string, flag *bool) {
				wasOpen := *flag
				g.ui.Checkbox(label, flag)
				if *flag && !wasOpen {
					g.ui.OpenWindow(name)
				}
			}

			windowCheckbox("Demo Window", "Demo Window", &g.demoWindowOpen)
			windowCheckbox("Input Controls", "Input Controls", &g.inputWindowOpen)
			windowCheckbox("Collapsible", "Collapsible Controls", &g.collapsibleWindowOpen)
			windowCheckbox("Popup Demo", "Popup Demo", &g.popupWindowOpen)
			windowCheckbox("Column Layout", "Column Layout", &g.columnWindowOpen)
			windowCheckbox("Window Features", "Window Features", &g.featuresWindowOpen)
			windowCheckbox("Event Log", "Event Log", &g.logWindowOpen)
			windowCheckbox("Enhanced", "Enhanced Controls", &g.enhancedWindowOpen)
			windowCheckbox("Fixed Size", "Fixed Size", &g.showNoClose)
			windowCheckbox("No Title", "NoTitle Window", &g.showNoTitle)

			g.ui.Space(10)
			g.ui.LayoutRow(1, []int{-1}, 0)
			if g.ui.Button("Show All") {
				g.demoWindowOpen = true
				g.inputWindowOpen = true
				g.collapsibleWindowOpen = true
				g.popupWindowOpen = true
				g.columnWindowOpen = true
				g.featuresWindowOpen = true
				g.logWindowOpen = true
				g.enhancedWindowOpen = true
				g.showNoClose = true
				g.showNoTitle = true
				// Must call OpenWindow for OptClosed windows
				g.ui.OpenWindow("Demo Window")
				g.ui.OpenWindow("Input Controls")
				g.ui.OpenWindow("Collapsible Controls")
				g.ui.OpenWindow("Popup Demo")
				g.ui.OpenWindow("Column Layout")
				g.ui.OpenWindow("Window Features")
				g.ui.OpenWindow("Event Log")
				g.ui.OpenWindow("Enhanced Controls")
			}
			if g.ui.Button("Hide All") {
				g.demoWindowOpen = false
				g.inputWindowOpen = false
				g.collapsibleWindowOpen = false
				g.popupWindowOpen = false
				g.columnWindowOpen = false
				g.featuresWindowOpen = false
				g.logWindowOpen = false
				g.enhancedWindowOpen = false
				g.showNoClose = false
				g.showNoTitle = false
			}

			g.ui.Space(10)
			if g.ui.Button("Close Menu") {
				g.showWindowsMenu = false
			}

			g.ui.EndWindow()
		}
	}

	g.ui.EndFrame()

	return nil
}

// Key repeat timing constants
const (
	keyRepeatDelay    = 400 * time.Millisecond // Initial delay before repeat starts
	keyRepeatInterval = 50 * time.Millisecond  // Interval between repeats
)

// handleKeyboard processes keyboard input for textbox with key repeat support
func (g *Game) handleKeyboard() {
	now := time.Now()

	// Handle Escape to toggle windows menu
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.showWindowsMenu = !g.showWindowsMenu
	}

	// Text input - handle initial press via AppendInputChars
	chars := ebiten.AppendInputChars(nil)
	for _, c := range chars {
		g.ui.TextInput(string(c))
	}

	// Helper for key handling with repeat support
	handleKeyWithRepeat := func(ebitenKey ebiten.Key, muiKey microui.Key) {
		if inpututil.IsKeyJustPressed(ebitenKey) {
			g.ui.KeyDown(muiKey)
			g.heldKeys[ebitenKey] = now
			g.lastRepeatTime[ebitenKey] = now
		} else if ebiten.IsKeyPressed(ebitenKey) {
			// Key is held - check for repeat
			if pressTime, ok := g.heldKeys[ebitenKey]; ok {
				timeSincePress := now.Sub(pressTime)
				timeSinceRepeat := now.Sub(g.lastRepeatTime[ebitenKey])

				// After initial delay, repeat at interval
				if timeSincePress >= keyRepeatDelay && timeSinceRepeat >= keyRepeatInterval {
					// Simulate fresh key press by releasing then pressing
					// This is needed because KeyDown only sets KeyPressed on initial press
					g.ui.KeyUp(muiKey)
					g.ui.KeyDown(muiKey)
					g.lastRepeatTime[ebitenKey] = now
				}
			}
		}
		if inpututil.IsKeyJustReleased(ebitenKey) {
			g.ui.KeyUp(muiKey)
			delete(g.heldKeys, ebitenKey)
			delete(g.lastRepeatTime, ebitenKey)
		}
	}

	// Special keys with repeat support
	handleKeyWithRepeat(ebiten.KeyBackspace, microui.KeyBackspace)
	handleKeyWithRepeat(ebiten.KeyEnter, microui.KeyEnter)
	handleKeyWithRepeat(ebiten.KeyDelete, microui.KeyDelete)
	handleKeyWithRepeat(ebiten.KeyLeft, microui.KeyLeft)
	handleKeyWithRepeat(ebiten.KeyRight, microui.KeyRight)
	handleKeyWithRepeat(ebiten.KeyHome, microui.KeyHome)
	handleKeyWithRepeat(ebiten.KeyEnd, microui.KeyEnd)

	// Character key repeat for text input
	// We need to handle printable characters separately since AppendInputChars
	// only gives us the initial press
	g.handleCharacterRepeat(now)
}

// handleCharacterRepeat handles key repeat for printable characters
func (g *Game) handleCharacterRepeat(now time.Time) {
	// List of printable character keys to check for repeat
	charKeys := []ebiten.Key{
		ebiten.KeyA, ebiten.KeyB, ebiten.KeyC, ebiten.KeyD, ebiten.KeyE,
		ebiten.KeyF, ebiten.KeyG, ebiten.KeyH, ebiten.KeyI, ebiten.KeyJ,
		ebiten.KeyK, ebiten.KeyL, ebiten.KeyM, ebiten.KeyN, ebiten.KeyO,
		ebiten.KeyP, ebiten.KeyQ, ebiten.KeyR, ebiten.KeyS, ebiten.KeyT,
		ebiten.KeyU, ebiten.KeyV, ebiten.KeyW, ebiten.KeyX, ebiten.KeyY,
		ebiten.KeyZ,
		ebiten.Key0, ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4,
		ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9,
		ebiten.KeySpace, ebiten.KeyMinus, ebiten.KeyEqual, ebiten.KeyBracketLeft,
		ebiten.KeyBracketRight, ebiten.KeyBackslash, ebiten.KeySemicolon,
		ebiten.KeyApostrophe, ebiten.KeyComma, ebiten.KeyPeriod, ebiten.KeySlash,
		ebiten.KeyGraveAccent,
	}

	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	for _, key := range charKeys {
		if inpututil.IsKeyJustPressed(key) {
			// Initial press is handled by AppendInputChars, just track timing
			g.heldKeys[key] = now
			g.lastRepeatTime[key] = now
		} else if ebiten.IsKeyPressed(key) {
			// Key is held - check for repeat
			if pressTime, ok := g.heldKeys[key]; ok {
				timeSincePress := now.Sub(pressTime)
				timeSinceRepeat := now.Sub(g.lastRepeatTime[key])

				if timeSincePress >= keyRepeatDelay && timeSinceRepeat >= keyRepeatInterval {
					// Generate the character for this key
					if char := keyToChar(key, shift); char != 0 {
						g.ui.TextInput(string(char))
					}
					g.lastRepeatTime[key] = now
				}
			}
		} else {
			// Key released
			delete(g.heldKeys, key)
			delete(g.lastRepeatTime, key)
		}
	}
}

// keyToChar converts an ebiten key to its character representation
func keyToChar(key ebiten.Key, shift bool) rune {
	// Letters
	if key >= ebiten.KeyA && key <= ebiten.KeyZ {
		base := 'a' + rune(key-ebiten.KeyA)
		if shift {
			return base - 32 // Convert to uppercase
		}
		return base
	}

	// Numbers and their shift symbols
	if key >= ebiten.Key0 && key <= ebiten.Key9 {
		if shift {
			symbols := []rune{')', '!', '@', '#', '$', '%', '^', '&', '*', '('}
			return symbols[key-ebiten.Key0]
		}
		return '0' + rune(key-ebiten.Key0)
	}

	// Special characters
	switch key {
	case ebiten.KeySpace:
		return ' '
	case ebiten.KeyMinus:
		if shift {
			return '_'
		}
		return '-'
	case ebiten.KeyEqual:
		if shift {
			return '+'
		}
		return '='
	case ebiten.KeyBracketLeft:
		if shift {
			return '{'
		}
		return '['
	case ebiten.KeyBracketRight:
		if shift {
			return '}'
		}
		return ']'
	case ebiten.KeyBackslash:
		if shift {
			return '|'
		}
		return '\\'
	case ebiten.KeySemicolon:
		if shift {
			return ':'
		}
		return ';'
	case ebiten.KeyApostrophe:
		if shift {
			return '"'
		}
		return '\''
	case ebiten.KeyComma:
		if shift {
			return '<'
		}
		return ','
	case ebiten.KeyPeriod:
		if shift {
			return '>'
		}
		return '.'
	case ebiten.KeySlash:
		if shift {
			return '?'
		}
		return '/'
	case ebiten.KeyGraveAccent:
		if shift {
			return '~'
		}
		return '`'
	}

	return 0
}

func (g *Game) Draw(screen *ebiten.Image) {
	baseColor := color.RGBA{
		R: uint8(g.bgColor[0]),
		G: uint8(g.bgColor[1]),
		B: uint8(g.bgColor[2]),
		A: 255,
	}

	if g.enableMetaballs && g.metaballs != nil {
		// Draw animated metaballs background
		g.metaballs.Draw(screen, baseColor)
	} else {
		// Use solid background color from sliders
		screen.Fill(baseColor)
	}

	g.renderer.SetTarget(screen)
	g.ui.Render(g.renderer)

	// Draw status bar at bottom
	g.drawStatusBar(screen)
}

func (g *Game) drawStatusBar(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	barHeight := 20
	barY := h - barHeight

	// Status bar background - use window bg color from style
	style := g.ui.Style()
	g.renderer.DrawRect(types.Vec2{X: 0, Y: barY}, types.Vec2{X: w, Y: barHeight}, style.Colors.WindowBg)

	// ESC hint on left side
	escText := "ESC: Windows Menu"
	g.renderer.DrawText(escText, types.Vec2{X: 8, Y: barY + 3}, nil, style.Colors.Text)

	// FPS on right side
	fpsText := fmt.Sprintf("FPS: %.0f", ebiten.ActualFPS())
	textWidth := style.Font.Width(fpsText)
	g.renderer.DrawText(fpsText, types.Vec2{X: w - textWidth - 8, Y: barY + 3}, nil, style.Colors.Text)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.screenW = outsideWidth
	g.screenH = outsideHeight
	return outsideWidth, outsideHeight
}
