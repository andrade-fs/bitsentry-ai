package tui

import (
	"fmt"
	"strings"

	"bitsentry-ai/internal/app"
	tea "github.com/charmbracelet/bubbletea"
)

func Run(a *app.App, dryRun bool) error {
	p := tea.NewProgram(newModel(a, dryRun))
	if _, err := p.Run(); err != nil {
		if strings.Contains(err.Error(), "could not open a new TTY") {
			return fmt.Errorf("this command needs an interactive terminal (TTY) to run the TUI")
		}
		return fmt.Errorf("start tui: %w", err)
	}
	return nil
}
