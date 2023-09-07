package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iuiq/do/internal/tasks"
	"github.com/rivo/tview"
)

type TUI struct {
	app      *tview.Application
	pages    *tview.Pages
	mainGrid *tview.Grid

	leftPanel      *tview.Grid
	leftPanelWidth int
	rightPanel     *tview.Grid

	focusedPanel *tview.Grid

	list     *tview.Table
	showDesc bool
	taskData *tasks.TodoList `yaml:"todo_list"`
}

// Init intializes the tview app and sets up the UI.
func (tui *TUI) Init(tl *tasks.TodoList) {
	tui.app = tview.NewApplication()
	tui.taskData = tl

	/*
		// Create a new list for todo tasks
		tui.list = tview.NewList().SetSelectable(true, false)
	*/

	tui.list = tview.NewTable().
		SetSelectable(true, false)

	// Populate tui list with initial tasks
	tui.Populate()

	// Create the left-hand side panel
	tui.leftPanel = tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		AddItem(tui.list, 0, 0, 1, 1, 0, 0, true)
	tui.leftPanel.SetTitle(tui.taskData.GetTitle())
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

	width := 25

	// Create the main parent grid
	tui.mainGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(width, 1, 0).
		AddItem(tui.leftPanel, 0, 0, 1, 1, 0, 0, true).
		AddItem(line, 0, 1, 1, 1, 0, 0, false).
		AddItem(tui.rightPanel, 0, 2, 1, 1, 0, 0, false)

		// Remove two from width to account for panel borders
	tui.leftPanelWidth = width - 2

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

// calcTaskIdx returns the calculated task index in a given task slice.
// This function takes into account whether the description for each
// task is shown, which would occupy an extra row in the task list
// table.
func (t *TUI) calcTaskIdx(selectedRow, columnWidth int) int {
	taskIdx := 0

	// Iterate through each row up to the selected row
	for i := 0; i < selectedRow; i++ {
		// If the task description is being shown, skip the next row
		if t.taskData.GetTask(taskIdx).GetShowDesc() {
			wrappedDesc := WordWrap(t.taskData.GetTask(taskIdx).GetDescription(), columnWidth)
			i += len(wrappedDesc) // Skip the row(s) meant for task description
		}
		taskIdx++
	}
	return taskIdx
}

// filterAndUpdateList filters out past completed tasks, marks todays
// completed tasks, and updates the tview todo list.
//
// Note: Since the list is being cleared and repopulated, the curor will
// always return to the first item in the list. This would be desired
// when the functionality for moving completed tasks to the end of the
// list is implemented.
func (t *TUI) filterAndUpdateList(columnWidth int) {
	t.list.Clear()

	if len(t.taskData.GetTasks()) == 0 {
		t.list.SetCellSimple(0, 0, "No tasks available")
	}

	currentRow := 0

	// TODO: should probably firt sort the list by priority
	for _, task := range t.taskData.GetTasks() {
		prefix := "[ []"
		createdToday := task.GetStarted().Format("2006-01-02") == time.Now().Format("2006-01-02")
		// If task is completed and not created today, then don't add it to
		// the tview list.
		if task.GetIsDone() && !createdToday {
			continue
		}

		// If task if completed and created today, then mark it complete.
		if task.GetIsDone() && createdToday {
			prefix = "[x[]"
		}

		// Add task name to the list
		t.list.SetCellSimple(currentRow, 0, prefix+task.GetName())

		// If task show description status is set to true, add the task
		// description to the list.
		if task.GetShowDesc() {
			wrappedDesc := WordWrap(task.Description, columnWidth)
			for _, line := range wrappedDesc {
				currentRow++
				t.list.SetCell(currentRow, 0, tview.NewTableCell(line).
					SetAlign(tview.AlignLeft).
					SetSelectable(false).
					SetTextColor(tcell.ColorYellow))
			}
		}
		currentRow++
	}
}

// WordWrap returns a slice of wrapped lines given the text to the specified
// length, breaking at word boundaries.
func WordWrap(text string, lineWidth int) []string {
	// Break text into individual words, split by spaces.
	words := strings.Fields(text)

	// wrappedLines will hold the lines of text after they've been wrapped.
	wrappedLines := []string{}

	// currentLine holds the words for the line being constructed.
	currentLine := ""

	// Loop through each word in the words slice.
	for _, word := range words {
		// If adding the new word to the current line would make it too long,
		// append currentLine to wrappedLines and start a new line.
		if len(currentLine)+len(word)+1 > lineWidth {
			wrappedLines = append(wrappedLines, currentLine)
			currentLine = ""
		}

		// If the current line isn't empty, add a space before the new word.
		if len(currentLine) > 0 {
			currentLine += " "
		}

		// Append the new word to the current line.
		currentLine += word
	}

	// Append any remaining text in currentLine to wrappedLines.
	if len(currentLine) > 0 {
		wrappedLines = append(wrappedLines, currentLine)
	}

	// Return the slice of wrapped lines.
	return wrappedLines
}

func (t *TUI) Populate() {
	t.filterAndUpdateList(t.leftPanelWidth)
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
		/*
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
		*/
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
func (t *TUI) listInputCapture(event *tcell.EventKey) {
	t.listBoardInputCapture(event)

	row, _ := t.list.GetSelection()
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'a': // Create new task
			form := t.createListForm(t.calcTaskIdx(row, t.leftPanelWidth))
			t.showModal(form)
			return
		case 'e': // Edit task
			form := t.editListForm(t.calcTaskIdx(row, t.leftPanelWidth))
			t.showModal(form)
		case 'x': // Toggle task completion status
			task := t.taskData.GetTask(t.calcTaskIdx(row, t.leftPanelWidth))
			task.SetDone(!task.GetIsDone())
			// TODO: handle toggle start/finish date fields
			//task.SetFinished(time.Now()) // set done date to current date
			t.filterAndUpdateList(t.leftPanelWidth)
		case 'd': // Delete task
			// Delete task from task data slice
			task, err := t.taskData.Remove(t.calcTaskIdx(row, t.leftPanelWidth))
			if err != nil {
				return
			}
			// Buffer returned deleted task
			t.taskData.SetBuffer(task)

			// Update task priorities
			t.taskData.UpdatePriorities(t.calcTaskIdx(row, t.leftPanelWidth))

			// Update tview todo list
			t.filterAndUpdateList(t.leftPanelWidth)
		case 'p': // Paste task
			currentIdx := t.calcTaskIdx(row, t.leftPanelWidth)
			// Read from buffer
			task := t.taskData.Buffer()

			// Add task to the todo list
			t.taskData.Add(task, currentIdx+1)

			// Update task priorities
			t.taskData.UpdatePriorities(currentIdx + 1)

			// Update tview todo list
			t.filterAndUpdateList(t.leftPanelWidth)
		case ' ': // Toggle task description
			selectedRow, _ := t.list.GetSelection()
			currentIdx := t.calcTaskIdx(selectedRow, t.leftPanelWidth)
			// If calculated task index is within bounds, toggle task show
			// description status, and update rendered list.
			if err := t.taskData.Bounds(currentIdx); err != nil {
				return
			}
			task := t.taskData.GetTask(currentIdx)
			task.ShowDesc = !task.ShowDesc
			t.filterAndUpdateList(t.leftPanelWidth)
		}
	}
}

// listBoardInputCapture captures shared input interactions between
// the list and kanban boards.
func (t *TUI) listBoardInputCapture(event *tcell.EventKey) {
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		}
	}
}

// showModal sets up a modal grid for the given form and displays it.
func (t *TUI) showModal(form *tview.Form) {
	// Returns a new primitive which puts the provided primitive in the center and
	// sets its size to the given width and height.
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}

	m := modal(form, 40, 30)
	t.pages.AddPage("modal", m, true, true)
	t.app.SetFocus(m)
}

// createListForm creates and returns a tview form for creating a new
// todo list task.
func (t *TUI) createListForm(currentIdx int) *tview.Form {
	var name, description string
	var isCore bool

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Create New Task")

	form.AddInputField("Name", "", 20, nil, func(text string) {
		name = text
	})
	form.AddInputField("Description", "", 50, nil, func(text string) {
		description = text
	})
	form.AddCheckbox("Is Core Task", false, func(checked bool) {
		isCore = checked
	})

	form.AddButton("Save", func() {
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
		t.filterAndUpdateList(t.leftPanelWidth)

		t.closeModal()
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
	})

	return form
}

// editListForm creates and returns a tview form for creating a new
// todo list task.
func (t *TUI) editListForm(currentIdx int) *tview.Form {
	tasks := t.taskData.GetTasks()
	task := &tasks[currentIdx]
	name := task.GetName()
	description := task.GetDescription()
	isCore := task.GetIsCore()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Edit Task")

	// Define the input fields for the forms and update field variables if
	// user makes any changes to the default values.
	form.AddInputField("Name", task.GetName(), 20, nil, func(text string) {
		name = text
	})
	form.AddInputField("Description", task.GetDescription(), 50, nil, func(text string) {
		description = text
	})
	form.AddCheckbox("Is Core Task", task.GetIsCore(), func(checked bool) {
		isCore = checked
	})

	form.AddButton("Save", func() {
		t.closeModal()

		// Update task in data slice
		task.SetName(name)
		task.SetDescription(description)
		task.SetCore(isCore)

		// Update tview list
		t.filterAndUpdateList(t.leftPanelWidth)
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
	})

	return form
}

// closeModal removes that modal page and sets the focus back to the
// main grid.
func (t *TUI) closeModal() {
	t.pages.RemovePage("modal")
	t.app.SetFocus(t.mainGrid)
}
