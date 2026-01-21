package microui

import (
	"image/color"

	"github.com/user/microui-go/types"
)

// CommandKind identifies the type of render command.
type CommandKind int

const (
	CmdRect CommandKind = iota
	CmdText
	CmdClip
	CmdIcon
	CmdBox         // Outline rectangle
	CmdScrollTrack // Scrollbar track (background)
	CmdScrollThumb // Scrollbar thumb (draggable)
)

// Icon IDs (matching original microui)
const (
	IconClose = iota + 1
	IconCheck
	IconCollapsed
	IconExpanded
	IconResize // Resize gripper (not in original microui)
	IconMax
)

// Command represents a single render command.
// Using a concrete struct (not interface) avoids heap allocations.
type Command struct {
	Kind  CommandKind
	Rect  types.Rect
	Pos   types.Vec2
	Size  types.Vec2
	Text  string
	Color color.Color
	Icon  int
	Font  types.Font
}

// CommandBuffer holds render commands for a frame.
// The buffer is pre-allocated and reused each frame.
type CommandBuffer struct {
	cmds []Command
}

// Init initializes the command buffer with the specified capacity.
func (cb *CommandBuffer) Init(capacity int) {
	cb.cmds = make([]Command, 0, capacity)
}

// Reset clears the buffer without releasing capacity.
func (cb *CommandBuffer) Reset() {
	cb.cmds = cb.cmds[:0]
}

// Push adds a command to the buffer.
func (cb *CommandBuffer) Push(cmd Command) {
	cb.cmds = append(cb.cmds, cmd)
}

// Each iterates over all commands in the buffer.
func (cb *CommandBuffer) Each(fn func(Command)) {
	for _, cmd := range cb.cmds {
		fn(cmd)
	}
}

// Len returns the number of commands in the buffer.
func (cb *CommandBuffer) Len() int {
	return len(cb.cmds)
}

// EachRange iterates over commands in the range [start, end).
func (cb *CommandBuffer) EachRange(start, end int, fn func(Command)) {
	if start < 0 {
		start = 0
	}
	if end > len(cb.cmds) {
		end = len(cb.cmds)
	}
	for i := start; i < end; i++ {
		fn(cb.cmds[i])
	}
}
