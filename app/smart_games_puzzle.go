package app

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/borko-rajkovic/smart_games_puzzle/app/board"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

func Main() {
	name, initial, err := chooseBoard(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	solveAndPrint(name, initial)
}

func chooseBoard(input io.Reader, output io.Writer) (string, board.Board, error) {
	scanner := bufio.NewScanner(input)
	for {
		fmt.Fprintln(output, "Choose a board to solve:")
		fmt.Fprintln(output, "1. Flat board")
		fmt.Fprintln(output, "2. Heart-shaped board")
		fmt.Fprint(output, "Enter 1 or 2: ")

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", board.Board{}, fmt.Errorf("read board selection: %w", err)
			}
			return "", board.Board{}, fmt.Errorf("no board selection provided")
		}
		fmt.Fprintln(output)

		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			return "Flat board", board.FlatBoard, nil
		case "2":
			return "Heart-shaped board", board.HeartBoard, nil
		default:
			fmt.Fprintln(output, "Invalid selection. Please enter 1 or 2.")
		}
	}
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
