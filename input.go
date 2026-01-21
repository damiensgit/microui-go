package microui

import "github.com/user/microui-go/types"

// MouseButton represents a mouse button.
type MouseButton int

const (
	MouseLeft MouseButton = iota
	MouseRight
	MouseMiddle
)

// Key represents a keyboard key.
type Key int

const (
	KeyShift Key = iota
	KeyCtrl
	KeyAlt
	KeyEnter
	KeyBackspace
	KeyDelete
	KeyEscape
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyTab
	KeySpace
)

// InputEvent is a union type for input events.
type InputEvent interface {
	isInput()
}

// MouseEvent represents a mouse event.
type MouseEvent struct {
	X, Y int
	Btn  MouseButton
	Down bool
}

func (MouseEvent) isInput() {}

// KeyEvent represents a keyboard event.
type KeyEvent struct {
	Key  Key
	Down bool
}

func (KeyEvent) isInput() {}

// TextEvent represents text input.
type TextEvent struct {
	Rune rune
}

func (TextEvent) isInput() {}

// MouseMove updates the mouse position.
func (u *UI) MouseMove(x, y int) {
	u.mu.Lock()
	u.input.MousePos = types.Vec2{X: x, Y: y}
	u.mu.Unlock()
}

// MouseDown handles a mouse button press.
func (u *UI) MouseDown(x, y int, btn MouseButton) {
	u.mu.Lock()
	u.input.MousePos = types.Vec2{X: x, Y: y}
	u.input.MouseDown[btn] = true
	u.input.MousePressed[btn] = true
	u.mu.Unlock()
}

// MouseUp handles a mouse button release.
func (u *UI) MouseUp(x, y int, btn MouseButton) {
	u.mu.Lock()
	u.input.MousePos = types.Vec2{X: x, Y: y}
	u.input.MouseDown[btn] = false
	u.mu.Unlock()
}

// Scroll handles mouse wheel scrolling.
// dx, dy are scroll deltas (positive = scroll right/down).
func (u *UI) Scroll(dx, dy int) {
	u.mu.Lock()
	u.input.ScrollDelta.X += dx
	u.input.ScrollDelta.Y += dy
	u.mu.Unlock()
}

// KeyDown handles a key press.
func (u *UI) KeyDown(key Key) {
	u.mu.Lock()
	if !u.input.KeyDown[key] {
		u.input.KeyPressed[key] = true // Only set on initial press
	}
	u.input.KeyDown[key] = true
	u.mu.Unlock()
}

// KeyUp handles a key release.
func (u *UI) KeyUp(key Key) {
	u.mu.Lock()
	delete(u.input.KeyDown, key)
	u.mu.Unlock()
}

// IsKeyDown returns true if the specified key is currently held down.
func (u *UI) IsKeyDown(key Key) bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.input.KeyDown[key]
}

// TextChar handles single character text input.
func (u *UI) TextChar(r rune) {
	u.mu.Lock()
	u.input.TextInput += string(r)
	u.mu.Unlock()
}

// TextInput adds text input for the current frame.
func (u *UI) TextInput(text string) {
	u.mu.Lock()
	u.input.TextInput += text
	u.mu.Unlock()
}

// InputChan returns the channel for sending input events.
func (u *UI) InputChan() chan InputEvent {
	return u.inputCh
}

func (u *UI) handleInput(ev InputEvent) {
	switch e := ev.(type) {
	case MouseEvent:
		if e.Down {
			u.MouseDown(e.X, e.Y, e.Btn)
		} else {
			u.MouseUp(e.X, e.Y, e.Btn)
		}
	case KeyEvent:
		if e.Down {
			u.KeyDown(e.Key)
		} else {
			u.KeyUp(e.Key)
		}
	case TextEvent:
		u.TextChar(e.Rune)
	}
}
