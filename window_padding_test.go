package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestWindow_PushPadding(t *testing.T) {
	ui := New(Config{})
	style := ui.Style()

	// Verify style has non-zero padding (precondition)
	if style.Padding.X == 0 && style.Padding.Y == 0 {
		t.Skip("Test requires non-zero default padding")
	}

	ui.BeginFrame()

	// Window WITH padding (default)
	ui.BeginWindow("WithPadding", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	layoutWith := ui.getLayout()
	bodyWith := layoutWith.body
	ui.EndWindow()

	// Window WITHOUT padding (using PushPadding)
	ui.PushPadding(types.Vec2{0, 0})
	ui.BeginWindow("NoPadding", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	layoutNo := ui.getLayout()
	bodyNo := layoutNo.body
	ui.EndWindow()
	ui.PopPadding()

	ui.EndFrame()

	// NoPadding body should be larger by 2*padding on each axis
	expectedWidthDiff := style.Padding.X * 2
	expectedHeightDiff := style.Padding.Y * 2

	actualWidthDiff := bodyNo.W - bodyWith.W
	actualHeightDiff := bodyNo.H - bodyWith.H

	if actualWidthDiff != expectedWidthDiff {
		t.Errorf("Width diff = %d, want %d (2*padding.X)", actualWidthDiff, expectedWidthDiff)
	}
	if actualHeightDiff != expectedHeightDiff {
		t.Errorf("Height diff = %d, want %d (2*padding.Y)", actualHeightDiff, expectedHeightDiff)
	}
}

func TestWindow_PushPaddingFullBody(t *testing.T) {
	ui := New(Config{})
	style := ui.Style()

	// For GUI style: BorderWidth=0, WindowBorder=0, TitleHeight=24
	// Body size given = 200x150
	// Window rect = 200x174 (150+24 title)
	// With zero padding: layout body should equal body rect

	ui.BeginFrame()
	ui.PushPadding(types.Vec2{0, 0})
	ui.BeginWindow("FullBody", types.Rect{X: 10, Y: 20, W: 200, H: 150})
	layout := ui.getLayout()
	ui.EndWindow()
	ui.PopPadding()
	ui.EndFrame()

	// Layout body should be full content area (no padding shrink)
	// X: 10 (window X, no border)
	// Y: 20 + 24 (window Y + title height)
	// W: 200 (full width)
	// H: 150 (full height, body = content)

	expectedBody := types.Rect{
		X: 10,
		Y: 20 + style.TitleHeight,
		W: 200,
		H: 150,
	}

	if layout.body != expectedBody {
		t.Errorf("Layout body = %v, want %v", layout.body, expectedBody)
	}
}

func TestWindow_PushPaddingRestores(t *testing.T) {
	ui := New(Config{})
	originalPadding := ui.Style().Padding

	// Push custom padding
	ui.PushPadding(types.Vec2{0, 0})
	if ui.Style().Padding.X != 0 || ui.Style().Padding.Y != 0 {
		t.Error("PushPadding should set padding to zero")
	}

	// Pop should restore
	ui.PopPadding()
	if ui.Style().Padding != originalPadding {
		t.Errorf("PopPadding should restore original: got %v, want %v", ui.Style().Padding, originalPadding)
	}
}

func TestWindow_PushPaddingNested(t *testing.T) {
	ui := New(Config{})
	original := ui.Style().Padding

	// Push first override
	ui.PushPadding(types.Vec2{10, 10})
	if ui.Style().Padding.X != 10 {
		t.Errorf("First push: got %d, want 10", ui.Style().Padding.X)
	}

	// Push second override
	ui.PushPadding(types.Vec2{0, 0})
	if ui.Style().Padding.X != 0 {
		t.Errorf("Second push: got %d, want 0", ui.Style().Padding.X)
	}

	// Pop second
	ui.PopPadding()
	if ui.Style().Padding.X != 10 {
		t.Errorf("After first pop: got %d, want 10", ui.Style().Padding.X)
	}

	// Pop first
	ui.PopPadding()
	if ui.Style().Padding != original {
		t.Errorf("After second pop: got %v, want %v", ui.Style().Padding, original)
	}
}
