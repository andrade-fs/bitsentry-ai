package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	title    lipgloss.Style
	section  lipgloss.Style
	selected lipgloss.Style
	muted    lipgloss.Style
	error    lipgloss.Style
	hint     lipgloss.Style
}

func newStyles() styles {
	return styles{
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		section:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")),
		selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		muted:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		error:    lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		hint:     lipgloss.NewStyle().Foreground(lipgloss.Color("14")),
	}
}
