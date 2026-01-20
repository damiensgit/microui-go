package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/user/microui-go/render/bubbletea"
)

func main() {
	// Color profile flag for testing on different terminals
	colors := flag.String("colors", "", "Force color mode: 16, 256, or true")
	flag.Parse()

	// Determine color mode for renderer (affects shadow style)
	var colorMode bubbletea.ColorMode
	var opts []tea.ProgramOption

	switch *colors {
	case "16":
		fmt.Println("Forcing 16-color mode (ANSI)")
		colorMode = bubbletea.Color16
		opts = append(opts, tea.WithColorProfile(colorprofile.ANSI))
	case "256":
		fmt.Println("Forcing 256-color mode")
		colorMode = bubbletea.Color256
		opts = append(opts, tea.WithColorProfile(colorprofile.ANSI256))
	case "true":
		fmt.Println("Forcing true color mode (24-bit)")
		colorMode = bubbletea.ColorTrueColor
		opts = append(opts, tea.WithColorProfile(colorprofile.TrueColor))
	default:
		// Auto-detect using same detection as Bubble Tea
		detected := colorprofile.Detect(os.Stdout, os.Environ())
		switch detected {
		case colorprofile.TrueColor:
			colorMode = bubbletea.ColorTrueColor
		case colorprofile.ANSI256:
			colorMode = bubbletea.Color256
		case colorprofile.ANSI:
			colorMode = bubbletea.Color16
		default:
			colorMode = bubbletea.ColorTrueColor // Fallback
		}
	}

	m := NewModel(colorMode)

	p := tea.NewProgram(m, opts...)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
