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

// A Task is the representation of a basic task.
type Task struct {
	Id          int       `yaml:"id"`          // unique identifier
	Name        string    `yaml:"name"`        // task name
	Description string    `yaml:"description"` // task description
	ShowDesc    bool      `yaml:"showDesc"`    // indicate whether or not to show task description
	Started     time.Time `yaml:"started"`     // date task was created
	Finished    time.Time `yaml:"finished"`    // date task was finished
	Priority    int       `yaml:"priority"`    // Determines task urgency. Lower numbers indicate higher priority.
	// TODO: move to TodoTask struct (?).
	Done bool `yaml:"done"` // used to signify when a task is done
}

// ID returns the unique identifier of the task.
func (t Task) GetID() int { return t.Id }

// SetID sets the unique identifier of the task.
func (t *Task) SetID(id int) { t.Id = id }

// GetName returns the name of the task.
func (t Task) GetName() string { return t.Name }

// SetName sets the name of the task.
func (t *Task) SetName(s string) { t.Name = s }

// GetDesc returns the description of the task.
func (t Task) GetDesc() string { return t.Description }

// SetDesc sets the description of the task.
func (t *Task) SetDesc(d string) { t.Description = d }

// GetShowDesc returns the show description status for the task.
func (t *Task) GetShowDesc() bool { return t.ShowDesc }

// SetShowDesc sets the show description status for the task.
func (t *Task) SetShowDesc(b bool) { t.ShowDesc = b }

// GetFinished returns the date when the task was last finished.
func (t Task) GetFinished() time.Time { return t.Finished }

// SetFinished sets the date when the task was last finished.
func (t *Task) SetFinished(d time.Time) { t.Finished = d }

// GetStarted returns the date when the task was last started.
func (t Task) GetStarted() time.Time { return t.Started }

// SetStarted sets the date when the task was last started.
func (t *Task) SetStarted(d time.Time) { t.Started = d }

// GetPriority returns the priority of the task.
func (t Task) GetPriority() int { return t.Priority }

// SetPriority sets the priority of the task.
func (t *Task) SetPriority(p int) { t.Priority = p }

// GetIsDone returns whether or not the task is finished.
func (t Task) GetIsDone() bool { return t.Done }

// SetDone sets the task "finished" status to true.
func (t *Task) SetDone(b bool) { t.Done = b }
