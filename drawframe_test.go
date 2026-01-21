package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestDrawFrame_CustomCallback(t *testing.T) {
	callCount := 0
	customDrawFrame := func(ui *UI, info FrameInfo) {
		callCount++
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Click") // Should trigger draw_frame
	ui.EndWindow()
	ui.EndFrame()

	if callCount == 0 {
		t.Error("Custom DrawFrame callback was never called")
	}
}

func TestDrawFrame_DefaultProducesCommands(t *testing.T) {
	ui := New(Config{})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Click")
	ui.EndWindow()
	ui.EndFrame()

	// Should have rect commands from default draw_frame
	hasRect := false
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdRect {
			hasRect = true
		}
	})

	if !hasRect {
		t.Error("Default DrawFrame should produce CmdRect commands")
	}
}

func TestDrawFrame_FrameKindProvided(t *testing.T) {
	frameKinds := make(map[FrameKind]bool)
	customDrawFrame := func(ui *UI, info FrameInfo) {
		frameKinds[info.Kind] = true
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Normal") // Button kind
	ui.EndWindow()
	ui.EndFrame()

	// Should have received button kind
	if !frameKinds[FrameButton] && !frameKinds[FrameWindow] {
		t.Error("DrawFrame should receive FrameKind values")
	}
}

func TestDrawFrame_FrameKindConstants(t *testing.T) {
	// Verify FrameKind constants are defined correctly
	expectedKinds := []struct {
		kind FrameKind
		name string
	}{
		{FrameWindow, "FrameWindow"},
		{FrameTitle, "FrameTitle"},
		{FramePanel, "FramePanel"},
		{FrameButton, "FrameButton"},
		{FrameInput, "FrameInput"},
		{FrameSliderThumb, "FrameSliderThumb"},
		{FrameScrollTrack, "FrameScrollTrack"},
		{FrameScrollThumb, "FrameScrollThumb"},
		{FrameHeader, "FrameHeader"},
	}

	// Check that each constant has a unique value
	seen := make(map[FrameKind]string)
	for _, tc := range expectedKinds {
		if existing, ok := seen[tc.kind]; ok {
			t.Errorf("FrameKind constant %s has same value as %s (%d)", tc.name, existing, tc.kind)
		}
		seen[tc.kind] = tc.name
	}

	// Check first constant is 0
	if FrameWindow != 0 {
		t.Errorf("FrameWindow should be 0, got %d", FrameWindow)
	}
}

func TestDrawFrame_FrameStateConstants(t *testing.T) {
	// Verify FrameState constants are defined correctly
	if StateNormal != 0 {
		t.Errorf("StateNormal should be 0, got %d", StateNormal)
	}
	if StateHover != 1 {
		t.Errorf("StateHover should be 1, got %d", StateHover)
	}
	if StateFocus != 2 {
		t.Errorf("StateFocus should be 2, got %d", StateFocus)
	}
}

func TestDrawFrame_GetColor(t *testing.T) {
	ui := New(Config{})

	// Test that GetColor returns appropriate colors for kind/state combinations
	tests := []struct {
		kind  FrameKind
		state FrameState
		name  string
	}{
		{FrameButton, StateNormal, "FrameButton Normal"},
		{FrameButton, StateHover, "FrameButton Hover"},
		{FrameButton, StateFocus, "FrameButton Focus"},
		{FrameInput, StateNormal, "FrameInput Normal"},
		{FrameInput, StateHover, "FrameInput Hover"},
		{FrameInput, StateFocus, "FrameInput Focus"},
		{FrameScrollTrack, StateNormal, "FrameScrollTrack Normal"},
		{FrameScrollThumb, StateNormal, "FrameScrollThumb Normal"},
		{FrameTitle, StateNormal, "FrameTitle Normal"},
		{FrameWindow, StateNormal, "FrameWindow Normal"},
		{FramePanel, StateNormal, "FramePanel Normal"},
		{FrameHeader, StateNormal, "FrameHeader Normal"},
		{FrameSliderThumb, StateNormal, "FrameSliderThumb Normal"},
	}

	for _, tt := range tests {
		c := ui.GetColor(tt.kind, tt.state)
		if c == nil {
			t.Errorf("GetColor(%s) returned nil", tt.name)
		}
	}
}

func TestDrawFrame_NilCallbackUsesDefault(t *testing.T) {
	ui := New(Config{
		DrawFrame: nil, // Explicitly nil
	})

	// Should not panic and should use default
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{-1}, 0)
	ui.Button("Click")
	ui.EndWindow()
	ui.EndFrame()

	// Should have rect commands from default draw_frame
	hasRect := false
	ui.commands.Each(func(cmd Command) {
		if cmd.Kind == CmdRect {
			hasRect = true
		}
	})

	if !hasRect {
		t.Error("Default DrawFrame (from nil config) should produce CmdRect commands")
	}
}

func TestDrawFrame_CalledWithCorrectRect(t *testing.T) {
	var capturedRect types.Rect
	customDrawFrame := func(ui *UI, info FrameInfo) {
		if info.Kind == FrameButton {
			capturedRect = info.Rect
		}
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Button("Click")
	ui.EndWindow()
	ui.EndFrame()

	// Button should have been rendered with non-zero dimensions
	if capturedRect.W == 0 || capturedRect.H == 0 {
		t.Errorf("DrawFrame should be called with valid rect, got %+v", capturedRect)
	}
}

func TestDrawFrame_PublicMethod(t *testing.T) {
	callCount := 0
	customDrawFrame := func(ui *UI, info FrameInfo) {
		callCount++
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	// Test that DrawFrame public method works
	testRect := types.Rect{X: 10, Y: 20, W: 100, H: 50}
	ui.DrawFrame(FrameInfo{Kind: FrameButton, State: StateNormal, Rect: testRect})

	if callCount != 1 {
		t.Errorf("DrawFrame method should call callback once, got %d calls", callCount)
	}
}

func TestDrawFrame_StateDetection(t *testing.T) {
	var capturedStates []FrameState
	customDrawFrame := func(ui *UI, info FrameInfo) {
		if info.Kind == FrameButton {
			capturedStates = append(capturedStates, info.State)
		}
	}

	ui := New(Config{
		DrawFrame: customDrawFrame,
	})

	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 200, H: 150})
	ui.LayoutRow(1, []int{100}, 30)
	ui.Button("Click")
	ui.EndWindow()
	ui.EndFrame()

	// Button should have StateNormal without hover/focus
	found := false
	for _, state := range capturedStates {
		if state == StateNormal {
			found = true
			break
		}
	}

	if !found {
		t.Error("DrawFrame should be called with StateNormal for unhovered button")
	}
}
