package microui

import "github.com/user/microui-go/types"

// Coordinate System Convention:
//
// ABSOLUTE: Screen coordinates (0,0 = top-left of screen)
//   - Layout.body
//   - Layout.max
//   - All Rect values in commands
//
// RELATIVE: Offset from body origin
//   - Layout.position
//   - Layout.nextRow (relative to body.Y)
//
// Conversion: absolute = body.{X,Y} + relative

const (
	nextTypeNone     = 0
	nextTypeAbsolute = 1
	nextTypeRelative = 2
)

// layoutBoundsSentinel is the initial value for max bounds tracking.
// Any real coordinate will exceed this, ensuring first update wins.
const layoutBoundsSentinel = -0x1000000

// pixelInclusiveBoundary adjusts for inclusive pixel boundaries.
// When a rect spans from X to X+W, the rightmost pixel is at X+W-1.
// This +1 ensures negative widths (fill remaining) calculate correctly.
const pixelInclusiveBoundary = 1

// ColumnLayout stores state for a layout column.
type ColumnLayout struct {
	columnRect types.Rect // This column's bounds
}

// Layout represents current layout state.
// Position is RELATIVE to body; body offset is added in LayoutNext.
type Layout struct {
	body            types.Rect // Container body rect (absolute coordinates)
	position        types.Vec2 // Current position RELATIVE to body
	size            types.Vec2 // Current item size
	max             types.Vec2 // Maximum content extent (absolute coordinates)
	widths          []int      // Column widths (stored from LayoutRow)
	items           int        // Number of items in row
	itemIndex       int        // Current item index
	nextRow         int        // Y position for next row (relative to body)
	indent          int        // Current indentation
	next            types.Rect // Override rect for next LayoutNext call
	nextType        int        // 0=none, 1=absolute, 2=relative (body-relative)
	minContentWidth int        // Minimum width for content (prevents shrinking)

	// Go extensions: explicit size overrides (cleared after each use)
	sizeOverrideW int // Width override from LayoutWidth (0 = not set)
	sizeOverrideH int // Height override from LayoutHeight (0 = not set)
}

// LayoutRow sets up a row layout with the specified columns.
func (u *UI) LayoutRow(columns int, widths []int, height int) {
	layout := u.getLayout()

	if widths != nil {
		layout.widths = make([]int, columns)
		copy(layout.widths, widths)
	}
	layout.items = columns
	layout.position = types.Vec2{X: layout.indent, Y: layout.nextRow}
	layout.size.Y = height
	layout.itemIndex = 0
}

// LayoutNext returns the next layout rectangle and advances the layout.
func (u *UI) LayoutNext() types.Rect {
	layout := u.getLayout()
	style := &u.style
	var res types.Rect

	if layout.nextType != nextTypeNone {
		nextType := layout.nextType
		layout.nextType = nextTypeNone
		res = layout.next
		if nextType == nextTypeAbsolute {
			u.lastRect = res
			return res
		}
	} else {
		if layout.itemIndex == layout.items {
			u.LayoutRow(layout.items, nil, layout.size.Y)
		}
		res.X = layout.position.X
		res.Y = layout.position.Y

		if layout.sizeOverrideW != 0 {
			res.W = layout.sizeOverrideW
			layout.sizeOverrideW = 0
		} else if layout.items > 0 && layout.itemIndex < len(layout.widths) {
			res.W = layout.widths[layout.itemIndex]
		} else {
			res.W = layout.size.X
		}

		if layout.sizeOverrideH != 0 {
			res.H = layout.sizeOverrideH
			layout.sizeOverrideH = 0
		} else {
			res.H = layout.size.Y
		}

		if res.W == 0 {
			res.W = style.Size.X + style.Padding.X*2
		}
		if res.H == 0 {
			res.H = style.Size.Y + style.Padding.Y*2
		}
		if res.W < 0 {
			// Use the larger of body width or minimum content width
			// This prevents controls from shrinking below established content
			effectiveWidth := layout.body.W
			if layout.minContentWidth > effectiveWidth {
				effectiveWidth = layout.minContentWidth
			}
			res.W += effectiveWidth - res.X + pixelInclusiveBoundary
		}
		if res.H < 0 {
			res.H += layout.body.H - res.Y + pixelInclusiveBoundary
		}
		layout.itemIndex++
	}

	layout.position.X += res.W + style.Spacing
	newNextRow := res.Y + res.H + style.Spacing
	if newNextRow > layout.nextRow {
		layout.nextRow = newNextRow
	}

	res.X += layout.body.X
	res.Y += layout.body.Y

	if res.X+res.W > layout.max.X {
		layout.max.X = res.X + res.W
	}
	if res.Y+res.H > layout.max.Y {
		layout.max.Y = res.Y + res.H
	}

	u.lastRect = res
	return res
}

func (u *UI) getLayout() *Layout {
	if u.layoutStack.Len() == 0 {
		layout := Layout{
			body:    u.currentWindowRect,
			max:     types.Vec2{X: layoutBoundsSentinel, Y: layoutBoundsSentinel},
		}
		if layout.body.Empty() {
			layout.body = types.Rect{X: 0, Y: 0, W: 800, H: 600}
		}
		u.layoutStack.Push(layout)
	}
	return &u.layoutStack.items[len(u.layoutStack.items)-1]
}

// pushLayout pushes a new layout context for a container.
// minContentWidth prevents controls from shrinking below established content width.
func (u *UI) pushLayout(body types.Rect, scroll types.Vec2, minContentWidth int) {
	layout := Layout{
		body: types.Rect{
			X: body.X - scroll.X,
			Y: body.Y - scroll.Y,
			W: body.W,
			H: body.H,
		},
		max:             types.Vec2{X: layoutBoundsSentinel, Y: layoutBoundsSentinel},
		minContentWidth: minContentWidth,
	}
	u.layoutStack.Push(layout)

	u.LayoutRow(1, []int{0}, 0)
}

// PushLayout is the public version of pushLayout.
func (u *UI) PushLayout(body types.Rect) {
	u.pushLayout(body, types.Vec2{}, 0)
}

// PopLayout pops the current layout context.
func (u *UI) PopLayout() {
	if u.layoutStack.Len() > 0 {
		u.layoutStack.Pop()
	}
}

// LayoutWidth sets the width for the next control only.
func (u *UI) LayoutWidth(width int) {
	u.getLayout().sizeOverrideW = width
}

// LayoutHeight sets the height for the next control only.
func (u *UI) LayoutHeight(height int) {
	u.getLayout().sizeOverrideH = height
}

// LayoutSetNext sets the rect for the next LayoutNext call.
// If relative is true, the rect is body-relative; otherwise absolute screen coordinates.
func (u *UI) LayoutSetNext(rect types.Rect, relative bool) {
	layout := u.getLayout()
	layout.next = rect
	if relative {
		layout.nextType = nextTypeRelative
	} else {
		layout.nextType = nextTypeAbsolute
	}
}

// LayoutBeginColumn starts a sub-layout column within the current row.
func (u *UI) LayoutBeginColumn() {
	columnRect := u.LayoutNext()
	u.columnStack.Push(ColumnLayout{columnRect: columnRect})
	u.pushLayout(columnRect, types.Vec2{}, 0)
}

// LayoutEndColumn ends the current column and restores parent layout.
func (u *UI) LayoutEndColumn() {
	if u.columnStack.Len() == 0 {
		return
	}
	childLayout := u.getLayout()
	u.PopLayout()
	parentLayout := u.getLayout()

	if newPosX := childLayout.position.X + childLayout.body.X - parentLayout.body.X; newPosX > parentLayout.position.X {
		parentLayout.position.X = newPosX
	}
	if newNextRow := childLayout.nextRow + childLayout.body.Y - parentLayout.body.Y; newNextRow > parentLayout.nextRow {
		parentLayout.nextRow = newNextRow
	}
	if childLayout.max.X > parentLayout.max.X {
		parentLayout.max.X = childLayout.max.X
	}
	if childLayout.max.Y > parentLayout.max.Y {
		parentLayout.max.Y = childLayout.max.Y
	}
	u.columnStack.Pop()
}

// expandRect expands a rect by n pixels on each side.
func expandRect(rect types.Rect, n int) types.Rect {
	return types.Rect{
		X: rect.X - n,
		Y: rect.Y - n,
		W: rect.W + n*2,
		H: rect.H + n*2,
	}
}

// expandRectXY expands a rect by separate X and Y values.
func expandRectXY(rect types.Rect, nx, ny int) types.Rect {
	return types.Rect{
		X: rect.X - nx,
		Y: rect.Y - ny,
		W: rect.W + nx*2,
		H: rect.H + ny*2,
	}
}
