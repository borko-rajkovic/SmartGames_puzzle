package tui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

const (
	defaultSpeed = 120 * time.Millisecond
	minSpeed     = 30 * time.Millisecond
	maxSpeed     = 2000 * time.Millisecond

	// allSolutionsLimit caps how many solutions ModeAll will collect
	// before stopping on its own; the user can also cancel early with esc.
	allSolutionsLimit = 130
)

var (
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	statusErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	listTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	placeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("83"))

	backtrackStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	stepCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))
)
