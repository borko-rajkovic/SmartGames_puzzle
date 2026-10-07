package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/borko-rajkovic/smart_games_puzzle/app/board"
	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

// pieceColors maps each puzzle piece's color name to a terminal color used
// to render the cells it occupies.
var pieceColors = map[string]lipgloss.Color{
	"Dark Blue":   lipgloss.Color("27"),
	"Red":         lipgloss.Color("196"),
	"Light Blue":  lipgloss.Color("45"),
	"Purple":      lipgloss.Color("129"),
	"Dark Green":  lipgloss.Color("22"),
	"Light Green": lipgloss.Color("83"),
	"Turquoise":   lipgloss.Color("80"),
	"Orange":      lipgloss.Color("214"),
	"Yellow":      lipgloss.Color("226"),
	"Pink":        lipgloss.Color("212"),
}

func colorFor(pieceColor string) lipgloss.Color {
	if c, ok := pieceColors[pieceColor]; ok {
		return c
	}
	return lipgloss.Color("245")
}

var (
	blockedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	emptyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	slotStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

func glyphFor(value cell.CellType) string {
	switch value {
	case cell.DownRight:
		return "◢"
	case cell.TopRight:
		return "◥"
	case cell.DownLeft:
		return "◣"
	case cell.TopLeft:
		return "◤"
	case cell.Complete:
		return "■"
	case cell.Blocked:
		return "×"
	case cell.TriangleUpSlot:
		return "△"
	case cell.TriangleDownSlot:
		return "▽"
	case cell.TriangleLeftSlot:
		return "▷"
	case cell.TriangleRightSlot:
		return "◁"
	}
	return " "
}

// pieceOverlay maps each flat (row*columns+column) board index to the
// color of the piece occupying it, considering only placements[:upTo].
func pieceOverlay(columns int, placements []board.Placement, upTo int) map[int]string {
	overlay := make(map[int]string)
	if upTo > len(placements) {
		upTo = len(placements)
	}
	for i := 0; i < upTo; i++ {
		p := placements[i]
		shape := p.Piece.Variations[p.VariationIndex].Cells()
		for r, rowCells := range shape {
			for c, value := range rowCells {
				if value == cell.Empty {
					continue
				}
				overlay[(p.Row+r)*columns+p.Column+c] = p.Piece.Color
			}
		}
	}
	return overlay
}

// lastPlacementCells returns the flat board indices touched by
// placements[upTo-1], used to briefly highlight the piece that was just
// placed during animation.
func lastPlacementCells(columns int, placements []board.Placement, upTo int) map[int]bool {
	cells := make(map[int]bool)
	if upTo <= 0 || upTo > len(placements) {
		return cells
	}
	p := placements[upTo-1]
	shape := p.Piece.Variations[p.VariationIndex].Cells()
	for r, rowCells := range shape {
		for c, value := range rowCells {
			if value == cell.Empty {
				continue
			}
			cells[(p.Row+r)*columns+p.Column+c] = true
		}
	}
	return cells
}

func styledGlyph(value, target cell.CellType, pieceColor string, highlighted bool) string {
	switch {
	case value == cell.Blocked || target == cell.Blocked:
		return blockedStyle.Render("×")
	case value >= cell.TriangleUpSlot && value <= cell.TriangleRightSlot:
		// A predefined, unsolved board encodes its triangle slots directly
		// as the cell value rather than via a separate target grid.
		return slotStyle.Render(glyphFor(value))
	case value == cell.Empty:
		if target >= cell.TriangleUpSlot && target <= cell.TriangleRightSlot {
			return slotStyle.Render(glyphFor(target))
		}
		return emptyStyle.Render("·")
	default:
		style := lipgloss.NewStyle().Foreground(colorFor(pieceColor)).Bold(true)
		if highlighted {
			style = style.Reverse(true)
		}
		return style.Render(glyphFor(value))
	}
}

// RenderBoard renders b, coloring each filled cell by the piece that
// occupies it according to placements[:upTo]. upTo may be less than
// len(placements) to render a mid-solve animation frame; the cells
// belonging to placements[upTo-1] (the most recently placed piece) are
// rendered in reverse video so they stand out.
func RenderBoard(b board.Board, placements []board.Placement, upTo int) string {
	rows, columns := b.Rows(), b.Columns()
	overlay := pieceOverlay(columns, placements, upTo)
	highlight := lastPlacementCells(columns, placements, upTo)

	cellString := func(row, column int) string {
		index := row*columns + column
		return styledGlyph(b.CellAt(row, column), b.TargetAt(row, column), overlay[index], highlight[index])
	}

	if displayRows := b.DisplayRows(); len(displayRows) > 0 {
		indent := b.DisplayIndent()
		lines := make([]string, 0, len(displayRows))
		for rowIndex, rowIndices := range displayRows {
			var line strings.Builder
			line.WriteString(strings.Repeat(" ", indent[rowIndex]))
			for i, flatIndex := range rowIndices {
				if i > 0 {
					line.WriteString(" ")
				}
				line.WriteString(cellString(flatIndex/columns, flatIndex%columns))
			}
			lines = append(lines, line.String())
		}
		return strings.Join(lines, "\n")
	}

	lines := make([]string, 0, rows)
	for row := 0; row < rows; row++ {
		var line strings.Builder
		line.WriteString("|")
		for column := 0; column < columns; column++ {
			line.WriteString(cellString(row, column))
			line.WriteString("|")
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

// Legend renders a color-coded key mapping each piece's swatch to its name.
func Legend(pieces []piece.Piece) string {
	parts := make([]string, len(pieces))
	for i, p := range pieces {
		swatch := lipgloss.NewStyle().Foreground(colorFor(p.Color)).Bold(true).Render("■")
		parts[i] = swatch + " " + p.Color
	}
	return strings.Join(parts, "   ")
}
