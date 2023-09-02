package tasks

import (
	"fmt"
	"time"
)

func ExampleID() {
	var t Task
	t = Task{id: 1}
	fmt.Println(t.ID())

	// Output:
	// 1
}

func ExampleSetID() {
	var t Task
	t = Task{id: 1}
	t.SetID(2)
	fmt.Println(t.id)

	// Output:
	// 2
}

func ExampleName() {
	var t Task
	t = Task{name: "Buy groceries"}
	fmt.Printf("Task name: %q\n", t.Name())

	// Output:
	// Task name: "Buy groceries"
}

func ExampleSetName() {
	var t Task
	t.SetName("Clean room")
	fmt.Printf("Task name: %q\n", t.name)

	// Output:
	// Task name: "Clean room"
}

func ExampleDescription() {
	var t Task
	t = Task{description: "Buying groceries entails taking inventory of what you have and what you need, and then taking the trip to the store."}
	fmt.Printf("Task description: %q\n", t.Description())

	// Output:
	// Task description: "Buying groceries entails taking inventory of what you have and what you need, and then taking the trip to the store."
}

func ExampleSetDescription() {
	var t Task
	t.SetDescription("Cleaning room consists of cleaing your desk and taking out the trash.")
	fmt.Printf("Task description: %q\n", t.description)

	// Output:
	// Task description: "Cleaning room consists of cleaing your desk and taking out the trash."
}

func ExampleFinished() {
	var t Task
	t = Task{finished: time.Date(2023, 8, 1, 12, 30, 0, 0, time.UTC)}
	d := t.finished.Format("2006-01-02")
	fmt.Printf("Date task last finished: %q\n", d)

	// Output:
	// Date task last finished: "2023-08-01"
}

func ExampleSetFinished() {
	var t Task
	t.SetFinished(time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC))
	d := t.finished.Format("2006-01-02")
	fmt.Printf("Date task last finished: %q\n", d)

	// Output:
	// Date task last finished: "2023-08-01"
}

func ExampleStarted() {
	var t Task
	t = Task{started: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)}
	d := t.Started().Format("2006-01-02")
	fmt.Printf("Date task last started: %q\n", d)

	// Output:
	// Date task last started: "2023-08-01"
}

func ExampleSetStarted() {
	var t Task
	t.SetStarted(time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC))
	d := t.started.Format("2006-01-02")
	fmt.Printf("Date task last started: %q\n", d)

	// Output:
	// Date task last started: "2023-08-01"
}

func ExamplePriority() {
	var t Task
	t = Task{priority: 1}
	fmt.Printf("Task priority: %d\n", t.priority)

	// Output:
	// Task priority: 1
}

func ExampleSetPriority() {
	var t Task
	t.SetPriority(1)
	fmt.Printf("Task priority: %d\n", t.priority)

	// Output:
	// Task priority: 1
}

func ExampleIsDone() {
	var t Task
	t = Task{done: true}
	fmt.Printf("Task done status: %v\n", t.IsDone())

	// Output:
	// Task done status: true
}

func ExampleSetDone() {
	var t Task
	t.SetDone(true)
	fmt.Printf("Task done status: %v\n", t.IsDone())

	// Output:
	// Task done status: true
}
