// Package tui implements an interactive terminal UI for choosing a board,
// solving it with a chosen strategy, and watching the full search process
// — including every backtrack — animated step by step.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/borko-rajkovic/smart_games_puzzle/app/board"
	"github.com/borko-rajkovic/smart_games_puzzle/app/piece"
)

type screen int

const (
	screenBoardSelect screen = iota
	screenModeSelect
	screenSolving
	screenSolutionList
	screenAnimate
)

// menuItem is a simple list.Item used for every menu in the TUI.
type menuItem struct {
	title string
	desc  string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

// Model is the root Bubble Tea model driving the whole application.
type Model struct {
	screen screen

	boardList    list.Model
	modeList     list.Model
	solutionList list.Model
	spinner      spinner.Model

	width, height int

	boardName    string
	initialBoard board.Board

	solveMode     string
	solveCancel   context.CancelFunc
	solveStart    time.Time
	statusMessage string

	solutions        []*board.Solution
	selectedSolution *board.Solution

	// stepIndex is the current position in the search-step trace.
	// 0 = initial board; 1..len(SearchSteps) = after that many search operations.
	stepIndex int
	playing   bool
	speed     time.Duration

	quitting bool
}

// NewModel builds the initial Model, ready to show the board-selection screen.
func NewModel() Model {
	boardItems := []list.Item{
		menuItem{title: "Flat board", desc: "A plain 5x6 rectangular board"},
		menuItem{title: "Heart-shaped board", desc: "A 6x6 board carved into a heart"},
	}
	modeItems := []list.Item{
		menuItem{title: "First solution", desc: "Stop as soon as a solution is found (fastest)"},
		menuItem{title: "Random solution", desc: "Find a randomized valid solution"},
		menuItem{
			title: "All solutions",
			desc:  fmt.Sprintf("Enumerate up to %d solutions (may take a while; esc cancels early)", allSolutionsLimit),
		},
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	return Model{
		screen:       screenBoardSelect,
		boardList:    newMenuList("Choose a board to solve", boardItems),
		modeList:     newMenuList("Choose a solve mode", modeItems),
		solutionList: newMenuList("Choose a solution to watch", nil),
		spinner:      sp,
		speed:        defaultSpeed,
	}
}

func newMenuList(title string, items []list.Item) list.Model {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = listTitleStyle
	return l
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		listHeight := msg.Height - 4
		m.boardList.SetSize(msg.Width, listHeight)
		m.modeList.SetSize(msg.Width, listHeight)
		m.solutionList.SetSize(msg.Width, listHeight)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.solveCancel != nil {
				m.solveCancel()
			}
			m.quitting = true
			return m, tea.Quit
		}
		switch m.screen {
		case screenBoardSelect:
			return m.updateBoardSelect(msg)
		case screenModeSelect:
			return m.updateModeSelect(msg)
		case screenSolving:
			return m.updateSolving(msg)
		case screenSolutionList:
			return m.updateSolutionList(msg)
		case screenAnimate:
			return m.updateAnimate(msg)
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case solveResultMsg:
		return m.handleSolveResult(msg)

	case animTickMsg:
		return m.handleAnimTick()
	}
	return m, nil
}

func (m Model) updateBoardSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		item, ok := m.boardList.SelectedItem().(menuItem)
		if !ok {
			return m, nil
		}
		switch item.title {
		case "Flat board":
			m.boardName, m.initialBoard = item.title, board.FlatBoard
		case "Heart-shaped board":
			m.boardName, m.initialBoard = item.title, board.HeartBoard
		}
		m.statusMessage = ""
		m.screen = screenModeSelect
		return m, nil
	case "q":
		m.quitting = true
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.boardList, cmd = m.boardList.Update(msg)
	return m, cmd
}

func (m Model) updateModeSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		item, ok := m.modeList.SelectedItem().(menuItem)
		if !ok {
			return m, nil
		}
		var mode board.Mode
		limit := 0
		switch item.title {
		case "First solution":
			mode = board.ModeFirst
		case "Random solution":
			mode = board.ModeRandom
		case "All solutions":
			mode, limit = board.ModeAll, allSolutionsLimit
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.solveCancel = cancel
		m.solveMode = item.title
		m.solveStart = time.Now()
		m.screen = screenSolving
		return m, tea.Batch(solveCmd(ctx, m.initialBoard, mode, limit), m.spinner.Tick)
	case "esc":
		m.screen = screenBoardSelect
		return m, nil
	case "q":
		m.quitting = true
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.modeList, cmd = m.modeList.Update(msg)
	return m, cmd
}

func (m Model) updateSolving(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		if m.solveCancel != nil {
			m.solveCancel()
		}
	}
	return m, nil
}

func (m Model) updateSolutionList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		index := m.solutionList.Index()
		if index < 0 || index >= len(m.solutions) {
			return m, nil
		}
		m.selectedSolution = m.solutions[index]
		return m.startAnimation()
	case "esc":
		m.screen = screenModeSelect
		return m, nil
	case "q":
		m.quitting = true
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.solutionList, cmd = m.solutionList.Update(msg)
	return m, cmd
}

func (m Model) updateAnimate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	total := len(m.selectedSolution.SearchSteps)
	switch msg.String() {
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.playing = false
		if len(m.solutions) > 1 {
			m.screen = screenSolutionList
		} else {
			m.screen = screenModeSelect
		}
		return m, nil
	case " ":
		m.playing = !m.playing
		if m.playing && m.stepIndex < total {
			return m, animTickCmd(m.speed)
		}
		return m, nil
	case "right", "l", "n":
		m.playing = false
		if m.stepIndex < total {
			m.stepIndex++
		}
		return m, nil
	case "left", "h", "p":
		m.playing = false
		if m.stepIndex > 0 {
			m.stepIndex--
		}
		return m, nil
	case "r":
		return m.startAnimation()
	case "+", "=":
		m.speed = max(m.speed/2, minSpeed)
		return m, nil
	case "-", "_":
		m.speed = min(m.speed*2, maxSpeed)
		return m, nil
	}
	return m, nil
}

func (m Model) handleSolveResult(msg solveResultMsg) (tea.Model, tea.Cmd) {
	m.solveCancel = nil

	if len(msg.solutions) == 0 {
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Search stopped: %v", msg.err)
		} else {
			m.statusMessage = "No solution exists for this board."
		}
		m.screen = screenModeSelect
		return m, nil
	}

	m.solutions = msg.solutions
	m.statusMessage = ""

	if len(msg.solutions) == 1 {
		m.selectedSolution = msg.solutions[0]
		return m.startAnimation()
	}

	items := make([]list.Item, len(msg.solutions))
	for i, solution := range msg.solutions {
		items[i] = menuItem{
			title: fmt.Sprintf("Solution %d", i+1),
			desc: fmt.Sprintf("%d placements, %d search steps",
				len(solution.Placements), len(solution.SearchSteps)),
		}
	}
	m.solutionList.SetItems(items)
	m.screen = screenSolutionList
	return m, nil
}

func (m Model) startAnimation() (tea.Model, tea.Cmd) {
	m.stepIndex = 0
	m.playing = true
	m.screen = screenAnimate
	return m, animTickCmd(m.speed)
}

func (m Model) handleAnimTick() (tea.Model, tea.Cmd) {
	if !m.playing || m.selectedSolution == nil {
		return m, nil
	}
	total := len(m.selectedSolution.SearchSteps)
	if m.stepIndex >= total {
		m.playing = false
		return m, nil
	}
	m.stepIndex++
	if m.stepIndex >= total {
		m.playing = false
		return m, nil
	}
	return m, animTickCmd(m.speed)
}

// currentBoard returns the board state to display at the current animation step.
// Step 0 shows the initial (empty) board; step k shows the state after the
// k-th search operation (place or backtrack).
func (m Model) currentBoard() board.Board {
	if m.selectedSolution == nil || m.stepIndex == 0 {
		return m.initialBoard
	}
	steps := m.selectedSolution.SearchSteps
	idx := m.stepIndex - 1
	if idx >= len(steps) {
		idx = len(steps) - 1
	}
	return steps[idx].Board
}

// placementsUpTo returns the placements that are active at the current step,
// so the board renderer can colour cells by piece.
func (m Model) placementsUpTo() []board.Placement {
	if m.selectedSolution == nil || m.stepIndex == 0 {
		return nil
	}
	steps := m.selectedSolution.SearchSteps
	idx := m.stepIndex - 1
	if idx >= len(steps) {
		idx = len(steps) - 1
	}
	// Replay forward from step 0 to collect the active placement set.
	active := make([]board.Placement, 0, len(m.selectedSolution.Placements))
	for i := 0; i <= idx; i++ {
		s := steps[i]
		if s.Kind == board.StepPlace {
			active = append(active, s.Placement)
		} else {
			// Remove the last placement of this piece.
			for j := len(active) - 1; j >= 0; j-- {
				if active[j].Piece.Color == s.Placement.Piece.Color {
					active = append(active[:j], active[j+1:]...)
					break
				}
			}
		}
	}
	return active
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case screenBoardSelect:
		return appStyle.Render(m.boardList.View())
	case screenModeSelect:
		view := m.modeList.View()
		if m.statusMessage != "" {
			view = statusErrorStyle.Render(m.statusMessage) + "\n\n" + view
		}
		return appStyle.Render(view)
	case screenSolving:
		return appStyle.Render(m.renderSolving())
	case screenSolutionList:
		return appStyle.Render(m.solutionList.View())
	case screenAnimate:
		return appStyle.Render(m.renderAnimate())
	}
	return ""
}

func (m Model) renderSolving() string {
	elapsed := time.Since(m.solveStart).Round(time.Second)
	return fmt.Sprintf(
		"%s Solving %s (%s)...\nElapsed: %s\n\n%s",
		m.spinner.View(),
		m.boardName,
		m.solveMode,
		elapsed,
		helpStyle.Render("esc: stop and use whatever was found so far   q: quit"),
	)
}

func (m Model) renderAnimate() string {
	solution := m.selectedSolution
	steps := solution.SearchSteps
	total := len(steps)

	header := titleStyle.Render(fmt.Sprintf("%s — %s", m.boardName, m.solveMode))

	// Step counter and current action label.
	var stepLabel, actionLine string
	if m.stepIndex == 0 {
		stepLabel = "Ready"
		actionLine = helpStyle.Render("Press space to start the search replay")
	} else {
		s := steps[m.stepIndex-1]
		if s.Kind == board.StepPlace {
			stepLabel = placeStyle.Render("▶ Place")
			actionLine = fmt.Sprintf("Placing %s at row %d, col %d (variation %d)",
				s.Placement.Piece.Color, s.Placement.Row, s.Placement.Column,
				s.Placement.VariationIndex+1)
		} else {
			stepLabel = backtrackStyle.Render("◀ Backtrack")
			actionLine = fmt.Sprintf("Removing %s — branch failed", s.Placement.Piece.Color)
		}
		if m.stepIndex >= total {
			stepLabel = placeStyle.Render("✓ Solved")
			actionLine = fmt.Sprintf("Solution found in %s search steps", stepCountStyle.Render(fmt.Sprintf("%d", total)))
		}
	}

	playState := "Paused"
	if m.playing {
		playState = "Playing"
	}

	progress := progressStyle.Render(fmt.Sprintf(
		"Search step %d / %d   %s   %s",
		m.stepIndex, total, stepLabel, playState,
	))

	placements := m.placementsUpTo()
	boardStr := RenderBoard(m.currentBoard(), placements, len(placements))

	help := helpStyle.Render(
		"space: play/pause   ←/→: step   r: restart   +/-: speed   esc: back   q: quit",
	)

	sections := []string{
		header,
		progress,
		"",
		boardStr,
		"",
		actionLine,
		"",
		Legend(piece.Pieces),
		"",
		help,
	}
	return strings.Join(sections, "\n")
}
