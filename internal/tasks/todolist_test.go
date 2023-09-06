package tasks

import "fmt"

func ExampleBounds() {
	t1 := &Task{Name: "Task 1"}
	task1 := TodoTask{Task: t1}
	t2 := &Task{Name: "Task 2"}
	task2 := TodoTask{Task: t2}
	t3 := &Task{Name: "Task 3"}
	task3 := TodoTask{Task: t3}
	list := TodoList{Tasks: []TodoTask{task1, task2, task3}}
	fmt.Println(list.Bounds(2))

	// Output:
	// <nil>
}

func ExampleBounds_outOfBounds() {
	t1 := &Task{Name: "Task 1"}
	task1 := TodoTask{Task: t1}
	t2 := &Task{Name: "Task 2"}
	task2 := TodoTask{Task: t2}
	t3 := &Task{Name: "Task 3"}
	task3 := TodoTask{Task: t3}
	list := TodoList{Tasks: []TodoTask{task1, task2, task3}}
	fmt.Println(list.Bounds(3))

	// Output:
	// index 3 out of range.
}

func ExampleUpdatePriorities() {
	t1 := &Task{Name: "code", Priority: 0}
	t2 := &Task{Name: "read", Priority: 1}
	t3 := &Task{Name: "eat", Priority: 2}

	task1 := TodoTask{Task: t1}
	task2 := TodoTask{Task: t2}
	task3 := TodoTask{Task: t3}

	// I now want to eat before I read and code.
	list := TodoList{Tasks: []TodoTask{task3, task1, task2}}

	// Update the tasks priority
	list.UpdatePriorities(0)

	for _, t := range list.Tasks {
		fmt.Printf("Task: %6q  Priority: %d\n", t.Name, t.Priority)
	}

	// Output:
	// Task:  "eat"  Priority: 0
	// Task: "code"  Priority: 1
	// Task: "read"  Priority: 2
}

func ExampleAdd() {
	t1 := &Task{Name: "code"}
	task1 := TodoTask{Task: t1}
	t2 := &Task{Name: "read"}
	task2 := TodoTask{Task: t2}
	list := TodoList{Tasks: []TodoTask{task1, task2}}

	t3 := &Task{Name: "eat"}
	task3 := TodoTask{Task: t3}

	if err := list.Add(&task3, 2); err != nil {
		fmt.Println(err)
	}

	for _, t := range list.Tasks {
		fmt.Printf("Task name: %q\n", t.Name)
	}

	// Output:
	// Task name: "code"
	// Task name: "read"
	// Task name: "eat"
}

func ExampleAdd_withinBounds() {
	t1 := &Task{Name: "code"}
	task1 := TodoTask{Task: t1}
	t2 := &Task{Name: "read"}
	task2 := TodoTask{Task: t2}
	list := TodoList{Tasks: []TodoTask{task1, task2}}

	t3 := &Task{Name: "eat"}
	task3 := TodoTask{Task: t3}

	if err := list.Add(&task3, 0); err != nil {
		fmt.Println(err)
	}

	for _, t := range list.Tasks {
		fmt.Printf("Task name: %q\n", t.Name)
	}

	// Output:
	// Task name: "eat"
	// Task name: "code"
	// Task name: "read"
}

func ExampleRemove() {
	t1 := &Task{Name: "code"}
	task1 := TodoTask{Task: t1}
	t2 := &Task{Name: "read"}
	task2 := TodoTask{Task: t2}
	t3 := &Task{Name: "eat"}
	task3 := TodoTask{Task: t3}
	list := TodoList{Tasks: []TodoTask{task1, task2, task3}}

	// Remove task 2 from the slice.
	cpy, err := list.Remove(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Removed task: %q\n", cpy.Name)
	fmt.Println("Remaining tasks:")

	for _, t := range list.Tasks {
		fmt.Printf("- %q\n", t.Name)
	}

	// Output:
	// Removed task: "read"
	// Remaining tasks:
	// - "code"
	// - "eat"
}
