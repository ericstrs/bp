package tasks

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

type BoardBuffer struct {
	BoardBuff   Board    `yaml:"board"`
	ChildBoards []*Board `yaml:"child_boards"`
}

type ColumnBuffer struct {
	ColumnBuff  BoardColumn `yaml:"column"`
	ChildBoards []*Board    `yaml:"child_boards"`
}

type TaskBuffer struct {
	TaskBuff    BoardTask `yaml:"task"`
	ChildBoards []*Board  `yaml:"child_boards"`
}

type BoardTree struct {
	RootBoards     []*Board `yaml:"root_boards"`
	ChildBoards    []*Board `yaml:"child_boards"`
	CurrentBoardID int      `yaml:"-"` //`yaml:"current_board_id"`

	BoardBuff BoardBuffer  `yaml:"board_buffer"`
	ColBuff   ColumnBuffer `yaml:"column_buffer"`
	TaskBuff  TaskBuffer   `yaml:"task_buffer"`

	BoardCounter int `yaml:"board_counter"`
	TaskCounter  int `yaml:"task_counter"`
}

type Board struct {
	ID         int           `yaml:"id"`
	Title      string        `yaml:"title"`
	ParentTask *BoardTask    `yaml:"-"` //`yaml:"parent_task_id"`
	Columns    []BoardColumn `yaml:"columns"`

	Children []int `yaml:"children"`
}

type BoardColumn struct {
	Title string      `yaml:"title"`
	Tasks []BoardTask `yaml:"tasks"`
}

type BoardTask struct {
	*Task
	ChildID  int  `yaml:"child_id"`
	HasChild bool `yaml:"has_child"`
}

var _ Taskable = &BoardTask{}

func (bb BoardBuffer) GetBoardBuff() Board { return bb.BoardBuff }

func (bb *BoardBuffer) SetBoardBuff(b Board) { bb.BoardBuff = b }

func (bb *BoardBuffer) Clear() { bb.ChildBoards = bb.ChildBoards[:0] }

func (bb BoardBuffer) GetChildBoards() []*Board { return bb.ChildBoards }

func (bb *BoardBuffer) AddChild(b *Board) { bb.InsertChild(b, -1) }

func (bb *BoardBuffer) InsertChild(b *Board, index int) {
	// If index out of range, then append child board to the slice.
	if index < 0 || index >= len(bb.ChildBoards) {
		bb.ChildBoards = append(bb.ChildBoards, b)
		return
	}
	// Otherwise, insert child board at the specified index.
	bb.ChildBoards = append(bb.ChildBoards[:index+1], bb.ChildBoards[index:]...)
	bb.ChildBoards[index] = b
}

func (bb *BoardBuffer) RemoveChild(b *Board) (Board, error) {
	// Loop through all child boards to find the match.
	for idx, board := range bb.ChildBoards {
		// Using pointer equality for comparison
		if board == b {
			// Copy board before removing.
			cpy := bb.ChildBoards[idx]

			// Remove the child board by slicing.
			bb.ChildBoards = append(bb.ChildBoards[:idx], bb.ChildBoards[idx+1:]...)
			return *cpy, nil
		}
	}

	// Return an error if the child board is not found.
	return Board{}, fmt.Errorf("child board with id %d not found", b.ID)
}

func (cb ColumnBuffer) GetColumnBuff() BoardColumn { return cb.ColumnBuff }

func (cb *ColumnBuffer) SetColumnBuff(b BoardColumn) { cb.ColumnBuff = b }

func (cb *ColumnBuffer) Clear() { cb.ChildBoards = cb.ChildBoards[:0] }

func (cb ColumnBuffer) GetChildBoards() []*Board { return cb.ChildBoards }

func (cb *ColumnBuffer) AddChild(b *Board) { cb.InsertChild(b, -1) }

func (cb *ColumnBuffer) InsertChild(b *Board, index int) {
	// If index out of range, then append child board to the slice.
	if index < 0 || index >= len(cb.ChildBoards) {
		cb.ChildBoards = append(cb.ChildBoards, b)
		return
	}
	// Otherwise, insert child board at the specified index.
	cb.ChildBoards = append(cb.ChildBoards[:index+1], cb.ChildBoards[index:]...)
	cb.ChildBoards[index] = b
}

func (cb *ColumnBuffer) RemoveChild(b *Board) (Board, error) {
	// Loop through all child boards to find the match.
	for idx, board := range cb.ChildBoards {
		// Using pointer equality for comparison
		if board == b {
			// Copy board before removing.
			cpy := cb.ChildBoards[idx]

			// Remove the child board by slicing.
			cb.ChildBoards = append(cb.ChildBoards[:idx], cb.ChildBoards[idx+1:]...)
			return *cpy, nil
		}
	}

	// Return an error if the child board is not found.
	return Board{}, fmt.Errorf("child board with id %d not found", b.ID)
}

func (tb TaskBuffer) GetTaskBuff() BoardTask { return tb.TaskBuff }

func (tb *TaskBuffer) SetTaskBuff(b BoardTask) { tb.TaskBuff = b }

func (tb *TaskBuffer) Clear() { tb.ChildBoards = tb.ChildBoards[:0] }

func (tb TaskBuffer) GetChildBoards() []*Board { return tb.ChildBoards }

func (tb *TaskBuffer) AddChild(b *Board) { tb.InsertChild(b, -1) }

func (tb *TaskBuffer) InsertChild(b *Board, index int) {
	// If index out of range, then append child board to the slice.
	if index < 0 || index >= len(tb.ChildBoards) {
		tb.ChildBoards = append(tb.ChildBoards, b)
		return
	}
	// Otherwise, insert child board at the specified index.
	tb.ChildBoards = append(tb.ChildBoards[:index+1], tb.ChildBoards[index:]...)
	tb.ChildBoards[index] = b
}

func (tb *TaskBuffer) RemoveChild(b *Board) (Board, error) {
	// Loop through all child boards to find the match.
	for idx, board := range tb.ChildBoards {
		// Using pointer equality for comparison
		if board == b {
			// Copy board before removing.
			cpy := tb.ChildBoards[idx]

			// Remove the child board by slicing.
			tb.ChildBoards = append(tb.ChildBoards[:idx], tb.ChildBoards[idx+1:]...)
			return *cpy, nil
		}
	}

	// Return an error if the child board is not found.
	return Board{}, fmt.Errorf("child board with id %d not found", b.ID)
}

func (tree BoardTree) GetRootBoards() []*Board { return tree.RootBoards }

func (tree BoardTree) GetChildBoards() []*Board { return tree.ChildBoards }

// GetBoard finds and returns a board. This function first searches
// through child board and then root boards.
func (tree *BoardTree) GetBoard(id int) (*Board, error) {
	for _, board := range tree.ChildBoards {
		if board.ID == id {
			return board, nil
		}
	}

	for _, board := range tree.RootBoards {
		if board.ID == id {
			return board, nil
		}
	}
	return nil, fmt.Errorf("couldn't find board with id = %d\n", id)
}

func (tree *BoardTree) AddRoot(b *Board) { tree.InsertRoot(b, -1) }

func (tree *BoardTree) InsertRoot(b *Board, index int) {
	// If index out of range, then append root board to the slice.
	if index < 0 || index >= len(tree.RootBoards) {
		tree.RootBoards = append(tree.RootBoards, b)
		return
	}
	// Otherwise, insert child board at the specified index.
	tree.RootBoards = append(tree.RootBoards[:index+1], tree.RootBoards[index:]...)
	tree.RootBoards[index] = b
}

func (tree *BoardTree) RemoveRoot(b *Board) (Board, error) {
	// Loop through all root boards to find the match.
	for idx, board := range tree.RootBoards {
		// Using pointer equality for comparison
		if board == b {
			// Copy board before removing.
			cpy := tree.RootBoards[idx]

			// Remove the root board by slicing.
			tree.RootBoards = append(tree.RootBoards[:idx], tree.RootBoards[idx+1:]...)
			return *cpy, nil
		}
	}

	// Return an error if the root board is not found.
	return Board{}, fmt.Errorf("root board with id %d not found", b.ID)
}

func (tree *BoardTree) AddChildBoard(b *Board) { tree.InsertChildBoard(b, -1) }

func (tree *BoardTree) InsertChildBoard(b *Board, index int) {
	// If index out of range, then append child board to the slice.
	if index < 0 || index >= len(tree.ChildBoards) {
		tree.ChildBoards = append(tree.ChildBoards, b)
		return
	}
	// Otherwise, insert child board at the specified index.
	tree.ChildBoards = append(tree.ChildBoards[:index+1], tree.ChildBoards[index:]...)
	tree.ChildBoards[index] = b
}

func (tree *BoardTree) RemoveChildBoard(b *Board) (Board, error) {
	// Loop through all child boards to find the match.
	for idx, board := range tree.ChildBoards {
		// Using pointer equality for comparison
		if board == b {
			// Copy board before removing.
			cpy := tree.ChildBoards[idx]

			// Remove the child board by slicing.
			tree.ChildBoards = append(tree.ChildBoards[:idx], tree.ChildBoards[idx+1:]...)
			return *cpy, nil
		}
	}

	// Return an error if the child board is not found.
	return Board{}, errors.New("child board not found")
}

func (tree BoardTree) GetCurrentBoardID() int { return tree.CurrentBoardID }

func (tree *BoardTree) SetCurrentBoardID(id int) { tree.CurrentBoardID = id }

// NewBoard creates and returns a default board. A default board
// consists of three empty columns: TODO, Working On, and Done.
func (tree *BoardTree) NewBoard(name string) *Board {
	// Create a new board
	board := new(Board)
	board.SetTitle(name)
	tree.BoardCounter++
	board.SetID(tree.BoardCounter)

	// Add default columns
	for _, title := range []string{"TODO", "Working On", "Done"} {
		column := new(BoardColumn)
		column.SetTitle(title)
		board.AddColumn(*column)
	}

	return board
}

func (tree BoardTree) GetBoardCtr() int { return tree.BoardCounter }

func (tree *BoardTree) IncrementBoardCtr() { tree.BoardCounter++ }

func (tree BoardTree) GetTaskCtr() int { return tree.TaskCounter }

func (tree *BoardTree) IncrementTaskCtr() { tree.TaskCounter++ }

// DeepCopy creates a deep copy of the board, its children, and columns.
// Returns a pointer to the new Board object, or an error if the
// original board is null.
//
// This function is both computationally and memory intensive due to its
// recursive nature, as it creates new instances for all children and
// grandchildren, and so on.
func (b *Board) DeepCopy(parentTask *BoardTask, tree *BoardTree, childBoards []*Board) (*Board, error) {
	// Handle null board
	if b == nil {
		return nil, errors.New("board is null")
	}

	tree.IncrementBoardCtr()
	// Create a new Board pointer and populate its fields with the data from the original board.
	newBoard := &Board{
		ID:         tree.GetBoardCtr(), // Assign newly generated ID
		Title:      b.Title,            // Copy the title
		ParentTask: parentTask,
	}

	// Deep copy columns
	for _, col := range b.Columns {
		cpy, err := col.DeepCopy(tree, newBoard, childBoards)
		if err != nil {
			return nil, err
		}
		newBoard.Columns = append(newBoard.Columns, cpy)
	}

	return newBoard, nil
}

func (c *BoardColumn) DeepCopy(tree *BoardTree, parentBoard *Board, childBoards []*Board) (BoardColumn, error) {
	newColmun := BoardColumn{
		Title: c.Title,
	}

	for _, task := range c.Tasks {
		cpy, err := task.DeepCopy(tree, parentBoard, childBoards)
		if err != nil {
			return BoardColumn{}, err
		}
		newColmun.Tasks = append(newColmun.Tasks, cpy)
	}

	return newColmun, nil
}

func (t *BoardTask) DeepCopy(tree *BoardTree, parentBoard *Board, childBoards []*Board) (BoardTask, error) {
	tree.IncrementTaskCtr()
	newTask := &Task{
		Id:          tree.GetTaskCtr(),
		Name:        t.Name,
		Description: t.Description,
		ShowDesc:    t.ShowDesc,
		Started:     t.Started,
		Finished:    t.Finished,
		Priority:    t.Priority,
	}

	newBoardTask := BoardTask{
		Task:     newTask,
		HasChild: t.HasChild,
	}

	if t.HasChild {
		var childBoard *Board
		childBoard = nil

		// Deep copying a root board is breaking here. childBoards is
		// RootBoards, but t.childID will live in BoardBuff.ChildBoards
		// That being said, need to iterate over both lists.
		for _, board := range childBoards {
			if board.ID == t.ChildID {
				childBoard = board
			}
		}

		if childBoard == nil {
			log.Printf("Failed to set child board field: couldn't find board with id = %d\n", t.ChildID)
			return newBoardTask, nil
		}
		cpy, err := childBoard.DeepCopy(&newBoardTask, tree, childBoards)
		if err != nil {
			log.Printf("Failed to set child board field: %v\n", err)
			return newBoardTask, nil
		}
		tree.AddChildBoard(cpy)
		parentBoard.AddChild(cpy.ID)
		newBoardTask.ChildID = cpy.ID
	}
	return newBoardTask, nil
}

func (b Board) GetID() int { return b.ID }

func (b *Board) SetID(id int) { b.ID = id }

func (b Board) GetTitle() string { return b.Title }

func (b *Board) SetTitle(t string) { b.Title = t }

func (b Board) GetParentTask() *BoardTask { return b.ParentTask }

func (b *Board) SetParentTask(pt *BoardTask) { b.ParentTask = pt }

func (b Board) GetColumns() []BoardColumn { return b.Columns }

func (b *Board) AddColumn(c BoardColumn) { b.InsertColumn(c, -1) }

func (b *Board) InsertColumn(c BoardColumn, index int) {
	// Ensure index in within the correct range.
	if index < 0 || index >= len(b.Columns) {
		b.Columns = append(b.Columns, c)
		return
	}
	// Otherwise, insert the column at the specified index.
	b.Columns = append(b.Columns[:index+1], b.Columns[index:]...)
	b.Columns[index] = c
}

func (b *Board) RemoveColumn(index int) (BoardColumn, error) {
	// Ensure index is in the correct range.
	if index < 0 || index >= len(b.Columns) {
		return BoardColumn{}, fmt.Errorf("index %d out of range", index)
	}

	// Copy column before removing.
	cpy := b.Columns[index]

	// Remove column
	b.Columns = append(b.Columns[:index], b.Columns[index+1:]...)

	return cpy, nil
}

func (b Board) GetChildren() []int { return b.Children }

func (b *Board) AddChild(id int) { b.InsertChild(id, -1) }

func (b *Board) InsertChild(id, index int) {
	// Ensure index in within the correct range.
	if index < 0 || index >= len(b.Children) {
		b.Children = append(b.Children, id)
		return
	}
	// Otherwise, insert the board id at the specified index.
	b.Children = append(b.Children[:index+1], b.Children[index:]...)
	b.Children[index] = id
}

func (b *Board) RemoveChild(id int) error {
	// Loop through all child boards ids to find the match.
	for idx, boardID := range b.Children {
		if boardID == id {
			// Remove the child board by slicing.
			b.Children = append(b.Children[:idx], b.Children[idx+1:]...)
			return nil
		}
	}
	// Return an error if the child board is not found.
	return errors.New("child board not found")
}

func (bc BoardColumn) GetTitle() string { return bc.Title }

func (bc *BoardColumn) SetTitle(t string) { bc.Title = t }

func (bc BoardColumn) GetTask(index int) (*BoardTask, error) {
	// Ensure index is in the correct range.
	if err := bc.Bounds(index); err != nil {
		return nil, err
	}
	return &bc.Tasks[index], nil
}

func (bc BoardColumn) GetTasks() []BoardTask { return bc.Tasks }

func (bc *BoardColumn) UpdatePriorities(start int) error {
	if err := bc.Bounds(start); err != nil {
		return fmt.Errorf("Failed to update task priorities: %v\n", err)
	}

	var wg sync.WaitGroup

	for i := start; i < len(bc.Tasks); i++ {
		// Increment the WaitGroup counter
		wg.Add(1)

		go func(i int) {
			// Decrement the WaitGroup counter when the goroutine completes
			defer wg.Done()

			// Update the task's priority
			bc.Tasks[i].SetPriority(i)
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	return nil
}

func (bc *BoardColumn) Add(task *BoardTask) error {
	return bc.InsertTask(task, -1)
}

func (bc *BoardColumn) InsertTask(task *BoardTask, index int) error {
	// If index is out of range, then append task to the slice.
	if err := bc.Bounds(index); err != nil {
		bc.Tasks = append(bc.Tasks, *task)
		return nil
	}

	// Otherwise, insert task at the specified index.
	bc.Tasks = append(bc.Tasks[:index+1], bc.Tasks[index:]...)
	bc.Tasks[index] = *task

	return nil
}

func (bc *BoardColumn) Remove(index int) (*BoardTask, error) {
	// Ensure index is in the correct range.
	if err := bc.Bounds(index); err != nil {
		return &BoardTask{}, fmt.Errorf("Failed to remove element from BoardColumn: %v", err)
	}

	// Copy task before removing.
	cpy := bc.Tasks[index]

	// Remove task from slice.
	bc.Tasks = append(bc.Tasks[:index], bc.Tasks[index+1:]...)

	return &cpy, nil
}

func (bc BoardColumn) Bounds(index int) error {
	if index < 0 || index >= len(bc.Tasks) {
		return fmt.Errorf("index %d out of range.", index)
	}
	return nil
}

func (bt *BoardTask) SetTask(t *Task) { bt.Task = t }

func (bt BoardTask) GetChildID() int {
	return bt.ChildID
}

func (bt *BoardTask) SetChildID(id int) { bt.ChildID = id }

func (bt BoardTask) GetHasChild() bool { return bt.HasChild }

func (bt *BoardTask) SetHasChild(b bool) { bt.HasChild = b }
