package board

import (
	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
)

type Board struct {
	cells         [][]cell.CellType
	targets       [][]cell.CellType
	displayRows   [][]int
	displayIndent []int
}

// Rows returns the number of rows in the board.
func (b Board) Rows() int {
	return len(b.cells)
}

// Columns returns the number of columns in the board.
func (b Board) Columns() int {
	if len(b.cells) == 0 {
		return 0
	}
	return len(b.cells[0])
}

// CellAt returns the current fill value of the cell at (row, column):
// Empty, a triangle half, Complete, or Blocked/a triangle slot for a
// predefined board that hasn't been solved yet.
func (b Board) CellAt(row, column int) cell.CellType {
	return b.cells[row][column]
}

// TargetAt returns what the cell at (row, column) must become. For a
// solved or in-progress board (produced by FindSolution/FindSolutions)
// this comes from the solver's target grid. For a predefined, unsolved
// board (no target grid yet), the cell's own value already encodes any
// Blocked/triangle-slot requirement, so that is returned instead.
func (b Board) TargetAt(row, column int) cell.CellType {
	if len(b.targets) == 0 {
		return b.cells[row][column]
	}
	return b.targets[row][column]
}

// DisplayRows returns the board's angled display layout (used by
// heart/diamond-shaped boards), or nil if the board uses the plain
// rectangular layout.
func (b Board) DisplayRows() [][]int {
	return b.displayRows
}

// DisplayIndent returns the per-row leading indent for the angled
// display layout returned by DisplayRows.
func (b Board) DisplayIndent() []int {
	return b.displayIndent
}

var FlatBoard = Board{
	cells: [][]cell.CellType{
		{cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
	},
}

var HeartBoard = Board{
	cells: [][]cell.CellType{
		{cell.Blocked, cell.Blocked, cell.TriangleDownSlot, cell.Empty, cell.Empty, cell.TriangleRightSlot},
		{cell.Blocked, cell.Blocked, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.TriangleDownSlot, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
		{cell.TriangleLeftSlot, cell.Empty, cell.Empty, cell.Empty, cell.Empty, cell.Empty},
	},
	displayRows: [][]int{
		{12, 2},
		{18, 13, 8, 3},
		{24, 19, 14, 9, 4},
		{30, 25, 20, 15, 10, 5},
		{31, 26, 21, 16, 11},
		{32, 27, 22, 17},
		{33, 28, 23},
		{34, 29},
		{35},
	},
	displayIndent: []int{3, 2, 1, 0, 1, 2, 3, 4, 5},
}
