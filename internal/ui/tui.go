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
	focusedPanel   *tview.Grid

	list     *tview.Table
	taskData *tasks.TodoList

	tree                      *tview.TreeView
	treeData                  []tasks.Board
	board                     *tview.Grid
	boardColumns              []*tview.Table
	boardColumnsData          []tasks.BoardColumn
	focusedColumn             int
	isNoColumnsTableDisplayed bool
}

// Init intializes the tview app and sets up the UI.
func (t *TUI) Init(tl *tasks.TodoList, b []tasks.Board) {
	t.app = tview.NewApplication()
	t.taskData = tl
	t.treeData = b

	width := 25
	// Remove two from width to account for panel borders
	t.leftPanelWidth = width - 2

	t.list = tview.NewTable().
		SetSelectable(true, false)

	t.board = tview.NewGrid().
		SetRows(0).
		SetColumns(0)

	t.boardInputCapture()

	root := tview.NewTreeNode("Board Trees")

	t.tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	t.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'l' {
			node := t.tree.GetCurrentNode()
			ref := node.GetReference()
			board, ok := ref.(tasks.Board)

			if !ok {
				return event
			}

			t.showBoard(&board)
		}
		return event
	})

	/*
		t.tree.SetSelectedFunc(func(node *tview.TreeNode) {
			// TODO: (?)
		})
	*/

	// Populate tui list with initial tasks
	t.Populate()

	// Create the left-hand side panel
	t.leftPanel = tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		AddItem(t.list, 0, 0, 1, 1, 0, 0, true)
	t.leftPanel.SetTitle(t.taskData.GetTitle())
	t.leftPanel.SetBorder(true)

	// Create right-hand side panel
	t.rightPanel = tview.NewGrid().
		SetRows(0).
		SetColumns(0)
		//AddItem(t.tree, 0, 0, 1, 1, 0, 0, true)
	t.rightPanel.SetTitle("Tree Navigation")
	t.rightPanel.SetBorder(false)
	t.showTreeView()

	// Create line to separate the todo list and the kanban boards
	line := tview.NewBox().
		SetBackgroundColor(tcell.ColorWhite)

	// Create the main parent grid
	t.mainGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(-1, 1, -4).
		AddItem(t.leftPanel, 0, 0, 1, 1, 0, 0, true).
		AddItem(line, 0, 1, 1, 1, 0, 0, false).
		AddItem(t.rightPanel, 0, 2, 1, 1, 0, 0, true)

	// Initialize panel focus to left panel
	t.focusedPanel = t.leftPanel

	// Add the main grid to page
	t.pages = tview.NewPages().
		AddPage("main", t.mainGrid, true, true)

	// Set input handling
	t.app.SetInputCapture(t.setupInputCapture())

	// Update left and right panel size before drawing. This won't affect
	// the current drawing, it sets the panel width variables for the next
	// draw operation.
	t.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		t.leftPanelWidth = int(float64(width)*0.2) - 2
		return false
	})

	if err := t.app.SetRoot(t.pages, true).Run(); err != nil {
		panic(err)
	}
}

// showBoard clears the right panel and sets the board.
func (t *TUI) showBoard(b *tasks.Board) {
	t.rightPanel.Clear()
	t.boardColumns = nil // Reset columns
	t.boardColumnsData = b.GetColumns()

	if len(t.boardColumnsData) == 0 {
		t.rightPanel.Clear()
		t.rightPanel.SetTitle("No Columns")
		noColumnTable := tview.NewTable()
		cell := tview.NewTableCell("This board has no columns.")

		noColumnTable.SetCell(0, 0, cell)

		t.board.AddItem(noColumnTable, 0, 0, 1, 1, 0, 0, false)
		t.rightPanel.AddItem(t.board, 0, 0, 1, 1, 0, 0, true)
		t.app.SetFocus(t.rightPanel)
		t.isNoColumnsTableDisplayed = true
		return
	}
	t.isNoColumnsTableDisplayed = false

	t.rightPanel.SetTitle(b.GetTitle())

	// Loop over all the columns in the board
	for i, column := range t.boardColumnsData {
		// Need to create the kanban board column equivalent of a todo list
		table := tview.NewTable().
			SetSelectable(false, false) // No selection by default
		table.SetBorder(true)
		table.SetTitle(column.GetTitle())

		t.boardColumns = append(t.boardColumns, table)
		t.updateColumn(i)
		t.board.AddItem(table, 0, i, 1, 1, 0, 0, false)
	}

	// Set right panel content to the board grid. This will override the
	// tree view being displayed.
	t.rightPanel.AddItem(t.board, 0, 0, 1, 1, 0, 0, false)

	t.focusedColumn = 0

	// Assert focus on the right panel. This is needed for board input
	// capture to work.
	t.app.SetFocus(t.boardColumns[0])
}

// updateColumn clears the focused column of the board and updates the
// focused column's contents.
func (t *TUI) updateColumn(colIdx int) {
	t.boardColumns[colIdx].Clear()

	if len(t.boardColumnsData[t.focusedColumn].GetTasks()) == 0 {
		t.boardColumns[colIdx].SetCellSimple(0, 0, "No tasks available")
		return
	}

	currentRow := 0
	for _, task := range t.boardColumnsData[t.focusedColumn].GetTasks() {
		t.boardColumns[colIdx].SetCellSimple(currentRow, 0, task.Name)

		// If task show description status is set to true, add the task
		// description to the list.
		if task.GetShowDesc() {
			// TODO: find righ panel width
			wrappedDesc := WordWrap(task.Description, 50)
			for _, line := range wrappedDesc {
				currentRow++
				t.boardColumns[colIdx].SetCell(currentRow, 0, tview.NewTableCell(line).
					SetAlign(tview.AlignLeft).
					SetSelectable(false).
					SetTextColor(tcell.ColorYellow))
			}
		}
		currentRow++
	}
}

// showTreeView clears the right panel and sets the tree view.
func (t *TUI) showTreeView() {
	t.rightPanel.SetTitle("Tree Navigation")
	t.rightPanel.Clear()
	// Set right panel content to the tree view. This will override a
	// board being dislayed.
	t.rightPanel.AddItem(t.tree, 0, 0, 1, 1, 0, 0, true)

	// Assert focus on the right panel. This is needed for tree input
	// capture to work.
	t.app.SetFocus(t.rightPanel)
}

// calcTaskIdx returns the calculated task index in a given task slice.
// This function takes into account whether the description for each
// task is shown, which would occupy an extra row in the task list
// table.
func (t *TUI) calcTaskIdx(selectedRow, columnWidth int) int {
	taskIdx := 0

	// Iterate through each row up to the selected row
	for i := 0; i < selectedRow; i++ {
		task := t.taskData.GetTask(taskIdx)

		// If the task description is being shown, skip the next row
		if task.GetShowDesc() {
			wrappedDesc := WordWrap(t.taskData.GetTask(taskIdx).GetDescription(), columnWidth)
			i += len(wrappedDesc) // Skip the row(s) meant for task description
		}
		taskIdx++
	}
	return taskIdx
}

// calcTaskIdxBoard returns the calculated task index in a given board
// column. This function takes into account whether the description for each
// task is shown, which would occupy an extra row in the column table.
func (t *TUI) calcTaskIdxBoard(selectedRow, columnWidth int) int {
	taskIdx := 0

	// Iterate through each row up to the selected row
	for i := 0; i < selectedRow; i++ {
		task := t.boardColumnsData[t.focusedColumn].GetTask(taskIdx)

		// If the task description is being shown, skip the next row
		if task.GetShowDesc() {
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
		return
	}

	// TODO: should probably firt sort the list by priority

	currentRow := 0
	for _, task := range t.taskData.GetTasks() {
		prefix := "[ []"

		// If task if completed, then mark it complete.
		if task.GetIsDone() {
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

// Populate takes the the tasks from the task list and populates the
// list that will be displayed to the user.
// This function assumes that the left panel width field has been set
// for the tui struct.
func (t *TUI) Populate() {
	t.filterAndUpdateList(t.leftPanelWidth)

	// For each rooted board tree, attach to root node.
	for _, board := range t.treeData {
		boardNode := tview.NewTreeNode(board.GetTitle()).
			SetReference(board).
			SetSelectable(true)
		t.tree.GetRoot().AddChild(boardNode)

		updateTree(boardNode, board)
	}
}

// updateTree recursively updates the entire tview tree view primitive
// given the board structure.
func updateTree(node *tview.TreeNode, board tasks.Board) {
	for _, column := range board.GetColumns() {
		columnNode := tview.NewTreeNode(column.GetTitle()).
			SetReference(column).
			SetSelectable(true)
		node.AddChild(columnNode)

		for _, task := range column.GetTasks() {
			taskNode := tview.NewTreeNode(task.GetName()).
				SetReference(task).
				SetSelectable(true)
			columnNode.AddChild(taskNode)

			if task.GetHasChild() {
				childBoard := task.GetChild()
				/*
					childNode := tview.NewTreeNode(childBoard.GetTitle()).
						SetReference(childBoard).
						SetSelectable(true)
					taskNode.AddChild(childNode)
				*/

				updateTree(taskNode, *childBoard)
			}
		}
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
			currentIdx := t.calcTaskIdx(row, t.leftPanelWidth)
			task := *t.taskData.GetTask(currentIdx)

			task.SetDone(!task.GetIsDone())

			// If the task was toggled to be done,
			if task.GetIsDone() {
				// Remove task from the task data slice
				_, err := t.taskData.Remove(currentIdx)
				if err != nil {
					return
				}

				// Append the task to the end of the slice
				t.taskData.Add(&task, len(t.taskData.GetTasks()))
				// Update task priorities
				t.taskData.UpdatePriorities(currentIdx)
			}

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

// boardInputCapture captures input interactions specific to the
// currently displayed board.
func (t *TUI) boardInputCapture() {
	t.board.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// If there a no columns table is being displayed, allow user to go
		// back to tree navigation with 'h'.
		if t.isNoColumnsTableDisplayed {
			switch event.Rune() {
			case 'h':
				t.showTreeView()
			}
			return event
		}

		isFocusedOnTable := false
		if rows, cols := t.boardColumns[t.focusedColumn].GetSelectable(); rows == false && cols == false {
			isFocusedOnTable = true
		}
		selectedRow, _ := t.boardColumns[t.focusedColumn].GetSelection()

		switch event.Rune() {
		case 'l': // move right
			if t.focusedColumn < len(t.boardColumns)-1 {
				t.boardColumns[t.focusedColumn].SetSelectable(false, false)
				t.focusedColumn++
				t.boardColumns[t.focusedColumn].SetSelectable(true, false)

				if isFocusedOnTable {
					t.boardColumns[t.focusedColumn].SetSelectable(false, false)
				}
				t.app.SetFocus(t.boardColumns[t.focusedColumn])
			}
		case 'h': // move left
			// If at the first column, switch back to TreeView
			if t.focusedColumn == 0 {
				t.boardColumns[t.focusedColumn].SetSelectable(false, false)
				t.showTreeView()
				return event
			}

			t.boardColumns[t.focusedColumn].SetSelectable(false, false)
			t.focusedColumn--
			t.boardColumns[t.focusedColumn].SetSelectable(true, false)

			if isFocusedOnTable {
				t.boardColumns[t.focusedColumn].SetSelectable(false, false)
			}
			t.app.SetFocus(t.boardColumns[t.focusedColumn])
		case 'j': // move down
			if isFocusedOnTable {
				// Enable task selection
				t.boardColumns[t.focusedColumn].SetSelectable(true, false)
			}
		case 'k': // move up
			if !isFocusedOnTable && selectedRow == 0 {
				// Disable task selection
				t.boardColumns[t.focusedColumn].SetSelectable(false, false)
			}
		case 'a':
			if isFocusedOnTable {
			} else {
				/*
					form := t.createListForm(t.calcTaskIdx(row, t.leftPanelWidth))
					t.showModal(form)
				*/
				return event
			}
			/* Note: dont forget to reflect changes to treeview */
			// case 'd'
		case ' ': // Toggle task description
			if !isFocusedOnTable {
				// TODO: replace 50 with right panel width
				currentIdx := t.calcTaskIdxBoard(selectedRow, 50)
				// If calculated task index is within bounds, toggle task show
				// description status, and update rendered list.
				if err := t.boardColumnsData[t.focusedColumn].Bounds(currentIdx); err != nil {
					return event
				}
				task := t.boardColumnsData[t.focusedColumn].GetTask(currentIdx)
				task.ShowDesc = !task.ShowDesc
				t.updateColumn(t.focusedColumn)
			}
		}
		return event
	})
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
		task.SetID(currentIdx + 1)
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

// createBoardTaskForm creates and returns a tview form for creating a
// new board task.
// This function assumes that a task is currently selected.
func (t *TUI) createBoardTaskForm(currentIdx int) *tview.Form {
	var name, description string
	var createChildBoard bool

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Create New Task")

	form.AddInputField("Name", "", 20, nil, func(text string) {
		name = text
	})
	form.AddInputField("Description", "", 50, nil, func(text string) {
		description = text
	})
	form.AddCheckbox("Create a Board?", false, func(checked bool) {
		createChildBoard = checked
	})

	form.AddButton("Save", func() {
		// Add task to task data slice
		task := new(tasks.BoardTask)
		task.SetTask(new(tasks.Task))
		task.SetPriority(currentIdx + 1)
		task.SetID(currentIdx + 1)
		task.SetStarted(time.Now())
		task.SetName(name)
		task.SetDescription(description)

		if createChildBoard {
			t.addChildBoard(name, currentIdx)
		}

		t.boardColumnsData[t.focusedColumn].InsertTask(task, currentIdx+1)
		t.boardColumnsData[t.focusedColumn].UpdatePriorities(currentIdx + 1)

		t.updateColumn(t.focusedColumn)

		t.closeModal()
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
	})

	return form
}

// addChildBoard creates and adds a child board for a newly constructed
// task.
func (t *TUI) addChildBoard(name string, currentIdx int) {
	// Create a new board
	// Set `Title` to name
	// Set `ParentBoard` to the t.board field
	// Set `ParentTask` to
	// t.boardColumnsData[t.focusedColumn].GetTask(currentIdx)

	// Add new board to the list of child boards for the currently focused
	// column.
	// Set the `child` BoardTask field to the new board
	// Set the `HasChild` BoardTask field to true

	// Update tree view to show newly added board.
	// t.updateTree()
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
