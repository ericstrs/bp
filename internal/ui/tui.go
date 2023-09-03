package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/iuiq/do/internal/tasks"
	"github.com/rivo/tview"
)

type TUI struct {
	app      *tview.Application
	mainGrid *tview.Grid

	leftPanel  *tview.Grid
	rightPanel *tview.Grid

	focusedPanel *tview.Grid

	list     *tview.List
	taskData *tasks.TodoList
}

// Init intializes the tview app and sets up the UI.
func (tui *TUI) Init(tl *tasks.TodoList) {
	tui.app = tview.NewApplication()
	tui.taskData = tl

	// Create a new list for todo tasks
	tui.list = tview.NewList().
		ShowSecondaryText(false)

	// Populate tui list with initial tasks
	tui.Populate()

	// Create the left-hand side panel
	tui.leftPanel = tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		AddItem(tui.list, 0, 0, 1, 1, 0, 0, true)
	tui.leftPanel.SetTitle("Daily TODOs")
	tui.leftPanel.SetBorder(true)

	// Initialize the right panel with a simple text view
	rightTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetText("This is the right panel")

	// Create right-hand side panel
	tui.rightPanel = tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		AddItem(rightTextView, 0, 0, 1, 1, 0, 0, false)
	tui.rightPanel.SetTitle("Right Panel")
	tui.rightPanel.SetBorder(false)

	// Create line to separate the todo list and the kanban boards
	line := tview.NewBox().
		SetBackgroundColor(tcell.ColorWhite)

	// Create the main parent grid
	tui.mainGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(20, 1, 0).
		AddItem(tui.leftPanel, 0, 0, 1, 1, 0, 0, true).
		AddItem(line, 0, 1, 1, 1, 0, 0, false).
		AddItem(tui.rightPanel, 0, 2, 1, 1, 0, 0, false)

	// Initialize panel focus to left panel
	tui.focusedPanel = tui.leftPanel

	// Set input handling
	tui.app.SetInputCapture(tui.setupInputCapture())

	if err := tui.app.SetRoot(tui.mainGrid, true).Run(); err != nil {
		panic(err)
	}
}

// TODO: func (tui *TUI) updateTodoList()

func (tui *TUI) Populate() {
	tui.list.Clear()
	// TODO: prepend [] and fill it in if the task is already completed
	for _, task := range tui.taskData.Tasks() {
		// Add a task to the list.
		tui.list.AddItem(task.Name(), "", 0, nil)
	}
}

// setupInputCapture sets up input capturing for the application.
func (tui *TUI) setupInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		// Global input handling
		if event = tui.globalInputCapture(event); event == nil {
			// Override the entered input by absorbing the event, stop it from
			// propogating further.
			return nil
		}

		// Context-specific input handling
		switch tui.app.GetFocus() {
		case tui.list:
			tui.listInputCapture(event)
			// TODO: case tui.boards:
			// TODO: case tui.tree:
		}

		return event
	}
}

// globalInputCapture captures input interactions across all displayed
// tview primatives.
func (tui *TUI) globalInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j': // Downward movement
		case 'k': // Upward movement
		case 'q': // Quit the program
			tui.app.Stop()
		}
	case tcell.KeyTab: // Switch panel focus
		switch tui.focusedPanel {
		case tui.leftPanel: // Switch focus to right panel
			tui.focusedPanel = tui.rightPanel
			tui.app.SetFocus(tui.rightPanel)
			tui.leftPanel.SetBorder(false)
			tui.rightPanel.SetBorder(true)
		case tui.rightPanel: // Switch focus to left panel
			tui.focusedPanel = tui.leftPanel
			tui.app.SetFocus(tui.leftPanel)
			tui.rightPanel.SetBorder(false)
			tui.leftPanel.SetBorder(true)
		}
		return nil // Override the tab key
	}
	return event
}

// listInputCapture captures input interactions specific to the list.
func (tui *TUI) listInputCapture(event *tcell.EventKey) {
	tui.listBoardInputCapture(event)

	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'x': // Toggle task completion status
		}
	}
}

// listBoardInputCapture captures shared input interactions between
// the list and kanban boards.
func (tui *TUI) listBoardInputCapture(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'a': // Create new task
			// TODO: code to run whenever "a" is pressed
			fmt.Println("Running to code to handle adding a new task...")
		case 'e': // Edit task
		case ' ': // Open task details or board
		case 'd': // Delete task
		case 'p': // Paste task
		}
	}
}
