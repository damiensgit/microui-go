package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestLayoutColumn_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	// Set up a 2-column row
	ui.LayoutRow(2, []int{150, -1}, 100)

	// Left column
	ui.LayoutBeginColumn()
	ui.LayoutRow(1, []int{-1}, 0)
	rect1 := ui.LayoutNext()
	rect2 := ui.LayoutNext()
	ui.LayoutEndColumn()

	// Right column
	ui.LayoutBeginColumn()
	ui.LayoutRow(1, []int{-1}, 0)
	rect3 := ui.LayoutNext()
	ui.LayoutEndColumn()

	// rect1 and rect2 should stack vertically in left column
	if rect2.Y <= rect1.Y {
		t.Errorf("rect2.Y=%d should be > rect1.Y=%d", rect2.Y, rect1.Y)
	}

	// rect3 should be in right column (X > rect1.X)
	if rect3.X <= rect1.X+rect1.W {
		t.Errorf("rect3.X=%d should be > rect1.X+W=%d", rect3.X, rect1.X+rect1.W)
	}

	// rect3 should be at same Y as rect1 (top of row)
	if rect3.Y != rect1.Y {
		t.Errorf("rect3.Y=%d should equal rect1.Y=%d", rect3.Y, rect1.Y)
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestLayoutColumn_HeightTracking(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	// Get initial position
	ui.LayoutRow(1, []int{-1}, 0)
	initialRect := ui.LayoutNext()
	initialY := initialRect.Y + initialRect.H

	ui.LayoutRow(2, []int{150, -1}, 30)

	// Left column with multiple items (should extend beyond row height)
	ui.LayoutBeginColumn()
	ui.LayoutRow(1, []int{-1}, 30)
	ui.LayoutNext() // Item 1
	ui.LayoutNext() // Item 2
	ui.LayoutNext() // Item 3 - extends beyond initial 30px row
	ui.LayoutEndColumn()

	// Right column with one item
	ui.LayoutBeginColumn()
	ui.LayoutRow(1, []int{-1}, 30)
	ui.LayoutNext()
	ui.LayoutEndColumn()

	// Get next row position
	ui.LayoutRow(1, []int{-1}, 30)
	nextRect := ui.LayoutNext()

	// nextRect.Y should be after the tallest column content
	// Left column had 3 items at 30px each = 90px of content
	expectedMinY := initialY + 90 // At least 90px below initial
	if nextRect.Y < expectedMinY {
		t.Errorf("nextRect.Y=%d is too early, should be >= %d (column content height not tracked)",
			nextRect.Y, expectedMinY)
	}

	ui.EndWindow()
	ui.EndFrame()
}
