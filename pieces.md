Each piece has maximum 3 squares in any direction.

What that effectively means is that any piece should fit into 3x3 squares:

| X | X | X |
| X | X | X |
| X | X | X |

Any position can be:

- 0 - Empty
- 1 - South-East portion of the piece
- 2 - North-East portion of the piece
- 3 - South-West portion of the piece
- 4 - North-West portion of the piece
- 5 - Complete piece

Numbers are chosen in this way to achieve that adding up 2 complementary pieces gives us 5:

SE, NE, SW, NW
 1,  2,  3,  4

1(SE) + 4(NW) = 5
2(NE) + 3(SW) = 5

That will help us later on in checking if piece can fill their cell on the board.

## Solver

`board.FindSolution(board.FlatBoard, piece.Pieces)` searches piece placements and
their rotations/reflections, returning the first arrangement that makes every
board cell equal `Complete`. It uses backtracking and chooses the unfinished
cell with the fewest currently compatible placements to reduce the search.