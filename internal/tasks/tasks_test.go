package tasks

import (
	"fmt"
	"time"
)

func ExampleGetID() {
	var t Task
	t = Task{Id: 1}
	fmt.Println(t.GetID())

	// Output:
	// 1
}

func ExampleSetID() {
	var t Task
	t = Task{Id: 1}
	t.SetID(2)
	fmt.Println(t.Id)

	// Output:
	// 2
}

func ExampleGetName() {
	var t Task
	t = Task{Name: "Buy groceries"}
	fmt.Printf("Task name: %q\n", t.GetName())

	// Output:
	// Task name: "Buy groceries"
}

func ExampleSetName() {
	var t Task
	t.SetName("Clean room")
	fmt.Printf("Task name: %q\n", t.Name)

	// Output:
	// Task name: "Clean room"
}

func ExampleGetDesc() {
	var t Task
	t = Task{Description: "Buying groceries entails taking inventory of what you have and what you need, and then taking the trip to the store."}
	fmt.Printf("Task description: %q\n", t.GetDesc())

	// Output:
	// Task description: "Buying groceries entails taking inventory of what you have and what you need, and then taking the trip to the store."
}

func ExampleSetDescription() {
	var t Task
	t.SetDesc("Cleaning room consists of cleaing your desk and taking out the trash.")
	fmt.Printf("Task description: %q\n", t.Description)

	// Output:
	// Task description: "Cleaning room consists of cleaing your desk and taking out the trash."
}

func ExampleGetFinished() {
	var t Task
	t = Task{Finished: time.Date(2023, 8, 1, 12, 30, 0, 0, time.UTC)}
	d := t.GetFinished().Format("2006-01-02")
	fmt.Printf("Date task last finished: %q\n", d)

	// Output:
	// Date task last finished: "2023-08-01"
}

func ExampleSetFinished() {
	var t Task
	t.SetFinished(time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC))
	d := t.Finished.Format("2006-01-02")
	fmt.Printf("Date task last finished: %q\n", d)

	// Output:
	// Date task last finished: "2023-08-01"
}

func ExampleGetStarted() {
	var t Task
	t = Task{Started: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)}
	d := t.GetStarted().Format("2006-01-02")
	fmt.Printf("Date task last started: %q\n", d)

	// Output:
	// Date task last started: "2023-08-01"
}

func ExampleSetStarted() {
	var t Task
	t.SetStarted(time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC))
	d := t.Started.Format("2006-01-02")
	fmt.Printf("Date task last started: %q\n", d)

	// Output:
	// Date task last started: "2023-08-01"
}

func ExampleGetPriority() {
	var t Task
	t = Task{Priority: 1}
	fmt.Printf("Task priority: %d\n", t.Priority)

	// Output:
	// Task priority: 1
}

func ExampleSetPriority() {
	var t Task
	t.SetPriority(1)
	fmt.Printf("Task priority: %d\n", t.Priority)

	// Output:
	// Task priority: 1
}

func ExampleGetIsDone() {
	var t Task
	t = Task{Done: true}
	fmt.Printf("Task done status: %v\n", t.GetIsDone())

	// Output:
	// Task done status: true
}

func ExampleSetDone() {
	var t Task
	t.SetDone(true)
	fmt.Printf("Task done status: %v\n", t.GetIsDone())

	// Output:
	// Task done status: true
}
