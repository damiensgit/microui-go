// Package types provides shared types for the microui-go library.
//
// This package exists to avoid import cycles between the core microui package
// and the render package. All shared geometry and color types are defined here.
//
// # Types
//
//   - Vec2: 2D vector/point
//   - Rect: Rectangle with position and size
//   - Font: Interface for text measurement
//   - ThemeColors: Predefined color themes (DarkTheme, LightTheme)
//
// # Usage
//
//	import "github.com/user/microui-go/types"
//
//	rect := types.Rect{X: 10, Y: 10, W: 200, H: 100}
//	theme := types.DarkTheme()
package types
