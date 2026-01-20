// Package types provides shared types for the microui-go library.
//
// This package exists to avoid import cycles between the core microui package
// and the render package. All shared geometry and color types are defined here.
//
// # Types
//
//   - Vec2: 2D vector/point
//   - Rect: Rectangle with position and size
//   - RGBA: Color in RGBA format (0-255)
//   - Font: Interface for text measurement
//   - ThemeColors: Predefined color themes
//
// # Usage
//
//	import "github.com/user/microui-go/types"
//
//	rect := types.Rect{X: 10, Y: 10, W: 200, H: 100}
//	color := types.RGBA{R: 255, G: 128, B: 64, A: 255}
package types
