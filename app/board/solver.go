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

// stackFrame is one level of the iterative DFS.
type stackFrame struct {
	options        []candidatePlacement
	idx            int
	hasEntry       bool
	entryCandidate candidatePlacement
	entryBDeltaMin int
	entryBDeltaMax int
}

// searchState holds all mutable state for a single DFS search run.
type searchState struct {
	initial Board
	rows    int
	columns int
	targets []cell.CellType
	values  []cell.CellType

	usedPieces []bool
	minTotals  []int
	maxTotals  []int

	path               []Placement
	intermediateBoards []Board
	searchSteps        []SearchStep
	results            []*Solution

	boardRemMin int
	boardRemMax int
	pieceRemMin int
	pieceRemMax int
}

func (s *searchState) currentBoard() Board {
	return boardFromValues(s.values, s.targets, s.rows, s.columns,
		s.initial.displayRows, s.initial.displayIndent)
}

func (s *searchState) isFeasible() bool {
	return s.boardRemMin <= s.pieceRemMax && s.boardRemMax >= s.pieceRemMin
}

// applyCandidate places a candidate, updates accumulators, and records a StepPlace.
func (s *searchState) applyCandidate(c candidatePlacement, bDeltaMin, bDeltaMax int) {
	s.boardRemMin += bDeltaMin
	s.boardRemMax += bDeltaMax
	s.pieceRemMin -= s.minTotals[c.pieceIndex]
	s.pieceRemMax -= s.maxTotals[c.pieceIndex]
	applyPlacement(s.values, s.usedPieces, c)
	snap := s.currentBoard()
	s.path = append(s.path, c.placement)
	s.intermediateBoards = append(s.intermediateBoards, snap)
	s.searchSteps = append(s.searchSteps, SearchStep{
		Kind:      StepPlace,
		Placement: c.placement,
		Board:     snap,
	})
}

// backtrack undoes a candidate, restores accumulators, and records a StepBacktrack.
func (s *searchState) backtrack(c candidatePlacement, bDeltaMin, bDeltaMax int) {
	s.intermediateBoards = s.intermediateBoards[:len(s.intermediateBoards)-1]
	s.path = s.path[:len(s.path)-1]
	undoPlacement(s.values, s.usedPieces, c)
	s.searchSteps = append(s.searchSteps, SearchStep{
		Kind:      StepBacktrack,
		Placement: c.placement,
		Board:     s.currentBoard(),
	})
	s.boardRemMin -= bDeltaMin
	s.boardRemMax -= bDeltaMax
	s.pieceRemMin += s.minTotals[c.pieceIndex]
	s.pieceRemMax += s.maxTotals[c.pieceIndex]
}

// recordSolution snapshots the current search state into a Solution.
func (s *searchState) recordSolution() {
	s.results = append(s.results, &Solution{
		Board:              s.currentBoard(),
		Placements:         append([]Placement(nil), s.path...),
		IntermediateBoards: append([]Board(nil), s.intermediateBoards...),
		SearchSteps:        append([]SearchStep(nil), s.searchSteps...),
	})
}

// allPiecesUsed reports whether every piece has been placed.
func allPiecesUsed(usedPieces []bool) bool {
	for _, used := range usedPieces {
		if !used {
			return false
		}
	}
	return true
}

// computeCandidateDelta returns how much this candidate reduces the board's
// remaining min and max requirement (both values are typically negative).
func computeCandidateDelta(cells []placementCell, targets []cell.CellType) (bDeltaMin, bDeltaMax int) {
	for _, sc := range cells {
		if isTriangleTarget(targets[sc.index]) {
			mn, mx := triangleContributionRange()
			bDeltaMin -= mn
			bDeltaMax -= mx
		} else {
			bDeltaMin -= int(sc.contribution)
			bDeltaMax -= int(sc.contribution)
		}
	}
	return
}

// reachedSolutionLimit reports whether the search should stop after the
// given number of results have been collected.
func reachedSolutionLimit(count int, opts Options) bool {
	if opts.Mode == ModeFirst {
		return true
	}
	limit := opts.Limit
	if opts.Mode == ModeRandom && limit == 0 {
		limit = 1
	}
	return limit > 0 && count >= limit
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

	values, targets, boardMin, boardMax, err := buildBoardState(initial.cells, rows, columns)
	if err != nil {
		return nil, err
	}

	candidatesByCell, minTotals, maxTotals, minPossible, maxPossible, err :=
		buildCandidates(pieces, rows, columns)
	if err != nil {
		return nil, err
	}

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

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	s := &searchState{
		initial:            initial,
		rows:               rows,
		columns:            columns,
		targets:            targets,
		values:             values,
		usedPieces:         make([]bool, len(pieces)),
		minTotals:          minTotals,
		maxTotals:          maxTotals,
		path:               make([]Placement, 0, len(pieces)),
		intermediateBoards: make([]Board, 0, len(pieces)),
		searchSteps:        make([]SearchStep, 0),
		results:            make([]*Solution, 0),
		boardRemMin:        boardMin,
		boardRemMax:        boardMax,
		pieceRemMin:        minPossible,
		pieceRemMax:        maxPossible,
	}

	s.runDFS(ctx, opts, rng, candidatesByCell)

	if len(s.results) == 0 {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, nil
	}
	return s.results, nil
}

// runDFS executes the iterative depth-first search, populating s.results.
func (s *searchState) runDFS(ctx context.Context, opts Options, rng *rand.Rand, candidatesByCell [][]candidatePlacement) {
	rootTargetIndex, rootOptions := findMostConstrained(s.values, s.targets, candidatesByCell, s.usedPieces)

	if rootTargetIndex == -1 {
		// Board already fully filled before any piece is placed.
		if allPiecesUsed(s.usedPieces) {
			s.results = append(s.results, &Solution{Board: s.currentBoard()})
		}
		return
	}
	if rootOptions == nil {
		return
	}

	if opts.Mode == ModeRandom {
		rng.Shuffle(len(rootOptions), func(i, j int) {
			rootOptions[i], rootOptions[j] = rootOptions[j], rootOptions[i]
		})
	}

	stack := []stackFrame{{options: rootOptions, idx: 0}}
	done := false

	for !done && len(stack) > 0 {
		top := &stack[len(stack)-1]

		if top.idx >= len(top.options) {
			// All candidates at this level tried — pop and undo the entry candidate.
			popped := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if popped.hasEntry {
				s.backtrack(popped.entryCandidate, popped.entryBDeltaMin, popped.entryBDeltaMax)
			}
			continue
		}

		if ctx.Err() != nil {
			done = true
			break
		}

		candidate := top.options[top.idx]
		top.idx++

		bDeltaMin, bDeltaMax := computeCandidateDelta(candidate.cells, s.targets)
		s.applyCandidate(candidate, bDeltaMin, bDeltaMax)

		if !s.isFeasible() {
			s.backtrack(candidate, bDeltaMin, bDeltaMax)
			continue
		}

		targetIndex, newOptions := findMostConstrained(s.values, s.targets, candidatesByCell, s.usedPieces)

		if targetIndex == -1 {
			// Every cell is filled — check that all pieces were used.
			if allPiecesUsed(s.usedPieces) {
				s.recordSolution()
				if reachedSolutionLimit(len(s.results), opts) {
					done = true
				}
			}
			s.backtrack(candidate, bDeltaMin, bDeltaMax)
			continue
		}

		if newOptions == nil {
			// Dead end: most-constrained cell has no valid candidates.
			s.backtrack(candidate, bDeltaMin, bDeltaMax)
			continue
		}

		if opts.Mode == ModeRandom {
			rng.Shuffle(len(newOptions), func(i, j int) {
				newOptions[i], newOptions[j] = newOptions[j], newOptions[i]
			})
		}
		stack = append(stack, stackFrame{
			options:        newOptions,
			idx:            0,
			hasEntry:       true,
			entryCandidate: candidate,
			entryBDeltaMin: bDeltaMin,
			entryBDeltaMax: bDeltaMax,
		})
	}
}
