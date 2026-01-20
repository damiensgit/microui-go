// Package ebiten provides an Ebiten v2 renderer for microui-go.
//
// This is the reference implementation showing how to implement
// the microui.Renderer interface.
//
// # Usage
//
//	import "github.com/user/microui-go/render/ebiten"
//
//	renderer := ebiten.NewRenderer()
//	renderer.SetTarget(screen) // *ebiten.Image
//	ui.Render(renderer)
package ebiten
