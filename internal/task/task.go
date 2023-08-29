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
	ID() string
	Name() string
	SetName(t string)
	Description() string
	SetDescription(d string)
	IsComplete() bool
	SetComplete(c bool)
	Priority() int
	SetPriority(p int)
}

// A Task is the representation of a basic task.
type Task struct {
	ID           int       // unique identifier
	Name         string    // task name
	Description  string    // task description
	CreationDate time.Time // date task was created
	Completion   time.Time // date task was completed
	Priority     int       // Determines task urgency. Lower numbers indicate higher priority.
	Complete     bool      // used to signify when a task has been completed
}
