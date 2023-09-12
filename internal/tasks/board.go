package tasks

import (
	"errors"
	"fmt"
	"sync"
)

type BoardTree struct {
	Root    *Board `yaml:"root"`
	Current *Board `yaml:"current"`
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
	Child *Board `yaml:"child"`
}

var _ Taskable = &BoardTask{}

func (tree BoardTree) GetRoot() *Board { return tree.Root }

func (tree *BoardTree) SetRoot(b *Board) { tree.Root = b }

func (tree BoardTree) GetCurrent() *Board { return tree.Current }

func (tree *BoardTree) SetCurrent(b *Board) { tree.Current = b }

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

func (b *Board) RemoveChild(index int) error {
	// Ensure index in within the correct range.
	if index < 0 || index >= len(b.Children) {
		return errors.New("failed to remove child board")
	}
	b.Children = append(b.Children[:index], b.Children[index+1:]...)
	return nil
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
