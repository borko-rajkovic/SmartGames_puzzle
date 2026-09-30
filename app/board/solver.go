package board

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

// Mode selects how FindSolutions explores the search space.
type Mode int

const (
	// ModeFirst stops as soon as a single solution is found.
	ModeFirst Mode = iota
	// ModeRandom shuffles candidate order at each branch so different
	// runs tend to discover different valid solutions.
	ModeRandom
	// ModeAll keeps searching after a solution is found, collecting every
	// solution it discovers (up to Options.Limit, if set).
	ModeAll
)

// Options configures a FindSolutions search.
type Options struct {
	// Mode selects the search strategy. Zero value is ModeFirst.
	Mode Mode
	// Limit caps the number of solutions collected. 0 means unlimited
	// (ModeFirst always stops after the first solution regardless).
	Limit int
	// Rand supplies randomness for ModeRandom. If nil, a time-seeded
	// source is used.
	Rand *rand.Rand
}

type Placement struct {
	Piece          piece.Piece
	VariationIndex int
	Row            int
	Column         int
}

type Solution struct {
	Board              Board
	Placements         []Placement
	IntermediateBoards []Board
}

type placementCell struct {
	index        int
	contribution cell.CellType
}

type candidatePlacement struct {
	placement  Placement
	pieceIndex int
	cells      []placementCell
}

// FindSolution returns the first solution found for the given board and
// piece set, or (nil, nil) if none exists.
func FindSolution(initial Board, pieces []piece.Piece) (*Solution, error) {
	solutions, err := FindSolutions(context.Background(), initial, pieces, Options{Mode: ModeFirst})
	if err != nil {
		return nil, err
	}
	if len(solutions) == 0 {
		return nil, nil
	}
	return solutions[0], nil
}

// FindSolutions searches for solutions to the given board using the given
// piece set, according to opts.Mode:
//
//   - ModeFirst stops at the first solution found (Options.Limit is ignored).
//   - ModeRandom shuffles candidate order at each branch, so repeated calls
//     tend to surface different valid solutions; it stops at the first
//     solution found unless Options.Limit > 1.
//   - ModeAll keeps searching for every solution, up to Options.Limit
//     (0 = unlimited).
//
// The search can be cancelled early via ctx; solutions collected before
// cancellation are still returned.
func FindSolutions(ctx context.Context, initial Board, pieces []piece.Piece, opts Options) ([]*Solution, error) {
	rows, columns, err := boardDimensions(initial.cells)
	if err != nil {
		return nil, fmt.Errorf("invalid board: %w", err)
	}

	values := make([]cell.CellType, rows*columns)
	targets := make([]cell.CellType, rows*columns)
	boardMinimum, boardMaximum := 0, 0
	for row := range initial.cells {
		for column, value := range initial.cells[row] {
			if value < cell.Empty || value > cell.TriangleRightSlot {
				return nil, fmt.Errorf("invalid board cell value %d at row %d, column %d", value, row, column)
			}
			index := row*columns + column
			switch value {
			case cell.Blocked:
				values[index] = cell.Blocked
				targets[index] = cell.Blocked
			case cell.TriangleUpSlot:
				targets[index] = cell.TriangleUpSlot
			case cell.TriangleDownSlot, cell.TriangleLeftSlot, cell.TriangleRightSlot:
				targets[index] = value
			default:
				values[index] = value
				targets[index] = cell.Complete
				boardMinimum += int(cell.Complete - value)
				boardMaximum += int(cell.Complete - value)
			}
		}
	}
	for index, target := range targets {
		if isTriangleTarget(target) && values[index] == cell.Empty {
			minimum, maximum := triangleContributionRange(target)
			boardMinimum += minimum
			boardMaximum += maximum
		}
	}

	candidatesByCell := make([][]candidatePlacement, rows*columns)
	minPieceTotals := make([]int, len(pieces))
	maxPieceTotals := make([]int, len(pieces))
	minimumPossibleTotal, maximumPossibleTotal := 0, 0
	for pieceIndex, currentPiece := range pieces {
		if len(currentPiece.Variations) == 0 {
			return nil, fmt.Errorf("piece %q has no variations", currentPiece.Color)
		}

		minimumPieceTotal, maximumPieceTotal := 0, 0
		for variationIndex, variation := range currentPiece.Variations {
			shape := variation.Cells()
			height, width, shapeCells, err := variationDimensions(shape)
			if err != nil {
				return nil, fmt.Errorf("piece %q variation %d: %w", currentPiece.Color, variationIndex, err)
			}
			variationTotal := 0
			for _, shapeCell := range shapeCells {
				variationTotal += int(shapeCell.contribution)
			}
			if variationIndex == 0 || variationTotal < minimumPieceTotal {
				minimumPieceTotal = variationTotal
			}
			if variationIndex == 0 || variationTotal > maximumPieceTotal {
				maximumPieceTotal = variationTotal
			}

			for top := 0; top+height <= rows; top++ {
				for left := 0; left+width <= columns; left++ {
					placement := candidatePlacement{
						placement: Placement{
							Piece:          currentPiece,
							VariationIndex: variationIndex,
							Row:            top,
							Column:         left,
						},
						pieceIndex: pieceIndex,
						cells:      make([]placementCell, len(shapeCells)),
					}
					for i, shapeCell := range shapeCells {
						row := top + shapeCell.index/len(shape[0])
						column := left + shapeCell.index%len(shape[0])
						index := row*columns + column
						placement.cells[i] = placementCell{index: index, contribution: shapeCell.contribution}
						candidatesByCell[index] = append(candidatesByCell[index], placement)
					}
				}
			}
		}
		minPieceTotals[pieceIndex] = minimumPieceTotal
		maxPieceTotals[pieceIndex] = maximumPieceTotal
		minimumPossibleTotal += minimumPieceTotal
		maximumPossibleTotal += maximumPieceTotal
	}

	if boardMinimum > maximumPossibleTotal || boardMaximum < minimumPossibleTotal {
		return nil, fmt.Errorf(
			"board requires between %d and %d, but pieces can contribute between %d and %d",
			boardMinimum,
			boardMaximum,
			minimumPossibleTotal,
			maximumPossibleTotal,
		)
	}

	rng := opts.Rand
	if opts.Mode == ModeRandom && rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	usedPieces := make([]bool, len(pieces))
	path := make([]Placement, 0, len(pieces))
	intermediateBoards := make([]Board, 0, len(pieces))
	results := make([]*Solution, 0)
	var search func() bool
	search = func() bool {
		if err := ctx.Err(); err != nil {
			return true
		}

		remainingBoardMinimum, remainingBoardMaximum := 0, 0
		for index, target := range targets {
			switch target {
			case cell.Complete:
				remaining := int(cell.Complete - values[index])
				remainingBoardMinimum += remaining
				remainingBoardMaximum += remaining
			default:
				if !isTriangleTarget(target) {
					continue
				}
				if values[index] == cell.Empty {
					minimum, maximum := triangleContributionRange(target)
					remainingBoardMinimum += minimum
					remainingBoardMaximum += maximum
				}
			}
		}
		remainingMinimum, remainingMaximum := 0, 0
		for pieceIndex, used := range usedPieces {
			if used {
				continue
			}
			remainingMinimum += minPieceTotals[pieceIndex]
			remainingMaximum += maxPieceTotals[pieceIndex]
		}
		if remainingBoardMinimum > remainingMaximum || remainingBoardMaximum < remainingMinimum {
			return false
		}

		targetIndex, options := -1, []candidatePlacement(nil)
		for index, value := range values {
			if isFilled(value, targets[index]) {
				continue
			}

			compatible := make([]candidatePlacement, 0)
			for _, candidate := range candidatesByCell[index] {
				if usedPieces[candidate.pieceIndex] || !canPlace(values, targets, candidate) {
					continue
				}
				compatible = append(compatible, candidate)
			}
			if len(compatible) == 0 {
				return false
			}
			if targetIndex == -1 || len(compatible) < len(options) {
				targetIndex, options = index, compatible
			}
		}

		if targetIndex == -1 {
			for _, used := range usedPieces {
				if !used {
					return false
				}
			}
			results = append(results, &Solution{
				Board: boardFromValues(
					values,
					targets,
					rows,
					columns,
					initial.displayRows,
					initial.displayIndent,
				),
				Placements:         append([]Placement(nil), path...),
				IntermediateBoards: append([]Board(nil), intermediateBoards...),
			})
			if opts.Mode == ModeFirst {
				return true
			}
			limit := opts.Limit
			if opts.Mode == ModeRandom && limit == 0 {
				limit = 1
			}
			if limit > 0 && len(results) >= limit {
				return true
			}
			return false
		}

		if opts.Mode == ModeRandom {
			rng.Shuffle(len(options), func(i, j int) {
				options[i], options[j] = options[j], options[i]
			})
		}

		for _, candidate := range options {
			usedPieces[candidate.pieceIndex] = true
			for _, shapeCell := range candidate.cells {
				values[shapeCell.index] += shapeCell.contribution
			}
			path = append(path, candidate.placement)
			intermediateBoards = append(intermediateBoards, boardFromValues(
				values,
				targets,
				rows,
				columns,
				initial.displayRows,
				initial.displayIndent,
			))

			if search() {
				return true
			}

			intermediateBoards = intermediateBoards[:len(intermediateBoards)-1]
			path = path[:len(path)-1]
			for _, shapeCell := range candidate.cells {
				values[shapeCell.index] -= shapeCell.contribution
			}
			usedPieces[candidate.pieceIndex] = false
		}
		return false
	}

	search()

	if len(results) == 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	return results, nil
}

func isFilled(value, target cell.CellType) bool {
	if isTriangleTarget(target) {
		return triangleFits(target, value)
	}
	return value == cell.Complete || target == cell.Blocked
}

func isTriangleTarget(target cell.CellType) bool {
	return target >= cell.TriangleUpSlot && target <= cell.TriangleRightSlot
}

func triangleContributionRange(target cell.CellType) (int, int) {
	return int(cell.DownRight), int(cell.TopLeft)
}

func triangleFits(target, contribution cell.CellType) bool {
	return isTriangleTarget(target) && contribution >= cell.DownRight && contribution <= cell.TopLeft
}

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

func canPlace(values, targets []cell.CellType, candidate candidatePlacement) bool {
	for _, shapeCell := range candidate.cells {
		index := shapeCell.index
		if targets[index] == cell.Blocked {
			return false
		}
		if isTriangleTarget(targets[index]) {
			if values[index] != cell.Empty || shapeCell.contribution == cell.Complete {
				return false
			}
			if !triangleFits(targets[index], shapeCell.contribution) {
				return false
			}
			continue
		}
		if values[index]+shapeCell.contribution > targets[index] {
			return false
		}
	}
	return true
}
