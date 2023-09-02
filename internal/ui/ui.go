// Package ui provides a set of interfaces and implementations for
// creating different types of user interfaces. This package aims to be
// a unified layer for interfacing with various UI types like TUI,
// WebUI, and GUI, while abstracting away the specific details ot each
// UI.
// This package defines a UI interface that all UI implementations
// should conform to. This allows the core application logic to interact
// with the UI in a consitent and decoupled manner.
package ui

import "github.com/iuiq/do/internal/tasks"

// A UI represents the user interface.
type UI interface {
	// Init sets up the UI.
	Init(tl *tasks.TaskList)
	// Populate takes the the tasks from the task list and populates the
	// list that will be displayed to the user.
	Populate(tl *tasks.TaskList)
}
