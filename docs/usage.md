# Usage

* [Overview](#overview)
* [Getting Started](#getting-started)
* [Main Loop](#main-loop)
* [Windows](#windows)
* [Panels](#panels)
* [Controls](#controls)
* [Layout](#layout)
* [IDs](#ids)
* [Input](#input)
* [Rendering](#rendering)
* [Style](#style)
* [Custom Controls](#custom-controls)

## Overview

microui-go is an immediate-mode UI library. This means the UI is built each frame by calling functions that return interaction results immediately. There is no retained widget tree — you describe what exists right now, and the library figures out the rest.

```go
if ui.Button("Click Me") {
    // button was clicked this frame
}
```

## Getting Started

Create a UI instance and configure the font measurement callbacks:

```go
import (
    "github.com/user/microui-go"
    "github.com/user/microui-go/types"
)

ui := microui.New()

// Set font for text measurement (required for proper layout)
ui.SetStyle(microui.Style{
    Font: myFont, // implements types.Font interface
    // ... other style options
})
```

The `types.Font` interface requires:
```go
type Font interface {
    Width(text string) int  // pixel width of text
    Height() int            // pixel height of a line
}
```

## Main Loop

Each frame follows this pattern:

```go
// 1. Provide input state
ui.SetMousePos(mouseX, mouseY)
if mousePressed {
    ui.SetMouseDown(microui.MouseLeft)
}
if mouseReleased {
    ui.SetMouseUp(microui.MouseLeft)
}
ui.SetScroll(scrollX, scrollY)
// ... keyboard input

// 2. Begin frame
ui.BeginFrame()

// 3. Build UI
if ui.BeginWindow("My Window", types.Rect{X: 10, Y: 10, W: 300, H: 400}) {
    // controls go here
    ui.EndWindow()
}

// 4. End frame
ui.EndFrame()

// 5. Render
ui.Render(renderer)
```

## Windows

Windows are the top-level containers. They can be dragged, resized, and closed.

```go
if ui.BeginWindow("Title", types.Rect{X: 100, Y: 100, W: 300, H: 200}) {
    // window content
    ui.EndWindow()
}
```

`BeginWindow` returns false if the window is closed (e.g., user clicked X). Use `BeginWindowOpt` for more control:

```go
// Window options
microui.OptNoResize    // disable resize handle
microui.OptNoScroll    // disable scrollbars
microui.OptNoClose     // hide close button
microui.OptNoTitle     // hide title bar
microui.OptAutoSize    // size to fit content
microui.OptPopup       // popup behavior (closes on outside click)
microui.OptClosed      // start closed, require OpenWindow() call
microui.OptNoInteract  // ignore input (HUD overlay)
```

To programmatically open a window that uses `OptClosed`:
```go
ui.OpenWindow("Title")
```

## Panels

Panels are scrollable regions within windows:

```go
ui.BeginPanel("panel-id")
// scrollable content
ui.EndPanel()
```

Panels get their size from the current layout row.

## Controls

### Labels
```go
ui.Label("Hello")
ui.LabelOpt("Centered", microui.OptAlignCenter)
```

### Buttons
```go
if ui.Button("Click") {
    // clicked
}

// With icon
if ui.ButtonOpt("Save", microui.IconCheck, 0) {
    // clicked
}
```

### Checkboxes
```go
var checked bool
if ui.Checkbox("Enable", &checked) {
    // state changed
}
```

### Sliders
```go
var value float64 = 0.5
if ui.Slider(&value, 0, 1) {
    // value changed
}

// With format string
ui.SliderOpt(&value, 0, 100, 1, "%.0f%%", 0)
```

### Number Input
Drag to change, shift+click to type directly:
```go
var num float64 = 42
if ui.Number(&num, 1) { // step = 1
    // value changed
}
```

### Text Input
```go
var buf []byte = []byte("initial text")
if ui.Textbox(&buf) & microui.ResSubmit != 0 {
    // Enter pressed
}
```

### Headers (Collapsible)
```go
if ui.Header("Section") {
    // content shown when expanded
}

// Expanded by default
if ui.HeaderEx("Open Section", microui.OptExpanded) {
    // content
}
```

### Tree Nodes
```go
if ui.TreeNode("Parent") {
    ui.Label("Child 1")
    if ui.TreeNode("Nested") {
        ui.Label("Grandchild")
    }
}
```

## Layout

Layout is row-based. Call `LayoutRow` to configure how subsequent controls are positioned:

```go
// 3 columns: 100px, 150px, fill remaining
ui.LayoutRow(3, []int{100, 150, -1}, 0)
ui.Button("A")  // 100px wide
ui.Button("B")  // 150px wide
ui.Button("C")  // fills remaining space
```

**Width values:**
* Positive: absolute pixel/cell width
* Zero: use style default
* Negative: relative to remaining space (-1 = fill, -2 = fill minus 1, etc.)

**Height:** The third parameter. Zero uses the style default.

### Columns

For side-by-side regions with independent layouts:

```go
ui.LayoutRow(2, []int{200, -1}, 300)

ui.LayoutBeginColumn()
ui.LayoutRow(1, []int{-1}, 0)
ui.Label("Left side")
ui.Button("L1")
ui.LayoutEndColumn()

ui.LayoutBeginColumn()
ui.LayoutRow(1, []int{-1}, 0)
ui.Label("Right side")
ui.Button("R1")
ui.LayoutEndColumn()
```

### Manual Positioning

For absolute positioning within the current container:

```go
ui.LayoutSetNext(types.Rect{X: 50, Y: 50, W: 100, H: 30}, false)
ui.Button("Absolute")
```

## IDs

Controls are identified by hashing their label. If you have multiple controls with the same label, use ID scoping:

```go
for i, item := range items {
    ui.PushID(i)
    if ui.Button("Delete") {
        // delete items[i]
    }
    ui.PopID()
}
```

Or use a unique pointer:
```go
ui.PushIDFromPtr(&items[i])
// controls
ui.PopID()
```

## Input

Provide input state before `BeginFrame`:

### Mouse
```go
ui.SetMousePos(x, y)
ui.SetMouseDown(microui.MouseLeft)   // button pressed this frame
ui.SetMouseUp(microui.MouseLeft)     // button released this frame
ui.SetScroll(deltaX, deltaY)         // scroll wheel
```

### Keyboard
```go
ui.SetKeyDown(microui.KeyBackspace)
ui.SetKeyUp(microui.KeyBackspace)
ui.InputText("typed characters")     // text input
```

**Key constants:**
```go
microui.KeyShift
microui.KeyCtrl
microui.KeyAlt
microui.KeyBackspace
microui.KeyDelete
microui.KeyReturn
microui.KeyLeft
microui.KeyRight
microui.KeyHome
microui.KeyEnd
```

## Rendering

After `EndFrame`, iterate the command buffer:

```go
ui.Render(renderer)
```

Your renderer must implement:
```go
type Renderer interface {
    DrawRect(pos, size types.Vec2, c color.Color)
    DrawText(text string, pos types.Vec2, font types.Font, c color.Color)
    SetClip(rect types.Rect)
}
```

Optional interfaces for extended features:
```go
// Icons (close button, checkmark, arrows)
DrawIcon(id int, rect types.Rect, c color.Color)

// Box outlines
DrawBox(rect types.Rect, c color.Color)

// Custom scrollbar appearance
DrawScrollTrack(rect types.Rect)
DrawScrollThumb(rect types.Rect)
```

**Command types:** `CmdRect`, `CmdText`, `CmdIcon`, `CmdClip`, `CmdBox`, `CmdScrollTrack`, `CmdScrollThumb`

## Style

Customize appearance through `ui.SetStyle()`:

```go
style := microui.GUIStyle() // or TUIStyle() for terminals

style.Font = myFont
style.Size = types.Vec2{X: 68, Y: 10}      // default control size
style.Padding = types.Vec2{X: 5, Y: 5}     // internal padding
style.Spacing = 4                           // between controls
style.Indent = 24                           // tree/header indent
style.TitleHeight = 24                      // window title bar
style.ScrollbarSize = 12
style.ThumbSize = 8

// Colors
style.Colors.Text = color.White
style.Colors.Border = color.Gray
style.Colors.WindowBg = color.RGBA{40, 40, 40, 255}
style.Colors.Button = color.RGBA{70, 70, 70, 255}
style.Colors.ButtonHover = color.RGBA{90, 90, 90, 255}
// ... etc

ui.SetStyle(style)
```

### Custom Frame Drawing

Override how control backgrounds are drawn:

```go
ui.SetDrawFrame(func(ui *microui.UI, rect types.Rect, colorID int) {
    // custom drawing logic
    // colorID indicates which element (button, base, window, etc.)
})
```

## Custom Controls

Build your own controls using the low-level API:

```go
func MyCustomControl(ui *microui.UI, data *MyData) bool {
    // 1. Get layout rect
    rect := ui.LayoutNext()

    // 2. Get unique ID
    id := ui.GetID("my-control")

    // 3. Register for input handling
    ui.UpdateControl(id, rect)

    // 4. Check state
    isHover := ui.Input().Hover == id
    isFocus := ui.Input().Focus == id

    // 5. Handle interaction
    changed := false
    if isFocus && ui.Input().MouseDown[microui.MouseLeft] {
        data.Value += ui.Input().MouseDelta.X
        changed = true
    }

    // 6. Draw
    ui.DrawRect(rect, myColor)
    ui.DrawControlText("label", rect, microui.ColorText, 0)

    return changed
}
```

Use `ui.SetFocus(id)` to grab keyboard focus, and check `ui.Input().KeyPressed[key]` for key events.
