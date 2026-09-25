package board

import (
	"fmt"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
)

type Board struct {
	cells [][]cell.CellType
}

func (b Board) Print() {
	colLength := len(b.cells[0])
	// rowLength := len(b.cells)

	for range colLength {
		fmt.Print(" _")
	}
	fmt.Println()
	for _, row := range b.cells {
		fmt.Print("|")
		for _, boardCell := range row {
			cellString := " "
			switch boardCell {
			case cell.DownRight:
				cellString = "◢"
			case cell.TopRight:
				cellString = "◣"
			case cell.DownLeft:
				cellString = "◤"
			case cell.TopLeft:
				cellString = "◥"
			case cell.Complete:
				cellString = "■"
			}
			fmt.Print(cellString)
			fmt.Print("|")
		}
		fmt.Println()
	}
	for range colLength {
		fmt.Print(" -")
	}
	println()
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
