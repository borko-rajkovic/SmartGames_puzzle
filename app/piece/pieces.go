package piece

import "github.com/borko-rajkovic/smart_games_puzzle/app/cell"

var PiecesMap = map[string]Piece{
	"a": DarkBluePiece,
	"b": RedPiece,
	"c": LightBluePiece,
	"d": PurplePiece,
	"e": DarkGreenPiece,
	"f": LightGreenPiece,
	"g": TurquoisePiece,
	"h": OrangePiece,
	"i": YellowPiece,
	"j": PinkPiece,
}

func PiecesMapToDigits(piecesMap map[string]Piece) []rune {
	digits := make([]rune, 0, len(piecesMap))
	for digit := range piecesMap {
		digits = append(digits, rune(digit[0]))
	}
	return digits
}

var Pieces = []Piece{
	DarkBluePiece,
	RedPiece,
	LightBluePiece,
	PurplePiece,
	DarkGreenPiece,
	LightGreenPiece,
	TurquoisePiece,
	OrangePiece,
	YellowPiece,
	PinkPiece,
}

var DarkBluePiece = createDarkBluePiece()
var RedPiece = createRedPiece()
var LightBluePiece = createLightBluePiece()
var PurplePiece = createPurplePiece()
var DarkGreenPiece = createDarkGreenPiece()
var LightGreenPiece = createLightGreenPiece()
var TurquoisePiece = createTurquoisePiece()
var OrangePiece = createOrangePiece()
var YellowPiece = createYellowPiece()
var PinkPiece = createPinkPiece()

func createDarkBluePiece() Piece {
	return createPiece("Dark Blue", [][]cell.CellType{
		{cell.DownRight, cell.Complete, cell.Complete},
		{cell.Empty, cell.Empty, cell.TopRight},
		{cell.Empty, cell.Empty, cell.Empty},
	})
}

func createRedPiece() Piece {
	return createPiece("Red", [][]cell.CellType{
		{cell.Complete, cell.Complete, cell.Empty},
		{cell.TopLeft, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty},
	})
}

func createLightBluePiece() Piece {
	return createPiece("Light Blue", [][]cell.CellType{
		{cell.DownLeft, cell.Empty, cell.Empty},
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.TopRight, cell.Empty, cell.Empty},
	})
}

func createPurplePiece() Piece {
	return createPiece("Purple", [][]cell.CellType{
		{cell.Empty, cell.DownRight, cell.Empty},
		{cell.DownRight, cell.Complete, cell.Empty},
		{cell.TopRight, cell.TopLeft, cell.Empty},
	})
}

func createDarkGreenPiece() Piece {
	return createPiece("Dark Green", [][]cell.CellType{
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.TopRight, cell.Empty, cell.Empty},
	})
}

func createLightGreenPiece() Piece {
	return createPiece("Light Green", [][]cell.CellType{
		{cell.DownRight, cell.Empty, cell.Empty},
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.Complete, cell.Complete, cell.Empty},
	})
}

func createTurquoisePiece() Piece {
	return createPiece("Turquoise", [][]cell.CellType{
		{cell.Complete, cell.DownLeft, cell.Empty},
		{cell.Complete, cell.TopLeft, cell.Empty},
		{cell.TopLeft, cell.Empty, cell.Empty},
	})
}

func createOrangePiece() Piece {
	return createPiece("Orange", [][]cell.CellType{
		{cell.Empty, cell.DownRight, cell.Complete},
		{cell.DownRight, cell.Complete, cell.TopLeft},
		{cell.Empty, cell.Empty, cell.Empty},
	})
}

func createYellowPiece() Piece {
	return createPiece("Yellow", [][]cell.CellType{
		{cell.Empty, cell.DownRight, cell.Empty},
		{cell.Empty, cell.Complete, cell.Empty},
		{cell.TopRight, cell.Complete, cell.Empty},
	})
}

func createPinkPiece() Piece {
	return createPiece("Pink", [][]cell.CellType{
		{cell.Complete, cell.Complete, cell.DownLeft},
		{cell.Empty, cell.TopRight, cell.TopLeft},
		{cell.Empty, cell.Empty, cell.Empty},
	})
}
