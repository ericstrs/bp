package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/iuiq/do/internal/tasks"
	"github.com/rivo/tview"
)

type TUI struct {
	list   *tview.List
	boards *tview.Grid
	grid   *tview.Grid
	app    *tview.Application
}

// Init intializes the tview app and sets up the UI.
func (tui *TUI) Init(tl *tasks.TodoList) {
	tui.app = tview.NewApplication()

	// Create a new list for todo tasks
	tui.list = tview.NewList().
		ShowSecondaryText(false)

	// Populate tui list with initial tasks
	tui.Populate(tl)

	/*
		// Create a grid for kanban boards
		tui.boards = tview.NewGrid().
			SetRows(0).
			SetColumns(10, 0, 10).
			// Placeholder cells, there will eventually contain the actual Kanban
			// boards.
			AddItem(tview.NewTextView().SetText("Col 1"), 0, 0, 1, 1, 0, 0, false).
			AddItem(tview.NewTextView().SetText("Col 2"), 0, 2, 1, 1, 0, 0, false)
	*/

	// Create right-hand side
	rightPanel := tview.NewBox().SetBorder(true).SetTitle("Right Panel")

	// Create line to separate the todo list and the kanban boards
	line := tview.NewBox().
		SetBackgroundColor(tcell.ColorWhite)

	// Create the main parent grid
	tui.grid = tview.NewGrid().
		SetRows(0).
		SetColumns(30, 1, 0).
		AddItem(tui.list, 0, 0, 1, 1, 0, 0, true).
		AddItem(line, 0, 1, 1, 1, 0, 0, false).
		//AddItem(tui.boards, 0, 2, 1, 1, 0, 0, false)
		AddItem(rightPanel, 0, 2, 1, 1, 0, 0, false)

	// Set input handling
	tui.grid.SetInputCapture(tui.setupInputCapture(tl))

	if err := tui.app.SetRoot(tui.grid, true).Run(); err != nil {
		panic(err)
	}
}

func (tui *TUI) Populate(tl *tasks.TodoList) {
	tui.list.Clear()
	for _, task := range tl.Tasks() {
		// Add a task to the list.
		tui.list.AddItem(task.Name(), "", 0, nil)
	}
}

func (tui *TUI) setupInputCapture(tl *tasks.TodoList) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab: // Change focus
		case tcell.KeyRune:
			switch event.Rune() {
			case 'a': // Create new task
				// TODO: code to run whenever "a" is pressed
				fmt.Println("Running to code to handle adding a new task...")
			case 'j': // Move to next task
			case 'k': // Move to previous task
			case 'e': // Edit task
			case ' ': // Open task details or board
			case 'd': // Delete task
			case 'p': // Paste task
			case 'q': // Quit the program
				tui.app.Stop()
			}
		}
		return event
	}
}
