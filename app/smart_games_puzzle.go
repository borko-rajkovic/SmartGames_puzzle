package app

import (
	"fmt"
	"log"

	"github.com/borko-rajkovic/smart_games_puzzle/app/board"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

func Main() {
	solution, err := board.FindSolution(board.FlatBoard, piece.Pieces)
	if err != nil {
		log.Fatal(err)
	}

	if solution == nil {
		fmt.Println("No solution found.")
		return
	}

	fmt.Printf("Solution found using %d pieces:\n", len(solution.Placements))
	for _, placement := range solution.Placements {
		fmt.Printf(
			"%s at row %d, column %d (variation %d)\n",
			placement.Piece.Color,
			placement.Row,
			placement.Column,
			placement.VariationIndex+1,
		)
		placement.Piece.Variations[placement.VariationIndex].Print()
	}
	solution.Board.Print()
}
