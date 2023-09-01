// Package tasks implements types and functions to manage tasks in
// to-do lists and kanban boards.
//
// This	package defines a Task struct that contains the common
// fields shared by all types of tasks, as well as a Taskable interface
// that specifies methods all task types must implement.
//
// This package currently implements task prioritization at the task
// level. That is, each task has a priority field, where a lower value
// indicates more importance. Task-level priority was chosen for its
// straightforward implementation.
//
// Future Possibilities:
//
//  1. Dynamic Priority Algorith: Implementing an algorithm that
//     automatically assigns and adjusts task priority based on various
//     factors such as deadlines, dependencies, etc.
//  2. Priority Queue: A more complex data structure like a priority
//     queue could be employed to make add/delete operations more efficient.
//  3. Custom Sort Queue: Allow custom sorting functions, allowing users
//     to define their own criteria for what makes a task "important."
//
// To swap out different prioritization methods, one could implement a
// Prioritization interface. By doing so, different sorting and
// prioritization algorithms can be swapped in and out without affecting
// the existing code base.
//
// Extensibility was kept in mind to enable developers to embed the
// Task struct into their own specific task type, adding additional
// fields and methods as needed.
package tasks

import "time"

// The Taskable type describes the methods a task must implement.
type Taskable interface {
	// ID returns the unique identifier of the task.
	ID() int
	// SetID sets the unique identifier of the task.
	SetID(id int)
	// Name returns the name of the task.
	Name() string
	// SetName sets the name of the task.
	SetName(s string)
	// Description returns the description of the task.
	Description() string
	// SetDescription sets the description of the task.
	SetDescription(d string)
	// DoneDate returns the date when the task was last finished.
	Finished() time.Time
	// SetFinished sets the date when the task was last finished.
	SetFinished(d time.Time)
	// Started returns the date when the task was last started.
	Started() time.Time
	// SetStarted sets the date when the task was last started.
	SetStarted(d time.Time)
	// Priority returns the priority of the task.
	Priority() int
	// SetPriority sets the priority of the task.
	SetPriority(p int)
	// IsDone returns whether or not the task is finished.
	IsDone() bool
	// SetDone sets the task "finished" status to true.
	SetDone(b bool)
}

// A Tasktlist represents a list of tasks.
type TaskList interface {
	// Bounds check if an index is within range.
	Bounds(index int) error
	// UpdatePriorities updates the priorities of tasks from the given
	// start index.
	UpdatePriorities(start int) error

	// Add adds a task to the list at a given index.
	//
	// Important Considerations:
	//
	//	1. Update Priority: The removal of a task can affect the
	//     priorities of the remaining tasks. Calling [Sort] post-removal
	//     should be done to update task priorities accordingly.
	//
	// Note: The priority updating of the tasks in the list is assumed to
	// be handled outside this function, and should be addressed post-add
	// operation.
	Add(task *Taskable, index int) error

	// Remove removes a task from the list by its index and returns the
	// returned task for buffering.
	//
	// Important Considerations:
	//
	//  1. Buffering. This function returns the removed task, which should
	//     be buffered for potential future reinsertion.
	//  2. Update Priority: The removal of a task can affect the
	//     priorities of the remaining tasks. Calling [Sort] post-removal
	//     should be done to update task priorities accordingly.
	//
	// Note: The buffering of the removed task and priority updating is
	// assumed to be handled outside this function, and they should be
	// addressed post-removal operation.
	Remove(index int) (*Taskable, error)
}

// A Task is the representation of a basic task.
type Task struct {
	id          int       // unique identifier
	name        string    // task name
	description string    // task description
	started     time.Time // date task was created
	finished    time.Time // date task was finished
	priority    int       // Determines task urgency. Lower numbers indicate higher priority.
	done        bool      // used to signify when a task it done
}

func (t Task) ID() int { return t.id }

func (t *Task) SetID(id int) { t.id = id }

func (t Task) Name() string { return t.name }

func (t *Task) SetName(s string) { t.name = s }

func (t Task) Description() string { return t.description }

func (t *Task) SetDescription(d string) { t.description = d }

func (t Task) Finished() time.Time { return t.finished }

func (t *Task) SetFinished(d time.Time) { t.finished = d }

func (t Task) Started() time.Time { return t.started }

func (t *Task) SetStarted(d time.Time) { t.started = d }

func (t Task) Priority() int { return t.priority }

func (t *Task) SetPriority(p int) { t.priority = p }

func (t Task) IsDone() bool { return t.done }

func (t *Task) SetDone(b bool) { t.done = b }
