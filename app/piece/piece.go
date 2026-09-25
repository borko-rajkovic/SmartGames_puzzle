package piece

import (
	"fmt"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
)

type Piece struct {
	Color      string
	Variations []Variation
}

func (piece Piece) alignVariations() {
	// move shape closest to the top-left corner for each variation
	for i := range piece.Variations {
		variation := &piece.Variations[i]
		minRow, minCol := len(variation.cells), len(variation.cells[0])

		for row := 0; row < len(variation.cells); row++ {
			for col := 0; col < len(variation.cells[row]); col++ {
				if variation.cells[row][col] != cell.Empty {
					if row < minRow {
						minRow = row
					}
					if col < minCol {
						minCol = col
					}
				}
			}
		}

		if minRow > 0 || minCol > 0 {
			newCells := make([][]cell.CellType, len(variation.cells))
			for row := range newCells {
				newCells[row] = make([]cell.CellType, len(variation.cells[0]))
				for col := range newCells[row] {
					if row+minRow < len(variation.cells) && col+minCol < len(variation.cells[row]) {
						newCells[row][col] = variation.cells[row+minRow][col+minCol]
					} else {
						newCells[row][col] = cell.Empty
					}
				}
			}
			variation.cells = newCells
		}
	}
}

type Variation struct {
	cells [][]cell.CellType
}

func (variation Variation) Cells() [][]cell.CellType {
	cells := make([][]cell.CellType, len(variation.cells))
	for row := range variation.cells {
		cells[row] = append([]cell.CellType(nil), variation.cells[row]...)
	}
	return cells
}

// Cells
//
// - 0 - Empty
// - 1 - South-East portion of the piece
// - 2 - North-East portion of the piece
// - 3 - South-West portion of the piece
// - 4 - North-West portion of the piece
// - 5 - Complete piece

func (piece *Piece) Print() {
	fmt.Printf("Color: %s\n", piece.Color)
	fmt.Println("Variations:")
	for i, variation := range piece.Variations {
		fmt.Printf("Variation %d:\n", i+1)
		variation.Print()
		fmt.Println()
	}
}

func (variation Variation) Print() {
	for _, row := range variation.cells {
		for _, pieceCell := range row {
			cellString := " "
			switch pieceCell {
			case cell.DownRight:
				cellString = "◢"
			case cell.TopRight:
				cellString = "◥"
			case cell.DownLeft:
				cellString = "◣"
			case cell.TopLeft:
				cellString = "◤"
			case cell.Complete:
				cellString = "■"
			}

			fmt.Printf("%s.", cellString)
		}
		fmt.Println()
	}
}

func (piece *Piece) calculateVariations(variation Variation) {
	piece.appendRotationVariations(variation)
	flippedVariation := flipVariation(variation)
	piece.appendRotationVariations(flippedVariation)
}

func flipVariation(variation Variation) Variation {
	flipped := make([][]cell.CellType, len(variation.cells))
	for i := range flipped {
		flipped[i] = make([]cell.CellType, len(variation.cells[0]))
		for j := range flipped[i] {
			flipped[i][j] = variation.cells[i][len(variation.cells[0])-1-j]
			switch flipped[i][j] {
			case cell.DownRight:
				flipped[i][j] = cell.DownLeft
			case cell.TopRight:
				flipped[i][j] = cell.TopLeft
			case cell.DownLeft:
				flipped[i][j] = cell.DownRight
			case cell.TopLeft:
				flipped[i][j] = cell.TopRight
			}
		}
	}
	return Variation{cells: flipped}
}

func (piece *Piece) appendRotationVariations(variation Variation) {
	piece.Variations = append(piece.Variations, variation)

	for range 3 {
		rotated := make([][]cell.CellType, len(variation.cells[0]))
		for i := range rotated {
			rotated[i] = make([]cell.CellType, len(variation.cells))
			for j := range rotated[i] {
				rotated[i][j] = piece.Variations[len(piece.Variations)-1].cells[len(variation.cells)-1-j][i]
				switch rotated[i][j] {
				case cell.DownRight:
					rotated[i][j] = cell.DownLeft
				case cell.TopRight:
					rotated[i][j] = cell.DownRight
				case cell.DownLeft:
					rotated[i][j] = cell.TopLeft
				case cell.TopLeft:
					rotated[i][j] = cell.TopRight
				}
			}
		}

		piece.Variations = append(piece.Variations, Variation{cells: rotated})
	}
}

func createPiece(color string, cells [][]cell.CellType) Piece {
	variation := Variation{
		cells: cells,
	}

	piece := Piece{
		Color: color,
	}

	piece.calculateVariations(variation)
	piece.alignVariations()

	return piece
}

// NewPiece creates a piece and generates its rotations and reflections.
func NewPiece(color string, cells [][]cell.CellType) (Piece, error) {
	if len(cells) == 0 || len(cells[0]) == 0 {
		return Piece{}, fmt.Errorf("piece must have at least one row and column")
	}
	columns := len(cells[0])
	copiedCells := make([][]cell.CellType, len(cells))
	hasOccupiedCell := false
	for row := range cells {
		if len(cells[row]) != columns {
			return Piece{}, fmt.Errorf("row %d has %d columns; expected %d", row, len(cells[row]), columns)
		}
		copiedCells[row] = append([]cell.CellType(nil), cells[row]...)
		for column, value := range cells[row] {
			if value < cell.Empty || value > cell.Complete {
				return Piece{}, fmt.Errorf("invalid cell value %d at row %d, column %d", value, row, column)
			}
			hasOccupiedCell = hasOccupiedCell || value != cell.Empty
		}
	}
	if !hasOccupiedCell {
		return Piece{}, fmt.Errorf("piece must have at least one occupied cell")
	}
	return createPiece(color, copiedCells), nil
}
