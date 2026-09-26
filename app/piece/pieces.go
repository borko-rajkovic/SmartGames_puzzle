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
	piece, error := NewPiece("Dark Blue", [][]cell.CellType{
		{cell.DownRight, cell.Complete, cell.Complete},
		{cell.Empty, cell.Empty, cell.TopRight},
		{cell.Empty, cell.Empty, cell.Empty},
	})
	if error != nil {
		panic(error)
	}
	return piece
}

func createRedPiece() Piece {
	piece, error := NewPiece("Red", [][]cell.CellType{
		{cell.Complete, cell.Complete, cell.Empty},
		{cell.TopLeft, cell.Empty, cell.Empty},
		{cell.Empty, cell.Empty, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createLightBluePiece() Piece {
	piece, error := NewPiece("Light Blue", [][]cell.CellType{
		{cell.DownLeft, cell.Empty, cell.Empty},
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.TopRight, cell.Empty, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createPurplePiece() Piece {
	piece, error := NewPiece("Purple", [][]cell.CellType{
		{cell.Empty, cell.DownRight, cell.Empty},
		{cell.DownRight, cell.Complete, cell.Empty},
		{cell.TopRight, cell.TopLeft, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createDarkGreenPiece() Piece {
	piece, error := NewPiece("Dark Green", [][]cell.CellType{
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.TopRight, cell.Empty, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createLightGreenPiece() Piece {
	piece, error := NewPiece("Light Green", [][]cell.CellType{
		{cell.DownRight, cell.Empty, cell.Empty},
		{cell.Complete, cell.Empty, cell.Empty},
		{cell.Complete, cell.Complete, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createTurquoisePiece() Piece {
	piece, error := NewPiece("Turquoise", [][]cell.CellType{
		{cell.Complete, cell.DownLeft, cell.Empty},
		{cell.Complete, cell.TopLeft, cell.Empty},
		{cell.TopLeft, cell.Empty, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createOrangePiece() Piece {
	piece, error := NewPiece("Orange", [][]cell.CellType{
		{cell.Empty, cell.DownRight, cell.Complete},
		{cell.DownRight, cell.Complete, cell.TopLeft},
		{cell.Empty, cell.Empty, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createYellowPiece() Piece {
	piece, error := NewPiece("Yellow", [][]cell.CellType{
		{cell.Empty, cell.DownRight, cell.Empty},
		{cell.Empty, cell.Complete, cell.Empty},
		{cell.TopRight, cell.Complete, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}

func createPinkPiece() Piece {
	piece, error := NewPiece("Pink", [][]cell.CellType{
		{cell.Complete, cell.Complete, cell.DownLeft},
		{cell.Empty, cell.TopRight, cell.TopLeft},
		{cell.Empty, cell.Empty, cell.Empty},
	})

	if error != nil {
		panic(error)
	}

	return piece
}
