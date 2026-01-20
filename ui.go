package microui

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/user/microui-go/types"
)

// Renderer interfaces for drawing commands.
// Renderers must implement BaseRenderer; other interfaces are optional.
type (
	BaseRenderer interface {
		DrawRect(pos, size types.Vec2, c color.Color)
		DrawText(text string, pos types.Vec2, font types.Font, c color.Color)
		SetClip(rect types.Rect)
	}
	IconRenderer interface {
		DrawIcon(id int, rect types.Rect, c color.Color)
	}
	BoxRenderer interface {
		DrawBox(rect types.Rect, c color.Color)
	}
	ScrollRenderer interface {
		DrawScrollTrack(rect types.Rect)
		DrawScrollThumb(rect types.Rect)
	}
)

// Config configures a new UI instance.
type Config struct {
	Style         Style
	CommandBuf    int
	InputChanSize int
	DrawFrame     func(ui *UI, rect types.Rect, colorID int) // Custom frame drawing callback
	ScreenWidth   int                                        // Screen width for snap-to-edge (0 = disabled)
	ScreenHeight  int                                        // Screen height for snap-to-edge (0 = disabled)
	SnapThreshold int                                        // Distance in pixels to trigger snap (default 20)
}

// UI is the main context for immediate-mode UI.
type UI struct {
	style    Style
	commands CommandBuffer
	input    InputState
	inputCh  chan InputEvent

	// Pools
	windowPool     growPool[Window]
	layoutStack    growStack[Layout]
	clipStack      growStack[types.Rect]
	idStack        growStack[ID]
	panelStack     growStack[Panel]
	columnStack    growStack[ColumnLayout]
	containerStack growStack[*Container]

	// Container management
	containers map[ID]*Container
	lastZIndex int

	// Root container system (for z-order and hover-root gating)
	rootList      []*Container // Containers rendered this frame (in submission order)
	hoverRoot     *Container   // Container that should receive input this frame
	nextHoverRoot *Container   // Candidate hover root for next frame
	scrollTarget  *Container   // Container receiving scroll input

	// Current state
	currentWindowRect types.Rect // Direct storage instead of pointer

	// State tracking
	treeNodeState map[ID]bool // Tracks expanded/collapsed state for headers/tree nodes

	// Textbox state
	textboxCursor   int // Cursor position in current textbox (byte offset)
	textboxScrollX  int // Horizontal scroll offset for current textbox (pixels)
	lastTextboxID   ID  // ID of last focused textbox (reset cursor on focus change)

	// Number textbox edit mode (shift-click)
	numberTextboxID  ID     // ID of number being edited as textbox
	numberTextboxBuf []byte // Buffer for textbox editing

	// Frame counter for pool management
	frame int

	// Window interaction state
	dragID           ID         // ID of container being dragged
	dragOffset       types.Vec2 // Offset from container origin to drag start point
	dragContainer    *Container // Container being dragged (for snap-to-edge)
	resizeID         ID         // ID of container being resized
	resizeStartRect  types.Rect // Window rect when resize started
	resizeStartMouse types.Vec2 // Mouse position when resize started

	// Screen bounds for snap-to-edge
	screenWidth   int
	screenHeight  int
	snapThreshold int

	// Custom drawing callback
	drawFrame func(ui *UI, rect types.Rect, colorID int)

	// Last layout rect returned
	lastRect types.Rect

	mu sync.Mutex

	// Debug support
	debug    bool
	debugLog func(format string, args ...any)
}

// Panel represents a scrollable panel state.
type Panel struct {
	rect    types.Rect
	scrollX int
	scrollY int
}

// New creates a new UI instance with the given configuration.
func New(cfg Config) *UI {
	if cfg.CommandBuf == 0 {
		cfg.CommandBuf = 1024
	}
	if cfg.InputChanSize == 0 {
		cfg.InputChanSize = 64
	}

	// Use default style if not provided
	if cfg.Style.Font == nil && cfg.Style.Colors.Text == nil {
		cfg.Style = DefaultStyle()
	}

	ui := &UI{
		style:   cfg.Style,
		inputCh: make(chan InputEvent, cfg.InputChanSize),
		input: InputState{
			KeyDown:    make(map[Key]bool),
			KeyPressed: make(map[Key]bool),
		},
	}

	ui.commands.Init(cfg.CommandBuf)
	ui.windowPool.Init(16, 64)
	ui.layoutStack.Init(16)
	ui.clipStack.Init(16)
	ui.idStack.Init(32)
	ui.panelStack.Init(8)
	ui.columnStack.Init(8)
	ui.containerStack.Init(8)
	ui.containers = make(map[ID]*Container)
	ui.treeNodeState = make(map[ID]bool)
	ui.rootList = make([]*Container, 0, 16)

	// Initialize DrawFrame callback
	if cfg.DrawFrame != nil {
		ui.drawFrame = cfg.DrawFrame
	} else {
		ui.drawFrame = defaultDrawFrame
	}

	// Initialize screen bounds for snap-to-edge
	ui.screenWidth = cfg.ScreenWidth
	ui.screenHeight = cfg.ScreenHeight
	ui.snapThreshold = cfg.SnapThreshold
	if ui.snapThreshold == 0 {
		ui.snapThreshold = 20 // Default snap threshold
	}

	return ui
}

// SetScreenSize updates the screen dimensions for snap-to-edge functionality.
func (u *UI) SetScreenSize(width, height int) {
	u.screenWidth = width
	u.screenHeight = height
}

// applySnapToEdge snaps a container to screen edges and other snap-target windows.
// Windows with OptSnapToEdge snap to windows with OptSnapTarget and to screen edges.
func (u *UI) applySnapToEdge(cnt *Container) {
	threshold := u.snapThreshold
	snappedX := false
	snappedY := false

	// Snap to windows marked as snap targets (uses containers map which persists across frames)
	for _, other := range u.containers {
		if other == cnt || !other.open {
			continue
		}

		// Only snap to windows marked as snap targets
		if other.opt&OptSnapTarget == 0 {
			continue
		}

		// Snap to right edge of other window (our left aligns with their right)
		if !snappedX && abs(cnt.rect.X-(other.rect.X+other.rect.W)) < threshold {
			if cnt.rect.Y < other.rect.Y+other.rect.H && cnt.rect.Y+cnt.rect.H > other.rect.Y {
				cnt.rect.X = other.rect.X + other.rect.W
				snappedX = true
			}
		}

		// Snap to left edge of other window (our right aligns with their left)
		if !snappedX && abs((cnt.rect.X+cnt.rect.W)-other.rect.X) < threshold {
			if cnt.rect.Y < other.rect.Y+other.rect.H && cnt.rect.Y+cnt.rect.H > other.rect.Y {
				cnt.rect.X = other.rect.X - cnt.rect.W
				snappedX = true
			}
		}

		// Snap to bottom edge of other window (our top aligns with their bottom)
		if !snappedY && abs(cnt.rect.Y-(other.rect.Y+other.rect.H)) < threshold {
			if cnt.rect.X < other.rect.X+other.rect.W && cnt.rect.X+cnt.rect.W > other.rect.X {
				cnt.rect.Y = other.rect.Y + other.rect.H
				snappedY = true
			}
		}

		// Snap to top edge of other window (our bottom aligns with their top)
		if !snappedY && abs((cnt.rect.Y+cnt.rect.H)-other.rect.Y) < threshold {
			if cnt.rect.X < other.rect.X+other.rect.W && cnt.rect.X+cnt.rect.W > other.rect.X {
				cnt.rect.Y = other.rect.Y - cnt.rect.H
				snappedY = true
			}
		}

		// Align left edges
		if !snappedX && abs(cnt.rect.X-other.rect.X) < threshold {
			if cnt.rect.Y < other.rect.Y+other.rect.H+threshold && cnt.rect.Y+cnt.rect.H > other.rect.Y-threshold {
				cnt.rect.X = other.rect.X
				snappedX = true
			}
		}

		// Align right edges
		if !snappedX && abs((cnt.rect.X+cnt.rect.W)-(other.rect.X+other.rect.W)) < threshold {
			if cnt.rect.Y < other.rect.Y+other.rect.H+threshold && cnt.rect.Y+cnt.rect.H > other.rect.Y-threshold {
				cnt.rect.X = other.rect.X + other.rect.W - cnt.rect.W
				snappedX = true
			}
		}

		// Align top edges
		if !snappedY && abs(cnt.rect.Y-other.rect.Y) < threshold {
			if cnt.rect.X < other.rect.X+other.rect.W+threshold && cnt.rect.X+cnt.rect.W > other.rect.X-threshold {
				cnt.rect.Y = other.rect.Y
				snappedY = true
			}
		}

		// Align bottom edges
		if !snappedY && abs((cnt.rect.Y+cnt.rect.H)-(other.rect.Y+other.rect.H)) < threshold {
			if cnt.rect.X < other.rect.X+other.rect.W+threshold && cnt.rect.X+cnt.rect.W > other.rect.X-threshold {
				cnt.rect.Y = other.rect.Y + other.rect.H - cnt.rect.H
				snappedY = true
			}
		}
	}

	// Snap to screen edges (if screen size is set and not already snapped)
	if u.screenWidth == 0 || u.screenHeight == 0 {
		return
	}

	// Snap to left edge
	if !snappedX && cnt.rect.X < threshold {
		cnt.rect.X = 0
	}

	// Snap to top edge
	if !snappedY && cnt.rect.Y < threshold {
		cnt.rect.Y = 0
	}

	// Snap to right edge
	if !snappedX {
		rightEdge := u.screenWidth - cnt.rect.W
		if cnt.rect.X > rightEdge-threshold && cnt.rect.X < rightEdge+threshold {
			cnt.rect.X = rightEdge
		}
	}

	// Snap to bottom edge
	if !snappedY {
		bottomEdge := u.screenHeight - cnt.rect.H
		if cnt.rect.Y > bottomEdge-threshold && cnt.rect.Y < bottomEdge+threshold {
			cnt.rect.Y = bottomEdge
		}
	}
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// SetDebug enables debug logging with the given callback.
func (u *UI) SetDebug(logFunc func(format string, args ...any)) {
	u.debug = logFunc != nil
	u.debugLog = logFunc
}

// BeginFrame prepares for a new frame of UI rendering.
func (u *UI) BeginFrame() {
	u.frame++
	u.commands.Reset()
	u.clipStack.Reset()
	u.input.TextInput = ""

	if !u.input.MouseDown[int(MouseLeft)] {
		u.dragID = 0
		u.dragContainer = nil
		u.resizeID = 0
		u.resizeStartRect = types.Rect{}
		u.resizeStartMouse = types.Vec2{}
	}

	u.hoverRoot = u.nextHoverRoot
	u.nextHoverRoot = nil
	u.scrollTarget = nil
	u.rootList = u.rootList[:0]

	u.input.MouseDelta = types.Vec2{
		X: u.input.MousePos.X - u.input.LastMousePos.X,
		Y: u.input.MousePos.Y - u.input.LastMousePos.Y,
	}
	u.input.LastMousePos = u.input.MousePos
	u.processInput()
}

// EndFrame finalizes the current frame.
func (u *UI) EndFrame() {
	if !u.input.UpdatedFocus {
		u.input.Focus = 0
	}
	u.input.UpdatedFocus = false
	u.input.MousePressed = [3]bool{}

	for k := range u.input.KeyPressed {
		delete(u.input.KeyPressed, k)
	}

	// Apply scroll wheel to target
	if u.scrollTarget != nil && (u.input.ScrollDelta.X != 0 || u.input.ScrollDelta.Y != 0) {
		u.scrollTarget.scroll.Y += u.input.ScrollDelta.Y
		u.scrollTarget.scroll.X += u.input.ScrollDelta.X

		maxScrollY := u.scrollTarget.contentSize.Y + u.style.Padding.Y*2 - u.scrollTarget.body.H
		maxScrollX := u.scrollTarget.contentSize.X + u.style.Padding.X*2 - u.scrollTarget.body.W
		if maxScrollY < 0 {
			maxScrollY = 0
		}
		if maxScrollX < 0 {
			maxScrollX = 0
		}
		if u.scrollTarget.scroll.Y < 0 {
			u.scrollTarget.scroll.Y = 0
		}
		if u.scrollTarget.scroll.Y > maxScrollY {
			u.scrollTarget.scroll.Y = maxScrollY
		}
		if u.scrollTarget.scroll.X < 0 {
			u.scrollTarget.scroll.X = 0
		}
		if u.scrollTarget.scroll.X > maxScrollX {
			u.scrollTarget.scroll.X = maxScrollX
		}
	}

	u.input.ScrollDelta = types.Vec2{}
}

// UpdateControl updates focus/hover state for a control.
func (u *UI) UpdateControl(id ID, rect types.Rect) (hover bool, active bool) {
	return u.UpdateControlOpt(id, rect, 0)
}

// UpdateControlOpt updates focus/hover state with options.
func (u *UI) UpdateControlOpt(id ID, rect types.Rect, opt int) (hover bool, active bool) {
	if opt&OptNoInteract != 0 {
		return false, false
	}

	clipped := u.CheckClip(rect)
	if clipped == ClipAll {
		return false, false
	}

	mouseOver := rect.Contains(u.input.MousePos)
	if clipped == ClipPart {
		clipRect := u.GetClipRect()
		mouseOver = mouseOver && clipRect.Contains(u.input.MousePos)
	}

	if u.input.Focus == id {
		u.input.UpdatedFocus = true
	}

	// Gate mouse input to hover root container
	inHR := u.inHoverRoot()
	if u.debug && u.input.MousePressed[int(MouseLeft)] {
		u.debugLog("UpdateControlOpt id=%d mouseOver=%v inHoverRoot=%v MousePressed=%v", id, mouseOver, inHR, u.input.MousePressed[int(MouseLeft)])
	}
	if !inHR {
		if u.input.Focus == id && u.input.MousePressed[int(MouseLeft)] {
			u.SetFocus(0)
		}
		return false, u.input.Focus == id
	}

	// Only set hover when mouse is not down (prevents stealing during drag)
	if mouseOver && !u.input.MouseDown[int(MouseLeft)] {
		u.input.Hover = id
	}

	if u.input.Focus == id {
		if u.input.MousePressed[int(MouseLeft)] && !mouseOver {
			u.SetFocus(0)
		}
		// If mouse released, lose focus (unless HOLDFOCUS option)
		if opt&OptHoldFocus == 0 && !u.input.MouseDown[int(MouseLeft)] {
			u.SetFocus(0)
		}
	}

	// If hovered and mouse pressed, gain focus (require mouseOver to prevent stale Hover)
	if u.input.Hover == id && mouseOver && u.input.MousePressed[int(MouseLeft)] {
		u.SetFocus(id)
	}

	// Instant click focus (mouse moved to control and clicked same frame)
	if mouseOver && u.input.MousePressed[int(MouseLeft)] && u.input.Focus != id {
		u.SetFocus(id)
	}

	u.input.LastID = id
	hover = u.input.Hover == id
	active = u.input.Focus == id
	return hover, active
}

// SetFocus sets the focused control.
func (u *UI) SetFocus(id ID) {
	u.input.Focus = id
	u.input.UpdatedFocus = true
}

// MouseOver returns true if the mouse is over the given rect AND
// the current container is the hover root (or we're in a valid input context).
func (u *UI) MouseOver(rect types.Rect) bool {
	return rect.Contains(u.input.MousePos) &&
		u.inHoverRoot() &&
		u.GetClipRect().Contains(u.input.MousePos)
}

// MousePos returns the current mouse position.
func (u *UI) MousePos() types.Vec2 {
	return u.input.MousePos
}

// MouseDelta returns the mouse movement since last frame.
func (u *UI) MouseDelta() types.Vec2 {
	return u.input.MouseDelta
}

// IsCapturingMouse returns true if the UI is capturing mouse input
// (window drag, window resize, scrollbar drag, or any control interaction in progress).
// Use this to avoid processing custom mouse input while UI is handling the mouse.
func (u *UI) IsCapturingMouse() bool {
	// Window drag or resize
	if u.dragID != 0 || u.resizeID != 0 {
		return true
	}
	// Any control has focus with mouse down (e.g., scrollbar drag, slider drag)
	if u.input.Focus != 0 && u.input.MouseDown[int(MouseLeft)] {
		return true
	}
	return false
}

// IsHoverRoot returns true if the given container name is the current hover root.
// Use this to check if a window should process mouse input.
func (u *UI) IsHoverRoot(name string) bool {
	if u.hoverRoot == nil {
		return false
	}
	return u.hoverRoot.name == name
}

// Render executes all queued commands using the given renderer.
// Commands are rendered in z-order by container (lowest zindex first).
func (u *UI) Render(renderer interface{}) {
	r, ok := renderer.(BaseRenderer)
	if !ok {
		return
	}
	ir, _ := renderer.(IconRenderer)
	br, _ := renderer.(BoxRenderer)
	sr, _ := renderer.(ScrollRenderer)

	renderCmd := func(cmd Command) {
		switch cmd.Kind {
		case CmdRect:
			r.DrawRect(cmd.Pos, cmd.Size, cmd.Color)
		case CmdText:
			r.DrawText(cmd.Text, cmd.Pos, cmd.Font, cmd.Color)
		case CmdClip:
			r.SetClip(cmd.Rect)
		case CmdIcon:
			if ir != nil {
				ir.DrawIcon(cmd.Icon, cmd.Rect, cmd.Color)
			}
		case CmdBox:
			if br != nil {
				br.DrawBox(cmd.Rect, cmd.Color)
			}
		case CmdScrollTrack:
			if sr != nil {
				sr.DrawScrollTrack(cmd.Rect)
			}
		case CmdScrollThumb:
			if sr != nil {
				sr.DrawScrollThumb(cmd.Rect)
			}
		}
	}

	if len(u.rootList) == 0 {
		u.commands.Each(renderCmd)
		return
	}

	sorted := make([]*Container, len(u.rootList))
	copy(sorted, u.rootList)
	sort.Slice(sorted, func(i, j int) bool {
		// Always-on-top windows render last (on top)
		iOnTop := sorted[i].opt&OptAlwaysOnTop != 0
		jOnTop := sorted[j].opt&OptAlwaysOnTop != 0
		if iOnTop != jOnTop {
			return !iOnTop // non-always-on-top first
		}
		return sorted[i].zindex < sorted[j].zindex
	})

	for _, cnt := range sorted {
		u.commands.EachRange(cnt.headIdx, cnt.tailIdx, renderCmd)
	}
}

// RootContainersSorted returns all root containers sorted by z-index (back to front).
// Always-on-top windows are sorted after regular windows.
// This is useful for custom rendering with per-container effects like shadows.
func (u *UI) RootContainersSorted() []*Container {
	sorted := make([]*Container, len(u.rootList))
	copy(sorted, u.rootList)
	sort.Slice(sorted, func(i, j int) bool {
		// Always-on-top windows render last (on top)
		iOnTop := sorted[i].opt&OptAlwaysOnTop != 0
		jOnTop := sorted[j].opt&OptAlwaysOnTop != 0
		if iOnTop != jOnTop {
			return !iOnTop // non-always-on-top first
		}
		return sorted[i].zindex < sorted[j].zindex
	})
	return sorted
}

// RenderContainer renders just the commands for a single container.
func (u *UI) RenderContainer(cnt *Container, renderer interface{}) {
	r, ok := renderer.(BaseRenderer)
	if !ok {
		return
	}
	ir, _ := renderer.(IconRenderer)
	br, _ := renderer.(BoxRenderer)
	sr, _ := renderer.(ScrollRenderer)

	u.commands.EachRange(cnt.headIdx, cnt.tailIdx, func(cmd Command) {
		switch cmd.Kind {
		case CmdRect:
			r.DrawRect(cmd.Pos, cmd.Size, cmd.Color)
		case CmdText:
			r.DrawText(cmd.Text, cmd.Pos, cmd.Font, cmd.Color)
		case CmdClip:
			r.SetClip(cmd.Rect)
		case CmdIcon:
			if ir != nil {
				ir.DrawIcon(cmd.Icon, cmd.Rect, cmd.Color)
			}
		case CmdBox:
			if br != nil {
				br.DrawBox(cmd.Rect, cmd.Color)
			}
		case CmdScrollTrack:
			if sr != nil {
				sr.DrawScrollTrack(cmd.Rect)
			}
		case CmdScrollThumb:
			if sr != nil {
				sr.DrawScrollThumb(cmd.Rect)
			}
		}
	})
}

// Style returns the current style.
func (u *UI) Style() Style {
	return u.style
}

// Frame returns the current frame number.
func (u *UI) Frame() int {
	return u.frame
}

// ScrollDelta returns the accumulated scroll delta for this frame.
func (u *UI) ScrollDelta() types.Vec2 {
	return u.input.ScrollDelta
}

// Label adds a text label to the current layout.
func (u *UI) Label(text string) {
	u.DrawControlText(text, u.LayoutNext(), ColorText, 0)
}

// Space adds vertical spacing without any control or extra spacing.
// height is the number of cells/pixels to skip.
func (u *UI) Space(height int) {
	layout := u.getLayout()
	layout.nextRow += height
	layout.position.Y = layout.nextRow
}

// LabelOpt adds a text label with alignment options.
func (u *UI) LabelOpt(text string, opt int) {
	u.DrawControlText(text, u.LayoutNext(), ColorText, opt)
}

// Button adds a button to the current layout.
// Returns true if the button was clicked this frame.
func (u *UI) Button(label string) bool {
	return u.ButtonOpt(label, 0, 0)
}

// ButtonOpt adds a button with icon and options.
func (u *UI) ButtonOpt(label string, icon int, opt int) bool {
	var id ID
	if label != "" {
		id = u.getID(label)
	} else {
		id = u.getIDFromInt(icon)
	}
	rect := u.LayoutNext()
	u.UpdateControlOpt(id, rect, opt)
	clicked := u.input.MousePressed[int(MouseLeft)] && u.input.Focus == id
	u.DrawControlFrame(id, rect, ColorButton, opt)
	if label != "" {
		u.DrawControlText(label, rect, ColorText, opt|OptAlignCenter)
	}
	if icon != 0 {
		u.DrawIcon(icon, rect, u.style.Colors.Text)
	}
	return clicked
}

// ToggleButton adds a toggle button that stays selected.
// Returns true if clicked (state should be toggled by caller).
func (u *UI) ToggleButton(label string, selected bool) bool {
	return u.ToggleButtonOpt(label, 0, selected, 0)
}

// ToggleButtonOpt adds a toggle button with icon and options.
// The selected parameter indicates if the button is currently active/pressed.
// Returns true if clicked (state should be toggled by caller).
func (u *UI) ToggleButtonOpt(label string, icon int, selected bool, opt int) bool {
	var id ID
	if label != "" {
		id = u.getID(label)
	} else {
		id = u.getIDFromInt(icon)
	}
	rect := u.LayoutNext()
	u.UpdateControlOpt(id, rect, opt)
	clicked := u.input.MousePressed[int(MouseLeft)] && u.input.Focus == id

	// Draw as pressed/focused if selected
	colorID := ColorButton
	if selected {
		colorID = ColorButtonFocus
	}
	u.DrawControlFrame(id, rect, colorID, opt)

	if label != "" {
		u.DrawControlText(label, rect, ColorText, opt|OptAlignCenter)
	}
	if icon != 0 {
		u.DrawIcon(icon, rect, u.style.Colors.Text)
	}
	return clicked
}

// BeginWindow starts a new window.
// Returns false if the window is closed.
func (u *UI) BeginWindow(title string, rect types.Rect) bool {
	return u.BeginWindowOpt(title, rect, 0)
}

// OpenWindow explicitly opens a window (useful before using OptClosed).
func (u *UI) OpenWindow(title string) {
	cnt := u.GetContainer(title)
	cnt.open = true
}

// BeginWindowOpt starts a new window with options.
// opt can include OptNoTitle, OptNoClose, OptNoResize, OptAutoSize, OptPopup, OptClosed.
// Returns false if the window is closed.
func (u *UI) BeginWindowOpt(title string, rect types.Rect, opt int) bool {
	// Get or create container BEFORE pushing ID (container ID should be stable)
	cnt := u.GetContainer(title)
	// Only set rect on first frame (when zindex is 0, meaning not yet initialized)
	// After that, the container maintains its own position (for dragging, etc.)
	if cnt.zindex == 0 {
		cnt.rect = rect
	}

	// Use container's rect for all subsequent operations (supports dragging/resizing)
	rect = cnt.rect

	// Store options for EndWindow to use (e.g., for AutoSize)
	cnt.opt = opt

	// Without OptClosed, auto-open the container (this is for regular windows)
	// Must happen BEFORE the close check below
	if !cnt.open && opt&OptClosed == 0 {
		cnt.open = true
	}

	if opt&OptPopup != 0 && opt&OptClosed != 0 {
		if u.input.MousePressed[int(MouseLeft)] && u.hoverRoot != cnt {
			cnt.open = false
		}
	}

	if !cnt.open {
		return false
	}

	u.PushID(title)
	if cnt.zindex == 0 {
		u.lastZIndex++
		cnt.zindex = u.lastZIndex
	}
	u.containerStack.Push(cnt)
	u.beginRootContainer(cnt)

	if cnt == u.hoverRoot && u.input.MousePressed[int(MouseLeft)] && opt&OptNoInteract == 0 {
		u.BringToFront(cnt)
	}

	if opt&OptPopup != 0 {
		u.clipStack.Push(unclippedRect)
		u.commands.Push(Command{Kind: CmdClip, Rect: unclippedRect})
	}

	if opt&OptNoFrame == 0 {
		u.DrawFrame(rect, ColorWindowBG)
	}
	u.PushClip(rect)

	titleHeight := u.style.TitleHeight
	borderWidth := u.style.BorderWidth
	contentRect := rect
	body := rect

	if borderWidth > 0 {
		contentRect.X += borderWidth
		contentRect.W -= borderWidth * 2
		contentRect.H -= borderWidth
		if opt&OptNoTitle != 0 {
			contentRect.Y += borderWidth
			contentRect.H -= borderWidth
		}
	}
	body = contentRect

	if opt&OptNoTitle == 0 {
		titleRect := types.Rect{X: contentRect.X, Y: rect.Y, W: contentRect.W, H: titleHeight}
		u.DrawFrame(titleRect, ColorTitleBG)
		titleID := u.GetID("!title")

		mouseOnTitle := titleRect.Contains(u.input.MousePos)
		if u.input.MousePressed[int(MouseLeft)] && mouseOnTitle && cnt == u.hoverRoot {
			if u.debug {
				u.debugLog("TitleBarClick: window=%q titleRect=%v mousePos=%v -> BringToFront", title, titleRect, u.input.MousePos)
			}
			u.BringToFront(cnt)
		}
		u.UpdateControlOpt(titleID, titleRect, opt)

		if u.input.Focus == titleID && u.input.MouseDown[int(MouseLeft)] {
			if u.input.MousePressed[int(MouseLeft)] {
				u.dragID = titleID
				u.dragContainer = cnt
				u.dragOffset = types.Vec2{
					X: u.input.MousePos.X - cnt.rect.X,
					Y: u.input.MousePos.Y - cnt.rect.Y,
				}
			}
			if u.dragID == titleID {
				newX := u.input.MousePos.X - u.dragOffset.X
				newY := u.input.MousePos.Y - u.dragOffset.Y
				if u.debug {
					u.debugLog("WindowDrag: pos=(%d,%d) offset=(%d,%d) newPos=(%d,%d)",
						u.input.MousePos.X, u.input.MousePos.Y, u.dragOffset.X, u.dragOffset.Y, newX, newY)
				}
				cnt.rect.X = newX
				cnt.rect.Y = newY

				// Apply real-time snapping while dragging
				if opt&OptSnapToEdge != 0 {
					u.applySnapToEdge(cnt)
				}
			}
		}

		body.Y += titleRect.H
		body.H -= titleRect.H

		if opt&OptNoClose == 0 {
			closeID := u.GetID("!close")
			closeRect := types.Rect{
				X: titleRect.X + titleRect.W - titleRect.H - 1,
				Y: titleRect.Y,
				W: titleRect.H,
				H: titleRect.H,
			}
			titleRect.W -= closeRect.W
			u.DrawIcon(IconClose, closeRect, u.style.Colors.TitleText)
			u.UpdateControlOpt(closeID, closeRect, opt)

			if u.debug && u.input.MousePressed[int(MouseLeft)] {
				mouseOver := closeRect.Contains(u.input.MousePos)
				u.debugLog("CloseButton: rect=%v mousePos=%v mouseOver=%v focus=%d closeID=%d MousePressed=%v",
					closeRect, u.input.MousePos, mouseOver, u.input.Focus, closeID, u.input.MousePressed[int(MouseLeft)])
			}

			if u.input.MousePressed[int(MouseLeft)] && u.input.Focus == closeID {
				if u.debug {
					u.debugLog("CloseButton: CLOSING WINDOW!")
				}
				cnt.open = false
			}
		}

	
		u.DrawControlText(title, titleRect, ColorTitleText, opt)

		contentRect = body
	}

	if opt&OptAutoSize != 0 {
		overheadW := rect.W - contentRect.W
		overheadH := rect.H - contentRect.H
		newW := cnt.contentSize.X + overheadW + u.style.Padding.X*2
		newH := cnt.contentSize.Y + overheadH + u.style.Padding.Y*2

		minW := u.style.Size.X + u.style.Padding.X*2
		minH := u.style.Size.Y + u.style.Padding.Y*2
		if minW < 10 {
			minW = 10
		}
		if minH < 3 {
			minH = 3
		}
		if newW < minW {
			newW = minW
		}
		if newH < minH {
			newH = minH
		}

		cnt.rect.W = newW
		cnt.rect.H = newH
		rect = cnt.rect
		contentRect = rect
		if borderWidth > 0 {
			contentRect.X += borderWidth
			contentRect.W -= borderWidth * 2
			contentRect.H -= borderWidth
			if opt&OptNoTitle != 0 {
				contentRect.Y += borderWidth
				contentRect.H -= borderWidth
			}
		}
		if opt&OptNoTitle == 0 {
			contentRect.Y += titleHeight
			contentRect.H -= titleHeight
		}
	}

	u.scrollbars(cnt, &contentRect)

	if opt&OptNoResize == 0 {
		sz := u.style.ScrollbarSize
		resizeID := u.GetID("!resize")
		resizeRect := types.Rect{
			X: rect.X + rect.W - sz,
			Y: rect.Y + rect.H - sz,
			W: sz,
			H: sz,
		}
		u.UpdateControlOpt(resizeID, resizeRect, opt)
		u.DrawIcon(IconResize, resizeRect, u.style.Colors.Text)

		if u.input.Focus == resizeID && u.input.MouseDown[int(MouseLeft)] {
			if u.input.MousePressed[int(MouseLeft)] {
				u.resizeID = resizeID
				u.resizeStartRect = cnt.rect
				u.resizeStartMouse = u.input.MousePos
			}

			if u.resizeID == resizeID {
				deltaX := u.input.MousePos.X - u.resizeStartMouse.X
				deltaY := u.input.MousePos.Y - u.resizeStartMouse.Y
				desiredW := u.resizeStartRect.W + deltaX
				desiredH := u.resizeStartRect.H + deltaY
				if desiredW < 10 {
					desiredW = 10
				}
				if desiredH < 5 {
					desiredH = 5
				}

				cnt.rect.W = desiredW
				cnt.rect.H = desiredH
			}
		}
	}

	cnt.body = contentRect
	u.currentWindowRect = contentRect
	u.PushClip(contentRect)

	paddedBody := expandRectXY(contentRect, -u.style.Padding.X, -u.style.Padding.Y)
	if paddedBody.W < 0 {
		paddedBody.W = 0
	}
	if paddedBody.H < 0 {
		paddedBody.H = 0
	}
	u.pushLayout(paddedBody, cnt.scroll)

	return true
}

// EndWindow finishes the current window.
func (u *UI) EndWindow() {
	cnt := u.GetCurrentContainer()
	if cnt != nil {
		layout := u.getLayout()
		cnt.contentSize.X = layout.max.X - layout.body.X
		cnt.contentSize.Y = layout.max.Y - layout.body.Y

		maxScrollX := cnt.contentSize.X + u.style.Padding.X*2 - cnt.body.W
		maxScrollY := cnt.contentSize.Y + u.style.Padding.Y*2 - cnt.body.H
		if maxScrollX < 0 {
			maxScrollX = 0
		}
		if maxScrollY < 0 {
			maxScrollY = 0
		}
		if cnt.scroll.X > maxScrollX {
			cnt.scroll.X = maxScrollX
		}
		if cnt.scroll.Y > maxScrollY {
			cnt.scroll.Y = maxScrollY
		}
	}

	u.PopLayout()
	u.PopClip()
	u.PopClip()

	if cnt != nil && cnt.opt&OptPopup != 0 {
		u.PopClip()
	}

	u.currentWindowRect = types.Rect{}
	if cnt != nil {
		u.endRootContainer(cnt)
	}

	u.containerStack.Pop()
	u.PopID() // Pop window ID scope
}

// GetCurrentContainer returns the current (topmost) container.
func (u *UI) GetCurrentContainer() *Container {
	if u.containerStack.Len() == 0 {
		return nil
	}
	return u.containerStack.Peek()
}

// GetContainer returns a container by name, creating it if needed.
// Container IDs are not affected by ID scoping - they are always stable.
func (u *UI) GetContainer(name string) *Container {
	id := u.getRawID(name) // Use raw ID - containers ignore ID stack
	if cnt, ok := u.containers[id]; ok {
		return cnt
	}
	// Create new container (starts closed)
	cnt := &Container{
		id:   id,
		name: name,
		open: false,
	}
	u.containers[id] = cnt
	return cnt
}

// BringToFront brings a container to the front of the z-order.
func (u *UI) BringToFront(cnt *Container) {
	u.lastZIndex++
	cnt.zindex = u.lastZIndex
}

// beginRootContainer marks the start of a root container (window/popup).
// It adds the container to rootList and tracks nextHoverRoot for input routing.
func (u *UI) beginRootContainer(cnt *Container) {
	// Add to root list
	u.rootList = append(u.rootList, cnt)

	// Record command buffer start index
	cnt.headIdx = u.commands.Len()

	// Non-interactive containers don't receive mouse input
	if cnt.opt&OptNoInteract != 0 {
		return
	}

	// Track hover root: if mouse is inside, check if this container should receive input
	mouseInRect := u.input.MousePos.X >= cnt.rect.X &&
		u.input.MousePos.X < cnt.rect.X+cnt.rect.W &&
		u.input.MousePos.Y >= cnt.rect.Y &&
		u.input.MousePos.Y < cnt.rect.Y+cnt.rect.H

	if mouseInRect {
		// Determine if this container should be the hover root
		// Always-on-top windows have input priority over regular windows
		shouldBeHoverRoot := false
		if u.nextHoverRoot == nil {
			shouldBeHoverRoot = true
		} else {
			cntOnTop := cnt.opt&OptAlwaysOnTop != 0
			hoverOnTop := u.nextHoverRoot.opt&OptAlwaysOnTop != 0
			if cntOnTop && !hoverOnTop {
				// Always-on-top wins over regular window
				shouldBeHoverRoot = true
			} else if cntOnTop == hoverOnTop {
				// Same category, use zindex
				shouldBeHoverRoot = cnt.zindex >= u.nextHoverRoot.zindex
			}
			// If hover is on-top and cnt is not, don't change
		}
		if shouldBeHoverRoot {
			u.nextHoverRoot = cnt
		}
	}

	// Track scroll target: same logic as hover root
	if mouseInRect {
		shouldBeScrollTarget := false
		if u.scrollTarget == nil {
			shouldBeScrollTarget = true
		} else {
			cntOnTop := cnt.opt&OptAlwaysOnTop != 0
			scrollOnTop := u.scrollTarget.opt&OptAlwaysOnTop != 0
			if cntOnTop && !scrollOnTop {
				shouldBeScrollTarget = true
			} else if cntOnTop == scrollOnTop {
				shouldBeScrollTarget = cnt.zindex >= u.scrollTarget.zindex
			}
		}
		if shouldBeScrollTarget {
			u.scrollTarget = cnt
		}
	}
}

// endRootContainer marks the end of a root container.
// It records the tail command index for z-order rendering.
func (u *UI) endRootContainer(cnt *Container) {
	// Record command buffer end index
	cnt.tailIdx = u.commands.Len()

	// Pop container from stack (for popups this tracks nesting)
	// Note: actual container stack popping is done by the caller (EndWindow/EndPopup)
}

// inHoverRoot returns true if the current container is in the hover root path.
func (u *UI) inHoverRoot() bool {
	if u.hoverRoot == nil {
		return true
	}
	for i := u.containerStack.Len() - 1; i >= 0; i-- {
		cnt := u.containerStack.items[i]
		if cnt == u.hoverRoot {
			return true
		}
		// Stop at root containers that aren't the hover root
		for _, root := range u.rootList {
			if cnt == root && cnt != u.hoverRoot {
				return false
			}
		}
	}
	return false
}

// PushCommand adds a command to the buffer.
func (u *UI) PushCommand(cmd Command) {
	u.commands.Push(cmd)
}

// DrawBox draws an outline rectangle at the specified position.
func (u *UI) DrawBox(rect types.Rect, c color.Color) {
	u.commands.Push(Command{
		Kind:  CmdBox,
		Rect:  rect,
		Pos:   types.Vec2{X: rect.X, Y: rect.Y},
		Size:  types.Vec2{X: rect.W, Y: rect.H},
		Color: c,
	})
}

// DrawRect draws a filled rectangle at the specified position.
func (u *UI) DrawRect(rect types.Rect, c color.Color) {
	u.commands.Push(Command{
		Kind:  CmdRect,
		Rect:  rect,
		Pos:   types.Vec2{X: rect.X, Y: rect.Y},
		Size:  types.Vec2{X: rect.W, Y: rect.H},
		Color: c,
	})
}

// drawScrollTrack adds a scrollbar track command.
func (u *UI) drawScrollTrack(rect types.Rect) {
	u.commands.Push(Command{
		Kind: CmdScrollTrack,
		Rect: rect,
	})
}

// drawScrollThumb adds a scrollbar thumb command.
func (u *UI) drawScrollThumb(rect types.Rect) {
	u.commands.Push(Command{
		Kind: CmdScrollThumb,
		Rect: rect,
	})
}

// DrawFrame draws a control frame using the configured callback.
// This allows users to customize how control backgrounds are rendered.
func (u *UI) DrawFrame(rect types.Rect, colorID int) {
	u.drawFrame(u, rect, colorID)
}

// DrawControlFrame draws a control frame with hover/focus color adjustment.
func (u *UI) DrawControlFrame(id ID, rect types.Rect, colorID int, opt int) {
	if opt&OptNoFrame != 0 {
		return
	}
	// Adjust color based on focus/hover state
	if u.input.Focus == id {
		colorID += 2
	} else if u.input.Hover == id {
		colorID += 1
	}
	u.DrawFrame(rect, colorID)
}

// DrawControlText draws text inside a control rect with alignment options.
func (u *UI) DrawControlText(text string, rect types.Rect, colorID int, opt int) {
	font := u.style.Font
	textWidth := font.Width(text)
	textHeight := font.Height()

	// Clip to rect
	u.PushClip(rect)

	// Calculate position based on alignment
	var pos types.Vec2
	pos.Y = rect.Y + (rect.H-textHeight)/2

	if opt&OptAlignCenter != 0 {
		pos.X = rect.X + (rect.W-textWidth)/2
	} else if opt&OptAlignRight != 0 {
		pos.X = rect.X + rect.W - textWidth - u.style.Padding.X
	} else {
		pos.X = rect.X + u.style.Padding.X
	}

	u.commands.Push(Command{
		Kind:  CmdText,
		Text:  text,
		Pos:   pos,
		Color: u.GetColorByID(colorID),
		Font:  font,
	})

	u.PopClip()
}

// defaultDrawFrame draws a filled rectangle with border.
func defaultDrawFrame(ui *UI, rect types.Rect, colorID int) {
	c := ui.GetColorByID(colorID)
	ui.DrawRect(rect, c)

	// Draw border if border color has non-zero alpha
	// Skip border for scrollbar elements and title bar
	if colorID == ColorScrollBase || colorID == ColorScrollThumb || colorID == ColorTitleBG {
		return
	}

	// Check if border color has alpha
	if ui.style.Colors.Border != nil {
		_, _, _, a := ui.style.Colors.Border.RGBA()
		if a > 0 {
			// Draw border
			borderRect := types.Rect{
				X: rect.X - 1,
				Y: rect.Y - 1,
				W: rect.W + 2,
				H: rect.H + 2,
			}
			ui.DrawBox(borderRect, ui.style.Colors.Border)
		}
	}
}

// GetColorByID returns the color for a given color ID.
// This is useful for custom DrawFrame callbacks.
func (u *UI) GetColorByID(colorID int) color.Color {
	switch colorID {
	case ColorText:
		return u.style.Colors.Text
	case ColorBorder:
		return u.style.Colors.Border
	case ColorWindowBG:
		return u.style.Colors.WindowBg
	case ColorTitleBG:
		return u.style.Colors.WindowTitle
	case ColorTitleText:
		if u.style.Colors.TitleText != nil {
			return u.style.Colors.TitleText
		}
		return u.style.Colors.Text
	case ColorPanelBG:
		return u.style.Colors.PanelBg
	case ColorButton:
		return u.style.Colors.Button
	case ColorButtonHover:
		return u.style.Colors.ButtonHover
	case ColorButtonFocus:
		return u.style.Colors.ButtonActive
	case ColorBase:
		return u.style.Colors.Base
	case ColorBaseHover:
		return u.style.Colors.BaseHover
	case ColorBaseFocus:
		return u.style.Colors.BaseFocus
	case ColorScrollBase:
		return u.style.Colors.ScrollBase
	case ColorScrollThumb:
		return u.style.Colors.ScrollThumb
	default:
		return u.style.Colors.Text
	}
}

// DrawIcon draws an icon at the specified rect.
func (u *UI) DrawIcon(iconID int, rect types.Rect, c color.Color) {
	// Check clipping
	clipped := u.CheckClip(rect)
	if clipped == ClipAll {
		return
	}

	// If partially clipped, set clip first
	if clipped == ClipPart {
		u.commands.Push(Command{
			Kind: CmdClip,
			Rect: u.GetClipRect(),
		})
	}

	u.commands.Push(Command{
		Kind:  CmdIcon,
		Icon:  iconID,
		Rect:  rect,
		Color: c,
	})

	// Restore clip if we changed it
	if clipped == ClipPart {
		u.commands.Push(Command{
			Kind: CmdClip,
			Rect: u.GetClipRect(), // Restore to current clip, not unclipped
		})
	}
}

// PushClip pushes a clip rectangle onto the stack.
// The new clip is intersected with the current clip, ensuring nested clips can only shrink.
func (u *UI) PushClip(rect types.Rect) {
	// Intersect with current clip
	if u.clipStack.Len() > 0 {
		current := u.clipStack.Peek()
		rect = intersectRect(rect, current)
	}
	u.clipStack.Push(rect)
	u.commands.Push(Command{
		Kind: CmdClip,
		Rect: rect,
	})
}

// intersectRect returns the intersection of two rectangles.
func intersectRect(a, b types.Rect) types.Rect {
	// Find the intersection
	x1 := max(a.X, b.X)
	y1 := max(a.Y, b.Y)
	x2 := min(a.X+a.W, b.X+b.W)
	y2 := min(a.Y+a.H, b.Y+b.H)

	// If no intersection, return empty rect
	if x2 <= x1 || y2 <= y1 {
		return types.Rect{}
	}

	return types.Rect{
		X: x1,
		Y: y1,
		W: x2 - x1,
		H: y2 - y1,
	}
}

// PopClip pops a clip rectangle from the stack.
func (u *UI) PopClip() {
	u.clipStack.Pop()
	// Restore previous clip
	if u.clipStack.Len() > 0 {
		prev := u.clipStack.Peek()
		u.commands.Push(Command{
			Kind: CmdClip,
			Rect: prev,
		})
	} else {
		// Clear clip
		u.commands.Push(Command{
			Kind: CmdClip,
			Rect: unclippedRect,
		})
	}
}

// unclippedRect is the default clip rect (effectively no clipping).
var unclippedRect = types.Rect{X: 0, Y: 0, W: 10000, H: 10000}

// GetClipRect returns the current clip rectangle.
func (u *UI) GetClipRect() types.Rect {
	if u.clipStack.Len() == 0 {
		return unclippedRect
	}
	return u.clipStack.Peek()
}

// CheckClip checks if a rectangle is clipped by the current clip rect.
// Returns ClipNone if fully visible, ClipPart if partially visible, ClipAll if invisible.
func (u *UI) CheckClip(rect types.Rect) int {
	cr := u.GetClipRect()

	// Fully outside (no intersection)
	if rect.X > cr.X+cr.W || rect.X+rect.W < cr.X ||
		rect.Y > cr.Y+cr.H || rect.Y+rect.H < cr.Y {
		return ClipAll
	}

	// Fully inside
	if rect.X >= cr.X && rect.X+rect.W <= cr.X+cr.W &&
		rect.Y >= cr.Y && rect.Y+rect.H <= cr.Y+cr.H {
		return ClipNone
	}

	// Partially visible
	return ClipPart
}

// PushWindowRect sets the current window content area.
func (u *UI) PushWindowRect(rect types.Rect) {
	u.currentWindowRect = rect
}

// PopWindowRect restores the previous window.
func (u *UI) PopWindowRect() {
	u.currentWindowRect = types.Rect{}
}

// Checkbox adds a checkbox to the current layout.
func (u *UI) Checkbox(label string, checked *bool) bool {
	id := u.getIDFromPtr(checked)
	rect := u.LayoutNext()
	box := types.Rect{X: rect.X, Y: rect.Y, W: rect.H, H: rect.H}
	u.UpdateControl(id, rect)

	changed := false
	if u.input.MousePressed[int(MouseLeft)] && u.input.Focus == id {
		*checked = !*checked
		changed = true
	}

	u.DrawControlFrame(id, box, ColorBase, 0)
	if *checked {
		u.DrawIcon(IconCheck, box, u.style.Colors.Text)
	}
	u.DrawControlText(label, types.Rect{X: rect.X + box.W, Y: rect.Y, W: rect.W - box.W, H: rect.H}, ColorText, 0)
	return changed
}

// Slider adds a horizontal slider to the current layout.
// Returns true if the value changed this frame.
func (u *UI) Slider(value *float64, low, high float64) bool {
	return u.SliderOpt(value, low, high, 0, "", 0)
}

// SliderOpt adds a slider with step, format, and options.
// step: value increment (0 for smooth), format: display format string (empty to hide value)
func (u *UI) SliderOpt(value *float64, low, high, step float64, format string, opt int) bool {
	rect := u.LayoutNext()
	id := u.getIDFromPtr(value)

	_, active := u.UpdateControl(id, rect)

	if opt&OptNoInteract != 0 {
		active = false
	}

	changed := false
	if active && u.input.MouseDown[int(MouseLeft)] {
		mousePos := u.input.MousePos
		relX := mousePos.X - rect.X
		// For discrete cells: clicking cell 0 = 0.0, clicking last cell (W-1) = 1.0
		denom := rect.W - 1
		if denom < 1 {
			denom = 1
		}
		ratio := float64(relX) / float64(denom)
		newValue := low + ratio*(high-low)

		// Apply step if specified
		if step > 0 {
			newValue = low + float64(int((newValue-low)/step+0.5))*step
		}

		// Clamp
		if newValue < low {
			newValue = low
		}
		if newValue > high {
			newValue = high
		}

		if *value != newValue {
			*value = newValue
			changed = true
		}
	}

	// Draw slider track
	u.DrawControlFrame(id, rect, ColorBase, opt)

	// Calculate thumb position
	ratio := 0.5
	if high != low {
		ratio = (*value - low) / (high - low)
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	thumbSize := u.style.ThumbSize
	thumbX := rect.X + int(ratio*float64(rect.W-thumbSize))
	thumbRect := types.Rect{X: thumbX, Y: rect.Y, W: thumbSize, H: rect.H}

	// Draw thumb with frame
	u.DrawControlFrame(id, thumbRect, ColorButton, opt)

	// Draw value text
	displayFormat := format
	if displayFormat == "" {
		displayFormat = "%.2f"
	}
	text := fmt.Sprintf(displayFormat, *value)
	u.DrawControlText(text, rect, ColorText, opt)

	return changed
}

// Number adds a draggable number input to the current layout.
// Drag left/right to decrease/increase value by step.
func (u *UI) Number(value *float64, step float64) bool {
	return u.NumberOpt(value, step, "%.2f", 0)
}

// NumberOpt adds a draggable number input with format and options.
// format controls how the number is displayed (e.g., "%.2f", "%d").
// opt can include OptAlignCenter, OptAlignRight, OptNoInteract.
// Shift+click enters textbox edit mode for direct value input.
func (u *UI) NumberOpt(value *float64, step float64, format string, opt int) bool {
	rect := u.LayoutNext()
	id := u.getIDFromPtr(value)

	// Check if we're in textbox edit mode
	if u.numberTextboxID == id {
		// Click outside exits textbox mode
		if u.input.MousePressed[int(MouseLeft)] && !rect.Contains(u.input.MousePos) {
			u.numberTextboxID = 0 // Exit textbox mode without applying value
			// Fall through to render as normal number control
		} else {
			// Render as textbox instead of number control
			result := u.numberTextboxRaw(&u.numberTextboxBuf, 64, id, rect, 0)
			if result&ResSubmit != 0 {
				// Parse and apply value on Enter
				if parsed, err := strconv.ParseFloat(string(u.numberTextboxBuf), 64); err == nil {
					*value = parsed
				}
				u.numberTextboxID = 0 // Exit textbox mode
				return true           // Value changed
			}
			// Also exit on focus loss through normal means
			if u.input.Focus != id {
				u.numberTextboxID = 0
			}
			return false
		}
	}

	// Update control state
	hover, active := u.UpdateControl(id, rect)

	changed := false

	// Check if interactive
	if opt&OptNoInteract == 0 {
		// Check for shift+click to enter textbox edit mode
		if u.input.MousePressed[int(MouseLeft)] && u.input.KeyDown[KeyShift] {
			if rect.Contains(u.input.MousePos) {
				u.numberTextboxID = id
				// Initialize buffer with current value
				u.numberTextboxBuf = []byte(fmt.Sprintf(format, *value))
				u.SetFocus(id)
				return false // Don't report change yet
			}
		}

		// Drag to change value (normal click without shift)
		if active && u.input.MouseDown[int(MouseLeft)] && !u.input.KeyDown[KeyShift] {
			*value += float64(u.input.MouseDelta.X) * step
			if u.input.MouseDelta.X != 0 {
				changed = true
			}
		}
	}

	// Draw background
	bgColor := u.style.Colors.Base
	if bgColor == nil {
		bgColor = u.style.Colors.CheckBg
	}
	if hover && opt&OptNoInteract == 0 {
		if u.style.Colors.BaseHover != nil {
			bgColor = u.style.Colors.BaseHover
		} else {
			bgColor = u.style.Colors.ButtonHover
		}
	}
	if active && opt&OptNoInteract == 0 {
		if u.style.Colors.BaseFocus != nil {
			bgColor = u.style.Colors.BaseFocus
		} else {
			bgColor = u.style.Colors.ButtonActive
		}
	}

	u.commands.Push(Command{
		Kind:  CmdRect,
		Rect:  rect,
		Pos:   types.Vec2{X: rect.X, Y: rect.Y},
		Size:  types.Vec2{X: rect.W, Y: rect.H},
		Color: bgColor,
	})

	// Draw value text
	text := fmt.Sprintf(format, *value)
	textWidth := u.style.Font.Width(text)
	textHeight := u.style.Font.Height()
	textX := rect.X + u.style.Padding.X
	if opt&OptAlignCenter != 0 {
		textX = rect.X + (rect.W-textWidth)/2
	} else if opt&OptAlignRight != 0 {
		textX = rect.X + rect.W - textWidth - u.style.Padding.X
	}
	textY := rect.Y + (rect.H-textHeight)/2 // Vertically centered

	// Clip text to rect bounds
	u.PushClip(rect)
	u.commands.Push(Command{
		Kind:  CmdText,
		Text:  text,
		Pos:   types.Vec2{X: textX, Y: textY},
		Color: u.style.Colors.Text,
		Font:  u.style.Font,
	})
	u.PopClip()

	return changed
}

// numberTextboxRaw renders an inline textbox for number editing.
// This is similar to TextboxOpt but takes the id and rect directly
// since LayoutNext() was already called by NumberOpt.
func (u *UI) numberTextboxRaw(buf *[]byte, maxLen int, id ID, rect types.Rect, opt int) int {
	// Update control state - textboxes need OptHoldFocus to keep focus after click
	hover, active := u.UpdateControlOpt(id, rect, opt|OptHoldFocus)

	result := 0

	// Handle focus change - position cursor at click location
	if active && u.lastTextboxID != id {
		u.lastTextboxID = id
		u.textboxScrollX = 0 // Reset scroll on focus change
		// Position cursor at click location (not just at end)
		u.textboxCursor = u.textboxCursorFromClick(buf, rect)
	}

	// Handle click-to-reposition cursor (clicking while already focused)
	if active && hover && u.input.MousePressed[int(MouseLeft)] && u.lastTextboxID == id {
		u.textboxCursor = u.textboxCursorFromClick(buf, rect)
	}

	// Clamp cursor to valid range - ONLY for active textbox!
	// Otherwise inactive textboxes with shorter buffers would clamp the cursor
	if active {
		if u.textboxCursor > len(*buf) {
			u.textboxCursor = len(*buf)
		}
		if u.textboxCursor < 0 {
			u.textboxCursor = 0
		}
	}

	// Handle text input when focused and interactive
	if active && opt&OptNoInteract == 0 {
		// Add typed text at cursor position (UTF-8 aware)
		if len(u.input.TextInput) > 0 {
			for _, r := range u.input.TextInput {
				runeBytes := []byte(string(r))
				if len(*buf)+len(runeBytes) <= maxLen-1 {
					// Insert at cursor position
					newBuf := make([]byte, len(*buf)+len(runeBytes))
					copy(newBuf, (*buf)[:u.textboxCursor])
					copy(newBuf[u.textboxCursor:], runeBytes)
					copy(newBuf[u.textboxCursor+len(runeBytes):], (*buf)[u.textboxCursor:])
					*buf = newBuf
					u.textboxCursor += len(runeBytes)
					result |= ResChange
				}
			}
		}

		// Handle backspace (delete character before cursor, UTF-8 aware)
		if u.input.KeyPressed[KeyBackspace] && u.textboxCursor > 0 {
			// Find start of previous UTF-8 character
			i := u.textboxCursor - 1
			for i > 0 && (*buf)[i]&0xC0 == 0x80 {
				i--
			}
			// Delete from i to cursor
			newBuf := make([]byte, len(*buf)-(u.textboxCursor-i))
			copy(newBuf, (*buf)[:i])
			copy(newBuf[i:], (*buf)[u.textboxCursor:])
			*buf = newBuf
			u.textboxCursor = i
			result |= ResChange
		}

		// Delete (UTF-8 aware)
		if u.input.KeyPressed[KeyDelete] && u.textboxCursor < len(*buf) {
			i := u.textboxCursor + 1
			for i < len(*buf) && (*buf)[i]&0xC0 == 0x80 {
				i++
			}
			newBuf := make([]byte, len(*buf)-(i-u.textboxCursor))
			copy(newBuf, (*buf)[:u.textboxCursor])
			copy(newBuf[u.textboxCursor:], (*buf)[i:])
			*buf = newBuf
			result |= ResChange
		}

		// Left/Right (UTF-8 aware)
		if u.input.KeyPressed[KeyLeft] && u.textboxCursor > 0 {
			u.textboxCursor--
			for u.textboxCursor > 0 && (*buf)[u.textboxCursor]&0xC0 == 0x80 {
				u.textboxCursor--
			}
		}
		if u.input.KeyPressed[KeyRight] && u.textboxCursor < len(*buf) {
			u.textboxCursor++
			for u.textboxCursor < len(*buf) && (*buf)[u.textboxCursor]&0xC0 == 0x80 {
				u.textboxCursor++
			}
		}

		if u.input.KeyPressed[KeyHome] {
			u.textboxCursor = 0
		}
		if u.input.KeyPressed[KeyEnd] {
			u.textboxCursor = len(*buf)
		}
		if u.input.KeyPressed[KeyEnter] {
			result |= ResSubmit
		}
	}

	if active {
		result |= ResActive
	}

	// Keep cursor visible
	if active {
		textWidth := rect.W - u.style.Padding.X*2
		cursorX := u.style.Font.Width(string((*buf)[:u.textboxCursor]))
		if cursorX-u.textboxScrollX > textWidth-10 {
			u.textboxScrollX = cursorX - textWidth + 20
		}
		if cursorX < u.textboxScrollX+10 {
			u.textboxScrollX = cursorX - 10
			if u.textboxScrollX < 0 {
				u.textboxScrollX = 0
			}
		}
	}

	// Draw textbox background
	bgColor := u.style.Colors.Base
	if bgColor == nil {
		bgColor = u.style.Colors.CheckBg
	}
	if hover && opt&OptNoInteract == 0 {
		if u.style.Colors.BaseHover != nil {
			bgColor = u.style.Colors.BaseHover
		} else {
			bgColor = u.style.Colors.ButtonHover
		}
	}
	if active {
		if u.style.Colors.BaseFocus != nil {
			bgColor = u.style.Colors.BaseFocus
		} else {
			bgColor = u.style.Colors.ButtonActive
		}
	}

	u.commands.Push(Command{
		Kind:  CmdRect,
		Rect:  rect,
		Pos:   types.Vec2{X: rect.X, Y: rect.Y},
		Size:  types.Vec2{X: rect.W, Y: rect.H},
		Color: bgColor,
	})

	// Push clip rect to prevent text drawing outside textbox bounds
	textClipRect := types.Rect{
		X: rect.X + u.style.Padding.X,
		Y: rect.Y,
		W: rect.W - u.style.Padding.X*2,
		H: rect.H,
	}
	u.PushClip(textClipRect)

	// Apply scroll offset to text position
	// Vertically center text within the control (like DrawControlText does)
	textX := rect.X + u.style.Padding.X - u.textboxScrollX
	textHeight := u.style.Font.Height()
	textY := rect.Y + (rect.H-textHeight)/2

	// Draw text content (without cursor - cursor drawn separately)
	text := string(*buf)
	u.commands.Push(Command{
		Kind:  CmdText,
		Text:  text,
		Pos:   types.Vec2{X: textX, Y: textY},
		Color: u.style.Colors.Text,
		Font:  u.style.Font,
	})

	// Pop clip rect before drawing cursor (cursor should overlay text)
	u.PopClip()

	// Draw cursor as thin vertical line (modern style, doesn't shift text)
	// Drawn after PopClip so it's not clipped by text area
	if active && opt&OptNoInteract == 0 {
		textBeforeCursor := string((*buf)[:u.textboxCursor])
		cursorPixelX := textX + u.style.Font.Width(textBeforeCursor)
		cursorHeight := u.style.Font.Height()
		cursorRect := types.Rect{X: cursorPixelX, Y: textY, W: 1, H: cursorHeight}
		u.DrawRect(cursorRect, u.style.Colors.Text)
	}

	return result
}

// BeginPanel starts a scrollable panel.
// Use a unique name for each panel.
func (u *UI) BeginPanel(name string) bool {
	return u.BeginPanelOpt(name, 0)
}

// BeginPanelOpt starts a panel with options.
// opt can include OptNoFrame (no background), OptNoScroll (disable scrolling).
func (u *UI) BeginPanelOpt(name string, opt int) bool {
	// Push panel name onto ID stack for scoping
	u.PushID(name)

	// Get rect from layout
	rect := u.LayoutNext()

	// Get or create container for this panel (for scroll persistence)
	cnt := u.GetContainer(name)

	// Update rect (panels use layout rect, not stored rect)
	cnt.rect = rect

	// Store options for scrollbar check
	cnt.opt = opt

	// Push container onto stack
	u.containerStack.Push(cnt)

	// Track scroll target: if mouse is inside panel, it takes priority over parent window
	if rect.Contains(u.input.MousePos) {
		u.scrollTarget = cnt
	}

	// Draw panel background unless OptNoFrame
	if opt&OptNoFrame == 0 {
		u.DrawFrame(rect, ColorPanelBG)
	}

	// Calculate body (content area)
	body := types.Rect{
		X: rect.X,
		Y: rect.Y,
		W: rect.W,
		H: rect.H,
	}

	u.scrollbars(cnt, &body)
	cnt.body = body
	u.PushClip(cnt.body)

	panel := Panel{
		rect:    rect,
		scrollX: cnt.scroll.X,
		scrollY: cnt.scroll.Y,
	}
	u.panelStack.Push(panel)

	paddedBody := expandRectXY(cnt.body, -u.style.Padding.X, -u.style.Padding.Y)
	u.pushLayout(paddedBody, cnt.scroll)

	return true
}

// EndPanel finishes the current panel.
func (u *UI) EndPanel() {
	cnt := u.GetCurrentContainer()
	if cnt != nil {
		layout := u.getLayout()
		cnt.contentSize.X = layout.max.X - layout.body.X
		cnt.contentSize.Y = layout.max.Y - layout.body.Y

		maxScrollX := cnt.contentSize.X + u.style.Padding.X*2 - cnt.body.W
		maxScrollY := cnt.contentSize.Y + u.style.Padding.Y*2 - cnt.body.H
		if maxScrollX < 0 {
			maxScrollX = 0
		}
		if maxScrollY < 0 {
			maxScrollY = 0
		}
		if cnt.scroll.X > maxScrollX {
			cnt.scroll.X = maxScrollX
		}
		if cnt.scroll.Y > maxScrollY {
			cnt.scroll.Y = maxScrollY
		}
	}

	u.PopLayout()
	u.panelStack.Pop()
	u.PopClip()
	if cnt != nil {
		u.containerStack.Pop()
	}
	u.PopID()
}

// Header adds a collapsible header to the current layout.
// Returns true if the header is expanded (content should be shown).
func (u *UI) Header(label string) bool {
	// Headers are expanded by default
	return u.HeaderEx(label, OptExpanded)
}

// HeaderEx adds a collapsible header with options.
func (u *UI) HeaderEx(label string, opt int) bool {
	u.LayoutRow(1, []int{-1}, 0)
	id := u.GetID(label)
	expanded, exists := u.treeNodeState[id]
	if !exists {
		expanded = (opt & OptExpanded) != 0
	}
	rect := u.LayoutNext()
	u.UpdateControl(id, rect)

	if u.input.MousePressed[int(MouseLeft)] && u.input.Focus == id {
		expanded = !expanded
	}
	u.treeNodeState[id] = expanded
	u.DrawControlFrame(id, rect, ColorButton, 0)

	iconID := IconCollapsed
	if expanded {
		iconID = IconExpanded
	}
	u.DrawIcon(iconID, types.Rect{X: rect.X, Y: rect.Y, W: rect.H, H: rect.H}, u.style.Colors.Text)

	iconOffset := rect.H - u.style.Padding.X
	if iconOffset < 2 {
		iconOffset = 2
	}
	u.DrawControlText(label, types.Rect{X: rect.X + iconOffset, Y: rect.Y, W: rect.W - iconOffset, H: rect.H}, ColorText, 0)
	return expanded
}

// BeginTreeNode starts a collapsible tree node.
// Returns true if the node is expanded. Must call EndTreeNode if true is returned.
func (u *UI) BeginTreeNode(label string) bool {
	return u.BeginTreeNodeEx(label, 0)
}

// BeginTreeNodeEx starts a tree node with options.
func (u *UI) BeginTreeNodeEx(label string, opt int) bool {
	u.LayoutRow(1, []int{-1}, 0)
	rect := u.LayoutNext()
	id := u.GetID(label)

	expanded, exists := u.treeNodeState[id]
	if !exists {
		expanded = (opt & OptExpanded) != 0
	}
	u.UpdateControl(id, rect)

	if u.input.MousePressed[int(MouseLeft)] && u.input.Focus == id {
		expanded = !expanded
	}
	u.treeNodeState[id] = expanded

	if u.input.Hover == id {
		u.DrawFrame(rect, ColorButtonHover)
	}

	iconID := IconCollapsed
	if expanded {
		iconID = IconExpanded
	}
	u.DrawIcon(iconID, types.Rect{X: rect.X, Y: rect.Y, W: rect.H, H: rect.H}, u.style.Colors.Text)

	iconOffset := rect.H - u.style.Padding.X
	if iconOffset < 2 {
		iconOffset = 2
	}
	u.DrawControlText(label, types.Rect{X: rect.X + iconOffset, Y: rect.Y, W: rect.W - iconOffset, H: rect.H}, ColorText, 0)

	if expanded {
		u.getLayout().indent += u.style.Indent
		u.PushID(label)
		return true
	}
	return false
}

// EndTreeNode ends the current tree node.
func (u *UI) EndTreeNode() {
	u.getLayout().indent -= u.style.Indent
	u.PopID()
}

// Textbox adds a text input field to the current layout.
// buf is the text buffer, maxLen is the maximum length.
// Returns ResChange if text changed, ResSubmit if Enter pressed.
func (u *UI) Textbox(buf *[]byte, maxLen int) int {
	return u.TextboxOpt(buf, maxLen, 0)
}

// TextboxOpt adds a text input field with options.
// opt can include OptNoInteract (read-only), OptHoldFocus (keep focus).
func (u *UI) TextboxOpt(buf *[]byte, maxLen int, opt int) int {
	rect := u.LayoutNext()
	id := u.getIDFromPtr(buf)

	// Update control state - textboxes need OptHoldFocus to keep focus after click
	hover, active := u.UpdateControlOpt(id, rect, opt|OptHoldFocus)

	result := 0

	// Handle focus change - position cursor at click location
	if active && u.lastTextboxID != id {
		u.lastTextboxID = id
		u.textboxScrollX = 0 // Reset scroll on focus change
		// Position cursor at click location (not just at end)
		u.textboxCursor = u.textboxCursorFromClick(buf, rect)
	}

	// Handle click-to-reposition cursor (clicking while already focused)
	if active && hover && u.input.MousePressed[int(MouseLeft)] && u.lastTextboxID == id {
		u.textboxCursor = u.textboxCursorFromClick(buf, rect)
	}

	// Clamp cursor to valid range - ONLY for active textbox!
	// Otherwise inactive textboxes with shorter buffers would clamp the cursor
	if active {
		if u.textboxCursor > len(*buf) {
			u.textboxCursor = len(*buf)
		}
		if u.textboxCursor < 0 {
			u.textboxCursor = 0
		}
	}

	// Handle text input when focused and interactive
	if active && opt&OptNoInteract == 0 {
		// Add typed text at cursor position (UTF-8 aware)
		if len(u.input.TextInput) > 0 {
			for _, r := range u.input.TextInput {
				runeBytes := []byte(string(r))
				if len(*buf)+len(runeBytes) <= maxLen-1 {
					// Insert at cursor position
					newBuf := make([]byte, len(*buf)+len(runeBytes))
					copy(newBuf, (*buf)[:u.textboxCursor])
					copy(newBuf[u.textboxCursor:], runeBytes)
					copy(newBuf[u.textboxCursor+len(runeBytes):], (*buf)[u.textboxCursor:])
					*buf = newBuf
					u.textboxCursor += len(runeBytes)
					result |= ResChange
				}
			}
		}

		// Handle backspace (delete character before cursor, UTF-8 aware)
		if u.input.KeyPressed[KeyBackspace] && u.textboxCursor > 0 {
			// Find start of previous UTF-8 character
			i := u.textboxCursor - 1
			for i > 0 && (*buf)[i]&0xC0 == 0x80 {
				i--
			}
			// Delete from i to cursor
			newBuf := make([]byte, len(*buf)-(u.textboxCursor-i))
			copy(newBuf, (*buf)[:i])
			copy(newBuf[i:], (*buf)[u.textboxCursor:])
			*buf = newBuf
			u.textboxCursor = i
			result |= ResChange
		}

		// Delete (UTF-8 aware)
		if u.input.KeyPressed[KeyDelete] && u.textboxCursor < len(*buf) {
			i := u.textboxCursor + 1
			for i < len(*buf) && (*buf)[i]&0xC0 == 0x80 {
				i++
			}
			newBuf := make([]byte, len(*buf)-(i-u.textboxCursor))
			copy(newBuf, (*buf)[:u.textboxCursor])
			copy(newBuf[u.textboxCursor:], (*buf)[i:])
			*buf = newBuf
			result |= ResChange
		}

		// Left/Right (UTF-8 aware)
		if u.input.KeyPressed[KeyLeft] && u.textboxCursor > 0 {
			u.textboxCursor--
			for u.textboxCursor > 0 && (*buf)[u.textboxCursor]&0xC0 == 0x80 {
				u.textboxCursor--
			}
		}
		if u.input.KeyPressed[KeyRight] && u.textboxCursor < len(*buf) {
			u.textboxCursor++
			for u.textboxCursor < len(*buf) && (*buf)[u.textboxCursor]&0xC0 == 0x80 {
				u.textboxCursor++
			}
		}

		if u.input.KeyPressed[KeyHome] {
			u.textboxCursor = 0
		}
		if u.input.KeyPressed[KeyEnd] {
			u.textboxCursor = len(*buf)
		}
		if u.input.KeyPressed[KeyEnter] {
			result |= ResSubmit
		}
	}

	if active {
		result |= ResActive
	}

	// Keep cursor visible
	if active {
		textWidth := rect.W - u.style.Padding.X*2
		cursorX := u.style.Font.Width(string((*buf)[:u.textboxCursor]))
		if cursorX-u.textboxScrollX > textWidth-10 {
			u.textboxScrollX = cursorX - textWidth + 20
		}
		if cursorX < u.textboxScrollX+10 {
			u.textboxScrollX = cursorX - 10
			if u.textboxScrollX < 0 {
				u.textboxScrollX = 0
			}
		}
	}

	// Draw textbox background
	bgColor := u.style.Colors.Base
	if bgColor == nil {
		bgColor = u.style.Colors.CheckBg
	}
	if hover && opt&OptNoInteract == 0 {
		if u.style.Colors.BaseHover != nil {
			bgColor = u.style.Colors.BaseHover
		} else {
			bgColor = u.style.Colors.ButtonHover
		}
	}
	if active {
		if u.style.Colors.BaseFocus != nil {
			bgColor = u.style.Colors.BaseFocus
		} else {
			bgColor = u.style.Colors.ButtonActive
		}
	}

	u.commands.Push(Command{
		Kind:  CmdRect,
		Rect:  rect,
		Pos:   types.Vec2{X: rect.X, Y: rect.Y},
		Size:  types.Vec2{X: rect.W, Y: rect.H},
		Color: bgColor,
	})

	// Push clip rect to prevent text drawing outside textbox bounds
	textClipRect := types.Rect{
		X: rect.X + u.style.Padding.X,
		Y: rect.Y,
		W: rect.W - u.style.Padding.X*2,
		H: rect.H,
	}
	u.PushClip(textClipRect)

	// Apply scroll offset to text position
	// Vertically center text within the control (like DrawControlText does)
	textX := rect.X + u.style.Padding.X - u.textboxScrollX
	textHeight := u.style.Font.Height()
	textY := rect.Y + (rect.H-textHeight)/2

	// Draw text content (without cursor - cursor drawn separately)
	text := string(*buf)
	u.commands.Push(Command{
		Kind:  CmdText,
		Text:  text,
		Pos:   types.Vec2{X: textX, Y: textY},
		Color: u.style.Colors.Text,
		Font:  u.style.Font,
	})

	// Pop clip rect before drawing cursor (cursor should overlay text)
	u.PopClip()

	// Draw cursor as thin vertical line (modern style, doesn't shift text)
	// Drawn after PopClip so it's not clipped by text area
	if active && opt&OptNoInteract == 0 {
		textBeforeCursor := string((*buf)[:u.textboxCursor])
		cursorPixelX := textX + u.style.Font.Width(textBeforeCursor)
		cursorHeight := u.style.Font.Height()
		cursorRect := types.Rect{X: cursorPixelX, Y: textY, W: 1, H: cursorHeight}
		u.DrawRect(cursorRect, u.style.Colors.Text)
	}

	return result
}

// textboxCursorFromClick calculates cursor position from mouse click location.
// It walks through the text measuring character widths to find the closest position.
func (u *UI) textboxCursorFromClick(buf *[]byte, rect types.Rect) int {
	// Calculate click X position relative to text start
	textStartX := rect.X + u.style.Padding.X - u.textboxScrollX
	clickX := u.input.MousePos.X - textStartX

	// If clicked before text start, cursor goes to beginning
	if clickX <= 0 {
		return 0
	}

	// Walk through text to find position closest to click
	text := string(*buf)
	font := u.style.Font
	bestPos := len(*buf)
	bestDist := clickX // Distance if cursor at end

	pos := 0
	for i, r := range text {
		// Measure width up to this character
		charWidth := font.Width(string(r))
		textWidthBefore := font.Width(text[:i])

		// Distance from click to position before this character
		dist := clickX - textWidthBefore
		if dist < 0 {
			dist = -dist
		}
		if dist < bestDist {
			bestDist = dist
			bestPos = pos
		}

		// Distance from click to position after this character
		dist = clickX - (textWidthBefore + charWidth)
		if dist < 0 {
			dist = -dist
		}
		if dist < bestDist {
			bestDist = dist
			bestPos = pos + len(string(r))
		}

		pos += len(string(r))
	}

	return bestPos
}

// GetID returns an ID for the given name, combined with current ID stack.
func (u *UI) GetID(name string) ID {
	// Start with base hash from ID stack
	var base uint32 = 2166136261
	if u.idStack.Len() > 0 {
		base = uint32(u.idStack.Peek())
	}

	// Hash the name
	for i := 0; i < len(name); i++ {
		base ^= uint32(name[i])
		base *= 16777619
	}
	return ID(base)
}

// getRawID returns an ID for the given name WITHOUT considering the ID stack.
// Used for container lookups where ID should be stable regardless of scope.
func (u *UI) getRawID(name string) ID {
	var base uint32 = 2166136261
	for i := 0; i < len(name); i++ {
		base ^= uint32(name[i])
		base *= 16777619
	}
	return ID(base)
}

// PushID pushes a new ID context onto the stack.
// All subsequent GetID calls will be relative to this context.
func (u *UI) PushID(name string) {
	id := u.GetID(name)
	u.idStack.Push(id)
}

// PopID removes the top ID context from the stack.
func (u *UI) PopID() {
	if u.idStack.Len() > 0 {
		u.idStack.Pop()
	}
}

// getID generates an ID from a string (internal, uses GetID).
func (u *UI) getID(name string) ID {
	return u.GetID(name)
}

// getIDFromPtr generates an ID from a pointer address.
func (u *UI) getIDFromPtr(ptr interface{}) ID {
	h := uint32(2166136261)
	s := fmt.Sprintf("%p", ptr)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return ID(h)
}

// getIDFromInt generates an ID from an integer (used for icon-only buttons).
func (u *UI) getIDFromInt(val int) ID {
	return u.GetID(fmt.Sprintf("!icon:%d", val))
}

// OpenPopup opens a popup at the current mouse position.
func (u *UI) OpenPopup(name string) {
	cnt := u.GetContainer(name)
	u.hoverRoot = cnt
	u.nextHoverRoot = cnt
	cnt.rect = types.Rect{
		X: u.input.MousePos.X,
		Y: u.input.MousePos.Y,
		W: 1,
		H: 1,
	}

	cnt.open = true
	u.BringToFront(cnt)
}

// BeginPopup begins a popup container.
func (u *UI) BeginPopup(name string) bool {
	opt := OptPopup | OptAutoSize | OptNoResize | OptNoScroll | OptNoTitle | OptClosed
	return u.BeginWindowOpt(name, types.Rect{}, opt)
}

// EndPopup ends the current popup.
func (u *UI) EndPopup() {
	u.EndWindow()
}

func (u *UI) processInput() {
	for {
		select {
		case ev := <-u.inputCh:
			u.handleInput(ev)
		default:
			return
		}
	}
}

// InputState tracks the current input state.
type InputState struct {
	MousePos      types.Vec2
	MouseDelta    types.Vec2   // Mouse movement this frame
	LastMousePos  types.Vec2   // Previous frame mouse position
	MouseDown     [3]bool
	MousePressed  [3]bool      // Cleared each frame
	ScrollDelta   types.Vec2   // Accumulated scroll this frame
	KeyDown       map[Key]bool
	KeyPressed    map[Key]bool // Key presses this frame (cleared each frame)
	Focus         ID           // Currently focused control (has input capture)
	Hover         ID           // Control under mouse (only when mouse not down)
	LastID        ID           // Last control ID processed
	UpdatedFocus  bool         // Was focus used this frame
	TextInput     string       // Text input this frame
}

// ID is a unique identifier for UI elements.
type ID uint32

// Window represents a window container.
type Window struct {
	rect  types.Rect
	title string
	open  bool
}

// Text adds word-wrapped text to the current layout.
// Unlike Label, Text wraps to fit the available width.
// Explicit newlines (\n) in the text create line breaks.
func (u *UI) Text(text string) {
	layout := u.getLayout()
	font := u.style.Font
	if font == nil {
		font = &types.MockFont{}
	}

	availWidth := layout.body.W - layout.indent - u.style.Padding.X*2

	relY := layout.position.Y
	paragraphs := strings.Split(text, "\n")
	for _, para := range paragraphs {
		if para == "" {
			relY += font.Height()
			continue
		}

		words := splitWords(para)
		line := ""

		for _, word := range words {
			testLine := line
			if len(testLine) > 0 {
				testLine += " "
			}
			testLine += word

			if font.Width(testLine) > availWidth && len(line) > 0 {
				u.commands.Push(Command{
					Kind:  CmdText,
					Text:  line,
					Pos:   types.Vec2{X: layout.body.X + layout.indent + u.style.Padding.X, Y: layout.body.Y + relY},
					Color: u.style.Colors.Text,
					Font:  font,
				})
				relY += font.Height()
				line = word
			} else {
				line = testLine
			}
		}

		if len(line) > 0 {
			u.commands.Push(Command{
				Kind:  CmdText,
				Text:  line,
				Pos:   types.Vec2{X: layout.body.X + layout.indent + u.style.Padding.X, Y: layout.body.Y + relY},
				Color: u.style.Colors.Text,
				Font:  font,
			})
			relY += font.Height()
		}
	}

	absX := layout.body.X + layout.indent + u.style.Padding.X
	absY := layout.body.Y + relY
	if absX+availWidth > layout.max.X {
		layout.max.X = absX + availWidth
	}
	if absY > layout.max.Y {
		layout.max.Y = absY
	}

	layout.nextRow = relY + u.style.Spacing
	layout.position.Y = layout.nextRow
}

// splitWords splits text into words.
func splitWords(text string) []string {
	var words []string
	word := ""
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			if len(word) > 0 {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}
	if len(word) > 0 {
		words = append(words, word)
	}
	return words
}

// scrollbars handles scrollbar rendering and interaction for containers.
func (u *UI) scrollbars(cnt *Container, body *types.Rect) {
	if cnt.opt&OptNoScroll != 0 {
		return
	}

	sz := u.style.ScrollbarSize

	cs := cnt.contentSize
	cs.X += u.style.Padding.X * 2
	cs.Y += u.style.Padding.Y * 2

	u.PushClip(*body)

	prevW := cnt.body.W
	prevH := cnt.body.H
	if prevW == 0 {
		prevW = body.W
	}
	if prevH == 0 {
		prevH = body.H
	}

	if cs.Y > prevH {
		body.W -= sz
	}
	if cs.X > prevW {
		body.H -= sz
	}

	maxScrollY := cs.Y - body.H
	if maxScrollY > 0 && body.H > 0 {
		base := types.Rect{
			X: body.X + body.W,
			Y: body.Y,
			W: sz,
			H: body.H,
		}
		scrollID := u.GetID("!scrollbary")
		u.UpdateControl(scrollID, base)
		if u.input.Focus == scrollID && u.input.MouseDown[int(MouseLeft)] {
			cnt.scroll.Y += u.input.MouseDelta.Y * cs.Y / base.H
		}
		if cnt.scroll.Y < 0 {
			cnt.scroll.Y = 0
		}
		if cnt.scroll.Y > maxScrollY {
			cnt.scroll.Y = maxScrollY
		}

		u.drawScrollTrack(base)

		thumb := base
		thumbMinSize := u.style.ThumbSize
		thumb.H = base.H * body.H / cs.Y
		if thumb.H < thumbMinSize {
			thumb.H = thumbMinSize
		}
		thumb.Y += cnt.scroll.Y * (base.H - thumb.H) / maxScrollY
		u.drawScrollThumb(thumb)

		if u.MouseOver(*body) {
			u.scrollTarget = cnt
		}
	} else {
		cnt.scroll.Y = 0
	}

	maxScrollX := cs.X - body.W
	if maxScrollX > 0 && body.W > 0 {
		base := types.Rect{
			X: body.X,
			Y: body.Y + body.H,
			W: body.W,
			H: sz,
		}
		scrollID := u.GetID("!scrollbarx")
		u.UpdateControl(scrollID, base)
		if u.input.Focus == scrollID && u.input.MouseDown[int(MouseLeft)] {
			cnt.scroll.X += u.input.MouseDelta.X * cs.X / base.W
		}
		if cnt.scroll.X < 0 {
			cnt.scroll.X = 0
		}
		if cnt.scroll.X > maxScrollX {
			cnt.scroll.X = maxScrollX
		}

		u.drawScrollTrack(base)

		thumb := base
		thumbMinSize := u.style.ThumbSize
		thumb.W = base.W * body.W / cs.X
		if thumb.W < thumbMinSize {
			thumb.W = thumbMinSize
		}
		thumb.X += cnt.scroll.X * (base.W - thumb.W) / maxScrollX
		u.drawScrollThumb(thumb)

		if u.MouseOver(*body) {
			u.scrollTarget = cnt
		}
	} else {
		cnt.scroll.X = 0
	}

	u.PopClip()
}

