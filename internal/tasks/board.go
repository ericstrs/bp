package tasks

import (
	"errors"
	"fmt"
	"sync"
)

type BoardTree struct {
	RootBoards   []*Board    `yaml:"root_boards"`
	Current      *Board      `yaml:"current"`
	BoardBuffer  Board       `yaml:"board_buffer"`
	ColumnBuffer BoardColumn `yaml:"column_buffer"`
}

type Board struct {
	ID          int           `yaml:"id"`
	Title       string        `yaml:"title"`
	ParentBoard *Board        `yaml:"parent_board"`
	ParentTask  *BoardTask    `yaml:"parent_task"`
	Children    []*Board      `yaml:"children"`
	Columns     []BoardColumn `yaml:"columns"`
	Buffer      *BoardTask    `yaml:"buffer"`
}

type BoardColumn struct {
	Title string      `yaml:"title"`
	Tasks []BoardTask `yaml:"tasks"`
}

type BoardTask struct {
	*Task
	Child    *Board `yaml:"child"`
	HasChild bool   `yaml:"has_child"`
}

var _ Taskable = &BoardTask{}

func (tree BoardTree) GetRootBoards() []*Board { return tree.RootBoards }

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
	return Board{}, errors.New("root board not found")
}

func (tree BoardTree) GetCurrent() *Board { return tree.Current }

func (tree *BoardTree) SetCurrent(b *Board) { tree.Current = b }

func (tree BoardTree) GetBoardBuffer() Board { return tree.BoardBuffer }

func (tree *BoardTree) SetBoardBuffer(b Board) { tree.BoardBuffer = b }

func (tree BoardTree) GetColumnBuffer() BoardColumn { return tree.ColumnBuffer }

func (tree *BoardTree) SetColumnBuffer(c BoardColumn) { tree.ColumnBuffer = c }

// NewBoard creates and returns a default board. A default board
// consists of three empty columns: TODO, Working On, and Done.
func NewBoard(name string) *Board {
	// Create a new board
	board := new(Board)
	board.SetTitle(name)

	// Add default columns
	for _, title := range []string{"TODO", "Working On", "Done"} {
		column := new(BoardColumn)
		column.SetTitle(title)
		board.AddColumn(column)
	}

	return board
}

// DeepCopy creates a deep copy of the board, its children, and columns.
// Returns a pointer to the new Board object, or an error if the
// original board is null.
//
// This function is both computationally and memory intensive due to its
// recursive nature, as it creates new instances for all children and
// grandchildren, and so on.
//
// Note: The fields 'ParentBoard' and 'ParentTask' are references to objects
// higher up in the tree, so they are not deep-copied by this function.
// Board buffers are also not copied.
func (b *Board) DeepCopy() (*Board, error) {
	// Handle null board
	if b == nil {
		return nil, errors.New("board is null")
	}

	// Create a new Board pointer and populate its fields with the data from the original board.
	newBoard := &Board{
		ID:    b.ID,    // TODO: generate a new ID for the new board
		Title: b.Title, // Copy the title
	}

	// Deep copy each child board.
	for _, child := range b.Children {
		copiedChild, err := child.DeepCopy()
		if err != nil {
			return nil, err
		}
		newBoard.Children = append(newBoard.Children, copiedChild)
	}

	// Shallow copy columns as these are value types
	newBoard.Columns = append(newBoard.Columns, b.Columns...)

	return newBoard, nil
}

func (b Board) GetTitle() string { return b.Title }

func (b *Board) SetTitle(t string) { b.Title = t }

func (b Board) GetParentBoard() *Board { return b.ParentBoard }

func (b *Board) SetParentBoard(pb *Board) { b.ParentBoard = pb }

func (b Board) GetParentTask() *BoardTask { return b.ParentTask }

func (b *Board) SetParentTask(pt *BoardTask) { b.ParentTask = pt }

func (b Board) GetChildren() []*Board { return b.Children }

func (b *Board) AddChild(c *Board) { b.InsertChild(c, -1) }

func (b *Board) InsertChild(c *Board, index int) {
	// If index out of range, then append child board to the slice.
	if index < 0 || index >= len(b.Children) {
		b.Children = append(b.Children, c)
		return
	}
	// Otherwise, insert child board at the specified index.
	b.Children = append(b.Children[:index+1], b.Children[index:]...)
	b.Children[index] = c
}

// RemoveChild removes a given child from the parent board's children
// slice. If the child board is not found, it return and error.
func (b *Board) RemoveChild(c *Board) error {
	// Loop through all children to find the match.
	for idx, child := range b.Children {
		// Using pointer equality for comparison
		if child == c {
			// Remove the child board by slicing.
			b.Children = append(b.Children[:idx], b.Children[idx+1:]...)
			return nil
		}
	}

	// Return an error if the child is not found.
	return errors.New("child not found")
}

func (b Board) GetColumns() []BoardColumn { return b.Columns }

func (b *Board) AddColumn(c *BoardColumn) { b.InsertColumn(c, -1) }

func (b *Board) InsertColumn(c *BoardColumn, index int) {
	// Ensure index in within the correct range.
	if index < 0 || index >= len(b.Columns) {
		b.Columns = append(b.Columns, *c)
		return
	}
	// Otherwise, insert the column at the specified index.
	b.Columns = append(b.Columns[:index+1], b.Columns[index:]...)
	b.Columns[index] = *c
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

func (b Board) GetBuffer() *BoardTask { return b.Buffer }

func (b *Board) SetBuffer(t *BoardTask) { b.Buffer = t }

func (bc BoardColumn) GetTitle() string { return bc.Title }

func (bc *BoardColumn) SetTitle(t string) { bc.Title = t }

func (bc BoardColumn) GetTask(index int) *BoardTask { return &bc.Tasks[index] }

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

func (bt BoardTask) GetChild() *Board { return bt.Child }

func (bt *BoardTask) SetChild(c *Board) { bt.Child = c }

func (bt BoardTask) GetHasChild() bool { return bt.HasChild }

func (bt *BoardTask) SetHasChild(b bool) { bt.HasChild = b }
