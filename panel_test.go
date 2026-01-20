package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestPanel_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	if ui.BeginPanel("test-panel") {
		ui.Label("Inside panel")
		ui.Label("More content")
		ui.EndPanel()
	}

	ui.EndFrame()
}

func TestPanel_WithWindow(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	rect := types.Rect{X: 10, Y: 10, W: 300, H: 200}
	if ui.BeginWindow("Test", rect) {
		if ui.BeginPanel("inner-panel") {
			ui.Label("Panel content")
			ui.EndPanel()
		}
		ui.EndWindow()
	}

	ui.EndFrame()
}

func TestPanel_MultiplePanels(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	if ui.BeginPanel("panel-1") {
		ui.Label("Panel 1")
		ui.EndPanel()
	}

	if ui.BeginPanel("panel-2") {
		ui.Label("Panel 2")
		ui.EndPanel()
	}

	ui.EndFrame()
}

func TestPanel_Nested(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	if ui.BeginPanel("outer") {
		ui.Label("Outer panel")

		if ui.BeginPanel("inner") {
			ui.Label("Inner panel")
			ui.EndPanel()
		}

		ui.EndPanel()
	}

	ui.EndFrame()
}

func TestPanel_GeneratesCommands(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	if ui.BeginPanel("test") {
		ui.Label("Content")
		ui.EndPanel()
	}

	ui.EndFrame()

	// Check that commands were generated (clip + rect + text)
	if ui.commands.Len() < 3 {
		t.Errorf("Panel should generate at least 3 commands, got %d", ui.commands.Len())
	}
}

func TestPanel_CreatesScrollableRegion(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Container", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	// Create a panel with fixed size
	ui.BeginPanel("ScrollPanel")

	// Add lots of content
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Panel content line")
	}

	ui.EndPanel()
	ui.EndWindow()
	ui.EndFrame()

	// Panel should exist as a container
	cnt := ui.GetContainer("ScrollPanel")
	if cnt == nil {
		t.Fatal("Panel should create a container")
	}
}

func TestPanel_ClipsContent(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Container", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	// Create panel with small height, lots of content
	ui.LayoutRow(1, []int{-1}, 100) // Panel height = 100
	ui.BeginPanel("ClipPanel")
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Should be clipped")
	}
	ui.EndPanel()

	ui.EndWindow()
	ui.EndFrame()

	// Test passed if no crash - clipping is visually verified
}

func TestPanel_HasScrollbars(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Container", types.Rect{X: 0, Y: 0, W: 400, H: 300})

	ui.LayoutRow(1, []int{200}, 100) // Panel 200x100
	ui.BeginPanel("ScrollablePanel")
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndPanel()

	ui.EndWindow()
	ui.EndFrame()

	// Panel should have tracked content size
	cnt := ui.GetContainer("ScrollablePanel")
	if cnt.ContentSize().Y <= 100 {
		t.Errorf("Panel ContentSize.Y = %d, want > 100 (content overflows)", cnt.ContentSize().Y)
	}
}

func TestPanel_ScrollPersists(t *testing.T) {
	ui := New(Config{})

	// Frame 1: Create panel and scroll it
	ui.BeginFrame()
	ui.BeginWindow("Container", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 100)
	ui.BeginPanel("PersistPanel")
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndPanel()
	ui.EndWindow()
	ui.EndFrame()

	// Set scroll manually for testing
	cnt := ui.GetContainer("PersistPanel")
	cnt.SetScroll(types.Vec2{X: 0, Y: 50})

	// Frame 2: Verify scroll persists
	ui.BeginFrame()
	ui.BeginWindow("Container", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{200}, 100)
	ui.BeginPanel("PersistPanel")
	for i := 0; i < 20; i++ {
		ui.LayoutRow(1, []int{-1}, 0)
		ui.Label("Content")
	}
	ui.EndPanel()
	ui.EndWindow()
	ui.EndFrame()

	if cnt.Scroll().Y != 50 {
		t.Errorf("Scroll.Y = %d, want 50 (should persist)", cnt.Scroll().Y)
	}
}
