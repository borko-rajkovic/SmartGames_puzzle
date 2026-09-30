package tui

import "github.com/borko-rajkovic/smart_games_puzzle/app/board"

// solveResultMsg carries the outcome of an asynchronous solve.
type solveResultMsg struct {
	solutions []*board.Solution
	err       error
}

// animTickMsg advances the placement animation by one step.
type animTickMsg struct{}
