// Package bubbletea provides a TUI renderer for microui-go using Bubble Tea v2.
//
// This renderer translates microui draw commands to terminal cells,
// supporting the "Ferocious Renderer" cell-based dirty tracking for
// efficient terminal updates.
//
// Usage:
//
//	renderer := bubbletea.NewRenderer(width, height)
//	ui := microui.New(microui.Config{
//	    Style: microui.Style{
//	        Font: &bubbletea.MonospaceFont{},
//	        // ... other style settings
//	    },
//	})
//
//	// In your Bubble Tea View():
//	ui.BeginFrame()
//	// ... build UI ...
//	ui.EndFrame()
//	ui.Render(renderer)
//	return tea.NewView(renderer) // renderer implements tea.Layer
package bubbletea
