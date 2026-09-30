package board

import (
	"context"
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

func TestFindSolutionsAllFindsEverySolution(t *testing.T) {
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

	solutions, err := FindSolutions(context.Background(), initial, pieces, Options{Mode: ModeAll})
	if err != nil {
		t.Fatalf("FindSolutions returned an error: %v", err)
	}
	// Every rotation/reflection variation of a single occupied cell is
	// geometrically identical, so a 1x1 piece placed on a single cell
	// contributes len(Variations) indistinguishable-looking solutions.
	// What matters is that both left-right orderings of the two distinct
	// pieces are discovered, and no other orderings.
	wantTotal := len(first.Variations) * len(second.Variations) * 2
	if len(solutions) != wantTotal {
		t.Fatalf("got %d solutions; want %d (every variation combination of both orderings)", len(solutions), wantTotal)
	}
	orderings := map[[2]string]bool{}
	for _, solution := range solutions {
		if len(solution.Placements) != 2 {
			t.Fatalf("solution has %d placements; want 2", len(solution.Placements))
		}
		orderings[[2]string{solution.Placements[0].Piece.Color, solution.Placements[1].Piece.Color}] = true
	}
	if len(orderings) != 2 {
		t.Fatalf("got %d distinct piece orderings; want 2 (first-then-second and second-then-first)", len(orderings))
	}
}

func TestFindSolutionsAllRespectsLimit(t *testing.T) {
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

	solutions, err := FindSolutions(context.Background(), initial, pieces, Options{Mode: ModeAll, Limit: 1})
	if err != nil {
		t.Fatalf("FindSolutions returned an error: %v", err)
	}
	if len(solutions) != 1 {
		t.Fatalf("got %d solutions; want 1 (limit)", len(solutions))
	}
}

func TestFindSolutionsRandomFindsAValidSolution(t *testing.T) {
	solutions, err := FindSolutions(context.Background(), FlatBoard, piece.Pieces, Options{Mode: ModeRandom})
	if err != nil {
		t.Fatalf("FindSolutions returned an error: %v", err)
	}
	if len(solutions) != 1 {
		t.Fatalf("got %d solutions; want 1", len(solutions))
	}
	if len(solutions[0].Placements) != len(piece.Pieces) {
		t.Fatalf("got %d placements; want %d", len(solutions[0].Placements), len(piece.Pieces))
	}
}

func TestFindSolutionsRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	solutions, err := FindSolutions(ctx, FlatBoard, piece.Pieces, Options{Mode: ModeAll})
	if err == nil {
		t.Fatal("FindSolutions did not report the cancellation")
	}
	if len(solutions) != 0 {
		t.Fatalf("got %d solutions; want 0 for an already-cancelled context", len(solutions))
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
