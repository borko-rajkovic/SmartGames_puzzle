# `board.go` and `solver.go` — how the puzzle is represented and solved

This document explains, in detail, the two core files of the `app/board` package:

- `app/board/board.go` — the `Board` data structure and how it's printed.
- `app/board/solver.go` — the backtracking search that finds a set of piece placements solving the board.

It also touches `app/cell/cell_type.go` and `app/piece/piece.go` where needed, since the solver's logic only makes sense in terms of those types.

## 1. The puzzle domain

This is a solver for a SmartGames-style triangle tiling puzzle: a board of square cells must be completely covered by a fixed set of pieces, where each piece is made of triangular halves and/or full squares. Every square cell is split diagonally into two triangles (a "South-East" + "North-West" pair, or a "North-East" + "South-West" pair, depending on orientation), and each piece contributes one or two triangle halves — or a full "complete" square — to each cell it occupies.

The board additionally can have:
- **Blocked cells** — cells that must stay empty (not part of the puzzle area at all — used for the heart-shaped board's rounded corners).
- **Triangle slot cells** — cells where only *one specific triangle orientation* is allowed to be placed (used to carve the heart shape out of a rectangular grid).

### `cell.CellType` — the alphabet everything is built from

```go
const (
	Empty CellType = iota
	DownRight
	TopRight
	DownLeft
	TopLeft
	Complete
	Blocked
	TriangleUpSlot
	TriangleDownSlot
	TriangleLeftSlot
	TriangleRightSlot
)
```

`CellType` is an `int` enum used for two different purposes depending on context:

1. **As a piece/board *value*** (how much of a cell is currently filled):
   - `Empty` (0) — nothing there yet.
   - `DownRight`, `TopRight`, `DownLeft`, `TopLeft` (1–4) — one diagonal half-triangle, named by which corner of the cell it points away from (e.g. `DownRight` = the triangle occupying the bottom-right half).
   - `Complete` (5) — the full cell is filled (either by a single "complete" piece cell, or by two complementary triangle halves added together — see below).
   - `Blocked` (6) — the cell doesn't exist / can never be filled.

2. **As a board *target*** (what the cell is allowed/required to end up as):
   - `Blocked` — must stay unfilled forever.
   - `TriangleUpSlot` / `TriangleDownSlot` / `TriangleLeftSlot` / `TriangleRightSlot` (7–10) — the cell must end up holding exactly one triangle, and only a triangle compatible with that slot orientation is accepted (see `triangleFits`).
   - `Complete` (implicitly, for ordinary cells) — the cell must end up fully filled.

The key trick that makes the solver's math work: **triangle values are additive**. Because `DownRight=1`, `TopRight=2`, `DownLeft=3`, `TopLeft=4`, and `Complete=5`, a `DownRight` (1) triangle plus a `TopLeft` (4) triangle sum to exactly `Complete` (5) — these two are the diagonally-complementary halves of a square, so adding their numeric values gives the "fully filled" sentinel for free, without any special-case code. This is why cell values are stored as `cell.CellType` (an int) and simply summed as pieces are placed (`solver.go:206`, `board.go` target math).

### `Board` (`board.go:9-14`)

```go
type Board struct {
	cells         [][]cell.CellType
	targets       [][]cell.CellType
	displayRows   [][]int
	displayIndent []int
}
```

- `cells` — the current fill state of the board, per row/column, using the "value" meaning of `CellType` above (`Empty`, `DownRight`, …, `Complete`, `Blocked`).
- `targets` — populated only on *solved/intermediate* boards produced by the solver (`boardFromValues`); tells the printer what a still-`Empty` cell is ultimately supposed to become (e.g. draw a faint △ for a `TriangleUpSlot` even though nothing has been placed there yet). Predefined boards like `FlatBoard`/`HeartBoard` don't set this — it's derived by the solver from the initial `cells`.
- `displayRows` / `displayIndent` — optional, used only by `HeartBoard`. The heart shape isn't a rectangle, so it can't be printed row-by-row from the underlying rectangular grid without gaps looking wrong. Instead, `displayRows` lists, for each *visual* row, which flat cell indices (`row*columns+column`) appear in that row, in left-to-right order, and `displayIndent` says how many leading spaces to print before that row so the diamond/heart shape lines up correctly. `Print()` uses `printAngled()` when `displayRows` is non-empty (`board.go:16-20`), otherwise it falls back to the plain rectangular renderer.

### Predefined boards (`board.go:124-155`)

- `FlatBoard` — a plain 5×6 grid of `Empty` cells, no shape constraints.
- `HeartBoard` — a 6×6 grid where:
  - The two top-left cells are `Blocked` (cut the corner off to help form the heart's dip).
  - Several cells are pre-set to `TriangleDownSlot`, `TriangleRightSlot`, `TriangleLeftSlot` — these force a specific triangle orientation at the rounded edges of the heart.
  - `displayRows`/`displayIndent` remap the 6×6 grid into the diamond/heart visual layout for printing.

### `Print()` / `printAngled()` (`board.go:16-122`)

Straightforward rendering: iterate cells, map each `CellType` to a glyph (`◢ ◣ ◤ ◥` for triangle halves, `■` for complete, `×` for blocked, `△ ▽ ◁ ▷` for empty triangle slots so you can see what's still needed). The only subtlety: for an `Empty` cell, if `targets` is populated and says this cell wants a specific triangle slot, that slot's glyph is printed instead of blank — so intermediate/solved boards visually show unmet triangle slots. `printAngled()` does the same cell→glyph mapping but walks `displayRows`/`displayIndent` instead of `cells` directly, for the heart layout.

## 2. Pieces and their variations (`piece.go`, `pieces.go`)

Each of the 10 physical puzzle pieces (`DarkBluePiece`, `RedPiece`, …) is defined once as a small grid of `CellType` values using the "half-triangle or complete" vocabulary, e.g.:

```go
{cell.DownRight, cell.Complete, cell.Complete},
{cell.Empty,     cell.Empty,    cell.TopRight},
{cell.Empty,     cell.Empty,    cell.Empty},
```

`NewPiece` → `createPiece` → `calculateVariations` then **automatically generates every rotation and mirror-image of that shape**:

- `appendRotationVariations` appends the given variation, then rotates it 90° three more times (4 rotations total), remapping each rotated triangle's orientation (`DownRight→DownLeft→TopLeft→TopRight→DownRight`, i.e. corners rotate consistently with the physical rotation).
- `flipVariation` mirrors the shape left-right and remaps `DownRight↔DownLeft`, `TopRight↔TopLeft` (the two "right" orientations flip to the corresponding "left" orientations).
- `calculateVariations` does rotations of the original shape *and* rotations of the flipped shape, giving up to 8 variations per piece (fewer if the piece has rotational/reflective symmetry — duplicates aren't de-duplicated, they're just redundant candidates the solver will try).
- `alignVariations` shifts every variation so its bounding box starts at row/column 0 (removes leading empty rows/columns), which keeps the shape's dimensions (`variationDimensions`) tight and consistent for placement math.

So by the time the solver receives `pieces []piece.Piece`, each piece already carries a full list of `Variations`, and each `Variation.Cells()` returns a small 2D grid (a copy) describing that orientation's shape and per-cell contribution values.

## 3. The solver (`solver.go`)

`FindSolution(initial Board, pieces []piece.Piece) (*Solution, error)` is an exact-cover-style backtracking search with heavy pruning. It works in three phases: **setup/flattening**, **candidate generation**, and **recursive search with feasibility pruning**.

### 3.1 Flattening the board into `values` and `targets`

The board's 2D `cells` grid is flattened into two parallel 1D slices of length `rows*columns`, indexed by `row*columns+column`:

- `values[index]` — current fill state (starts `Empty`, `Blocked`, or, for a pre-filled cell, its existing value).
- `targets[index]` — what that cell needs to become:
  - `Blocked` cells (both `values` and `targets`) stay `Blocked`/`Blocked` and are never touched again.
  - Cells that started as a triangle slot (`TriangleUpSlot`/`Down`/`Left`/`Right`) keep `values=Empty` but get `targets=<that slot type>`.
  - Every other cell (ordinary board cell, initially `Empty` or occasionally pre-filled) gets `targets=Complete` — it must end up fully filled.

While doing this, the code also accumulates `boardMinimum`/`boardMaximum`: the total numeric "contribution" the board still needs across all cells, as a range. For an ordinary cell needing `Complete`, min=max=`Complete - value` (how much more needs to be added). For a triangle-slot cell, `triangleContributionRange` returns `(int(cell.DownRight), int(cell.TopLeft))` = `(1, 4)` — *any* single triangle value between 1 and 4 could in principle satisfy it, so the range is wide (this is a coarse bound, refined later by the actual placement legality check).

### 3.2 Generating every legal candidate placement, indexed by cell

For every piece, every variation of that piece, and every top-left board position `(top,left)` where the variation's bounding box fits on the board, the solver builds a `candidatePlacement`:

```go
type candidatePlacement struct {
	placement  Placement          // piece, variation index, row, column
	pieceIndex int
	cells      []placementCell    // {index: flat board index, contribution: cell.CellType}
}
```

Crucially, this candidate is **not yet checked against blocked cells or triangle slots** — it's just "if this variation's non-empty shape cells were overlaid at this position, here's what flat board indices they'd touch and how much value they'd add." Each candidate is appended into `candidatesByCell[index]` for *every* cell index it touches (`solver.go:112-118`), so later the solver can ask "which candidates could possibly fill cell X?" in O(1) lookup instead of re-scanning all pieces.

While doing this, it also tracks, per piece, `minPieceTotals`/`maxPieceTotals` — the minimum and maximum total contribution value that piece could add across all its variations (used for a global feasibility check) — and sums these into `minimumPossibleTotal`/`maximumPossibleTotal` across all pieces.

### 3.3 Early impossibility check

```go
if boardMinimum > maximumPossibleTotal || boardMaximum < minimumPossibleTotal {
    return nil, fmt.Errorf(...)
}
```

If the total "demand" of the board (sum of what every cell needs) can't possibly be matched by the total "supply" all pieces could provide (even in the best case), the solver fails immediately with a descriptive error rather than searching. This is a coarse necessary-but-not-sufficient sanity check, done once before the recursive search even starts.

### 3.4 The recursive backtracking search

State carried across recursive calls (closed over by the `search` closure):
- `values` — mutated in place as pieces are placed/undone.
- `usedPieces []bool` — which pieces (by index) have already been placed somewhere (each piece is used exactly once, since a puzzle set has one copy of each).
- `path []Placement` — the sequence of placements made so far (the emerging solution).
- `intermediateBoards []Board` — a snapshot of the board's state *after* each placement, for later playback/animation of the solving process.

Each call to `search()` does:

1. **Global feasibility pruning (branch-and-bound):** recomputes, from scratch, `remainingBoardMinimum/Maximum` (sum over all not-yet-satisfied cells of how much more they still need) and `remainingMinimum/Maximum` (sum of `minPieceTotals`/`maxPieceTotals` over all pieces not yet used). If the unplaced pieces *can't possibly* supply what the unfilled board still needs (`remainingBoardMinimum > remainingMaximum`), or would necessarily *overshoot* even in the best case (`remainingBoardMaximum < remainingMinimum`), the branch is abandoned immediately (`return false`) without trying any more placements. This is the same kind of min/max interval check as the setup phase, but re-evaluated at every node so bad branches die as early as possible instead of only being caught much later by an explicit conflict.

2. **Most-constrained-cell selection (MRV heuristic):** rather than picking the next piece to place, the algorithm picks the next **cell** to fill — specifically the unfilled cell with the *fewest* legal candidate placements that could fill it (`solver.go:173-192`). For every cell not yet `isFilled`, it filters `candidatesByCell[index]` down to those whose piece isn't already used and which are legal right now (`canPlace`), then keeps track of whichever cell has the smallest non-empty `compatible` list. If any not-yet-filled cell has **zero** compatible candidates, the whole branch is dead (`return false`) — this is a strong, cheap-to-check dead-end detector (similar to "naked single has no candidates" in Sudoku solvers), and it's what makes this solver fast: instead of blindly trying every piece in every position, it always attacks the most constrained/most likely to fail cell first, so contradictions are discovered as early as possible.

   This "most constrained variable" strategy is the single most important idea in the whole search: on an open board with few placed pieces, many cells might have dozens of candidates, but corners, cells next to `Blocked` cells, or cells needing an already-used piece often have very few — the solver homes in on those first.

3. **Success/failure at a leaf:** if every cell is filled (`targetIndex == -1`, meaning the loop above found no more unfilled cells), the code double-checks that *every* piece got used (`for _, used := range usedPieces { if !used { return false } }`) — this matters because it's possible (in principle) for the board to be entirely filled/blocked while some piece was never placed, which wouldn't be a valid solution for a puzzle where all pieces must be used. Only if all cells are filled *and* all pieces are used does it return `true`.

4. **Trying candidates for the chosen cell:** for each candidate placement that could fill the chosen cell (`options`), it:
   - Marks `usedPieces[candidate.pieceIndex] = true`.
   - Adds each shape cell's `contribution` into `values` at the corresponding index (this is the "triangle halves sum to Complete" trick from section 1 in action).
   - Appends the placement to `path`, and snapshots the resulting board into `intermediateBoards` (via `boardFromValues`) for later display.
   - Recurses (`search()`). If it returns `true`, the solution is found and propagates back up unchanged (no undo needed — the placement stays).
   - If it returns `false`, **backtracks**: pops the last `intermediateBoards`/`path` entries, subtracts the contributions back out of `values`, and clears `usedPieces[candidate.pieceIndex]`, then tries the next candidate.
   - If no candidate works, `search()` returns `false` for this cell/branch, causing the caller one level up to backtrack further.

### 3.5 Legality checks: `isFilled`, `canPlace`, `triangleFits`

- `isFilled(value, target)` — a cell counts as already satisfied if: its target is a triangle slot and the current value is a triangle that fits that slot (`triangleFits`), OR its value is `Complete`, OR its target is `Blocked` (blocked cells are always considered "filled"/untouchable).
- `canPlace(values, targets, candidate)` — checks, for every cell a candidate would touch:
  - If the target is `Blocked`, reject outright — nothing may ever be placed on a blocked cell.
  - If the target is a triangle slot: the cell must currently be `Empty` (a triangle slot can only be filled once, by a single triangle — since a `Complete` piece cell there would overshoot, and a second triangle would corrupt the slot), the contribution can't itself be `Complete` (a "complete" piece cell can't legally go into a slot that wants exactly one triangle), and the contribution's orientation must match what the slot allows (`triangleFits`).
  - Otherwise (an ordinary `Complete`-target cell): reject if `values[index] + contribution > targets[index]` — i.e. don't let a placement overfill a cell past `Complete` (this is what prevents, e.g., two triangle halves that aren't complementary, or a triangle plus a complete cell, from being stacked illegally — the sum must never exceed 5).
- `triangleFits(target, contribution)` — true only if `target` is one of the 4 triangle slot types AND `contribution` is one of the 4 half-triangle orientations (`DownRight..TopLeft`). Note it does *not* check that the contribution's specific orientation (e.g. `TriangleUpSlot` vs `DownRight`) semantically matches — any of the 4 triangle halves is accepted for any of the 4 slot types under this check. The real-world semantic matching (which triangle orientation visually satisfies which slot glyph) is presumably guaranteed by how the puzzle's fixed boards/pieces are authored rather than enforced here — this function is a coarser "is it a slot and is it a half-triangle" gate, not a strict orientation match.

### 3.6 Result construction

If `search()` returns `true`, `FindSolution` builds the final `Solution`:

```go
type Solution struct {
	Board              Board     // the fully solved board
	Placements         []Placement // ordered list of {piece, variation, row, column}
	IntermediateBoards []Board     // board state after each successive placement
}
```

`Board` is rebuilt from the final `values`/`targets` via `boardFromValues`, carrying over the original `displayRows`/`displayIndent` so the heart-shaped board still prints correctly. `Placements` and `IntermediateBoards` are copies of the internal `path`/`intermediateBoards` slices (defensive copies via `append([]T(nil), ...)`, so later mutation of internal solver state — which doesn't happen after return, but as a matter of API hygiene — can't corrupt the returned solution).

If `search()` returns `false` (board fully explored, no arrangement of all 10 pieces satisfies all constraints), `FindSolution` returns `(nil, nil)` — no error, just "no solution exists" for this particular board.

## 4. Why this approach works well for this puzzle

- **Numeric cell values let "does this overfill/is this complete" collapse into simple integer arithmetic** (`values[index] + contribution > targets[index]`, and the `DownRight+TopLeft=Complete` identity) instead of needing an explicit geometric/orientation model at solve time.
- **Precomputing every (piece, variation, position) candidate once, indexed by the cells it touches**, turns "which pieces could go here?" into an array lookup instead of a fresh geometric scan at every search step.
- **Cell-based (not piece-based) branching, always picking the most constrained cell**, is what keeps the search tree small — it's the same idea as constraint propagation / MRV in Sudoku or exact-cover (DLX/Algorithm X) solvers: fail fast, fail on the most constrained part of the board first, rather than trying pieces in an arbitrary fixed order.
- **Min/max interval pruning at both the global level (once, in `FindSolution`) and at every recursion step (in `search`)** cuts off entire subtrees before the "zero candidates for some cell" check even needs to run, giving a second, cheaper layer of pruning.
- **Snapshotting `intermediateBoards` at every successful placement** means the caller (`app/smart_games_puzzle.go`) can replay the solve step-by-step for the user, not just show the final answer.
