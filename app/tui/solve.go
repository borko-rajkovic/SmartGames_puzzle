package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/borko-rajkovic/smart_games_puzzle/app/board"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

// solveCmd runs board.FindSolutions in the background and reports the
// result as a solveResultMsg once it completes or ctx is cancelled.
func solveCmd(ctx context.Context, initial board.Board, mode board.Mode, limit int) tea.Cmd {
	return func() tea.Msg {
		solutions, err := board.FindSolutions(ctx, initial, piece.Pieces, board.Options{
			Mode:  mode,
			Limit: limit,
		})
		return solveResultMsg{solutions: solutions, err: err}
	}
}

func animTickCmd(speed time.Duration) tea.Cmd {
	return tea.Tick(speed, func(time.Time) tea.Msg {
		return animTickMsg{}
	})
}
