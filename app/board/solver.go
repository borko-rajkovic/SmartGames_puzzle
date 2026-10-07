package board

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/borko-rajkovic/smart_games_puzzle/app/cell"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

// Mode selects how FindSolutions explores the search space.
type Mode int

const (
	// ModeFirst stops as soon as a single solution is found.
	ModeFirst Mode = iota
	// ModeRandom shuffles candidate order at each branch so different
	// runs tend to discover different valid solutions.
	ModeRandom
	// ModeAll keeps searching after a solution is found, collecting every
	// solution it discovers (up to Options.Limit, if set).
	ModeAll
)

// Options configures a FindSolutions search.
type Options struct {
	// Mode selects the search strategy. Zero value is ModeFirst.
	Mode Mode
	// Limit caps the number of solutions collected. 0 means unlimited
	// (ModeFirst always stops after the first solution regardless).
	Limit int
	// Rand supplies randomness for ModeRandom. If nil, a time-seeded
	// source is used.
	Rand *rand.Rand
}

// Placement describes one piece placed on the board at a specific position
// and orientation (variation).
type Placement struct {
	Piece          piece.Piece
	VariationIndex int
	Row            int
	Column         int
}

// StepKind classifies what happened at one step of the backtracking search.
type StepKind int

const (
	// StepPlace records a candidate placement being tried.
	StepPlace StepKind = iota
	// StepBacktrack records a placement being undone because the branch failed.
	StepBacktrack
)

// SearchStep is one snapshot captured during the backtracking search: the
// board after an apply or undo, together with which piece was involved and
// whether this was a forward attempt or a rollback.
type SearchStep struct {
	Kind      StepKind
	Placement Placement
	Board     Board
}

// Solution holds one complete solution together with its search trace.
type Solution struct {
	Board Board
	// Placements lists the pieces in the order they were placed to reach
	// this solution.
	Placements []Placement
	// IntermediateBoards holds the board state after each successful
	// placement (one entry per element of Placements).
	IntermediateBoards []Board
	// SearchSteps records every place and backtrack operation the solver
	// performed while finding this solution, suitable for full-trace replay.
	SearchSteps []SearchStep
}

// placementCell links a flat board index to the contribution a piece cell adds.
type placementCell struct {
	index        int
	contribution cell.CellType
}

// candidatePlacement is a fully resolved placement candidate: which piece,
// which variation, where on the board, and the (index, contribution) pair
// for each of the piece's occupied cells.
type candidatePlacement struct {
	placement  Placement
	pieceIndex int
	cells      []placementCell
}

// FindSolution returns the first solution found for the given board and
// piece set, or (nil, nil) if none exists.
func FindSolution(initial Board, pieces []piece.Piece) (*Solution, error) {
	solutions, err := FindSolutions(context.Background(), initial, pieces, Options{Mode: ModeFirst})
	if err != nil {
		return nil, err
	}
	if len(solutions) == 0 {
		return nil, nil
	}
	return solutions[0], nil
}

// FindSolutions searches for solutions to the given board using the given
// piece set, according to opts. The search uses two pruning strategies:
//
//  1. Contribution-range pruning: before each recursive call, the solver
//     checks whether the remaining unused pieces can still contribute
//     enough (or not too much) to satisfy the remaining unfilled cells.
//     If not, the branch is cut immediately.
//
//  2. MRV (Minimum Remaining Values): at each level the solver picks the
//     unfilled cell that has the fewest compatible candidates, which keeps
//     the branching factor low and surfaces dead ends early.
//
// Every place and backtrack operation is recorded in Solution.SearchSteps
// so callers can replay the full search trace.
func FindSolutions(ctx context.Context, initial Board, pieces []piece.Piece, opts Options) ([]*Solution, error) {
	rows, columns, err := boardDimensions(initial.cells)
	if err != nil {
		return nil, fmt.Errorf("invalid board: %w", err)
	}

	// Parse the initial board into flat value/target slices.
	// values[i] is the current fill level; targets[i] is the required level.
	values, targets, boardMin, boardMax, err := buildBoardState(initial.cells, rows, columns)
	if err != nil {
		return nil, err
	}

	// Precompute every valid (piece × variation × position) candidate,
	// indexed by which board cell they touch. Also derive per-piece
	// contribution bounds used for feasibility pruning.
	candidatesByCell, minTotals, maxTotals, minPossible, maxPossible, err :=
		buildCandidates(pieces, rows, columns)
	if err != nil {
		return nil, err
	}

	// Fast global feasibility check: if the pieces cannot possibly satisfy
	// the board's total contribution requirement, fail immediately.
	if boardMin > maxPossible || boardMax < minPossible {
		return nil, fmt.Errorf(
			"board requires between %d and %d, but pieces can contribute between %d and %d",
			boardMin, boardMax, minPossible, maxPossible,
		)
	}

	rng := opts.Rand
	if opts.Mode == ModeRandom && rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// Mutable solver state — all shared across the recursive search closure.
	usedPieces := make([]bool, len(pieces))
	path := make([]Placement, 0, len(pieces))
	intermediateBoards := make([]Board, 0, len(pieces))
	searchSteps := make([]SearchStep, 0)
	results := make([]*Solution, 0)

	var search func() bool
	search = func() bool {
		if ctx.Err() != nil {
			// Context cancelled — stop and surface whatever was found.
			return true
		}

		// Pruning: bail early when remaining pieces cannot fill remaining cells.
		if !feasible(values, targets, usedPieces, minTotals, maxTotals) {
			return false
		}

		// MRV: pick the unfilled cell with the fewest compatible candidates.
		// Returns (-1, non-nil): all cells are filled — solution candidate.
		// Returns (>=0, nil):   some cell has zero candidates — dead end.
		// Returns (>=0, non-nil): normal case — proceed with those candidates.
		targetIndex, options := findMostConstrained(values, targets, candidatesByCell, usedPieces)

		if targetIndex == -1 {
			// Every cell is filled. Check all pieces were actually used.
			for _, used := range usedPieces {
				if !used {
					return false
				}
			}
			results = append(results, &Solution{
				Board: boardFromValues(values, targets, rows, columns,
					initial.displayRows, initial.displayIndent),
				Placements:         append([]Placement(nil), path...),
				IntermediateBoards: append([]Board(nil), intermediateBoards...),
				SearchSteps:        append([]SearchStep(nil), searchSteps...),
			})
			if opts.Mode == ModeFirst {
				return true
			}
			limit := opts.Limit
			if opts.Mode == ModeRandom && limit == 0 {
				limit = 1
			}
			if limit > 0 && len(results) >= limit {
				return true
			}
			return false
		}

		if options == nil {
			// Dead end: the most-constrained cell has no valid candidates.
			return false
		}

		if opts.Mode == ModeRandom {
			rng.Shuffle(len(options), func(i, j int) {
				options[i], options[j] = options[j], options[i]
			})
		}

		for _, candidate := range options {
			// Forward step: apply this candidate and snapshot the result.
			applyPlacement(values, usedPieces, candidate)
			snap := boardFromValues(values, targets, rows, columns,
				initial.displayRows, initial.displayIndent)
			path = append(path, candidate.placement)
			intermediateBoards = append(intermediateBoards, snap)
			searchSteps = append(searchSteps, SearchStep{
				Kind:      StepPlace,
				Placement: candidate.placement,
				Board:     snap,
			})

			if search() {
				return true
			}

			// Backtrack step: undo the candidate and snapshot the rolled-back board.
			intermediateBoards = intermediateBoards[:len(intermediateBoards)-1]
			path = path[:len(path)-1]
			undoPlacement(values, usedPieces, candidate)
			searchSteps = append(searchSteps, SearchStep{
				Kind:      StepBacktrack,
				Placement: candidate.placement,
				Board: boardFromValues(values, targets, rows, columns,
					initial.displayRows, initial.displayIndent),
			})
		}
		return false
	}

	search()

	if len(results) == 0 {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, nil
	}
	return results, nil
}
