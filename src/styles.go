package main

import "github.com/charmbracelet/lipgloss"

var (
	CorrectTextStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	WrongTextStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	OvertypedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ba2222"))
	UnreachedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	CursorTextStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f5e614"))
)
