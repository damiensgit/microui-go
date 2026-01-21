package microui

import "github.com/user/microui-go/types"

// Container represents a UI container (window, panel, popup).
type Container struct {
	id          ID
	name        string
	rect        types.Rect
	body           types.Rect // Content area
	contentSize    types.Vec2 // Tracks actual content size for scrolling
	minContentWidth int       // Minimum content width (prevents shrinking below established content)
	scroll         types.Vec2
	zindex      int
	open        bool
	opt         int // Options passed to container (for AutoSize, etc.)

	// Command buffer indices for z-order rendering
	headIdx int // Command buffer index at container start
	tailIdx int // Command buffer index at container end
}

// ID returns the container's ID.
func (c *Container) ID() ID {
	return c.id
}

// Name returns the container's name.
func (c *Container) Name() string {
	return c.name
}

// Rect returns the container's bounds.
func (c *Container) Rect() types.Rect {
	return c.rect
}

// SetRect sets the container's bounds.
func (c *Container) SetRect(r types.Rect) {
	c.rect = r
}

// Body returns the container's content area.
func (c *Container) Body() types.Rect {
	return c.body
}

// Scroll returns the container's scroll offset.
func (c *Container) Scroll() types.Vec2 {
	return c.scroll
}

// SetScroll sets the container's scroll offset.
func (c *Container) SetScroll(s types.Vec2) {
	c.scroll = s
}

// ZIndex returns the container's z-order.
func (c *Container) ZIndex() int {
	return c.zindex
}

// Open returns whether the container is open.
func (c *Container) Open() bool {
	return c.open
}

// Opt returns the container's option flags.
func (c *Container) Opt() int {
	return c.opt
}

// ContentSize returns the container's actual content size.
// This is useful for calculating scroll ranges.
func (c *Container) ContentSize() types.Vec2 {
	return c.contentSize
}

// SetContentSize sets the container's content size.
func (c *Container) SetContentSize(s types.Vec2) {
	c.contentSize = s
}
