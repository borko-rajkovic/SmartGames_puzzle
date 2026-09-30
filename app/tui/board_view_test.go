package tui

import (
	"strings"
	"testing"

	"github.com/borko-rajkovic/smart_games_puzzle/app/board"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

func TestRenderBoardFlat(t *testing.T) {
	solution, err := board.FindSolution(board.FlatBoard, piece.Pieces)
	if err != nil || solution == nil {
		t.Fatalf("could not solve flat board: %v", err)
	}

	before := RenderBoard(board.FlatBoard, solution.Placements, 0)
	if !strings.Contains(before, "·") {
		t.Errorf("expected unplaced cells to render as ·, got:\n%s", before)
	}
	if strings.Contains(before, "■") {
		t.Errorf("expected no filled cells before any placement, got:\n%s", before)
	}

	afterFirst := RenderBoard(solution.IntermediateBoards[0], solution.Placements, 1)
	if !strings.Contains(afterFirst, "■") {
		t.Errorf("expected a filled cell after the first placement, got:\n%s", afterFirst)
	}

	final := RenderBoard(solution.Board, solution.Placements, len(solution.Placements))
	if strings.Contains(final, "·") {
		t.Errorf("expected a fully solved board to have no empty cells left, got:\n%s", final)
	}
	if !strings.Contains(final, "|") {
		t.Errorf("expected the flat board's rectangular border, got:\n%s", final)
	}
}

func TestRenderBoardHeartUsesAngledLayoutAndSlots(t *testing.T) {
	// The heart's blocked corner cells are cut out of DisplayRows entirely
	// (that's how the rectangular grid becomes heart-shaped), so they
	// never appear in the angled rendering at all - unlike the flat
	// board's rectangular layout, which renders every cell.
	unsolved := RenderBoard(board.HeartBoard, nil, 0)
	if strings.Contains(unsolved, "|") {
		t.Errorf("expected the angled heart layout, not the rectangular border, got:\n%s", unsolved)
	}
	for _, slotGlyph := range []string{"▽", "◁", "▷"} {
		if !strings.Contains(unsolved, slotGlyph) {
			t.Errorf("expected unmet triangle slot %q to be shown, got:\n%s", slotGlyph, unsolved)
		}
	}

	solution, err := board.FindSolution(board.HeartBoard, piece.Pieces)
	if err != nil || solution == nil {
		t.Fatalf("could not solve heart board: %v", err)
	}
	final := RenderBoard(solution.Board, solution.Placements, len(solution.Placements))
	if strings.Contains(final, "·") || strings.Contains(final, "▽") || strings.Contains(final, "◁") || strings.Contains(final, "▷") {
		t.Errorf("expected a fully solved heart board to have no unmet cells left, got:\n%s", final)
	}
}

func TestLegendListsEveryPiece(t *testing.T) {
	legend := Legend(piece.Pieces)
	for _, p := range piece.Pieces {
		if !strings.Contains(legend, p.Color) {
			t.Errorf("legend is missing piece color %q:\n%s", p.Color, legend)
		}
	}
}
