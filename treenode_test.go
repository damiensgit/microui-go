package microui

import (
	"testing"

	"github.com/user/microui-go/types"
)

func TestTreeNode_Basic(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)

	if ui.BeginTreeNode("Node 1") {
		ui.Label("Child 1")
		ui.Label("Child 2")
		ui.EndTreeNode()
	}

	if ui.BeginTreeNode("Node 2") {
		ui.Label("Child A")
		ui.EndTreeNode()
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestTreeNode_Nested(t *testing.T) {
	ui := New(Config{})
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 0)

	if ui.BeginTreeNode("Parent") {
		ui.Label("Direct child")

		if ui.BeginTreeNode("Nested") {
			ui.Label("Nested child")
			ui.EndTreeNode()
		}

		ui.EndTreeNode()
	}

	ui.EndWindow()
	ui.EndFrame()
}

func TestTreeNode_DefaultCollapsed(t *testing.T) {
	ui := New(Config{})

	// First frame - tree node should be collapsed by default
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 30)

	expanded := ui.BeginTreeNode("Node")
	if expanded {
		ui.EndTreeNode()
	}

	ui.EndWindow()
	ui.EndFrame()

	// Default should be collapsed (unlike Header which is expanded by default)
	if expanded {
		t.Error("TreeNode should be collapsed by default")
	}
}

func TestTreeNode_OptExpanded(t *testing.T) {
	ui := New(Config{})

	// Use OptExpanded to start expanded
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 30)

	expanded := ui.BeginTreeNodeEx("Node", OptExpanded)
	if expanded {
		ui.EndTreeNode()
	}

	ui.EndWindow()
	ui.EndFrame()

	// Should be expanded due to OptExpanded
	if !expanded {
		t.Error("TreeNode should be expanded when OptExpanded is set")
	}
}

func TestTreeNode_Toggle(t *testing.T) {
	ui := New(Config{})

	// First frame - tree node collapsed by default
	ui.BeginFrame()
	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 30)

	expanded1 := ui.BeginTreeNode("Node")
	if expanded1 {
		ui.EndTreeNode()
	}

	ui.EndWindow()
	ui.EndFrame()

	// Verify collapsed
	if expanded1 {
		t.Error("TreeNode should be collapsed initially")
	}

	// Click on tree node to toggle
	// TreeNode is at Y ~= 29 (after title 24 + padding 5), height ~= 20 (size.Y 10 + padding*2)
	// So click at Y=35 to be inside the control
	ui.MouseMove(50, 35)
	ui.MouseDown(50, 35, MouseLeft)
	ui.BeginFrame()

	ui.BeginWindow("Test", types.Rect{X: 0, Y: 0, W: 400, H: 300})
	ui.LayoutRow(1, []int{-1}, 30)

	expanded2 := ui.BeginTreeNode("Node")
	if expanded2 {
		ui.EndTreeNode()
	}

	ui.EndWindow()
	ui.EndFrame()

	// Should now be expanded after click
	if !expanded2 {
		t.Error("TreeNode should be expanded after click")
	}
}
