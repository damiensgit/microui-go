package microui

import "testing"

func TestIDStack_PushPop(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Same name should give same ID at same level
	id1 := ui.GetID("button")
	id2 := ui.GetID("button")
	if id1 != id2 {
		t.Error("Same name at same level should give same ID")
	}

	// Push ID context
	ui.PushID("container1")

	// Same name in different context should give different ID
	id3 := ui.GetID("button")
	if id3 == id1 {
		t.Error("Same name in different context should give different ID")
	}

	ui.PopID()

	// After pop, same context again
	id4 := ui.GetID("button")
	if id4 != id1 {
		t.Error("After PopID, should return to original context")
	}

	ui.EndFrame()
}

func TestIDStack_Nested(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.PushID("level1")
	id1 := ui.GetID("item")

	ui.PushID("level2")
	id2 := ui.GetID("item")

	if id1 == id2 {
		t.Error("Different nesting levels should give different IDs")
	}

	ui.PopID()
	ui.PopID()

	ui.EndFrame()
}
