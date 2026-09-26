package cell

type CellType int

const (
	Empty CellType = iota
	DownRight
	TopRight
	DownLeft
	TopLeft
	Complete
	Blocked
	TriangleSlot
	TriangleDownSlot
	TriangleLeftSlot
	TriangleRightSlot
)
