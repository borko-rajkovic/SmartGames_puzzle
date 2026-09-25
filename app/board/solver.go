package board

import (
	"fmt"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

type Placement struct {
	Piece          piece.Piece
	VariationIndex int
	Row            int
	Column         int
}

type Solution struct {
	Board      Board
	Placements []Placement
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

func FindSolution(initial Board, pieces []piece.Piece) (*Solution, error) {
	rows, columns, err := boardDimensions(initial.cells)
	if err != nil {
		return nil, fmt.Errorf("invalid board: %w", err)
	}

	values := make([]cell.CellType, rows*columns)
	neededTotal := 0
	for row := range initial.cells {
		for column, value := range initial.cells[row] {
			if value < cell.Empty || value > cell.Complete {
				return nil, fmt.Errorf("invalid board cell value %d at row %d, column %d", value, row, column)
			}
			values[row*columns+column] = value
			neededTotal += int(cell.Complete - value)
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

	if neededTotal < minimumPossibleTotal || neededTotal > maximumPossibleTotal {
		return nil, fmt.Errorf(
			"pieces can contribute between %d and %d, but the board requires %d",
			minimumPossibleTotal,
			maximumPossibleTotal,
			neededTotal,
		)
	}

	usedPieces := make([]bool, len(pieces))
	path := make([]Placement, 0, len(pieces))
	var search func() bool
	search = func() bool {
		remainingNeeded := 0
		for _, value := range values {
			remainingNeeded += int(cell.Complete - value)
		}
		remainingMinimum, remainingMaximum := 0, 0
		for pieceIndex, used := range usedPieces {
			if used {
				continue
			}
			remainingMinimum += minPieceTotals[pieceIndex]
			remainingMaximum += maxPieceTotals[pieceIndex]
		}
		if remainingNeeded < remainingMinimum || remainingNeeded > remainingMaximum {
			return false
		}

		targetIndex, options := -1, []candidatePlacement(nil)
		for index, value := range values {
			if value == cell.Complete {
				continue
			}

			compatible := make([]candidatePlacement, 0)
			for _, candidate := range candidatesByCell[index] {
				if usedPieces[candidate.pieceIndex] || !canPlace(values, candidate) {
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
			return true
		}

		for _, candidate := range options {
			usedPieces[candidate.pieceIndex] = true
			for _, shapeCell := range candidate.cells {
				values[shapeCell.index] += shapeCell.contribution
			}
			path = append(path, candidate.placement)

			if search() {
				return true
			}

			path = path[:len(path)-1]
			for _, shapeCell := range candidate.cells {
				values[shapeCell.index] -= shapeCell.contribution
			}
			usedPieces[candidate.pieceIndex] = false
		}
		return false
	}

	if !search() {
		return nil, nil
	}

	solvedBoard := Board{cells: make([][]cell.CellType, rows)}
	for row := range solvedBoard.cells {
		solvedBoard.cells[row] = append([]cell.CellType(nil), values[row*columns:(row+1)*columns]...)
	}
	return &Solution{
		Board:      solvedBoard,
		Placements: append([]Placement(nil), path...),
	}, nil
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

func canPlace(values []cell.CellType, candidate candidatePlacement) bool {
	for _, shapeCell := range candidate.cells {
		if values[shapeCell.index]+shapeCell.contribution > cell.Complete {
			return false
		}
	}
	return true
}
