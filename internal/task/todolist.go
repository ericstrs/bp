package tasks

import (
	"fmt"
	"sync"
)

type TodoTask struct {
	*Task
	isCore bool // Indicates if this task is a "core" task, meaning it recurs daily
}

var _ Taskable = &TodoTask{} // Compile-time check for interface satisfaction.

type TodoList struct {
	title  string
	tasks  []TodoTask
	buffer *TodoTask
}

func (t *TodoList) UpdatePriorities(start int) error {
	if err := t.Bounds(start); err != nil {
		return fmt.Errorf("Failed to update task priorities: %v\n", err)
	}

	var wg sync.WaitGroup

	for i := start; i < len(t.tasks); i++ {
		// Increment the WaitGroup counter
		wg.Add(1)

		go func(i int) {
			// Decrement the WaitGroup counter when the goroutine completes
			defer wg.Done()

			// Update the task's priority
			t.tasks[i].SetPriority(i)
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

func (t *TodoList) Add(task *TodoTask, index int) error {
	// If index is out of range, then append task to the slice.
	if err := t.Bounds(index); err != nil {
		t.tasks = append(t.tasks, *task)
		return nil
	}

	// Otherwise, insert task at the specified index.
	t.tasks = append(t.tasks[:index+1], t.tasks[index:]...)
	t.tasks[index] = *task

	return nil
}

func (t *TodoList) Remove(index int) (*TodoTask, error) {
	// Ensure index is in the correct range.
	if err := t.Bounds(index); err != nil {
		return &TodoTask{}, fmt.Errorf("Failed to remove element from TodoList: %v", err)
	}

	// Copy task before removing.
	cpy := t.tasks[index]

	// Remove task from slice.
	t.tasks = append(t.tasks[:index], t.tasks[index+1:]...)

	return &cpy, nil
}

func (t *TodoList) Bounds(index int) error {
	if index < 0 || index >= len(t.tasks) {
		return fmt.Errorf("index %d out of range.", index)
	}
	return nil
}
