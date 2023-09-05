package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iuiq/do/internal/tasks"
	"github.com/rivo/tview"
)

type TUI struct {
	app      *tview.Application
	pages    *tview.Pages
	mainGrid *tview.Grid

	leftPanel  *tview.Grid
	rightPanel *tview.Grid

	focusedPanel *tview.Grid

	list     *tview.List
	taskData *tasks.TodoList

	listTaskForm *tview.Form
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

	// Add the main grid to page
	tui.pages = tview.NewPages().
		AddPage("main", tui.mainGrid, true, true)

	// Set input handling
	tui.app.SetInputCapture(tui.setupInputCapture())

	if err := tui.app.SetRoot(tui.pages, true).Run(); err != nil {
		panic(err)
	}
}

// filterAndUpdateList filters out past completed tasks, marks todays
// completed tasks, and updates the tview todo list.
//
// Note: Since the list is being cleared and repopulated, the curor will
// always return to the first item in the list. This would be desired
// when the functionality for moving completed tasks to the end of the
// list is implemented.
func (tui *TUI) filterAndUpdateList() {
	tui.list.Clear()
	for _, task := range tui.taskData.Tasks() {
		prefix := "[ []"
		createdToday := task.Started().Format("2006-01-02") == time.Now().Format("2006-01-02")
		// If task is completed and not created today, then don't add it to
		// the tview list.
		if task.IsDone() && !createdToday {
			continue
		}

		// If task if completed and created today, then mark it complete.
		if task.IsDone() && createdToday {
			prefix = "[x[]"
		}

		// Add a task to the list
		tui.list.AddItem(fmt.Sprintf("%s %s", prefix, task.Name()), "", 0, nil)
	}
}

func (tui *TUI) Populate() {
	tui.filterAndUpdateList()
}

// setupInputCapture sets up input capturing for the application.
func (tui *TUI) setupInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		tasks := tui.taskData.Tasks()
		task := tasks[tui.list.GetCurrentItem()]

		// Global input handling
		if event = tui.globalInputCapture(event, &task); event == nil {
			// Override the entered input by absorbing the event, stop it from
			// propogating further.
			return nil
		}

		// Context-specific input handling
		switch tui.app.GetFocus() {
		case tui.list:
			tui.listInputCapture(event, &task)
			// TODO: case tui.boards:
			// TODO: case tui.tree:
		}

		return event
	}
}

// globalInputCapture captures input interactions across all displayed
// tview primatives.
func (tui *TUI) globalInputCapture(event *tcell.EventKey, task *tasks.TodoTask) *tcell.EventKey {
	// If tview primatives for user input are currently focused, ignore
	// any global input captures. This prevents applicaton side effects.
	// For example, this allows the user to type "q" in an input field
	// without quiting the application.
	switch tui.app.GetFocus().(type) {
	case *tview.InputField, *tview.DropDown, *tview.Checkbox, *tview.Button:
		return event
	}

	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j':
			// Account for wrap around when moving down. For some reason, upward
			// movement using k is already accounted for.
			total := tui.list.GetItemCount()
			idx := tui.list.GetCurrentItem() + 1
			if idx >= total {
				idx = 0
			}
			tui.list.SetCurrentItem(idx)
		case 'k': // Upward movement
			currentIdx := tui.list.GetCurrentItem()
			tui.list.SetCurrentItem(currentIdx - 1)
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
func (tui *TUI) listInputCapture(event *tcell.EventKey, task *tasks.TodoTask) {
	tui.listBoardInputCapture(event)

	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'a': // Create new task
			// Create a new modal and add it to pages.
			modal := tui.createListFormModal(tui.list.GetCurrentItem())
			tui.pages.AddPage("modal", modal, true, true)
			tui.app.SetFocus(modal)
			return
		case 'x': // Toggle task completion status
			task.SetDone(!task.IsDone())
			task.SetFinished(time.Now()) // set done date to current date
			tui.filterAndUpdateList()
		}
	}
}

// listBoardInputCapture captures shared input interactions between
// the list and kanban boards.
func (tui *TUI) listBoardInputCapture(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'e': // Edit task
		case 'd': // Delete task
		case 'p': // Paste task
		case ' ': // Open task details or board
		}
	}
}

// createListFormModal creates a tview form and returns a tview modal
// containing the form.
func (t *TUI) createListFormModal(currentIdx int) tview.Primitive {
	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}

	var name, description string
	var isCore bool

	t.listTaskForm = tview.NewForm()
	t.listTaskForm.SetBorder(true)
	t.listTaskForm.SetTitle("Create New Task")

	t.listTaskForm.AddInputField("Name", "", 20, nil, func(text string) {
		name = text
	})
	t.listTaskForm.AddInputField("Description", "", 50, nil, func(text string) {
		description = text
	})
	t.listTaskForm.AddCheckbox("Is Core Task", false, func(checked bool) {
		isCore = checked
	})

	t.listTaskForm.AddButton("Save", func() {
		t.closeModal()
		// Add task to task data slice
		task := new(tasks.TodoTask)
		task.SetTask(new(tasks.Task))
		task.SetStarted(time.Now())
		task.SetName(name)
		task.SetDescription(description)
		task.SetCore(isCore)
		task.SetPriority(currentIdx + 1)
		t.taskData.Add(task, currentIdx+1)
		t.taskData.UpdatePriorities(currentIdx + 1)

		// Update tview list
		t.filterAndUpdateList()
	})

	t.listTaskForm.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
	})

	return modal(t.listTaskForm, 40, 30)
}

// closeModal removes that modal page and sets the focus back to the
// main grid.
func (t *TUI) closeModal() {
	t.pages.RemovePage("modal")
	t.app.SetFocus(t.mainGrid)
}
