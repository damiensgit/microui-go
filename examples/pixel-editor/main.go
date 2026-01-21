package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/user/microui-go"
	"github.com/user/microui-go/extras/snap"
	"github.com/user/microui-go/render/retro"
	"github.com/user/microui-go/types"
)

const (
	screenWidth  = 900
	screenHeight = 700
)

type Game struct {
	ui               *microui.UI
	renderer         *retro.Renderer
	editor           *Editor
	lastMouse        bool // Track mouse button state for proper down/up events
	lastDrawMouse    bool // Track mouse state for canvas drawing (separate from UI)
	lastInsideCanvas bool // Track if mouse was inside canvas last frame
	// For centered zoom - track panel body position from last frame
	canvasPanelBody types.Rect

	// Demo controls state
	demoChecks  [3]bool
	demoSlider  float64
	demoNumber  float64
	demoTextBuf []byte

	// Font preview - body rect saved during buildUI for drawing after render
	fontPreviewRect types.Rect
}

func NewGame() (*Game, error) {
	// Create retro renderer with pixel art editor theme
	theme := retro.PixelTheme()
	renderer, err := retro.NewRenderer(theme)
	if err != nil {
		return nil, err
	}

	// Create UI style from theme - all layout values defined in theme.go
	style := microui.GUIStyle()
	style.Font = renderer.Font()
	style.TitleHeight = theme.StyleTitleHeight()
	style.Size = types.Vec2{X: 68, Y: theme.StyleControlHeight()}
	style.ScrollbarSize = theme.StyleScrollbarWidth()
	style.ScrollbarMargin = theme.StyleScrollbarMargin()
	style.ScrollbarBorder = theme.StyleScrollbarBorder()
	style.WindowBorder = theme.StyleWindowBorder() // Visual border - body is inside this
	style.ControlMargin = theme.StyleControlMargin()   // Control border width (clip boundary)
	style.ControlPadding = theme.StyleControlPadding() // Additional padding inside border
	style.Spacing = theme.StyleSpacing()
	style.ThumbSize = theme.StyleThumbSize()
	padX, padY := theme.StylePadding()
	style.Padding = types.Vec2{X: padX, Y: padY}

	ui := microui.New(microui.Config{
		Style:     style,
		DrawFrame: renderer.DrawFrame,
		OnWindowDrag: snap.Handler(snap.Config{
			Threshold:  20,
			ScreenSize: ebiten.WindowSize, // Dynamic - updates on resize
		}),
	})

	// Create editor with 32x32 canvas
	editor := NewEditor(32, 32)

	return &Game{
		ui:          ui,
		renderer:    renderer,
		editor:      editor,
		demoChecks:  [3]bool{true, false, true},
		demoSlider:  0.5,
		demoNumber:  42,
		demoTextBuf: make([]byte, 0, 128),
	}, nil
}

func (g *Game) Update() error {
	// Handle input
	mx, my := ebiten.CursorPosition()
	g.ui.MouseMove(mx, my)

	// Only send MouseDown/MouseUp on transitions (not every frame)
	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if pressed && !g.lastMouse {
		g.ui.MouseDown(mx, my, microui.MouseLeft)
	} else if !pressed && g.lastMouse {
		g.ui.MouseUp(mx, my, microui.MouseLeft)
	}
	g.lastMouse = pressed

	// Scroll or Zoom (Ctrl+scroll = zoom centered on mouse)
	_, sy := ebiten.Wheel()
	if sy != 0 {
		if ebiten.IsKeyPressed(ebiten.KeyControl) {
			// Ctrl+scroll = zoom centered on mouse position
			g.zoomAtMouse(sy > 0)
		} else {
			// Normal scroll
			g.ui.Scroll(0, int(sy*-30))
		}
	}

	// Text input
	for _, r := range ebiten.AppendInputChars(nil) {
		g.ui.TextInput(string(r))
	}

	// Key handling
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.ui.KeyDown(microui.KeyShift)
	} else {
		g.ui.KeyUp(microui.KeyShift)
	}
	if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
		g.ui.KeyDown(microui.KeyBackspace)
	} else {
		g.ui.KeyUp(microui.KeyBackspace)
	}
	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.ui.KeyDown(microui.KeyEnter)
	} else {
		g.ui.KeyUp(microui.KeyEnter)
	}

	return nil
}

// zoomAtMouse zooms in/out centered on the current mouse position.
func (g *Game) zoomAtMouse(zoomIn bool) {
	mx, my := ebiten.CursorPosition()

	// Get the canvas panel
	panel := g.ui.GetContainer("canvas_panel")
	if panel == nil {
		// Fallback to simple zoom if panel not found
		if zoomIn {
			g.editor.ZoomIn()
		} else {
			g.editor.ZoomOut()
		}
		return
	}

	// Get current state
	oldZoom := g.editor.zoom
	oldScroll := panel.Scroll()
	body := g.canvasPanelBody

	// Calculate old canvas size and centering offset
	oldCanvasW := g.editor.width * oldZoom
	oldCanvasH := g.editor.height * oldZoom
	oldOffsetX := 0
	oldOffsetY := 0
	if oldCanvasW < body.W {
		oldOffsetX = (body.W - oldCanvasW) / 2
	}
	if oldCanvasH < body.H {
		oldOffsetY = (body.H - oldCanvasH) / 2
	}

	// Calculate canvas coordinate under mouse before zoom
	// Account for centering offset
	canvasX := float64(mx-body.X-oldOffsetX+oldScroll.X) / float64(oldZoom)
	canvasY := float64(my-body.Y-oldOffsetY+oldScroll.Y) / float64(oldZoom)

	// Perform zoom
	if zoomIn {
		g.editor.ZoomIn()
	} else {
		g.editor.ZoomOut()
	}
	newZoom := g.editor.zoom

	// If zoom didn't change (at min/max), nothing to adjust
	if newZoom == oldZoom {
		return
	}

	// Calculate new canvas size and centering offset
	newCanvasW := g.editor.width * newZoom
	newCanvasH := g.editor.height * newZoom
	newOffsetX := 0
	newOffsetY := 0
	if newCanvasW < body.W {
		newOffsetX = (body.W - newCanvasW) / 2
	}
	if newCanvasH < body.H {
		newOffsetY = (body.H - newCanvasH) / 2
	}

	// Calculate new scroll to keep same canvas coordinate under mouse
	// newScroll = canvasCoord * newZoom - (mousePos - bodyPos - newOffset)
	newScrollX := int(canvasX*float64(newZoom)) - (mx - body.X - newOffsetX)
	newScrollY := int(canvasY*float64(newZoom)) - (my - body.Y - newOffsetY)

	// Clamp scroll to valid range
	maxScrollX := newCanvasW - body.W
	maxScrollY := newCanvasH - body.H
	if newScrollX < 0 {
		newScrollX = 0
	}
	if newScrollY < 0 {
		newScrollY = 0
	}
	if maxScrollX > 0 && newScrollX > maxScrollX {
		newScrollX = maxScrollX
	}
	if maxScrollY > 0 && newScrollY > maxScrollY {
		newScrollY = maxScrollY
	}

	panel.SetScroll(types.Vec2{X: newScrollX, Y: newScrollY})
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.SetTarget(screen)
	g.renderer.Clear()

	g.ui.BeginFrame()
	g.buildUI()
	g.ui.EndFrame()

	g.renderer.Render(g.ui)

	// Draw font preview after render (bypasses UI clipping)
	if g.fontPreviewRect.W > 0 {
		g.drawFontPreviewContent()
	}
}

func (g *Game) buildUI() {
	// Toolbar options: always on top + can snap + can be snapped to + no resize
	toolbarOpts := microui.OptNoResize | microui.OptAlwaysOnTop | snap.OptSnapToEdge | snap.OptSnapTarget

	// Tools window (size = content area, chrome added automatically)
	// Content: 3 buttons + label + slider, need enough height for all
	if g.ui.BeginWindowOpt("Tools", types.Rect{X: 10, Y: 10, W: 140, H: 180}, toolbarOpts) {
		g.ui.LayoutRow(1, []int{-1}, 0)

		if g.ui.ToggleButton("Pencil", g.editor.tool == ToolPencil) {
			g.editor.SetTool(ToolPencil)
		}
		if g.ui.ToggleButton("Eraser", g.editor.tool == ToolEraser) {
			g.editor.SetTool(ToolEraser)
		}
		if g.ui.ToggleButton("Fill", g.editor.tool == ToolFill) {
			g.editor.SetTool(ToolFill)
		}

		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Brush Size:")
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.SliderOpt(&g.editor.brushSize, 1, 8, 1, "%.0f", 0)

		g.ui.EndWindow()
	}

	// Color palette window - compact grid, edge-to-edge, no margin
	// 4 cols × 8 rows = 32 colors, 24×24 screen px each (larger boxes)
	paletteCols := 4
	paletteRows := 8
	cellSize := 24 // Larger color boxes
	gridW := paletteCols * cellSize
	gridH := paletteRows * cellSize
	// Body exactly fits the grid - colors fill edge to edge
	g.ui.PushPadding(types.Vec2{0, 0}) // Zero padding for edge-to-edge content
	if g.ui.BeginWindowOpt("Palette", types.Rect{X: 10, Y: 240, W: gridW, H: gridH}, toolbarOpts) {
		// Get body position for absolute drawing
		cnt := g.ui.GetCurrentContainer()
		body := cnt.Body()

		// Draw directly to body without layout system (avoids contentSize tracking)
		for i, c := range g.editor.palette {
			col := i % paletteCols
			row := i / paletteCols
			rect := types.Rect{
				X: body.X + col*cellSize,
				Y: body.Y + row*cellSize,
				W: cellSize,
				H: cellSize,
			}
			g.ui.DrawRect(rect, c)

			// Check for click
			if g.ui.MouseOver(rect) && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				g.editor.SetColor(c)
			}

			// Draw selection indicator: black+white double border (visible on ANY color)
			if colorsEqual(c, g.editor.currentColor) {
				px := g.renderer.Theme().Px(1) // 1 logical pixel in screen pixels
				black := color.RGBA{0, 0, 0, 255}
				white := color.RGBA{255, 255, 255, 255}
				// Outer: black border
				g.ui.DrawBox(rect, black)
				// Inner: white border (1 logical pixel inside the black)
				inner := types.Rect{X: rect.X + px, Y: rect.Y + px, W: rect.W - px*2, H: rect.H - px*2}
				g.ui.DrawBox(inner, white)
			}
		}
		g.ui.EndWindow()
	}
	g.ui.PopPadding()

	// Canvas window
	if g.ui.BeginWindowOpt("Canvas", types.Rect{X: 140, Y: 10, W: 540, H: 515}, 0) {
		g.drawCanvas()
		g.ui.EndWindow()
	}

	// Info window - sized to fit content without scrollbars
	if g.ui.BeginWindowOpt("Info", types.Rect{X: 700, Y: 10, W: 190, H: 200}, toolbarOpts) {
		g.ui.LayoutRow(2, []int{60, -1}, 0)
		g.ui.Label("Size:")
		g.ui.Label("32 x 32")
		g.ui.Label("Zoom:")
		g.ui.Label("x" + itoa(g.editor.zoom))
		g.ui.Label("Tool:")
		g.ui.Label(g.editor.ToolName())

		g.ui.LayoutRow(2, []int{80, 80}, 0)
		if g.ui.Button("Zoom +") {
			g.editor.ZoomIn()
		}
		if g.ui.Button("Zoom -") {
			g.editor.ZoomOut()
		}

		g.ui.LayoutRow(1, []int{-1}, 0)
		if g.ui.Button("Clear") {
			g.editor.Clear()
		}

		g.ui.EndWindow()
	}

	// Font Preview window - shows different font styles and sizes
	// Save body rect for drawing after render (to bypass clipping)
	if g.ui.BeginWindowOpt("Fonts", types.Rect{X: 10, Y: 450, W: 400, H: 240}, 0) {
		cnt := g.ui.GetCurrentContainer()
		g.fontPreviewRect = cnt.Body()
		g.ui.EndWindow()
	} else {
		g.fontPreviewRect = types.Rect{} // Window closed
	}

	// Controls Demo window - showcases all UI controls
	if g.ui.BeginWindowOpt("Controls Demo", types.Rect{X: 700, Y: 260, W: 190, H: 300}, 0) {
		// Buttons section
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Buttons:")
		g.ui.LayoutRow(2, []int{85, -1}, 0)
		g.ui.Button("Normal")
		g.ui.Button("Button 2")

		// Checkboxes section
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Checkboxes:")
		g.ui.Checkbox("Option A", &g.demoChecks[0])
		g.ui.Checkbox("Option B", &g.demoChecks[1])
		g.ui.Checkbox("Option C", &g.demoChecks[2])

		// Slider section
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Slider:")
		g.ui.Slider(&g.demoSlider, 0, 1)

		// Number input section
		g.ui.LayoutRow(2, []int{60, -1}, 0)
		g.ui.Label("Number:")
		g.ui.Number(&g.demoNumber, 1)

		// Text input section
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Text Input:")
		g.ui.Textbox(&g.demoTextBuf, 128)

		// Tree node / Header section (shows expand/collapse icons)
		g.ui.LayoutRow(1, []int{-1}, 0)
		if g.ui.Header("Expandable Section") {
			g.ui.Label("  Hidden content 1")
			g.ui.Label("  Hidden content 2")
		}

		// Icons demo - buttons with icons
		g.ui.LayoutRow(1, []int{-1}, 0)
		g.ui.Label("Icons (in buttons):")
		g.ui.LayoutRow(4, []int{40, 40, 40, 40}, 0)
		g.ui.ButtonOpt("", microui.IconClose, 0)
		g.ui.ButtonOpt("", microui.IconCheck, 0)
		g.ui.ButtonOpt("", microui.IconCollapsed, 0)
		g.ui.ButtonOpt("", microui.IconExpanded, 0)

		g.ui.EndWindow()
	}
}

func (g *Game) drawCanvas() {
	// Use a scrollable panel for the canvas content
	g.ui.LayoutRow(1, []int{-1}, -1)
	g.ui.BeginPanel("canvas_panel")

	// Save panel body for centered zoom calculations
	var panelBody types.Rect
	if panel := g.ui.GetContainer("canvas_panel"); panel != nil {
		panelBody = panel.Body()
		g.canvasPanelBody = panelBody
	}

	// Calculate pixel size based on zoom
	pixelSize := g.editor.zoom
	canvasW := g.editor.width * pixelSize
	canvasH := g.editor.height * pixelSize

	// Calculate centering offset - center canvas within panel when smaller than panel
	offsetX := 0
	offsetY := 0
	if canvasW < panelBody.W {
		offsetX = (panelBody.W - canvasW) / 2
	}
	if canvasH < panelBody.H {
		offsetY = (panelBody.H - canvasH) / 2
	}

	// Reserve space for the full canvas (enables scrolling when zoomed)
	// Add centering margins to the layout
	contentW := canvasW
	contentH := canvasH
	if canvasW < panelBody.W {
		contentW = panelBody.W // Use full width when canvas is smaller
	}
	if canvasH < panelBody.H {
		contentH = panelBody.H // Use full height when canvas is smaller
	}
	g.ui.LayoutRow(1, []int{contentW}, contentH)
	layoutRect := g.ui.LayoutNext()

	// Calculate actual canvas rect with centering
	canvasRect := types.Rect{
		X: layoutRect.X + offsetX,
		Y: layoutRect.Y + offsetY,
		W: canvasW,
		H: canvasH,
	}

	// Draw canvas background
	g.ui.DrawRect(canvasRect, g.renderer.Theme().Canvas.Base)

	// Draw pixels
	clipRect := g.ui.GetClipRect()
	for y := 0; y < g.editor.height; y++ {
		for x := 0; x < g.editor.width; x++ {
			px := canvasRect.X + x*pixelSize
			py := canvasRect.Y + y*pixelSize

			// Skip pixels outside clip rect for performance
			if px+pixelSize < clipRect.X || px > clipRect.X+clipRect.W ||
				py+pixelSize < clipRect.Y || py > clipRect.Y+clipRect.H {
				continue
			}

			c := g.editor.GetPixel(x, y)
			g.ui.DrawRect(types.Rect{X: px, Y: py, W: pixelSize, H: pixelSize}, c)
		}
	}

	// Draw grid if zoomed in enough
	if g.editor.zoom >= 8 {
		gridColor := color.RGBA{R: 60, G: 60, B: 60, A: 100}
		for y := 0; y <= g.editor.height; y++ {
			py := canvasRect.Y + y*pixelSize
			if py >= clipRect.Y && py <= clipRect.Y+clipRect.H {
				g.ui.DrawRect(types.Rect{X: canvasRect.X, Y: py, W: canvasW, H: 1}, gridColor)
			}
		}
		for x := 0; x <= g.editor.width; x++ {
			px := canvasRect.X + x*pixelSize
			if px >= clipRect.X && px <= clipRect.X+clipRect.W {
				g.ui.DrawRect(types.Rect{X: px, Y: canvasRect.Y, W: 1, H: canvasH}, gridColor)
			}
		}
	}

	// Save the visible area (clip rect) for hit testing before EndPanel
	visibleRect := clipRect

	g.ui.EndPanel()

	// Handle drawing input only if:
	// 1. UI isn't capturing the mouse for drag/resize/scrollbar
	// 2. Canvas window is the hover root (no other window is under the mouse)
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	canDraw := !g.ui.IsCapturingMouse() && g.ui.IsHoverRoot("Canvas")

	// Determine if mouse is inside drawable canvas area
	insideCanvas := false
	var cx, cy int

	if canDraw {
		mx, my := ebiten.CursorPosition()
		// Check if mouse is within the VISIBLE panel area (not the full canvas)
		if mx >= visibleRect.X && mx < visibleRect.X+visibleRect.W &&
			my >= visibleRect.Y && my < visibleRect.Y+visibleRect.H {
			// Convert to canvas coordinates (canvasRect includes centering offset)
			cx = (mx - canvasRect.X) / pixelSize
			cy = (my - canvasRect.Y) / pixelSize

			if cx >= 0 && cx < g.editor.width && cy >= 0 && cy < g.editor.height {
				insideCanvas = true
			}
		}
	}

	// Handle stroke state transitions
	if mousePressed && insideCanvas {
		if !g.lastDrawMouse || !g.lastInsideCanvas {
			// Mouse just pressed OR just entered canvas - start new stroke
			g.editor.StartStroke(cx, cy)
		} else {
			// Mouse held inside canvas - continue stroke with line interpolation
			g.editor.ContinueStroke(cx, cy)
		}
	}

	// End stroke when mouse released OR left canvas area
	if g.lastDrawMouse && (!mousePressed || !insideCanvas) {
		g.editor.EndStroke()
	}

	g.lastDrawMouse = mousePressed && canDraw
	g.lastInsideCanvas = insideCanvas
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) drawFontPreviewContent() {
	// Use saved body rect from buildUI
	body := g.fontPreviewRect

	// Colors
	textColor := color.RGBA{220, 220, 220, 255}
	labelColor := color.RGBA{150, 150, 150, 255}
	headerColor := color.RGBA{100, 180, 255, 255}

	// Font styles to preview
	styles := []retro.FontStyle{retro.FontSans, retro.FontM6x11, retro.FontM5x7}
	styleNames := []string{"Pixeloid", "m6x11", "m5x7"}

	// Sizes to preview
	// Pixeloid: clean at 9, 18, 27
	// m6x11/m5x7: clean at 16, 32, 48
	sizes := []float64{16, 18, 32}

	// Column positions
	col1 := body.X + 8   // Size label
	col2 := body.X + 50  // Pixeloid
	col3 := body.X + 170 // m6x11
	col4 := body.X + 290 // m5x7

	// Draw column headers
	y := body.Y + 4
	g.renderer.DrawFontSample("Size", col1, y, retro.FontMono, 12, headerColor)
	g.renderer.DrawFontSample(styleNames[0], col2, y, retro.FontMono, 12, headerColor)
	g.renderer.DrawFontSample(styleNames[1], col3, y, retro.FontMono, 12, headerColor)
	g.renderer.DrawFontSample(styleNames[2], col4, y, retro.FontMono, 12, headerColor)
	y += 20

	// Sample text
	sampleText := "Abc123"

	// Draw each size row
	for _, size := range sizes {
		// Size label
		label := itoa(int(size)) + "px"
		g.renderer.DrawFontSample(label, col1, y, retro.FontMono, 12, labelColor)

		// Each style
		cols := []int{col2, col3, col4}
		for i, style := range styles {
			g.renderer.DrawFontSample(sampleText, cols[i], y, style, size, textColor)
		}

		// Variable row height based on font size
		y += int(size) + 8
	}
}

func colorsEqual(a, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Pixel Editor - microui-go retro demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(120) // Higher TPS for smoother mouse tracking

	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
