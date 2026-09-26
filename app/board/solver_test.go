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
	if err == nil || !strings.Contains(err.Error(), "board requires between 10 and 10, but pieces can contribute between 5 and 5") {
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

func TestFindSolutionSolvesHeartBoard(t *testing.T) {
	squareSlots, triangleSlots := 0, 0
	for _, row := range HeartBoard.cells {
		for _, value := range row {
			switch {
			case value == cell.Empty:
				squareSlots++
			case isTriangleTarget(value):
				triangleSlots++
			}
		}
	}
	if squareSlots != 28 || triangleSlots != 4 {
		t.Fatalf("heart board has %d square slots and %d triangle slots; want 28 and 4", squareSlots, triangleSlots)
	}
	displayedSlots := make(map[int]bool, squareSlots+triangleSlots)
	for _, row := range HeartBoard.displayRows {
		for _, index := range row {
			if displayedSlots[index] {
				t.Fatalf("heart display repeats board cell %d", index)
			}
			displayedSlots[index] = true
		}
	}
	for row, cells := range HeartBoard.cells {
		for column, value := range cells {
			index := row*len(cells) + column
			if (value == cell.Blocked) == displayedSlots[index] {
				t.Errorf("heart display membership does not match board cell (%d, %d)", row, column)
			}
		}
	}

	solution, err := FindSolution(HeartBoard, piece.Pieces)
	if err != nil {
		t.Fatalf("FindSolution returned an error: %v", err)
	}
	if solution == nil {
		t.Fatal("FindSolution did not solve the heart board")
	}
	if len(solution.Placements) != len(piece.Pieces) {
		t.Fatalf("got %d placements; want %d", len(solution.Placements), len(piece.Pieces))
	}
	if len(solution.IntermediateBoards) != len(solution.Placements) {
		t.Fatalf("got %d intermediate boards; want %d", len(solution.IntermediateBoards), len(solution.Placements))
	}
	for row, cells := range solution.Board.cells {
		for column, value := range cells {
			if HeartBoard.cells[row][column] == cell.Blocked {
				if value != cell.Blocked {
					t.Errorf("blocked board cell (%d, %d) = %d; want Blocked", row, column, value)
				}
				continue
			}
			if isTriangleTarget(HeartBoard.cells[row][column]) {
				if value < cell.DownRight || value > cell.TopLeft {
					t.Errorf("triangle board cell (%d, %d) = %d; want a triangle contribution", row, column, value)
				}
				continue
			}
			if value != cell.Complete {
				t.Errorf("playable board cell (%d, %d) = %d; want %d", row, column, value, cell.Complete)
			}
		}
	}
}
