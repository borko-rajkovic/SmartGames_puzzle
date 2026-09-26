package app

import (
	"fmt"
	"log"

	"github.com/borko-rajkovic/smart_games_puzzle/app/board"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

func Main() {
	solveAndPrint("Heart-shaped board", board.HeartBoard)
}

func solveAndPrint(name string, initial board.Board) {
	solution, err := board.FindSolution(initial, piece.Pieces)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s:\n", name)
	if solution == nil {
		fmt.Println("No solution found.")
		return
	}

	fmt.Printf("Solution found using %d pieces:\n", len(solution.Placements))
	fmt.Println("Initial board:")
	initial.Print()
	for index, placement := range solution.Placements {
		fmt.Printf(
			"After placement %d/%d: %s at row %d, column %d (variation %d)\n",
			index+1,
			len(solution.Placements),
			placement.Piece.Color,
			placement.Row,
			placement.Column,
			placement.VariationIndex+1,
		)
		placement.Piece.Variations[placement.VariationIndex].Print()
		solution.IntermediateBoards[index].Print()
	}
}
