package board

import (
	"fmt"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
)

type Board struct {
	cells         [][]cell.CellType
	targets       [][]cell.CellType
	displayRows   [][]int
	displayIndent []int
}

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
				cellString = "◣"
			case cell.DownLeft:
				cellString = "◤"
			case cell.TopLeft:
				cellString = "◥"
			case cell.Complete:
				cellString = "■"
			case cell.Blocked:
				cellString = "×"
			case cell.TriangleSlot:
				cellString = "△"
			case cell.TriangleDownSlot:
				cellString = "▽"
			case cell.TriangleLeftSlot:
				cellString = "◁"
			case cell.TriangleRightSlot:
				cellString = "▷"
			}
			if boardCell == cell.Empty && len(b.targets) > 0 {
				switch b.targets[rowIndex][column] {
				case cell.TriangleSlot:
					cellString = "△"
				case cell.TriangleDownSlot:
					cellString = "▽"
				case cell.TriangleLeftSlot:
					cellString = "◁"
				case cell.TriangleRightSlot:
					cellString = "▷"
				}
			}
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

func (b Board) printAngled() {
	for rowIndex, row := range b.displayRows {
		for range b.displayIndent[rowIndex] {
			fmt.Print(" ")
		}
		for column, index := range row {
			if column > 0 {
				fmt.Print(" ")
			}
			value := b.cells[index/len(b.cells[0])][index%len(b.cells[0])]
			target := cell.Empty
			if len(b.targets) > 0 {
				target = b.targets[index/len(b.cells[0])][index%len(b.cells[0])]
			}
			switch value {
			case cell.Empty:
				if target == cell.TriangleDownSlot {
					fmt.Print("▽")
				} else if target == cell.TriangleLeftSlot {
					fmt.Print("◁")
				} else if target == cell.TriangleRightSlot {
					fmt.Print("▷")
				} else {
					fmt.Print("□")
				}
			case cell.TriangleDownSlot:
				fmt.Print("▽")
			case cell.TriangleLeftSlot:
				fmt.Print("◁")
			case cell.TriangleRightSlot:
				fmt.Print("▷")
			case cell.DownRight:
				fmt.Print("◢")
			case cell.TopRight:
				fmt.Print("◣")
			case cell.DownLeft:
				fmt.Print("◤")
			case cell.TopLeft:
				fmt.Print("◥")
			case cell.Complete:
				fmt.Print("■")
			}
		}
		fmt.Println()
	}
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
