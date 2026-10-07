package board

import (
	"fmt"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

// buildBoardState converts the initial board cell grid into the flat value
// and target arrays the solver operates on.
//
// values[i] — current fill level of cell i (starts at what the board provides).
// targets[i] — what cell i must reach (Complete, Blocked, or a triangle slot).
//
// Also returns the total contribution range [boardMin, boardMax] that pieces
// must provide in aggregate to solve the board.
func buildBoardState(cells [][]cell.CellType, rows, columns int) (
	values, targets []cell.CellType, boardMin, boardMax int, err error,
) {
	values = make([]cell.CellType, rows*columns)
	targets = make([]cell.CellType, rows*columns)

	for row := range cells {
		for column, value := range cells[row] {
			if value < cell.Empty || value > cell.TriangleRightSlot {
				return nil, nil, 0, 0, fmt.Errorf(
					"invalid board cell value %d at row %d, column %d", value, row, column)
			}
			index := row*columns + column
			switch value {
			case cell.Blocked:
				// Blocked cells are permanently off-limits.
				values[index] = cell.Blocked
				targets[index] = cell.Blocked
			case cell.TriangleUpSlot, cell.TriangleDownSlot,
				cell.TriangleLeftSlot, cell.TriangleRightSlot:
				// Triangle slot: the target encodes which orientation is required;
				// the value starts empty until a compatible piece fills it.
				targets[index] = value
			default:
				// Regular cell: already carries some fill from the board definition;
				// target is always Complete (5). The piece must supply the remainder.
				values[index] = value
				targets[index] = cell.Complete
				boardMin += int(cell.Complete - value)
				boardMax += int(cell.Complete - value)
			}
		}
	}

	// Triangle slots accept contributions in the range [DownRight, TopLeft],
	// so add that range to the board's total requirement.
	for index, target := range targets {
		if isTriangleTarget(target) && values[index] == cell.Empty {
			minimum, maximum := triangleContributionRange()
			boardMin += minimum
			boardMax += maximum
		}
	}
	return values, targets, boardMin, boardMax, nil
}

// buildCandidates generates every valid (piece × variation × board position)
// candidate and groups them by the board cell each placement touches.
//
// Also returns per-piece [min, max] contribution bounds (across all variations)
// used by the feasibility pruner, and the global [minPossible, maxPossible]
// sum across all pieces.
func buildCandidates(pieces []piece.Piece, rows, columns int) (
	candidatesByCell [][]candidatePlacement,
	minTotals, maxTotals []int,
	minPossible, maxPossible int,
	err error,
) {
	candidatesByCell = make([][]candidatePlacement, rows*columns)
	minTotals = make([]int, len(pieces))
	maxTotals = make([]int, len(pieces))

	for pieceIndex, currentPiece := range pieces {
		if len(currentPiece.Variations) == 0 {
			return nil, nil, nil, 0, 0,
				fmt.Errorf("piece %q has no variations", currentPiece.Color)
		}

		minPiece, maxPiece := 0, 0
		for variationIndex, variation := range currentPiece.Variations {
			shape := variation.Cells()
			height, width, shapeCells, vErr := variationDimensions(shape)
			if vErr != nil {
				return nil, nil, nil, 0, 0,
					fmt.Errorf("piece %q variation %d: %w", currentPiece.Color, variationIndex, vErr)
			}

			// Total contribution this variation adds to the board.
			total := 0
			for _, sc := range shapeCells {
				total += int(sc.contribution)
			}
			if variationIndex == 0 || total < minPiece {
				minPiece = total
			}
			if variationIndex == 0 || total > maxPiece {
				maxPiece = total
			}

			// Register this variation at every valid board position.
			for top := 0; top+height <= rows; top++ {
				for left := 0; left+width <= columns; left++ {
					cp := candidatePlacement{
						placement: Placement{
							Piece:          currentPiece,
							VariationIndex: variationIndex,
							Row:            top,
							Column:         left,
						},
						pieceIndex: pieceIndex,
						cells:      make([]placementCell, len(shapeCells)),
					}
					for i, sc := range shapeCells {
						row := top + sc.index/len(shape[0])
						col := left + sc.index%len(shape[0])
						idx := row*columns + col
						cp.cells[i] = placementCell{index: idx, contribution: sc.contribution}
						// Index this candidate under every cell it touches so the
						// MRV search can quickly look up candidates for any cell.
						candidatesByCell[idx] = append(candidatesByCell[idx], cp)
					}
				}
			}
		}
		minTotals[pieceIndex] = minPiece
		maxTotals[pieceIndex] = maxPiece
		minPossible += minPiece
		maxPossible += maxPiece
	}
	return candidatesByCell, minTotals, maxTotals, minPossible, maxPossible, nil
}

// feasible returns false when it can already be proven that no completion is
// possible: the total remaining need of unfilled cells falls outside the range
// of what the remaining unused pieces can contribute.
func feasible(
	values, targets []cell.CellType,
	usedPieces []bool,
	minTotals, maxTotals []int,
) bool {
	// Sum what the board still needs from pieces.
	boardRemMin, boardRemMax := 0, 0
	for index, target := range targets {
		switch {
		case target == cell.Complete:
			rem := int(cell.Complete - values[index])
			boardRemMin += rem
			boardRemMax += rem
		case isTriangleTarget(target) && values[index] == cell.Empty:
			min, max := triangleContributionRange()
			boardRemMin += min
			boardRemMax += max
		}
	}
	// Sum what the unused pieces can still provide.
	pieceRemMin, pieceRemMax := 0, 0
	for i, used := range usedPieces {
		if !used {
			pieceRemMin += minTotals[i]
			pieceRemMax += maxTotals[i]
		}
	}
	return boardRemMin <= pieceRemMax && boardRemMax >= pieceRemMin
}

// findMostConstrained scans all unfilled cells and returns the one with the
// fewest compatible candidate placements (MRV heuristic), together with
// its candidate list.
//
// Return values:
//   - (-1, non-nil empty): all cells are filled — caller should check for a solution.
//   - (>=0, nil):           some cell has zero candidates — dead end, prune.
//   - (>=0, non-nil):       normal case — caller should try the returned candidates.
func findMostConstrained(
	values, targets []cell.CellType,
	candidatesByCell [][]candidatePlacement,
	usedPieces []bool,
) (int, []candidatePlacement) {
	bestIndex, bestCount := -1, 0

	// Pass 1: find the unfilled cell with the fewest compatible candidates.
	// Only count — no slice allocation for non-winning cells.
	for index, value := range values {
		if isFilled(value, targets[index]) {
			continue
		}
		count := 0
		for _, candidate := range candidatesByCell[index] {
			if !usedPieces[candidate.pieceIndex] && canPlace(values, targets, candidate) {
				count++
			}
		}
		if count == 0 {
			return index, nil // dead end
		}
		if bestIndex == -1 || count < bestCount {
			bestIndex, bestCount = index, count
		}
	}
	if bestIndex == -1 {
		return -1, nil // all cells filled
	}
	// Pass 2: collect candidates for the winning cell only.
	best := make([]candidatePlacement, 0, bestCount)
	for _, candidate := range candidatesByCell[bestIndex] {
		if !usedPieces[candidate.pieceIndex] && canPlace(values, targets, candidate) {
			best = append(best, candidate)
		}
	}
	return bestIndex, best
}

// applyPlacement marks the piece as used and adds its per-cell contributions.
func applyPlacement(values []cell.CellType, usedPieces []bool, c candidatePlacement) {
	usedPieces[c.pieceIndex] = true
	for _, sc := range c.cells {
		values[sc.index] += sc.contribution
	}
}

// undoPlacement reverses applyPlacement.
func undoPlacement(values []cell.CellType, usedPieces []bool, c candidatePlacement) {
	for _, sc := range c.cells {
		values[sc.index] -= sc.contribution
	}
	usedPieces[c.pieceIndex] = false
}

// isFilled reports whether cell index is already at its required fill level.
func isFilled(value, target cell.CellType) bool {
	if isTriangleTarget(target) {
		return triangleFits(target, value)
	}
	return value == cell.Complete || target == cell.Blocked
}

// isTriangleTarget reports whether a target value encodes a triangle slot.
func isTriangleTarget(target cell.CellType) bool {
	return target >= cell.TriangleUpSlot && target <= cell.TriangleRightSlot
}

// triangleContributionRange returns the [min, max] contribution a single
// triangle piece cell can provide. Used for feasibility bounds on slots.
func triangleContributionRange() (int, int) {
	return int(cell.DownRight), int(cell.TopLeft)
}

// triangleFits reports whether the given contribution is geometrically
// compatible with the given triangle slot orientation.
//
// Each slot is a half-square triangle; only the two quarter-triangles that
// lie within that half can fill it.
func triangleFits(target, contribution cell.CellType) bool {
	switch target {
	case cell.TriangleUpSlot: // upper half — accepts top-corner pieces
		return contribution == cell.TopRight || contribution == cell.TopLeft
	case cell.TriangleDownSlot: // lower half — accepts bottom-corner pieces
		return contribution == cell.DownRight || contribution == cell.DownLeft
	case cell.TriangleLeftSlot: // right half (slot at left boundary) — right-side pieces
		return contribution == cell.TopRight || contribution == cell.DownRight
	case cell.TriangleRightSlot: // left half (slot at right boundary) — left-side pieces
		return contribution == cell.TopLeft || contribution == cell.DownLeft
	}
	return false
}

// canPlace checks whether a candidate placement is compatible with the
// current board state: no blocked cells, no overfill of regular cells,
// and correct triangle orientation for triangle slots.
func canPlace(values, targets []cell.CellType, candidate candidatePlacement) bool {
	for _, sc := range candidate.cells {
		index := sc.index
		if targets[index] == cell.Blocked {
			return false
		}
		if isTriangleTarget(targets[index]) {
			// Triangle slot: must be empty, piece must not be Complete,
			// and the contribution must match the slot's orientation.
			if values[index] != cell.Empty || sc.contribution == cell.Complete {
				return false
			}
			if !triangleFits(targets[index], sc.contribution) {
				return false
			}
			continue
		}
		// Regular cell: adding this contribution must not exceed the target.
		if values[index]+sc.contribution > targets[index] {
			return false
		}
	}
	return true
}

// boardFromValues builds a Board snapshot from the current flat value/target
// arrays, copying the display layout from the initial board.
func boardFromValues(
	values, targets []cell.CellType,
	rows, columns int,
	displayRows [][]int,
	displayIndent []int,
) Board {
	result := Board{
		cells:         make([][]cell.CellType, rows),
		targets:       make([][]cell.CellType, rows),
		displayRows:   make([][]int, len(displayRows)),
		displayIndent: append([]int(nil), displayIndent...),
	}
	for row := range result.cells {
		start, end := row*columns, (row+1)*columns
		result.cells[row] = append([]cell.CellType(nil), values[start:end]...)
		result.targets[row] = append([]cell.CellType(nil), targets[start:end]...)
	}
	for row := range displayRows {
		result.displayRows[row] = append([]int(nil), displayRows[row]...)
	}
	return result
}

// boardDimensions validates and returns the row and column count of a cell grid.
func boardDimensions(cells [][]cell.CellType) (int, int, error) {
	if len(cells) == 0 || len(cells[0]) == 0 {
		return 0, 0, fmt.Errorf("must have at least one row and column")
	}
	columns := len(cells[0])
	for row := range cells {
		if len(cells[row]) != columns {
			return 0, 0, fmt.Errorf("row %d has %d columns; expected %d", row, len(cells[row]), columns)
		}
	}
	return len(cells), columns, nil
}

// variationDimensions validates a piece variation's cell grid and returns its
// bounding box dimensions together with the list of occupied cells
// (index within the variation grid, contribution value).
func variationDimensions(cells [][]cell.CellType) (int, int, []placementCell, error) {
	if len(cells) == 0 || len(cells[0]) == 0 {
		return 0, 0, nil, fmt.Errorf("must have at least one row and column")
	}
	columns := len(cells[0])
	maxRow, maxColumn := -1, -1
	shapeCells := make([]placementCell, 0)
	for row := range cells {
		if len(cells[row]) != columns {
			return 0, 0, nil, fmt.Errorf("row %d has %d columns; expected %d", row, len(cells[row]), columns)
		}
		for column, value := range cells[row] {
			if value < cell.Empty || value > cell.Complete {
				return 0, 0, nil, fmt.Errorf("invalid cell value %d at row %d, column %d", value, row, column)
			}
			if value == cell.Empty {
				continue
			}
			shapeCells = append(shapeCells, placementCell{
				index:        row*columns + column,
				contribution: value,
			})
			if row > maxRow {
				maxRow = row
			}
			if column > maxColumn {
				maxColumn = column
			}
		}
	}
	if len(shapeCells) == 0 {
		return 0, 0, nil, fmt.Errorf("has no occupied cells")
	}
	return maxRow + 1, maxColumn + 1, shapeCells, nil
}
