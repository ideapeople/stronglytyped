package main

import "github.com/charmbracelet/lipgloss"

var (
	CorrectTextStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	WrongTextStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	UnreachedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
)
