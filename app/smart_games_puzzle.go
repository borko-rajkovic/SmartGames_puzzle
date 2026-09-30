package app

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/borko-rajkovic/smart_games_puzzle/app/tui"
)

func Main() {
	program := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
