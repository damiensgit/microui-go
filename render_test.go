package microui

import (
	"image/color"
	"testing"

	"github.com/user/microui-go/types"
)

// mockRenderer tracks draw calls for verification
type mockRenderer struct {
	rects []types.Rect
	texts []string
	clips []types.Rect
}

func (m *mockRenderer) DrawRect(pos, size types.Vec2, c color.Color) {
	m.rects = append(m.rects, types.Rect{X: pos.X, Y: pos.Y, W: size.X, H: size.Y})
}

func (m *mockRenderer) DrawText(text string, pos types.Vec2, font types.Font, c color.Color) {
	m.texts = append(m.texts, text)
}

func (m *mockRenderer) SetClip(rect types.Rect) {
	m.clips = append(m.clips, rect)
}

func TestRender_DrawsWindowContent(t *testing.T) {
	ui := New(Config{})
	r := &mockRenderer{}

	ui.BeginFrame()
	ui.BeginWindow("Win", types.Rect{X: 10, Y: 10, W: 200, H: 100})
	ui.Label("Hello")
	ui.Label("World")
	ui.EndWindow()
	ui.EndFrame()

	ui.Render(r)

	// Should have rendered both labels
	found := 0
	for _, text := range r.texts {
		if text == "Hello" || text == "World" {
			found++
		}
	}
	if found != 2 {
		t.Errorf("expected both labels rendered, got texts: %v", r.texts)
	}
}

func TestRender_ZOrderRespected(t *testing.T) {
	ui := New(Config{})

	// Create two overlapping windows
	ui.BeginFrame()
	ui.BeginWindow("First", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.BeginWindow("Second", types.Rect{X: 50, Y: 50, W: 100, H: 100})
	ui.EndWindow()
	ui.EndFrame()

	sorted := ui.RootContainersSorted()
	if len(sorted) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(sorted))
	}

	// Second window created later should have higher z-index (rendered on top)
	if sorted[0].Name() != "First" || sorted[1].Name() != "Second" {
		t.Errorf("expected [First, Second], got [%s, %s]", sorted[0].Name(), sorted[1].Name())
	}

	// Now click on First to bring it to front
	ui.BeginFrame()
	ui.MouseMove(25, 25)
	ui.MouseDown(25, 25, MouseLeft)
	ui.BeginWindow("First", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.BeginWindow("Second", types.Rect{X: 50, Y: 50, W: 100, H: 100})
	ui.EndWindow()
	ui.EndFrame()

	sorted = ui.RootContainersSorted()
	// First should now be on top after click
	if sorted[1].Name() != "First" {
		t.Errorf("after click, expected First on top, got [%s, %s]", sorted[0].Name(), sorted[1].Name())
	}
}

func TestRender_AlwaysOnTopWindowsRenderLast(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	// Create always-on-top window first
	ui.BeginWindowOpt("Overlay", types.Rect{X: 0, Y: 0, W: 50, H: 50}, OptAlwaysOnTop)
	ui.EndWindow()
	// Create regular window second (would normally be on top due to z-order)
	ui.BeginWindow("Normal", types.Rect{X: 25, Y: 25, W: 50, H: 50})
	ui.EndWindow()
	ui.EndFrame()

	// Bring Normal to front by clicking
	ui.BeginFrame()
	ui.MouseMove(40, 40)
	ui.MouseDown(40, 40, MouseLeft)
	ui.BeginWindowOpt("Overlay", types.Rect{X: 0, Y: 0, W: 50, H: 50}, OptAlwaysOnTop)
	ui.EndWindow()
	ui.BeginWindow("Normal", types.Rect{X: 25, Y: 25, W: 50, H: 50})
	ui.EndWindow()
	ui.EndFrame()

	sorted := ui.RootContainersSorted()

	// Overlay should still be last despite lower z-index
	if sorted[len(sorted)-1].Name() != "Overlay" {
		t.Errorf("always-on-top window should render last, got: %v", containerNames(sorted))
	}
}

func TestRenderContainer_OnlyDrawsOneWindow(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("A", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.Label("window A")
	ui.EndWindow()
	ui.BeginWindow("B", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.Label("window B")
	ui.EndWindow()
	ui.EndFrame()

	// Render only window A
	r := &mockRenderer{}
	cntA := ui.GetContainer("A")
	ui.RenderContainer(cntA, r)

	hasA, hasB := false, false
	for _, text := range r.texts {
		if text == "window A" {
			hasA = true
		}
		if text == "window B" {
			hasB = true
		}
	}

	if !hasA {
		t.Error("should have rendered window A content")
	}
	if hasB {
		t.Error("should NOT have rendered window B content")
	}
}

func TestRender_InvalidRendererIgnored(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 100, H: 100})
	ui.EndWindow()
	ui.EndFrame()

	// Pass something that doesn't implement BaseRenderer
	ui.Render("not a renderer")
	ui.RenderContainer(ui.GetContainer("Test"), 12345)
	// No panic = pass
}

func containerNames(containers []*Container) []string {
	names := make([]string, len(containers))
	for i, c := range containers {
		names[i] = c.Name()
	}
	return names
}
