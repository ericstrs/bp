package ui

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iuiq/do/internal/tasks"
	"github.com/rivo/tview"
)

type TUI struct {
	app           *tview.Application
	pages         *tview.Pages
	mainGrid      *tview.Grid
	zoomedInPanel bool

	leftPanel       *tview.Grid
	leftPanelWidth  int
	rightPanel      *tview.Grid
	rightPanelWidth int
	focusedPanel    *tview.Grid

	list     *tview.Table
	taskData *tasks.TodoList

	tree                      *tview.TreeView
	treeData                  *tasks.BoardTree
	board                     *tview.Grid
	boardColumns              []*tview.Table
	boardColumnsData          []tasks.BoardColumn
	focusedColumn             int
	isNoColumnsTableDisplayed bool
}

type NodeRef struct {
	ID   int
	Type string // This could be 'Board'
}

// Init intializes the tview app and sets up the UI.
func (t *TUI) Init(tl *tasks.TodoList, tree *tasks.BoardTree) {
	t.app = tview.NewApplication()
	t.appInputCapture()
	// Update left and right panel size before drawing. This won't affect
	// the current drawing, it sets the panel width variables for the next
	// draw operation.
	t.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		if t.zoomedInPanel {
			t.leftPanelWidth = width
			t.rightPanelWidth = width
			return false
		}
		t.leftPanelWidth = int(float64(width)*0.2) - 2
		t.rightPanelWidth = width - t.leftPanelWidth
		return false
	})

	t.taskData = tl
	t.treeData = tree
	width := 25
	// Remove two from default width to account for panel borders
	t.leftPanelWidth = width - 2

	t.list = tview.NewTable().
		SetSelectable(true, false)
	t.listInputCapture()

	t.board = tview.NewGrid().
		SetRows(0).
		SetColumns(0)
	t.boardInputCapture()

	root := tview.NewTreeNode("Board Trees")
	t.tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)
	t.treeInputCapture()
	t.tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if node.IsExpanded() {
			node.Collapse()
		} else {
			node.Expand()
		}
	})

	// Populate tui list and tree view.
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
	t.rightPanel.SetBorder(false)
	t.showTreeView()

	// Create the main parent grid
	t.mainGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(-1, -4).
		AddItem(t.leftPanel, 0, 0, 1, 1, 0, 0, true).
		AddItem(t.rightPanel, 0, 1, 1, 1, 0, 0, false)

	// Initialize panel focus to left panel
	t.focusedPanel = t.leftPanel

	// Add the main grid to page
	t.pages = tview.NewPages().
		AddPage("main", t.mainGrid, true, true)

	if err := t.app.SetRoot(t.pages, true).Run(); err != nil {
		panic(err)
	}
}

// showBoard clears the right panel and sets the board.
func (t *TUI) showBoard(b *tasks.Board) {
	t.rightPanel.Clear()
	t.board.Clear()
	t.boardColumns = nil // Reset columns
	t.boardColumnsData = b.GetColumns()

	if len(t.boardColumnsData) == 0 {
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
	for i := range t.boardColumnsData {
		table := tview.NewTable().
			SetSelectable(false, false) // No selection by default
		table.SetBorder(true)

		t.boardColumns = append(t.boardColumns, table)
		t.updateColumn(i)
		t.board.AddItem(table, 0, i, 1, 1, 0, 0, true)
	}

	// Set right panel content to the board grid. This will override the
	// tree view being displayed.
	t.rightPanel.AddItem(t.board, 0, 0, 1, 1, 0, 0, true)

	// Assert focus on the right panel. This is needed for board input
	// capture to work.
	t.app.SetFocus(t.boardColumns[0])
	t.focusedColumn = 0
}

// updateColumn clears the focused column of the board and updates the
// focused column's contents.
func (t *TUI) updateColumn(colIdx int) {
	t.boardColumns[colIdx].Clear()
	t.boardColumns[colIdx].SetTitle(t.boardColumnsData[colIdx].GetTitle())

	if len(t.boardColumnsData[colIdx].GetTasks()) == 0 {
		t.boardColumns[colIdx].SetCellSimple(0, 0, "No tasks available")
		return
	}

	currentRow := 0
	for _, task := range t.boardColumnsData[colIdx].GetTasks() {
		task := task
		prefix := ""
		if task.GetHasChild() {
			prefix = "# "
		}
		name := prefix + task.GetName()
		t.boardColumns[colIdx].SetCellSimple(currentRow, 0, name)

		// If task show description status is set to true, add the task
		// description to the list.
		if task.GetShowDesc() {
			lineWidth := (t.rightPanelWidth / len(t.boardColumnsData)) - 4
			wrappedDesc := WordWrap(task.Description, lineWidth)
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
	t.rightPanel.Clear()
	t.rightPanel.SetTitle("Tree Navigation")
	// Set right panel content to the tree view. This will override a
	// board being dislayed.
	t.rightPanel.AddItem(t.tree, 0, 0, 1, 1, 0, 0, true)

	// Assert focus on the right panel. This is needed for tree input
	// capture to work.
	t.app.SetFocus(t.tree)
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
		task, err := t.boardColumnsData[t.focusedColumn].GetTask(taskIdx)
		if err != nil {
			log.Printf("Failed to calculate board task index: %v\n", err)
			return taskIdx
		}

		// If the task description is being shown, skip the next row
		if task.GetShowDesc() {
			wrappedDesc := WordWrap(task.GetDescription(), columnWidth)
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
	t.updateTree()
}

// updateTree updates the entire tview tree view primitive.
func (t *TUI) updateTree() {
	t.tree.GetRoot().ClearChildren()

	rootBoards := t.treeData.GetRootBoards()

	// If there are no root boards,
	if len(rootBoards) == 0 {
		noBoardsNode := tview.NewTreeNode("No boards available")
		t.tree.GetRoot().AddChild(noBoardsNode)
		return
	}

	// For each rooted board tree, attach to root node.
	for i := range rootBoards {
		board := rootBoards[i]
		nr := NodeRef{ID: board.GetID(), Type: "Board"}
		boardNode := tview.NewTreeNode("# " + board.GetTitle()).
			SetReference(nr).
			SetColor(tcell.ColorGreen).
			SetExpanded(false).
			SetSelectable(true)
		t.tree.GetRoot().AddChild(boardNode)

		t.addBoardToTree(boardNode, board)
	}
}

// addBoardToTree recursively adds a given board and all its children to the tree
// view.
func (t *TUI) addBoardToTree(node *tview.TreeNode, board *tasks.Board) {
	columns := board.GetColumns()
	for i := range columns {
		columnNode := tview.NewTreeNode(columns[i].GetTitle()).
			SetReference(&columns[i]).
			SetSelectable(true)
		node.AddChild(columnNode)

		tasks := columns[i].GetTasks()
		for i := range tasks {
			task := tasks[i]
			if task.GetHasChild() {
				id := task.GetChildID()
				childBoard, err := t.treeData.GetBoardByID(id)
				if err != nil {
					log.Println("Failed to populate tree view:", err)
					return
				}
				nr := NodeRef{ID: childBoard.GetID(), Type: "Board"}
				childNode := tview.NewTreeNode("# " + childBoard.GetTitle()).
					SetReference(nr).
					SetColor(tcell.ColorGreen).
					SetSelectable(true)
				columnNode.AddChild(childNode)
				t.addBoardToTree(childNode, childBoard)
				continue
			}

			taskNode := tview.NewTreeNode(task.GetName()).
				SetReference(&task).
				SetSelectable(true)
			columnNode.AddChild(taskNode)
		}
	}
}

func (t *TUI) treeInputCapture() {
	t.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'l':
			node := t.tree.GetCurrentNode()
			board, ok := t.getBoardRef(node)
			if !ok {
				return event
			}

			t.showBoard(board)
		case 'a': // Add a root board
			node := t.tree.GetCurrentNode()
			// If node is the root node, create a new root board.
			if node.GetLevel() == 0 {
				form := t.createRootBoardForm()
				t.showModal(form)
				return event
			}
		case 'e': // Edit root board
			node := t.tree.GetCurrentNode()
			// If node is the root node, edit new root board.
			if node.GetLevel() == 1 {
				node := t.tree.GetCurrentNode()
				board, ok := t.getBoardRef(node)
				if !ok {
					log.Println("Failed to edit root board: current tree view node isn't of type Board")
					return event
				}
				form := t.editRootBoardForm(board, node)
				t.showModal(form)
				return event
			}
		case 'd': // Delete a root board
			node := t.tree.GetCurrentNode()
			if node.GetLevel() != 1 {
				return event
			}
			board, ok := t.getBoardRef(node)
			if !ok {
				log.Println("Failed to remove root board: current tree view node isn't of type Board.")
				return event
			}

			// Remove root board
			b, err := t.treeData.RemoveRoot(board)
			if err != nil {
				log.Printf("Failed to remove root board: %v\n", err)
				return event
			}
			// Buffer deleted board
			t.treeData.BoardBuffer.Clear()
			t.treeData.BoardBuffer.SetBoardBuffer(b)

			t.removeRefBoard(&b)

			// Update tree view
			t.updateTree()
			// Show updated tree view
			t.showTreeView()
		case 'p': // Paste buffered root board
			node := t.tree.GetCurrentNode()
			if node.GetLevel() != 0 {
				return event
			}

			board := t.treeData.BoardBuffer.GetBoardBuffer()
			cpy, err := board.DeepCopy(nil, t.treeData, t.treeData.BoardBuffer.GetChildBoards())
			if err != nil {
				log.Printf("Failed to paste board: %v\n", err)
				return event
			}

			// Append root board
			t.treeData.AddRoot(cpy)

			// Update tree view
			t.updateTree()
			// Show updated tree view
			t.showTreeView()
		}
		return event
	})
}

// appInputCapture captures input interactions across all displayed
// tview primatives.
func (tui *TUI) appInputCapture() {
	tui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
			case 'z': // Toggle panel zoom
				tui.mainGrid.Clear()
				switch tui.focusedPanel {
				case tui.leftPanel:
					if tui.zoomedInPanel {
						tui.mainGrid.AddItem(tui.leftPanel, 0, 0, 1, 1, 0, 0, true).
							AddItem(tui.rightPanel, 0, 1, 1, 1, 0, 0, false)
						tui.mainGrid.SetColumns(-1, -4)
					} else {
						tui.mainGrid.AddItem(tui.leftPanel, 0, 0, 1, 1, 0, 0, true)
						tui.mainGrid.SetColumns(0)
					}
				case tui.rightPanel:
					if tui.zoomedInPanel {
						tui.mainGrid.AddItem(tui.leftPanel, 0, 0, 1, 1, 0, 0, true).
							AddItem(tui.rightPanel, 0, 1, 1, 1, 0, 0, false)
						tui.mainGrid.SetColumns(-1, -4)
					} else {
						tui.mainGrid.AddItem(tui.rightPanel, 0, 0, 1, 1, 0, 0, true)
						tui.mainGrid.SetColumns(0)
					}
				}
				tui.zoomedInPanel = !tui.zoomedInPanel
			}
		case tcell.KeyTab: // Switch panel focus
			switch tui.focusedPanel {
			case tui.leftPanel: // Switch focus to right panel
				if tui.zoomedInPanel {
					tui.mainGrid.Clear()
					tui.mainGrid.AddItem(tui.rightPanel, 0, 0, 1, 1, 0, 0, true)
				}
				tui.app.SetFocus(tui.rightPanel)
				tui.focusedPanel = tui.rightPanel
				tui.list.SetSelectable(false, false)
				tui.leftPanel.SetBorder(false)
				tui.rightPanel.SetBorder(true)
			case tui.rightPanel: // Switch focus to left panel
				if tui.zoomedInPanel {
					tui.mainGrid.Clear()
					tui.mainGrid.AddItem(tui.leftPanel, 0, 0, 1, 1, 0, 0, true)
				}
				tui.app.SetFocus(tui.leftPanel)
				tui.focusedPanel = tui.leftPanel
				tui.list.SetSelectable(true, false)
				tui.rightPanel.SetBorder(false)
				tui.leftPanel.SetBorder(true)
			}
			return nil // Override the tab key
		}
		return event
	})
}

// listInputCapture captures input interactions specific to the list.
func (t *TUI) listInputCapture() {
	t.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.list.GetSelection()
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'a': // Create new task
				form := t.createListForm(t.calcTaskIdx(row, t.leftPanelWidth))
				t.showModal(form)
				return event
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
						return event
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
					return event
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
					return event
				}
				task := t.taskData.GetTask(currentIdx)
				task.ShowDesc = !task.ShowDesc
				t.filterAndUpdateList(t.leftPanelWidth)
			}
		}
		return event
	})
}

// boardInputCapture captures input interactions specific to the
// currently displayed board.
func (t *TUI) boardInputCapture() {
	t.board.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// If there a no columns table is being displayed,
		if t.isNoColumnsTableDisplayed {
			switch event.Rune() {
			case 'h': // go back to tree navigation
				t.showTreeView()
			case 'a': // add a new column
				form, err := t.createColumnForm(t.focusedColumn)
				if err != nil {
					log.Println("Failed to add a column to the board: ", err)
					return event
				}
				t.showModal(form)
			}
			return event
		}

		isFocusedOnTable := false
		if rows, cols := t.boardColumns[t.focusedColumn].GetSelectable(); rows == false && cols == false {
			isFocusedOnTable = true
		}
		selectedRow, _ := t.boardColumns[t.focusedColumn].GetSelection()

		if event.Rune() == 'L' || event.Key() == tcell.KeyEnter {
			// If focused on a task
			if !isFocusedOnTable {
				task, err := t.boardColumnsData[t.focusedColumn].GetTask(t.calcTaskIdxBoard(selectedRow, t.rightPanelWidth))
				if err != nil {
					log.Printf("Failed to enter sub-board: %v\n", err)
					return event
				}

				// If task has a child
				if task.GetHasChild() {
					parentNode := t.tree.GetCurrentNode()
					_, ok := t.getBoardRef(parentNode)

					if !ok {
						//log.Println("Failed to enter sub-board: current tree view node isn't of type Board.")
						return event
					}

					id := task.GetChildID()
					childBoard, err := t.treeData.GetBoardByID(id)
					if err != nil {
						log.Println("Failed to enter sub-board: ", err)
						return event
					}
					// Find tree view node that references focused board column
					for _, node := range parentNode.GetChildren() {
						column, ok := getColumnRef(node)
						if !ok {
							log.Println("Failed to enter sub-board: board child tree view node isn't of type Column.")
							return event
						}
						if column == &t.boardColumnsData[t.focusedColumn] {
							// Find tree view node that references focused board task
							for _, n := range node.GetChildren() {
								board, ok := t.getBoardRef(n)
								if !ok {
									log.Println("Failed to enter sub-board: found tree view node isn't of type Board.")
									return event
								}
								if board == childBoard {
									t.tree.SetCurrentNode(n)
									t.showBoard(childBoard)
									return event
								}
							}
						}
					}
				}
			}
		}

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
			// If focus is on the entire table, create a new board column to
			// the right of the currently focused column.
			if isFocusedOnTable {
				form, err := t.createColumnForm(t.focusedColumn)
				if err != nil {
					log.Println("Failed to add a column to the board: ", err)
					return event
				}
				t.showModal(form)
			} else {
				// Otherwise, focus is on a task in the column, create a new a new
				// board task underneath the currently focused task.
				form := t.createBoardTaskForm(t.calcTaskIdxBoard(selectedRow, t.rightPanelWidth))
				t.showModal(form)
			}
		case 'e':
			if isFocusedOnTable { // Edit board column
				form := t.editColumnForm()
				t.showModal(form)
			} else { // Edit board tasks
				form, err := t.editBoardTaskForm(t.calcTaskIdxBoard(selectedRow, t.rightPanelWidth))
				if err != nil {
					log.Printf("Failed to edit board task: %v\n", err)
					return event
				}
				t.showModal(form)
			}
		case 'd':
			if isFocusedOnTable { // Delete and buffer board column
				node := t.tree.GetCurrentNode()
				parentBoard, ok := t.getBoardRef(node)
				if !ok {
					log.Println("Failed to remove board column: current tree view node isn't of type Board.")
					return event
				}

				t.treeData.ColBuffer.Clear()

				// If a task in the column references a board, remove that board from
				// child board slice. Buffer the task and board and its chilren.
				for _, task := range t.boardColumnsData[t.focusedColumn].GetTasks() {
					if task.GetHasChild() {
						board, err := t.treeData.GetBoardByID(task.GetChildID())
						if err != nil {
							log.Printf("Failed to buffer a referenced board in the column: %v\n", err)
							continue
						}
						_, err = t.treeData.RemoveChildBoard(board)
						if err != nil {
							log.Printf("Failed to remove and buffer child board: %v\n", err)
							continue
						}
						// Buffer and remove immediate referenced child
						t.treeData.ColBuffer.AddChild(board)
						// Sever connection from parent board and child board
						log.Println("Parentboard ", parentBoard.GetTitle(), " ; removing child ", board.GetID())
						if err = parentBoard.RemoveChild(board.GetID()); err != nil {
							log.Printf("Failed to remove child board from parent child boards slice: %v\n", err)
							continue
						}

						// Buffer and remove any potential children
						t.removeRefColumn(board)
					}
				}

				column, err := parentBoard.RemoveColumn(t.focusedColumn)
				if err != nil {
					log.Printf("Failed to remove board column: %v\n", err)
					return event
				}
				t.treeData.ColBuffer.SetColumnBuffer(column)

				// Update and show board
				t.showBoard(parentBoard)
				t.updateTree()
			} else { // Delete and buffer board task
				node := t.tree.GetCurrentNode()
				parentBoard, ok := t.getBoardRef(node)
				if !ok {
					log.Println("Failed to remove board task: current tree view node isn't of type Board.")
					return event
				}

				// Delete task from focused column
				taskIdx := t.calcTaskIdxBoard(selectedRow, t.rightPanelWidth)
				task, err := t.boardColumnsData[t.focusedColumn].Remove(taskIdx)
				if err != nil {
					return event
				}
				t.treeData.TaskBuffer.Clear()
				t.treeData.TaskBuffer.SetTaskBuffer(*task)

				// If task being deleted references a board, from it from child
				// boards slice. Buffer the task and board and its children.
				if task.GetHasChild() {
					board, err := t.treeData.GetBoardByID(task.GetChildID())
					if err != nil {
						log.Printf("Failed to buffer referenced board: %v\n", err)
					} else {
						// Buffer and remove immediate referenced child
						t.treeData.TaskBuffer.AddChild(board)
						t.treeData.RemoveChildBoard(board)
						// Sever connection from parent board and child board
						parentBoard.RemoveChild(board.GetID())

						// Buffer and remove any potential children
						t.removeRefTask(board)
					}
				}

				// Update the focused column and tree view
				t.boardColumnsData[t.focusedColumn].UpdatePriorities(taskIdx)
				t.updateColumn(t.focusedColumn)
				t.updateTree()
			}
		case 'p':
			if isFocusedOnTable { // Paste board column
				// Get current board
				node := t.tree.GetCurrentNode()
				board, ok := t.getBoardRef(node)
				if !ok {
					log.Println("Couldn't paste column: current tree view node isn't of type Board.")
					return event
				}

				// Get buffered column
				column := t.treeData.ColBuffer.GetColumnBuffer()

				cpy, err := column.DeepCopy(t.treeData, board, t.treeData.ColBuffer.GetChildBoards())
				if err != nil {
					log.Printf("Failed to paste board column: %v\n", err)
					return event
				}

				// Insert column into board
				board.InsertColumn(cpy, t.focusedColumn+1)

				// Update board and tree view
				t.showBoard(board)
				t.updateTree()
			} else { // Paste board task
				// Get current board
				node := t.tree.GetCurrentNode()
				board, ok := t.getBoardRef(node)
				if !ok {
					log.Println("Couldn't paste task: current tree view node isn't of type Board.")
					return event
				}

				currentIdx := t.calcTaskIdxBoard(selectedRow, t.rightPanelWidth)

				// Read from buffer
				task := t.treeData.TaskBuffer.GetTaskBuffer()
				// If this board has no buffered task, return early.
				if task.Task == nil {
					return event
				}

				cpy, err := task.DeepCopy(t.treeData, board, t.treeData.TaskBuffer.GetChildBoards())
				if err != nil {
					log.Printf("Failed to paste board task: %v\n", err)
					return event
				}

				// Add task to the todo list
				t.boardColumnsData[t.focusedColumn].InsertTask(&cpy, currentIdx+1)

				// Update the focused column and tree view
				t.boardColumnsData[t.focusedColumn].UpdatePriorities(currentIdx)
				t.updateColumn(t.focusedColumn)
				t.updateTree()
			}
		case ' ': // Toggle task description
			if !isFocusedOnTable {
				currentIdx := t.calcTaskIdxBoard(selectedRow, t.rightPanelWidth)
				// If calculated task index is within bounds, toggle task show
				// description status, and update rendered list.
				task, err := t.boardColumnsData[t.focusedColumn].GetTask(currentIdx)
				if err != nil {
					return event
				}
				task.ShowDesc = !task.ShowDesc
				t.updateColumn(t.focusedColumn)
			}
		}
		return event
	})
}

// removeRefBoard moves a given board and all its children from the main
// list of boards to a buffer board list.
func (t *TUI) removeRefBoard(board *tasks.Board) {
	for _, id := range board.GetChildren() {
		childBoard, err := t.treeData.GetBoardByID(id)
		if err != nil {
			log.Printf("Failed to buffer a referenced child board: %v\n", err)
			continue
		}
		t.treeData.BoardBuffer.AddChild(childBoard)
		t.treeData.RemoveChildBoard(childBoard)
		t.removeRefBoard(childBoard) // remove childBoard's children
	}
}

func (t *TUI) removeRefColumn(board *tasks.Board) {
	for _, id := range board.GetChildren() {
		childBoard, err := t.treeData.GetBoardByID(id)
		if err != nil {
			log.Printf("Failed to buffer a referenced child board: %v\n", err)
			continue
		}
		t.treeData.ColBuffer.AddChild(childBoard)
		t.treeData.RemoveChildBoard(childBoard)
		t.removeRefColumn(childBoard) // remove childBoard's children
	}
}

func (t *TUI) removeRefTask(board *tasks.Board) {
	for _, id := range board.GetChildren() {
		childBoard, err := t.treeData.GetBoardByID(id)
		if err != nil {
			log.Printf("Failed to buffer a referenced child board: %v\n", err)
			continue
		}
		t.treeData.TaskBuffer.AddChild(childBoard)
		t.treeData.RemoveChildBoard(childBoard)
		t.removeRefTask(childBoard) // remove childBoard's children
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
		task.SetID(currentIdx + 1)
		t.taskData.Add(task, currentIdx+1)
		t.taskData.UpdatePriorities(currentIdx + 1)

		// Update tview list
		t.filterAndUpdateList(t.leftPanelWidth)

		t.closeModal()
		t.app.SetFocus(t.leftPanel)
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.leftPanel)
	})

	return form
}

// createRootBoardForm creates and returns a tview form for creating
// a new root board.
func (t *TUI) createRootBoardForm() *tview.Form {
	var name string

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Create New Root Board")

	form.AddInputField("Name", name, 20, nil, func(text string) {
		name = text
	})

	form.AddButton("Save", func() {
		board := t.treeData.NewBoard(name)
		t.treeData.AddRoot(board)

		// Update tree view
		t.updateTree()
		// Show updated tree view
		t.showTreeView()

		t.closeModal()
		t.app.SetFocus(t.tree)
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.tree)
	})

	return form

}

// createColumnForm creates and returns a tview form for creating a
// new board column.
// This function assumes that a task is currently selected.
func (t *TUI) createColumnForm(focusedColumn int) (*tview.Form, error) {
	var name string

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Create New Board Column")

	form.AddInputField("Name", "", 20, nil, func(text string) {
		name = text
	})

	form.AddButton("Save", func() {
		node := t.tree.GetCurrentNode()
		board, ok := t.getBoardRef(node)

		if !ok {
			//return errors.New("current tree view node isn't of type Board")
			return
		}

		column := new(tasks.BoardColumn)
		column.SetTitle(name)

		// Insert new column
		board.InsertColumn(*column, t.focusedColumn+1)

		t.showBoard(board)

		// Update tree view to include the newly board column
		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	return form, nil
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
		t.treeData.IncrementTaskCtr()
		task.SetID(t.treeData.GetTaskCtr())
		task.SetStarted(time.Now())
		task.SetName(name)
		task.SetDescription(description)

		if createChildBoard {
			if err := t.createAndAddChildBoard(name, task); err != nil {
				log.Printf("Failed to create and add child board for %q task: %v\n", name, err)
			}
		}

		t.boardColumnsData[t.focusedColumn].InsertTask(task, currentIdx+1)
		t.boardColumnsData[t.focusedColumn].UpdatePriorities(currentIdx)

		// Update the column to show the newly added task
		t.updateColumn(t.focusedColumn)

		// Update tree view to include the newly added task and its
		// child if created.
		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	return form
}

// createAndAddChildBoard creates and adds a child board for a
// newly constructed task.
func (t *TUI) createAndAddChildBoard(name string, parentTask *tasks.BoardTask) error {
	node := t.tree.GetCurrentNode()
	parentBoard, ok := t.getBoardRef(node)
	if !ok {
		return errors.New("node doesn't reference a board")
	}
	newBoard := t.treeData.NewBoard(name)
	t.treeData.AddChildBoard(newBoard)
	createConnection(parentTask, parentBoard, newBoard)

	return nil
}

// editListForm creates and returns a tview form for editing a
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
		// Update task in data slice
		task.SetName(name)
		task.SetDescription(description)
		task.SetCore(isCore)

		// Update tview list
		t.filterAndUpdateList(t.leftPanelWidth)

		t.closeModal()
		t.app.SetFocus(t.leftPanel)
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.leftPanel)
	})

	return form
}

// editRootBoardForm creates and returns a tview form for editing a
// root board.
func (t *TUI) editRootBoardForm(board *tasks.Board, node *tview.TreeNode) *tview.Form {
	name := board.GetTitle()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Edit Root Board")

	// Define the input fields for the forms and update field variables if
	// user makes any changes to the default values.
	form.AddInputField("Name", name, 20, nil, func(text string) {
		name = text
	})

	form.AddButton("Save", func() {
		board.SetTitle(name)
		// Update tree node that references the root board
		node.SetText(name)
		t.closeModal()
		t.app.SetFocus(t.tree)
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.tree)
	})

	return form
}

// editColumnForm creates and returns a tview form for editing a
// board column.
func (t *TUI) editColumnForm() *tview.Form {
	column := &t.boardColumnsData[t.focusedColumn]
	name := column.GetTitle()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Edit Column")

	// Define the input fields for the forms and update field variables if
	// user makes any changes to the default values.
	form.AddInputField("Name", name, 20, nil, func(text string) {
		name = text
	})

	form.AddButton("Save", func() {
		// Update task in data slice
		column.SetTitle(name)

		// Update tview list
		t.updateColumn(t.focusedColumn)

		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	return form
}

// editBoardTaskForm creates and returns a tview form for editing a
// todo list task.
func (t *TUI) editBoardTaskForm(currentIdx int) (*tview.Form, error) {
	var createChildBoard bool
	task, err := t.boardColumnsData[t.focusedColumn].GetTask(currentIdx)
	if err != nil {
		return nil, err
	}
	name := task.GetName()
	description := task.GetDescription()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Edit Task")

	// Define the input fields for the forms and update field variables if
	// user makes any changes to the default values.
	form.AddInputField("Name", name, 20, nil, func(text string) {
		name = text
	})
	form.AddInputField("Description", description, 50, nil, func(text string) {
		description = text
	})

	if !task.GetHasChild() {
		form.AddCheckbox("Create a Board?", false, func(checked bool) {
			createChildBoard = checked
		})
	}

	form.AddButton("Save", func() {
		task.SetName(name)
		if task.GetHasChild() {
			id := task.GetChildID()
			childBoard, err := t.treeData.GetBoardByID(id)
			if err != nil {
				log.Printf("Failed to create and add child board for %q task: %v\n", err)
				return
			}
			childBoard.SetTitle(name)
		}
		task.SetDescription(description)

		if createChildBoard {
			if err := t.createAndAddChildBoard(name, task); err != nil {
				log.Printf("Failed to create and add child board for %q task: %v\n", name, err)
				return
			}
		}

		// Update tview list
		t.updateColumn(t.focusedColumn)

		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardColumns[t.focusedColumn])
	})

	return form, nil
}

// closeModal removes that modal page and sets the focus back to the
// main grid.
func (t *TUI) closeModal() {
	t.pages.RemovePage("modal")
}

// getBoardRef returns the Board referenced by a TreeNode.
func (t *TUI) getBoardRef(n *tview.TreeNode) (*tasks.Board, bool) {
	nr, ok := n.GetReference().(NodeRef)
	if !ok {
		return nil, false
	}
	if nr.Type != "Board" {
		return nil, false
	}

	board, err := t.treeData.GetBoardByID(nr.ID)
	if err != nil {
		log.Println("Failed to get board reference: ", err)
		return nil, false
	}
	return board, true
}

// getColumnRef returns the board column referenced by a TreeNode.
func getColumnRef(n *tview.TreeNode) (*tasks.BoardColumn, bool) {
	if column, ok := n.GetReference().(*tasks.BoardColumn); ok {
		return column, true
	}
	return nil, false
}

// getTaskRef returns board task referenced by a TreeNode.
func getTaskRef(n *tview.TreeNode) (*tasks.BoardTask, bool) {
	if task, ok := n.GetReference().(*tasks.BoardTask); ok {
		return task, true
	}
	return nil, false
}

// createConnection establishes a link between a board task and a board.
func createConnection(parentTask *tasks.BoardTask, parentBoard, board *tasks.Board) {
	// Estabilsh connection from board to parent task
	board.SetParentTask(parentTask)

	// Establish connection from parent task to a board
	parentTask.SetChildID(board.GetID())
	parentTask.SetHasChild(true)

	parentBoard.AddChild(board.GetID())
}

// severConnection removes the link between a board task and a board.
func severConnection(parentTask *tasks.BoardTask, board *tasks.Board) {
	// Sever the connection from board to parent task
	board.SetParentTask(nil)

	// Remove connection from parent task to a board
	parentTask.SetChildID(-1)
	parentTask.SetHasChild(false)
}
