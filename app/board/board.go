package board

import (
	"fmt"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
)


func (b Board) Print() {
	if len(b.displayRows) > 0 {
		b.printAngled()
		return
	}

	colLength := len(b.cells[0])

	for range colLength {
		fmt.Print(" _")
	}
	fmt.Println()
	for rowIndex, row := range b.cells {
		fmt.Print("|")
		for column, boardCell := range row {
			cellString := " "
			switch boardCell {
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
			case cell.Blocked:
				cellString = "×"
			case cell.TriangleUpSlot:
				cellString = "△"
			case cell.TriangleDownSlot:
				cellString = "▽"
			case cell.TriangleLeftSlot:
				cellString = "◁"
			case cell.TriangleRightSlot:
				cellString = "▷"
			}
			if boardCell == cell.Empty && len(b.targets) > 0 {
			fmt.Print(cellString)
			fmt.Print("|")
		}
		fmt.Println()
			}
	for range colLength {
		fmt.Print(" -")
	}
	fmt.Println()
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

