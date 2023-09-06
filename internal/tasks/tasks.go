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
	GetID() int
	// SetID sets the unique identifier of the task.
	SetID(id int)
	// Name returns the name of the task.
	GetName() string
	// SetName sets the name of the task.
	SetName(s string)
	// Description returns the description of the task.
	GetDescription() string
	// SetDescription sets the description of the task.
	SetDescription(d string)
	// DoneDate returns the date when the task was last finished.
	GetFinished() time.Time
	// SetFinished sets the date when the task was last finished.
	SetFinished(d time.Time)
	// Started returns the date when the task was last started.
	GetStarted() time.Time
	// SetStarted sets the date when the task was last started.
	SetStarted(d time.Time)
	// Priority returns the priority of the task.
	GetPriority() int
	// SetPriority sets the priority of the task.
	SetPriority(p int)
	// IsDone returns whether or not the task is finished.
	GetIsDone() bool
	// SetDone sets the task "finished" status to true.
	SetDone(b bool)
}

// A Tasktlist represents a list of tasks.
type TaskList interface {
	// Title returns the title of the task list.
	GetTitle() string
	// SetTitle sets the title of the task list.
	SetTitle(s string)
	// Tasks returns the list of tasks.
	GetTasks() []Taskable
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
	//     priorities of the remaining tasks. Calling [UpdatePriorities]
	//     post-removal should be done to update task priorities
	//     accordingly.
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
	Id          int       `yaml:"id"`          // unique identifier
	Name        string    `yaml:"name"`        // task name
	Description string    `yaml:"description"` // task description
	ShowDesc    bool      `yaml:"showDesc"`    // indicate whether or not to show task description
	Started     time.Time `yaml:"started"`     // date task was created
	Finished    time.Time `yaml:"finished"`    // date task was finished
	Priority    int       `yaml:"priority"`    // Determines task urgency. Lower numbers indicate higher priority.
	Done        bool      `yaml:"done"`        // used to signify when a task it done
}

func (t Task) GetID() int { return t.Id }

func (t *Task) SetID(id int) { t.Id = id }

func (t Task) GetName() string { return t.Name }

func (t *Task) SetName(s string) { t.Name = s }

func (t Task) GetDescription() string { return t.Description }

func (t *Task) SetDescription(d string) { t.Description = d }

func (t Task) GetFinished() time.Time { return t.Finished }

func (t *Task) SetFinished(d time.Time) { t.Finished = d }

func (t Task) GetStarted() time.Time { return t.Started }

func (t *Task) SetStarted(d time.Time) { t.Started = d }

func (t Task) GetPriority() int { return t.Priority }

func (t *Task) SetPriority(p int) { t.Priority = p }

func (t Task) GetIsDone() bool { return t.Done }

func (t *Task) SetDone(b bool) { t.Done = b }
