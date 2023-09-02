package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/iuiq/do/internal/tasks"
	"github.com/rivo/tview"
)

type TUI struct {
	list *tview.List
	app  *tview.Application
}

// Init intializes the tview app and sets up the UI.
func (tui *TUI) Init(tl *tasks.TodoList) {
	tui.app = tview.NewApplication()

	tui.list = tview.NewList().
		ShowSecondaryText(false)

	// Populate tui list with initial tasks
	tui.Populate(tl)

	// Set up user input handling
	tui.list.SetInputCapture(tui.setupInputCapture(tl))

	if err := tui.app.SetRoot(tui.list, true).Run(); err != nil {
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
		case tcell.KeyRune:
			switch event.Rune() {
			case 'a':
				// TODO: code to run whenever "a" is pressed
				fmt.Println("Running to code to handle adding a new task...")
			case 'q':
				tui.app.Stop()
			}
		}
		return event
	}
}
