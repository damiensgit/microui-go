package microui

import "testing"

func TestScroll_Accumulates(t *testing.T) {
	ui := New(Config{})

	// Add scroll input
	ui.Scroll(10, 20)
	ui.Scroll(5, -10)

	ui.BeginFrame()

	// Should have accumulated deltas
	if ui.ScrollDelta().X != 15 {
		t.Errorf("ScrollDelta().X = %d, want 15", ui.ScrollDelta().X)
	}
	if ui.ScrollDelta().Y != 10 {
		t.Errorf("ScrollDelta().Y = %d, want 10", ui.ScrollDelta().Y)
	}

	ui.EndFrame()

	// Next frame should start fresh
	ui.BeginFrame()
	if ui.ScrollDelta().X != 0 {
		t.Errorf("ScrollDelta should reset each frame, got X=%d", ui.ScrollDelta().X)
	}

	ui.EndFrame()
}
