// Package snap provides optional window snap-to-edge functionality for microui.
//
// This is an optional extension. Import it only if you need window snapping.
//
// Dynamic screen size (recommended for resizable windows):
//
//	ui := microui.New(microui.Config{
//	    OnWindowDrag: snap.Handler(snap.Config{
//	        Threshold:  20,
//	        ScreenSize: ebiten.WindowSize, // Called each drag to get current size
//	    }),
//	})
//
// Static screen size (for fixed-size applications):
//
//	ui := microui.New(microui.Config{
//	    OnWindowDrag: snap.StaticHandler(snap.StaticConfig{
//	        Threshold:    20,
//	        ScreenWidth:  800,
//	        ScreenHeight: 600,
//	    }),
//	})
//
// Mark windows that should snap:
//
//	ui.BeginWindowOpt("Tools", rect, snap.OptSnapToEdge|snap.OptSnapTarget)
package snap

import (
	microui "github.com/user/microui-go"
)

// Option flags for snap behavior.
// These use high bits (1 << 20+) to avoid collision with core microui flags.
const (
	OptSnapToEdge = 1 << 20 // Window snaps to screen edges and other windows when dragging
	OptSnapTarget = 1 << 21 // Window can be snapped TO by other windows
)

// Config configures snap behavior.
type Config struct {
	Threshold  int               // Snap distance in pixels (default 20)
	ScreenSize func() (w, h int) // Returns current screen dimensions (for resize support)
}

// StaticConfig is a convenience for static screen dimensions.
// For resizable windows, use Config with ScreenSize function instead.
type StaticConfig struct {
	Threshold    int
	ScreenWidth  int
	ScreenHeight int
}

// Handler returns a window drag handler that implements snap-to-edge.
// Pass this to microui.Config.OnWindowDrag.
// Hold Shift while dragging to temporarily disable snapping.
func Handler(cfg Config) func(*microui.UI, *microui.Container) {
	if cfg.Threshold == 0 {
		cfg.Threshold = 20
	}
	return func(ui *microui.UI, cnt *microui.Container) {
		// Shift bypasses snapping for free placement
		if ui.IsKeyDown(microui.KeyShift) {
			return
		}
		var screenW, screenH int
		if cfg.ScreenSize != nil {
			screenW, screenH = cfg.ScreenSize()
		}
		applySnapToEdge(ui, cnt, cfg.Threshold, screenW, screenH)
	}
}

// StaticHandler is a convenience for static screen dimensions.
// For resizable windows, use Handler with Config.ScreenSize instead.
func StaticHandler(cfg StaticConfig) func(*microui.UI, *microui.Container) {
	return Handler(Config{
		Threshold:  cfg.Threshold,
		ScreenSize: func() (int, int) { return cfg.ScreenWidth, cfg.ScreenHeight },
	})
}

// applySnapToEdge snaps a container to screen edges and other snap-target windows.
func applySnapToEdge(ui *microui.UI, cnt *microui.Container, threshold, screenW, screenH int) {
	// Only snap windows that have OptSnapToEdge flag
	if cnt.Opt()&OptSnapToEdge == 0 {
		return
	}
	snappedX := false
	snappedY := false

	rect := cnt.Rect()

	// Snap to other windows marked as snap targets
	ui.EachContainer(func(other *microui.Container) bool {
		if other == cnt || !other.Open() {
			return true // continue
		}

		// Only snap to windows marked as snap targets
		if other.Opt()&OptSnapTarget == 0 {
			return true // continue
		}

		otherRect := other.Rect()

		// Check if windows have vertical overlap (for horizontal snapping)
		vOverlap := rect.Y < otherRect.Y+otherRect.H && rect.Y+rect.H > otherRect.Y
		// Check if windows have horizontal overlap (for vertical snapping)
		hOverlap := rect.X < otherRect.X+otherRect.W && rect.X+rect.W > otherRect.X

		// Snap our left edge to their right edge (dock to right side)
		if !snappedX && vOverlap && abs(rect.X-(otherRect.X+otherRect.W)) < threshold {
			rect.X = otherRect.X + otherRect.W
			snappedX = true
		}

		// Snap our right edge to their left edge (dock to left side)
		if !snappedX && vOverlap && abs((rect.X+rect.W)-otherRect.X) < threshold {
			rect.X = otherRect.X - rect.W
			snappedX = true
		}

		// Snap our top edge to their bottom edge (dock below)
		if !snappedY && hOverlap && abs(rect.Y-(otherRect.Y+otherRect.H)) < threshold {
			rect.Y = otherRect.Y + otherRect.H
			snappedY = true
		}

		// Snap our bottom edge to their top edge (dock above)
		if !snappedY && hOverlap && abs((rect.Y+rect.H)-otherRect.Y) < threshold {
			rect.Y = otherRect.Y - rect.H
			snappedY = true
		}

		return true // continue iteration
	})

	// Snap to screen edges (if screen size is set and not already snapped)
	if screenW == 0 || screenH == 0 {
		cnt.SetRect(rect)
		return
	}

	// Snap to left edge
	if !snappedX && rect.X < threshold {
		rect.X = 0
	}

	// Snap to top edge
	if !snappedY && rect.Y < threshold {
		rect.Y = 0
	}

	// Snap to right edge
	if !snappedX {
		rightEdge := screenW - rect.W
		if rect.X > rightEdge-threshold && rect.X < rightEdge+threshold {
			rect.X = rightEdge
		}
	}

	// Snap to bottom edge
	if !snappedY {
		bottomEdge := screenH - rect.H
		if rect.Y > bottomEdge-threshold && rect.Y < bottomEdge+threshold {
			rect.Y = bottomEdge
		}
	}

	cnt.SetRect(rect)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
