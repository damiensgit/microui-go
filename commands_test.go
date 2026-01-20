package microui

import (
	"image/color"
	"testing"

	"github.com/user/microui-go/types"
)

func TestCommandBuffer_Reset(t *testing.T) {
	cb := &CommandBuffer{}
	cb.Init(10)

	cb.Push(Command{Kind: CmdRect})
	cb.Push(Command{Kind: CmdText})

	if len(cb.cmds) != 2 {
		t.Fatalf("Push() resulted in %d commands, want 2", len(cb.cmds))
	}

	cb.Reset()

	if len(cb.cmds) != 0 {
		t.Errorf("Reset() left %d commands, want 0", len(cb.cmds))
	}

	// Capacity should be preserved
	if cap(cb.cmds) != 10 {
		t.Errorf("Reset() changed capacity to %d, want 10", cap(cb.cmds))
	}
}

func TestCommandBuffer_Push(t *testing.T) {
	cb := &CommandBuffer{}
	cb.Init(4)

	for i := 0; i < 6; i++ {
		cb.Push(Command{Kind: CmdRect})
	}

	if len(cb.cmds) != 6 {
		t.Errorf("Push() resulted in %d commands, want 6", len(cb.cmds))
	}
}

func TestCommandBuffer_Each(t *testing.T) {
	cb := &CommandBuffer{}
	cb.Init(10)

	cb.Push(Command{Kind: CmdRect})
	cb.Push(Command{Kind: CmdText})
	cb.Push(Command{Kind: CmdClip})

	count := 0
	cb.Each(func(cmd Command) {
		count++
	})

	if count != 3 {
		t.Errorf("Each() called %d times, want 3", count)
	}
}

func TestCommandBuffer_Len(t *testing.T) {
	cb := &CommandBuffer{}
	cb.Init(10)

	if cb.Len() != 0 {
		t.Errorf("Len() = %d, want 0", cb.Len())
	}

	cb.Push(Command{Kind: CmdRect})

	if cb.Len() != 1 {
		t.Errorf("Len() = %d, want 1", cb.Len())
	}
}

func TestCommand_AllFields(t *testing.T) {
	cmd := Command{
		Kind:  CmdText,
		Rect:  types.Rect{X: 1, Y: 2, W: 3, H: 4},
		Pos:   types.Vec2{X: 10, Y: 20},
		Size:  types.Vec2{X: 30, Y: 40},
		Text:  "test",
		Color: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Icon:  42,
		Font:  nil,
	}

	if cmd.Kind != CmdText {
		t.Error("Kind field not set")
	}
	if cmd.Text != "test" {
		t.Error("Text field not set")
	}
}
