package tasks

import (
	"fmt"
	"sync"
)

type TodoTask struct {
	*Task
	IsCore bool `yaml:"isCore"` // Indicates if this task is a "core" task, meaning it recurs daily
}

var _ Taskable = &TodoTask{} // Compile-time check for interface satisfaction.

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

func (t *TodoList) GetTitle() string { return t.Title }

func (t *TodoList) SetTitle(s string) { t.Title = s }

func (t *TodoList) GetTask(index int) *TodoTask { return &t.Tasks[index] }

func (t *TodoList) GetTasks() []TodoTask { return t.Tasks } /*[]Taskable {
	taskables := make([]Taskable, len(t.Tasks))
	for i, task := range t.Tasks {
		taskables[i] = Taskable(task)
	}
	return taskables
}
*/

// Buffer returns the buffered task.
func (t *TodoList) Buffer() *TodoTask { return t.buffer }

// SetBuffer sets the buffered task.
func (t *TodoList) SetBuffer(task *TodoTask) { t.buffer = task }

func (t *TodoList) UpdatePriorities(start int) error {
	if err := t.Bounds(start); err != nil {
		return fmt.Errorf("Failed to update task priorities: %v\n", err)
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

func (t *TodoList) Bounds(index int) error {
	if index < 0 || index >= len(t.Tasks) {
		return fmt.Errorf("index %d out of range.", index)
	}
	return nil
}
