package tasks

import (
	"fmt"
	"sync"
)

type TodoTask struct {
	*Task
	IsCore bool `yaml:"isCore"` // Indicates if this task is a "core" task, meaning it recurs daily
}

type TodoList struct {
	Title  string     `yaml:"title"`
	Tasks  []TodoTask `yaml:"tasks"`
	buffer *TodoTask
}

//var _ TaskList = &TodoList{}

// SetTask sets the general task.
func (task *TodoTask) SetTask(t *Task) { task.Task = t }

// IsCore returns whether or not a task is a "core" task.
func (task *TodoTask) GetIsCore() bool { return task.IsCore }

// SetCore sets whether or not a task is a "core" task.
func (task *TodoTask) SetCore(b bool) { task.IsCore = b }

// GetTitle returns the title of the todo list.
func (t *TodoList) GetTitle() string { return t.Title }

// SetTitle sets the title of the todo list.
func (t *TodoList) SetTitle(s string) { t.Title = s }

// GetTask returns the task at the given index.
func (t *TodoList) GetTask(index int) (*TodoTask, error) {
	if err := t.Bounds(index); err != nil {
		return nil, fmt.Errorf("failed to get task: %v\n", err)
	}
	return &t.Tasks[index], nil
}

// GetTasks returns the list of tasks.
func (t *TodoList) GetTasks() []TodoTask { return t.Tasks }

// Buffer returns the buffered task.
func (t *TodoList) Buffer() *TodoTask { return t.buffer }

// SetBuff sets the buffered task.
func (t *TodoList) SetBuff(task *TodoTask) { t.buffer = task }

// UpdatePriorities updates the priorities of tasks from the given
// start index.
func (t *TodoList) UpdatePriorities(start int) error {
	if err := t.Bounds(start); err != nil {
		return fmt.Errorf("failed to update task priorities: %v\n", err)
	}

	var wg sync.WaitGroup

	for i := start; i < len(t.Tasks); i++ {
		// Increment the WaitGroup counter
		wg.Add(1)

		go func(i int) {
			// Decrement the WaitGroup counter when the goroutine completes
			defer wg.Done()

			// Update the task's priority
			t.Tasks[i].SetPriority(i)
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	return nil
}

// Add adds a task to the todo list at the given index.
//
// Important Considerations:
//
//  1. Update Priority: The removal of a task can affect the
//     priorities of the remaining tasks. Calling [UpdatePriorities]
//     post-removal should be done to update task priorities
//     accordingly.
//
// Note: The priority updating of the tasks in the list is assumed to
// be handled outside this function, and should be addressed post-add
// operation.
func (t *TodoList) Add(task *TodoTask, index int) error {
	// If index is out of range, then append task to the slice.
	if err := t.Bounds(index); err != nil {
		t.Tasks = append(t.Tasks, *task)
		return nil
	}
	// Otherwise, insert task at the specified index.
	t.Tasks = append(t.Tasks[:index+1], t.Tasks[index:]...)
	t.Tasks[index] = *task

	return nil
}

// Remove removes a task from the todo list by its index and returns the
// removed task for buffering.
//
// Important Considerations:
//
//  1. Buffering. This function returns the removed task, which should
//     be buffered.
//  2. Update Priority: The removal of a task can affect the
//     priorities of the remaining tasks. Calling a sort function
//     post-removal should be done to update task priorities accordingly.
//
// Note: The buffering of the removed task and priority updating is
// assumed to be handled outside this function, and they should be
// addressed post-removal operation.
func (t *TodoList) Remove(index int) (*TodoTask, error) {
	// Ensure index is in the correct range.
	if err := t.Bounds(index); err != nil {
		return &TodoTask{}, fmt.Errorf("Failed to remove element from TodoList: %v", err)
	}

	// Copy task before removing.
	cpy := t.Tasks[index]

	// Remove task from slice.
	t.Tasks = append(t.Tasks[:index], t.Tasks[index+1:]...)

	return &cpy, nil
}

// Bounds checks if an index is within range.
func (t *TodoList) Bounds(index int) error {
	if index < 0 || index >= len(t.Tasks) {
		return fmt.Errorf("index %d out of range.", index)
	}
	return nil
}
