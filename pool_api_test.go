package microui

import "testing"

func TestPoolItem_InitGetUpdate(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	// Create a pool of items
	items := make([]PoolItem, 8)

	// Init should find a slot and return its index
	id := ui.GetID("test_item")
	idx := ui.PoolInit(items, id)
	if idx < 0 || idx >= len(items) {
		t.Fatalf("PoolInit returned invalid index: %d", idx)
	}

	// Get should find the item
	foundIdx := ui.PoolGet(items, id)
	if foundIdx != idx {
		t.Errorf("PoolGet returned %d, want %d", foundIdx, idx)
	}

	// Unknown ID should return -1
	unknownID := ui.GetID("unknown")
	notFound := ui.PoolGet(items, unknownID)
	if notFound != -1 {
		t.Errorf("PoolGet for unknown ID returned %d, want -1", notFound)
	}

	ui.EndFrame()
}

func TestPoolItem_LRUReplacement(t *testing.T) {
	ui := New(Config{})

	items := make([]PoolItem, 2) // Small pool

	// Fill pool
	ui.BeginFrame()
	id1 := ui.GetID("item1")
	idx1 := ui.PoolInit(items, id1)
	ui.EndFrame()

	ui.BeginFrame()
	id2 := ui.GetID("item2")
	idx2 := ui.PoolInit(items, id2)
	ui.EndFrame()

	// Pool is full - next init should replace LRU (item1)
	ui.BeginFrame()
	ui.PoolUpdate(items, idx2) // Touch item2 to make it more recent
	ui.EndFrame()

	ui.BeginFrame()
	id3 := ui.GetID("item3")
	idx3 := ui.PoolInit(items, id3)

	// idx3 should have replaced idx1 (LRU)
	if idx3 != idx1 {
		t.Errorf("PoolInit should replace LRU item, idx3=%d, idx1=%d", idx3, idx1)
	}

	// item1 should no longer be found
	if ui.PoolGet(items, id1) != -1 {
		t.Error("item1 should have been replaced")
	}

	ui.EndFrame()
}

func TestPoolItem_FrameCounter(t *testing.T) {
	ui := New(Config{})

	// Frame should start at 0
	if ui.Frame() != 0 {
		t.Errorf("Frame() should start at 0, got %d", ui.Frame())
	}

	// Frame should increment with BeginFrame
	ui.BeginFrame()
	if ui.Frame() != 1 {
		t.Errorf("Frame() should be 1 after first BeginFrame, got %d", ui.Frame())
	}
	ui.EndFrame()

	ui.BeginFrame()
	if ui.Frame() != 2 {
		t.Errorf("Frame() should be 2 after second BeginFrame, got %d", ui.Frame())
	}
	ui.EndFrame()
}

func TestPoolItem_UpdateMarksCurrentFrame(t *testing.T) {
	ui := New(Config{})
	items := make([]PoolItem, 4)

	ui.BeginFrame()
	id := ui.GetID("test")
	idx := ui.PoolInit(items, id)

	// LastUpdate should be current frame
	if items[idx].LastUpdate != ui.Frame() {
		t.Errorf("LastUpdate should be %d, got %d", ui.Frame(), items[idx].LastUpdate)
	}
	ui.EndFrame()

	// After another frame, update should set to new frame
	ui.BeginFrame()
	ui.BeginFrame() // Frame is now 3
	ui.PoolUpdate(items, idx)
	if items[idx].LastUpdate != ui.Frame() {
		t.Errorf("After PoolUpdate, LastUpdate should be %d, got %d", ui.Frame(), items[idx].LastUpdate)
	}
	ui.EndFrame()
}
