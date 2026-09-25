package board

import (
	"strings"
	"testing"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

func TestFindSolutionFillsBoard(t *testing.T) {
	first, err := piece.NewPiece("first", [][]cell.CellType{{cell.Complete}})
	if err != nil {
		t.Fatalf("could not create first piece: %v", err)
	}
	second, err := piece.NewPiece("second", [][]cell.CellType{{cell.Complete}})
	if err != nil {
		t.Fatalf("could not create second piece: %v", err)
	}
	pieces := []piece.Piece{first, second}
	initial := Board{cells: [][]cell.CellType{{cell.Empty, cell.Empty}}}

	solution, err := FindSolution(initial, pieces)
	if err != nil {
		t.Fatalf("FindSolution returned an error: %v", err)
	}
	if solution == nil {
		t.Fatal("FindSolution did not find a solution")
	}
	if len(solution.Placements) != len(pieces) {
		t.Fatalf("got %d placements; want %d", len(solution.Placements), len(pieces))
	}
	if len(solution.IntermediateBoards) != len(pieces) {
		t.Fatalf("got %d intermediate boards; want %d", len(solution.IntermediateBoards), len(pieces))
	}
	for column, value := range solution.Board.cells[0] {
		if value != cell.Complete {
			t.Errorf("board cell (0, %d) = %d; want %d", column, value, cell.Complete)
		}
	}
	if solution.IntermediateBoards[0].cells[0][0] != cell.Complete ||
		solution.IntermediateBoards[0].cells[0][1] != cell.Empty {
		t.Errorf("first intermediate board = %v; want first cell filled and second empty", solution.IntermediateBoards[0].cells)
	}
	if solution.IntermediateBoards[1].cells[0][0] != cell.Complete ||
		solution.IntermediateBoards[1].cells[0][1] != cell.Complete {
		t.Errorf("second intermediate board = %v; want both cells filled", solution.IntermediateBoards[1].cells)
	}
}

func TestFindSolutionReportsContributionMismatch(t *testing.T) {
	single, err := piece.NewPiece("single", [][]cell.CellType{{cell.Complete}})
	if err != nil {
		t.Fatalf("could not create piece: %v", err)
	}
	pieces := []piece.Piece{single}
	initial := Board{cells: [][]cell.CellType{{cell.Empty, cell.Empty}}}

	_, err = FindSolution(initial, pieces)
	if err == nil || !strings.Contains(err.Error(), "pieces can contribute between 5 and 5, but the board requires 10") {
		t.Fatalf("got error %v; want contribution mismatch", err)
	}
}

func TestFindSolutionSolvesConfiguredPuzzle(t *testing.T) {
	solution, err := FindSolution(FlatBoard, piece.Pieces)
	if err != nil {
		t.Fatalf("FindSolution returned an error: %v", err)
	}
	if solution == nil {
		t.Fatal("FindSolution did not solve the configured puzzle")
	}
	if len(solution.Placements) != len(piece.Pieces) {
		t.Fatalf("got %d placements; want %d", len(solution.Placements), len(piece.Pieces))
	}
	if len(solution.IntermediateBoards) != len(solution.Placements) {
		t.Fatalf("got %d intermediate boards; want %d", len(solution.IntermediateBoards), len(solution.Placements))
	}
	for row, cells := range solution.Board.cells {
		for column, value := range cells {
			if value != cell.Complete {
				t.Errorf("board cell (%d, %d) = %d; want %d", row, column, value, cell.Complete)
			}
		}
	}
}
