package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	lastEvent string
	eventCount int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		m.lastEvent = fmt.Sprintf("Key: %s", msg.String())
		m.eventCount++

	case tea.MouseClickMsg:
		m.lastEvent = fmt.Sprintf("Click: x=%d y=%d btn=%v", msg.X, msg.Y, msg.Button)
		m.eventCount++

	case tea.MouseReleaseMsg:
		m.lastEvent = fmt.Sprintf("Release: x=%d y=%d btn=%v", msg.X, msg.Y, msg.Button)
		m.eventCount++

	case tea.MouseMotionMsg:
		m.lastEvent = fmt.Sprintf("Motion: x=%d y=%d", msg.X, msg.Y)
		m.eventCount++

	case tea.MouseWheelMsg:
		m.lastEvent = fmt.Sprintf("Wheel: x=%d y=%d btn=%v", msg.X, msg.Y, msg.Button)
		m.eventCount++

	case tea.WindowSizeMsg:
		m.lastEvent = fmt.Sprintf("Resize: %dx%d", msg.Width, msg.Height)
		m.eventCount++
	}

	return m, nil
}

func (m model) View() tea.View {
	content := fmt.Sprintf(`Mouse Test - Press q to quit

Event count: %d
Last event: %s

Try clicking, moving mouse, scrolling...
`, m.eventCount, m.lastEvent)

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
