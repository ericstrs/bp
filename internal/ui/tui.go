package ui

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iuiq/do/internal/tasks"
	"github.com/rivo/tview"
)

const (
	borderWidth = 2 // Boarder width is about two cells
)

type TUI struct {
	app      *tview.Application
	pages    *tview.Pages
	mainGrid *tview.Grid
	zoomedIn bool

	leftPanel       *tview.Grid
	leftPanelWidth  int
	rightPanel      *tview.Grid
	rightPanelWidth int
	focusedPanel    *tview.Grid

	list     *tview.Table
	taskData *tasks.TodoList

	tree          *tview.TreeView
	treeData      *tasks.BoardTree
	board         *tview.Grid
	boardCols     []*tview.Table
	boardColsData []tasks.BoardColumn
	focusedCol    int
	isEmptyTable  bool
}

type NodeRef struct {
	ID   int
	Type string // This could be 'Board'
}

// Init intializes the tview app and sets up the UI.
func (t *TUI) Init(tl *tasks.TodoList, tree *tasks.BoardTree) {
	t.taskData = tl
	t.treeData = tree
	t.InitApp()
	t.InitList()
	t.InitBoard()
	t.InitTree()

	// Populate tui list and tree view.
	t.Populate()

	t.InitLeftPanel()
	t.InitRightPanel()
	// Initialize panel focus to left panel
	t.focusedPanel = t.leftPanel

	// Create the main parent grid
	t.mainGrid = tview.NewGrid().
		SetRows(0).
		SetColumns(-1, -4).
		AddItem(t.leftPanel, 0, 0, 1, 1, 0, 0, true).
		AddItem(t.rightPanel, 0, 1, 1, 1, 0, 0, false)

	// Add the main grid to page
	t.pages = tview.NewPages().
		AddPage("main", t.mainGrid, true, true)

	if err := t.app.SetRoot(t.pages, true).Run(); err != nil {
		panic(err)
	}
}

// InitApp initialzes the application.
func (t *TUI) InitApp() {
	t.app = tview.NewApplication()
	t.appInputCapture()
	// Update left and right panel size before drawing. This won't affect
	// the current drawing, it sets the panel width variables for the next
	// draw operation.
	t.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		width, _ := screen.Size()
		if t.zoomedIn {
			t.leftPanelWidth = width
			t.rightPanelWidth = width
			return false
		}
		t.leftPanelWidth = int(float64(width)*0.2) - borderWidth
		t.rightPanelWidth = width - t.leftPanelWidth
		return false
	})
}

// InitList initialzes the list.
func (t *TUI) InitList() {
	t.list = tview.NewTable().
		SetSelectable(true, false)
	t.listInputCapture()
}

// InitBoard initialzes the board.
func (t *TUI) InitBoard() {
	t.board = tview.NewGrid().
		SetRows(0).
		SetColumns(0)
	t.boardInputCapture()
}

// InitTree initialzes the tree.
func (t *TUI) InitTree() {
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
}

// InitLeftPanel initialzes the left panel.
func (t *TUI) InitLeftPanel() {
	width := 25
	// Remove two from default width to account for panel borders
	t.leftPanelWidth = width - borderWidth
	// Create the left-hand side panel
	t.leftPanel = tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		AddItem(t.list, 0, 0, 1, 1, 0, 0, true)
	t.leftPanel.SetTitle(t.taskData.GetTitle())
	t.leftPanel.SetBorder(true)
}

// InitRightPanel initialzes the right panel.
func (t *TUI) InitRightPanel() {
	// Create right-hand side panel
	t.rightPanel = tview.NewGrid().
		SetRows(0).
		SetColumns(0)
	t.rightPanel.SetBorder(false)
	t.showTreeView()
}

// showBoard clears the right panel and sets the board.
func (t *TUI) showBoard(b *tasks.Board) {
	t.rightPanel.Clear()
	t.board.Clear()
	t.boardCols = nil // Reset columns
	t.boardColsData = b.GetColumns()

	t.isEmptyTable = false
	if len(t.boardColsData) == 0 {
		t.rightPanel.SetTitle("No Columns")
		noColTable := tview.NewTable()
		cell := tview.NewTableCell("This board has no columns.")

		noColTable.SetCell(0, 0, cell)

		t.board.AddItem(noColTable, 0, 0, 1, 1, 0, 0, false)
		t.rightPanel.AddItem(t.board, 0, 0, 1, 1, 0, 0, true)
		t.app.SetFocus(t.rightPanel)
		t.isEmptyTable = true
		return
	}

	t.rightPanel.SetTitle(b.GetTitle())

	// Create a table for each column in the board
	for i := range t.boardColsData {
		table := tview.NewTable().
			SetSelectable(false, false) // No selection by default
		table.SetBorder(true)

		t.boardCols = append(t.boardCols, table)
		t.updateColumn(i)
		t.board.AddItem(table, 0, i, 1, 1, 0, 0, true)
	}

	// Set right panel content to the board grid. This will override the
	// tree view being displayed.
	t.rightPanel.AddItem(t.board, 0, 0, 1, 1, 0, 0, true)

	// Assert focus on the right panel. This is needed for board input
	// capture to work.
	t.app.SetFocus(t.boardCols[0])
	t.focusedCol = 0
}

// updateColumn clears the focused column of the board and updates the
// focused column's contents.
func (t *TUI) updateColumn(colIdx int) {
	col := &t.boardColsData[colIdx]
	table := t.boardCols[colIdx]

	table.Clear()
	table.SetTitle(col.GetTitle())

	if len(col.GetTasks()) == 0 {
		table.SetCellSimple(0, 0, "No tasks available")
		return
	}

	currentRow := 0
	for _, task := range col.GetTasks() {
		task := task
		prefix := ""
		if task.GetHasChild() {
			prefix = "# "
		}
		name := prefix + task.GetName()
		table.SetCellSimple(currentRow, 0, name)

		// If task show description status is set to true, add the task
		// description to the list.
		if task.GetShowDesc() {
			lineWidth := (t.rightPanelWidth / len(t.boardColsData)) - 4
			wrappedDesc := WordWrap(task.GetDesc(), lineWidth)
			for _, line := range wrappedDesc {
				currentRow++
				table.SetCell(currentRow, 0, tview.NewTableCell(line).
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
// task is shown, which would occupy one or more rows in the task list
// table.
func (t *TUI) calcTaskIdx(row, colWidth int) int {
	taskIdx := 0

	// Iterate through each row up to the selected row
	for i := 0; i < row; i++ {
		task, err := t.taskData.GetTask(taskIdx)
		if err != nil {
			log.Printf("Failed to calculate list task index: %v\n", err)
			return taskIdx
		}

		// If the task description is being shown, skip the next row
		if task.GetShowDesc() {
			wd := WordWrap(task.GetDesc(), colWidth)
			i += len(wd) // Skip the row(s) meant for task description
		}
		taskIdx++
	}
	return taskIdx
}

// calcTaskIdxBoard returns the calculated task index in a given board
// column. This function takes into account whether the description for each
// task is shown, which would occupy one or more rows in the column table.
func (t *TUI) calcTaskIdxBoard(row, colWidth int) int {
	taskIdx := 0

	// Iterate through each row up to the selected row
	for i := 0; i < row; i++ {
		task, err := t.boardColsData[t.focusedCol].GetTask(taskIdx)
		if err != nil {
			log.Printf("Failed to calculate board task index: %v\n", err)
			return taskIdx
		}

		// If the task description is being shown, skip the next row
		if task.GetShowDesc() {
			wd := WordWrap(task.GetDesc(), colWidth)
			i += len(wd) // Skip the row(s) meant for task description
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
func (t *TUI) filterAndUpdateList(colWidth int) {
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
			wd := WordWrap(task.Description, colWidth)
			for _, line := range wd {
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

	// If there are no root boards, add message node and return
	if len(rootBoards) == 0 {
		noBoardsNode := tview.NewTreeNode("No boards available")
		t.tree.GetRoot().AddChild(noBoardsNode)
		return
	}

	// For each rooted board tree, attach to root node
	for i := range rootBoards {
		board := rootBoards[i]
		nr := NodeRef{ID: board.GetID(), Type: "Board"}
		boardNode := tview.NewTreeNode("# " + board.GetTitle()).
			SetReference(nr).
			SetColor(tcell.ColorGreen).
			SetSelectable(true)
		t.tree.GetRoot().AddChild(boardNode)

		// Add board's children
		t.addBoardToTree(boardNode, board)
	}
}

// addBoardToTree recursively adds a given board and all its children
// to the tree view.
func (t *TUI) addBoardToTree(n *tview.TreeNode, b *tasks.Board) {
	columns := b.GetColumns()
	for i := range columns {
		columnNode := tview.NewTreeNode(columns[i].GetTitle()).
			SetReference(&columns[i]).
			SetSelectable(true)
		n.AddChild(columnNode)

		tasks := columns[i].GetTasks()
		for i := range tasks {
			task := tasks[i]
			if task.GetHasChild() {
				childBoard, err := t.treeData.GetBoard(task.GetChildID())
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
				tui.toggleZoom()
			}
		case tcell.KeyTab: // Switch panel focus
			tui.switchPanel()
			return nil // Override the tab key
		}
		return event
	})
}

// toggleZoom toggles the panel zoom
func (t *TUI) toggleZoom() {
	t.mainGrid.Clear()
	switch t.focusedPanel {
	case t.leftPanel:
		if t.zoomedIn {
			t.mainGrid.AddItem(t.leftPanel, 0, 0, 1, 1, 0, 0, true).
				AddItem(t.rightPanel, 0, 1, 1, 1, 0, 0, false)
			t.mainGrid.SetColumns(-1, -4)
		} else {
			t.mainGrid.AddItem(t.leftPanel, 0, 0, 1, 1, 0, 0, true)
			t.mainGrid.SetColumns(0)
		}
	case t.rightPanel:
		if t.zoomedIn {
			t.mainGrid.AddItem(t.leftPanel, 0, 0, 1, 1, 0, 0, true).
				AddItem(t.rightPanel, 0, 1, 1, 1, 0, 0, false)
			t.mainGrid.SetColumns(-1, -4)
		} else {
			t.mainGrid.AddItem(t.rightPanel, 0, 0, 1, 1, 0, 0, true)
			t.mainGrid.SetColumns(0)
		}
	}
	t.zoomedIn = !t.zoomedIn
}

// switchPanel switches panel focus.
func (t *TUI) switchPanel() {
	switch t.focusedPanel {
	case t.leftPanel: // switch focus to right panel
		if t.zoomedIn {
			t.mainGrid.Clear()
			t.mainGrid.AddItem(t.rightPanel, 0, 0, 1, 1, 0, 0, true)
		}
		t.app.SetFocus(t.rightPanel)
		t.focusedPanel = t.rightPanel
		t.list.SetSelectable(false, false)
		t.leftPanel.SetBorder(false)
		t.rightPanel.SetBorder(true)
	case t.rightPanel: // switch focus to left panel
		if t.zoomedIn {
			t.mainGrid.Clear()
			t.mainGrid.AddItem(t.leftPanel, 0, 0, 1, 1, 0, 0, true)
		}
		t.app.SetFocus(t.leftPanel)
		t.focusedPanel = t.leftPanel
		t.list.SetSelectable(true, false)
		t.rightPanel.SetBorder(false)
		t.leftPanel.SetBorder(true)
	}
}

// listInputCapture captures input interactions specific to the list.
func (t *TUI) listInputCapture() {
	t.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := t.list.GetSelection()
		idx := t.calcTaskIdx(row, t.leftPanelWidth)
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'a': // create new task
				form := t.createListForm(idx)
				t.showModal(form)
				return event
			case 'e': // edit task
				form, err := t.editListForm(idx)
				if err != nil {
					log.Printf("failed to edit list: %v\n", err)
					return event
				}
				t.showModal(form)
			case 'x': // toggle task completion status
				err := t.toggleTaskDone(idx)
				if err != nil {
					return event
				}
			case 'd': // delete task
				err := t.deleteListTask(idx)
				if err != nil {
					return event
				}
			case 'p': // paste task
				t.pasteListTask(idx)
			case ' ': // toggle task description
				err := t.toggleTaskDesc(idx)
				if err != nil {
					return event
				}
			}
		}
		return event
	})
}

// toggleTaskDone toggles a tasks completion status. Toggling a task
// from done to start does not restart the start date.
func (t *TUI) toggleTaskDone(idx int) error {
	tk, err := t.taskData.GetTask(idx)
	if err != nil {
		return err
	}
	task := *tk // Create local copy to prevent duplication
	task.SetDone(!task.GetIsDone())
	// If task toggled to done, remove task from task slice
	if task.GetIsDone() {
		if _, err := t.taskData.Remove(idx); err != nil {
			return err
		}
		// Append task to end of slice and update task priorities
		t.taskData.Add(&task, len(t.taskData.GetTasks()))
		t.taskData.UpdatePriorities(idx)
		task.SetFinished(time.Now()) // update done date
	}
	t.filterAndUpdateList(t.leftPanelWidth)
	return nil
}

// deleteListTask deletes and buffers a task from the list.
func (t *TUI) deleteListTask(idx int) error {
	task, err := t.taskData.Remove(idx)
	if err != nil {
		return err
	}
	t.taskData.SetBuff(task)
	t.taskData.UpdatePriorities(idx)
	t.filterAndUpdateList(t.leftPanelWidth)
	return nil
}

// pasteListTask pastes buffered list task.
func (t *TUI) pasteListTask(idx int) {
	task := t.taskData.Buffer()
	t.taskData.Add(task, idx+1)
	t.taskData.UpdatePriorities(idx + 1)
	t.filterAndUpdateList(t.leftPanelWidth)
}

// toggleTaskDesc toggles a list task description.
func (t *TUI) toggleTaskDesc(idx int) error {
	task, err := t.taskData.GetTask(idx)
	if err != nil {
		return err
	}
	task.ShowDesc = !task.ShowDesc
	t.filterAndUpdateList(t.leftPanelWidth)
	return nil
}

// treeInputCapture captures input interactions specific to the tree
// view.
func (t *TUI) treeInputCapture() {
	t.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'l': // Enter board
			node := t.tree.GetCurrentNode()
			board, ok := t.getBoardRef(node)
			if !ok {
				return event
			}
			t.showBoard(board)
		case 'a':
			// If node is the root node, create a new root board.
			node := t.tree.GetCurrentNode()
			if node.GetLevel() == 0 {
				t.showModal(t.createRootBoardForm())
			}
		case 'e':
			// If node is root board, edit root board.
			node := t.tree.GetCurrentNode()
			if node.GetLevel() == 1 {
				board, ok := t.getBoardRef(node)
				if !ok {
					return event
				}
				t.showModal(t.editRootBoardForm(board, node))
			}
		case 'd': // Delete a root board
			t.deleteRootBoard()
		case 'p': // Paste buffered root board
			t.pasteRootBoard()
		}
		return event
	})
}

// deleteRootBoard deletes a root board.
func (t *TUI) deleteRootBoard() {
	node := t.tree.GetCurrentNode()
	if node.GetLevel() != 1 {
		return
	}
	board, ok := t.getBoardRef(node)
	if !ok {
		log.Println("Failed to remove root board: current tree view node isn't of type Board.")
		return
	}

	b, err := t.treeData.RemoveRoot(board)
	if err != nil {
		log.Printf("Failed to remove root board: %v\n", err)
		return
	}
	// Buffer deleted board
	t.treeData.BoardBuff.Clear()
	t.treeData.BoardBuff.SetBoardBuff(b)

	// Remove root board's children
	t.removeRefBoard(&b)

	// Update and show tree view
	t.tree.GetRoot().RemoveChild(node)
	t.showTreeView()
}

// pasteRootBoard reads buffered root board and pastes it.
func (t *TUI) pasteRootBoard() {
	node := t.tree.GetCurrentNode()
	if node.GetLevel() != 0 {
		return
	}

	board := t.treeData.BoardBuff.GetBoardBuff()
	cpy, err := board.DeepCopy(nil, t.treeData, t.treeData.BoardBuff.GetChildBoards())
	if err != nil {
		log.Printf("Failed to paste board: %v\n", err)
		return
	}

	// Append root board
	t.treeData.AddRoot(cpy)

	// Update and show tree view
	t.updateTree()
	t.showTreeView()
}

// boardInputCapture captures input interactions specific to the
// currently displayed board.
func (t *TUI) boardInputCapture() {
	t.board.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Input capture for the no columns table
		if t.isEmptyTable {
			switch event.Rune() {
			case 'h': // go back to tree navigation
				t.showTreeView()
			case 'a': // add a new column
				form, err := t.createColForm(t.focusedCol)
				if err != nil {
					log.Println("Failed to add a column to the board: ", err)
					return event
				}
				t.showModal(form)
			}
			return event
		}

		focusedOnTable := false
		rows, cols := t.boardCols[t.focusedCol].GetSelectable()
		if rows == false && cols == false {
			focusedOnTable = true
		}
		row, _ := t.boardCols[t.focusedCol].GetSelection()

		if event.Rune() == 'L' || event.Key() == tcell.KeyEnter {
			t.enterSubBoard(focusedOnTable, row)
		}

		switch event.Rune() {
		case 'l': // move right
			if t.focusedCol < len(t.boardCols)-1 {
				t.boardCols[t.focusedCol].SetSelectable(false, false)
				t.focusedCol++
				t.boardCols[t.focusedCol].SetSelectable(true, false)

				if focusedOnTable {
					t.boardCols[t.focusedCol].SetSelectable(false, false)
				}
				t.app.SetFocus(t.boardCols[t.focusedCol])
			}
		case 'h': // move left
			// If at the first column, switch back to TreeView
			if t.focusedCol == 0 {
				t.boardCols[t.focusedCol].SetSelectable(false, false)
				t.showTreeView()
				return event
			}

			t.boardCols[t.focusedCol].SetSelectable(false, false)
			t.focusedCol--
			t.boardCols[t.focusedCol].SetSelectable(true, false)

			if focusedOnTable {
				t.boardCols[t.focusedCol].SetSelectable(false, false)
			}
			t.app.SetFocus(t.boardCols[t.focusedCol])
		}

		switch focusedOnTable {
		case true:
			t.boardColInputCapture(event)
		case false:
			t.boardTaskInputCapture(event, row)
		}

		return event
	})
}

// enterSubBoard enters a sub board.
func (t *TUI) enterSubBoard(focusedOnTable bool, row int) {
	// Ensure focus is on a task
	if focusedOnTable {
		return
	}
	col := &t.boardColsData[t.focusedCol]
	task, err := col.GetTask(t.calcTaskIdxBoard(row, t.rightPanelWidth))
	if err != nil {
		log.Printf("Failed to enter sub-board: %v\n", err)
		return
	}

	// If task has a child
	if task.GetHasChild() {
		parentNode := t.tree.GetCurrentNode()
		if _, ok := t.getBoardRef(parentNode); !ok {
			return
		}

		childBoard, err := t.treeData.GetBoard(task.GetChildID())
		if err != nil {
			log.Println("Failed to enter sub-board: ", err)
			return
		}

		// Find tree view node that references focused board column
		for _, node := range parentNode.GetChildren() {
			c, ok := getColRef(node)
			if !ok {
				log.Println("Failed to enter sub-board: tree view node isn't of type Column.")
				return
			}
			if c == col {
				// Find tree view node that references focused board task
				for _, n := range node.GetChildren() {
					board, ok := t.getBoardRef(n)
					if !ok {
						log.Println("Failed to enter sub-board: tree view node isn't of type Board.")
						return
					}
					if board == childBoard {
						t.tree.SetCurrentNode(n)
						t.showBoard(childBoard)
						return
					}
				}
			}
		}
	}
}

// boardColInputCapture captures input interactions specific to the
// currently focused board column.
func (t *TUI) boardColInputCapture(e *tcell.EventKey) *tcell.EventKey {
	switch e.Rune() {
	case 'j': // move down
		// Enable task selection
		t.boardCols[t.focusedCol].SetSelectable(true, false)
	case 'a':
		// If focus is on the entire table, create a new board column to
		// the right of the currently focused column.
		form, err := t.createColForm(t.focusedCol)
		if err != nil {
			log.Println("Failed to add a column to the board: ", err)
			return e
		}
		t.showModal(form)
	case 'e': // edit board column
		form := t.editColForm()
		t.showModal(form)
	case 'd': // Delete and buffer board column
		t.removeBoardCol()
	case 'p': // Paste board column
		t.pasteBoardCol()
	}
	return e
}

// removeBoardCol deletes and buffers a board column. This includes the
// removal of a board thats referenced by a task in the column and all
// its children.
func (t *TUI) removeBoardCol() {
	node := t.tree.GetCurrentNode()
	parentBoard, ok := t.getBoardRef(node)
	if !ok {
		log.Println("Failed to remove board column: current tree view node isn't of type Board.")
		return
	}
	t.treeData.ColBuff.Clear()

	// If a task in the column references a board, remove that board from
	// child board slice. Buffer the task and board and its chilren.
	for _, task := range t.boardColsData[t.focusedCol].GetTasks() {
		if task.GetHasChild() {
			t.removeRefCol(task, parentBoard)
		}
	}

	col, err := parentBoard.RemoveColumn(t.focusedCol)
	if err != nil {
		log.Printf("Failed to remove board column: %v\n", err)
		return
	}
	t.treeData.ColBuff.SetColumnBuff(col)

	// Update and show board
	t.showBoard(parentBoard)
	t.updateTree()
}

func (t *TUI) pasteBoardCol() {
	// Get current board
	node := t.tree.GetCurrentNode()
	board, ok := t.getBoardRef(node)
	if !ok {
		log.Println("Couldn't paste column: current tree view node isn't of type Board.")
		return
	}

	// Get buffered column
	column := t.treeData.ColBuff.GetColumnBuff()

	cpy, err := column.DeepCopy(t.treeData, board, t.treeData.ColBuff.GetChildBoards())
	if err != nil {
		log.Printf("Failed to paste board column: %v\n", err)
		return
	}

	// Insert column into board
	board.InsertColumn(cpy, t.focusedCol+1)

	// Update board and tree view
	t.showBoard(board)
	t.updateTree()
}

// boardTaskInputCapture captures input interactions specific to the
// current board task.
func (t *TUI) boardTaskInputCapture(e *tcell.EventKey, row int) *tcell.EventKey {
	switch e.Rune() {
	case 'k': // move up
		if row == 0 {
			// Disable task selection
			t.boardCols[t.focusedCol].SetSelectable(false, false)
		}
	case 'a': // add board task underneath current task
		form := t.createBoardTaskForm(t.calcTaskIdxBoard(row, t.rightPanelWidth))
		t.showModal(form)
	case 'e': // edit board tasks
		form, err := t.editBoardTaskForm(t.calcTaskIdxBoard(row, t.rightPanelWidth))
		if err != nil {
			log.Printf("Failed to edit board task: %v\n", err)
			return e
		}
		t.showModal(form)
	case 'd': // delete and buffer board task
		t.removeBoardTask(row)
	case 'p': // paste board task
		t.pasteBoardTask(row)
	case ' ': // Toggle task description
		t.toggleBoardTaskDesc(row)
	}
	return e
}

// removeBoardTask deletes and buffers a board task.
func (t *TUI) removeBoardTask(row int) {
	node := t.tree.GetCurrentNode()
	parentBoard, ok := t.getBoardRef(node)
	if !ok {
		log.Println("Failed to remove board task: current tree view node isn't of type Board.")
		return
	}
	col := &t.boardColsData[t.focusedCol]

	// Delete task from focused column
	idx := t.calcTaskIdxBoard(row, t.rightPanelWidth)
	task, err := col.Remove(idx)
	if err != nil {
		log.Printf("Failed to remove and buffer task: %v\n", err)
		return
	}
	t.treeData.TaskBuff.Clear()
	t.treeData.TaskBuff.SetTaskBuff(*task)

	// If task being deleted references a board, from it from child
	// boards slice. Buffer the task and board and its children.
	if task.GetHasChild() {
		t.removeRefTask(task, parentBoard)
	}

	// Update the focused column and tree view
	col.UpdatePriorities(idx)
	t.updateColumn(t.focusedCol)
	t.updateTree()
}

// pasteBoardTask reads buffered task and pastes it.
func (t *TUI) pasteBoardTask(row int) {
	// Get current board
	node := t.tree.GetCurrentNode()
	board, ok := t.getBoardRef(node)
	if !ok {
		log.Println("Couldn't paste task: current tree view node isn't of type Board.")
		return
	}

	idx := t.calcTaskIdxBoard(row, t.rightPanelWidth)

	// Read from buffer
	task := t.treeData.TaskBuff.GetTaskBuff()
	// If this board has no buffered task, return early.
	if task.Task == nil {
		return
	}

	cpy, err := task.DeepCopy(t.treeData, board, t.treeData.TaskBuff.GetChildBoards())
	if err != nil {
		log.Printf("Failed to paste board task: %v\n", err)
		return
	}

	col := &t.boardColsData[t.focusedCol]
	col.InsertTask(&cpy, idx+1)
	col.UpdatePriorities(idx)
	t.updateColumn(t.focusedCol)
	t.updateTree()
}

// toggleBoardTaskDesc toggles a board task description.
func (t *TUI) toggleBoardTaskDesc(row int) {
	idx := t.calcTaskIdxBoard(row, t.rightPanelWidth)
	// If calculated task index is within bounds, toggle task show
	// description status, and update rendered list.
	task, err := t.boardColsData[t.focusedCol].GetTask(idx)
	if err != nil {
		return
	}
	task.ShowDesc = !task.ShowDesc
	t.updateColumn(t.focusedCol)
}

// removeRefBoard moves a given board and all its children from the main
// list of boards to a buffer board list.
func (t *TUI) removeRefBoard(b *tasks.Board) {
	for _, id := range b.GetChildren() {
		board, err := t.treeData.GetBoard(id)
		if err != nil {
			log.Printf("Failed to buffer a referenced child board: %v\n", err)
			continue
		}
		t.treeData.BoardBuff.AddChild(board)
		t.treeData.RemoveChildBoard(board)
		t.removeRefBoard(board) // remove child board's children
	}
}

// removeRefCol removes a board referenced by a given task in a column.
func (t *TUI) removeRefCol(task tasks.BoardTask, parentBoard *tasks.Board) {
	board, err := t.treeData.GetBoard(task.GetChildID())
	if err != nil {
		log.Printf("Failed to buffer a referenced board in the column: %v\n", err)
		return
	}
	// Remove and buffer immediate referenced child
	if _, err = t.treeData.RemoveChildBoard(board); err != nil {
		log.Printf("Failed to remove and buffer child board: %v\n", err)
		return
	}
	t.treeData.ColBuff.AddChild(board)

	// Sever connection from parent board and referenced child board.
	// Note: Connection from task to child board must not be severed to
	// allow pasting in the future.
	if err = parentBoard.RemoveChild(board.GetID()); err != nil {
		log.Printf("Failed to remove child board from parent child boards slice: %v\n", err)
		return
	}

	// Buffer and remove any potential children
	t.removeBoardRefCol(board)
}

// removeBoardRefCol recursively buffers and removes a boards children.
// The board is referenced by a board task in a given board column.
func (t *TUI) removeBoardRefCol(b *tasks.Board) {
	for _, id := range b.GetChildren() {
		board, err := t.treeData.GetBoard(id)
		if err != nil {
			log.Printf("Failed to buffer a referenced child board: %v\n", err)
			continue
		}
		t.treeData.ColBuff.AddChild(board)
		t.treeData.RemoveChildBoard(board)
		t.removeBoardRefCol(board) // remove child board's children
	}
}

// removeRefTask removes a board referenced by a given task.
func (t *TUI) removeRefTask(task *tasks.BoardTask, parentBoard *tasks.Board) {
	board, err := t.treeData.GetBoard(task.GetChildID())
	if err != nil {
		log.Printf("Failed to buffer referenced board: %v\n", err)
		return
	}

	// Remove and buffer immediate referenced child
	_, err = t.treeData.RemoveChildBoard(board)
	if err != nil {
		log.Printf("Failed to remove and buffer child board: %v\n", err)
		return
	}
	t.treeData.TaskBuff.AddChild(board)

	// Sever connection from parent board and child board.
	// Note: Connection from task to child board must not be severed to
	// allow pasting in the future.
	parentBoard.RemoveChild(board.GetID())
	// Buffer and remove any potential children
	t.removeBoardRefTask(board)
}

// removeBoardRefTask recursively buffers and removes a boards children.
// The board is referenced by a board task.
func (t *TUI) removeBoardRefTask(b *tasks.Board) {
	for _, id := range b.GetChildren() {
		board, err := t.treeData.GetBoard(id)
		if err != nil {
			log.Printf("Failed to buffer a referenced child board: %v\n", err)
			continue
		}
		t.treeData.TaskBuff.AddChild(board)
		t.treeData.RemoveChildBoard(board)
		t.removeBoardRefTask(board) // remove child board's children
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
func (t *TUI) createListForm(idx int) *tview.Form {
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
		task.SetPriority(idx + 1)
		task.SetID(idx + 1)
		t.taskData.Add(task, idx+1)
		t.taskData.UpdatePriorities(idx + 1)

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

// createColForm creates and returns a tview form for creating a
// new board column.
// This function assumes that a task is currently selected.
func (t *TUI) createColForm(focusedCol int) (*tview.Form, error) {
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
		board.InsertColumn(*column, t.focusedCol+1)

		t.showBoard(board)

		// Update tree view to include the newly board column
		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
	})

	return form, nil
}

// createBoardTaskForm creates and returns a tview form for creating a
// new board task.
// This function assumes that a task is currently selected.
func (t *TUI) createBoardTaskForm(idx int) *tview.Form {
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
		task.SetPriority(idx + 1)
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

		col := &t.boardColsData[t.focusedCol]
		col.InsertTask(task, idx+1)
		col.UpdatePriorities(idx)

		// Update the column to show the newly added task
		t.updateColumn(t.focusedCol)

		// Update tree view to include the newly added task and its
		// child if created.
		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
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
func (t *TUI) editListForm(idx int) (*tview.Form, error) {
	task, err := t.taskData.GetTask(idx)
	if err != nil {
		return nil, err
	}
	name := task.GetName()
	description := task.GetDesc()
	isCore := task.GetIsCore()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Edit Task")

	// Define the input fields for the forms and update field variables if
	// user makes any changes to the default values.
	form.AddInputField("Name", task.GetName(), 20, nil, func(text string) {
		name = text
	})
	form.AddInputField("Description", task.GetDesc(), 50, nil, func(text string) {
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

	return form, nil
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

// editColForm creates and returns a tview form for editing a
// board column.
func (t *TUI) editColForm() *tview.Form {
	col := &t.boardColsData[t.focusedCol]
	name := col.GetTitle()

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
		col.SetTitle(name)

		// Update tview list
		t.updateColumn(t.focusedCol)

		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
	})

	return form
}

// editBoardTaskForm creates and returns a tview form for editing a
// todo list task.
func (t *TUI) editBoardTaskForm(idx int) (*tview.Form, error) {
	var createChildBoard bool
	task, err := t.boardColsData[t.focusedCol].GetTask(idx)
	if err != nil {
		return nil, err
	}
	name := task.GetName()
	desc := task.GetDesc()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Edit Task")

	// Define the input fields for the forms and update field variables if
	// user makes any changes to the default values.
	form.AddInputField("Name", name, 20, nil, func(text string) {
		name = text
	})
	form.AddInputField("Description", desc, 50, nil, func(text string) {
		desc = text
	})

	if !task.GetHasChild() {
		form.AddCheckbox("Create a Board?", false, func(checked bool) {
			createChildBoard = checked
		})
	}

	form.AddButton("Save", func() {
		task.SetName(name)
		if task.GetHasChild() {
			childBoard, err := t.treeData.GetBoard(task.GetChildID())
			if err != nil {
				log.Printf("Failed to create and add child board for %q task: %v\n", err)
				return
			}
			childBoard.SetTitle(name)
		}
		task.SetDescription(desc)

		if createChildBoard {
			if err := t.createAndAddChildBoard(name, task); err != nil {
				log.Printf("Failed to create and add child board for %q task: %v\n", name, err)
				return
			}
		}

		// Update tview list
		t.updateColumn(t.focusedCol)

		t.updateTree()

		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
	})

	form.AddButton("Cancel", func() {
		// Close the modal without doing anything
		t.closeModal()
		t.app.SetFocus(t.boardCols[t.focusedCol])
	})

	return form, nil
}

// closeModal removes the modal page
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

	board, err := t.treeData.GetBoard(nr.ID)
	if err != nil {
		log.Println("Failed to get board reference: ", err)
		return nil, false
	}
	return board, true
}

// getColRef returns the board column referenced by a TreeNode.
func getColRef(n *tview.TreeNode) (*tasks.BoardColumn, bool) {
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
func severConnection(parentTask *tasks.BoardTask, parentBoard, board *tasks.Board) {
	// Sever the connection from board to parent task
	board.SetParentTask(nil)

	// Remove connection from parent task to a board
	parentTask.SetChildID(-1)
	parentTask.SetHasChild(false)

	parentBoard.RemoveChild(board.GetID())
}

func InitForm() {
	app := tview.NewApplication()
	form := tview.NewForm()

	form.AddInputField("Task Title", "", 20, nil, nil).
		AddInputField("Task Description", "", 20, nil, nil).
		AddInputField("Expected Output", "", 20, nil, nil).
		AddTextArea("Constraints", "", 40, 0, 0, nil).
		AddInputField("Additional Information", "", 20, nil, nil)

	// Add button to submit form
	form.AddButton("Submit", func() {

		taskTitle := form.GetFormItem(0).(*tview.InputField).GetText()
		taskDescription := form.GetFormItem(1).(*tview.InputField).GetText()
		expectedOutput := form.GetFormItem(2).(*tview.InputField).GetText()
		constraints := form.GetFormItem(3).(*tview.TextArea).GetText()
		additionalInfo := form.GetFormItem(4).(*tview.InputField).GetText()

		app.Stop()

		// Print the form details
		fmt.Printf("# Task Breakdown Request\n\n")
		fmt.Printf("Task Title: %s\n\n", taskTitle)
		fmt.Printf("Task Description: %s\n\n", taskDescription)
		fmt.Printf("Expected Output: %s\n\n", expectedOutput)
		fmt.Printf("Additional Information: %s\n\n", additionalInfo)
		fmt.Printf("Constraints:\n%s\n\n", constraints)
		fmt.Println("Please provide a detailed breakdown of this task into main tasks, sub-tasks, and if possible, even further sub-tasks. Also, identify any dependencies or priority considerations.")
	})

	// Add a button to cancel form and go back to the previous view
	form.AddButton("Cancel", func() {
		// Close form and stop the application
		app.Stop()
	})

	// Handle keyboard shortcuts
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			// Quit the application when 'q' is pressed
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.SetRoot(form, true).Run(); err != nil {
		panic(err)
	}
}
