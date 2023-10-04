package tasks

import "fmt"

func ExampleGetTitle_Board() {
	b := Board{Title: "Build a project manager"}
	fmt.Printf("Board title: %q\n", b.GetTitle())

	// Output:
	// Board title: "Build a project manager"
}

func ExampleSetTitle_Board() {
	b := Board{}
	b.SetTitle("Build a project manager")
	fmt.Printf("Board title: %q\n", b.Title)

	// Output:
	// Board title: "Build a project manager"
}

func ExampleGetParentTask() {
	b := Board{ParentTask: &BoardTask{
		Task: &Task{Id: 1},
	},
	}
	fmt.Printf("Parent Task ID: %d\n", b.GetParentTask().Id)

	// Output:
	// Parent Task ID: 1
}

func ExampleSetParentTask() {
	b := Board{}
	pt := BoardTask{
		Task: &Task{Id: 1},
	}
	b.SetParentTask(&pt)
	fmt.Printf("Parent Task ID: %d\n", b.ParentTask.Id)

	// Output:
	// Parent Task ID: 1
}

func ExampleGetColumns() {
	bc1 := BoardColumn{Title: "TODO"}
	bc2 := BoardColumn{Title: "Working On"}
	bc3 := BoardColumn{Title: "Done"}

	b := Board{Columns: []BoardColumn{bc1, bc2, bc3}}

	for i, column := range b.GetColumns() {
		fmt.Printf("Column %d Title: %q\n", i, column.Title)
	}

	// Output:
	// Column 0 Title: "TODO"
	// Column 1 Title: "Working On"
	// Column 2 Title: "Done"
}

func ExampleAddColumn() {
	bc1 := BoardColumn{Title: "TODO"}
	bc2 := BoardColumn{Title: "Working On"}

	b := Board{Columns: []BoardColumn{bc1, bc2}}

	bc3 := BoardColumn{Title: "Done"}
	b.AddColumn(bc3)

	for i, column := range b.GetColumns() {
		fmt.Printf("Column %d Title: %q\n", i, column.Title)
	}

	// Output:
	// Column 0 Title: "TODO"
	// Column 1 Title: "Working On"
	// Column 2 Title: "Done"
}

func ExampleInsertColumn() {
	bc1 := BoardColumn{Title: "TODO"}
	bc2 := BoardColumn{Title: "Working On"}
	bc3 := BoardColumn{Title: "Done"}

	b := Board{Columns: []BoardColumn{bc1, bc2, bc3}}

	bc4 := BoardColumn{Title: "Testing"}
	b.InsertColumn(bc4, 2)

	for i, column := range b.GetColumns() {
		fmt.Printf("Column %d Title: %q\n", i, column.Title)
	}

	// Output:
	// Column 0 Title: "TODO"
	// Column 1 Title: "Working On"
	// Column 2 Title: "Testing"
	// Column 3 Title: "Done"
}

func ExampleGetTask() {
	bt1 := BoardTask{Task: &Task{Id: 1}}
	bt2 := BoardTask{Task: &Task{Id: 2}}
	bc := BoardColumn{Title: "TODO", Tasks: []BoardTask{bt1, bt2}}
	task, err := bc.GetTask(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Column Task ID:", task.Id)

	// Output:
	// Column Task ID: 2
}

func ExampleGetTasks() {
	bt1 := BoardTask{Task: &Task{Id: 1}}
	bt2 := BoardTask{Task: &Task{Id: 2}}
	bc := BoardColumn{Title: "TODO", Tasks: []BoardTask{bt1, bt2}}

	for _, bt := range bc.GetTasks() {
		fmt.Println("Column Task ID:", bt.Id)
	}

	// Output:
	// Column Task ID: 1
	// Column Task ID: 2
}

func ExampleUpdatePriorities_BoardColumn() {
	t1 := &Task{Name: "code", Priority: 0}
	t2 := &Task{Name: "read", Priority: 1}
	t3 := &Task{Name: "eat", Priority: 2}

	task1 := BoardTask{Task: t1}
	task2 := BoardTask{Task: t2}
	task3 := BoardTask{Task: t3}

	// I now want to eat before I read and code.
	bc := BoardColumn{Tasks: []BoardTask{task3, task1, task2}}

	// Update the tasks priority
	bc.UpdatePriorities(0)

	for _, t := range bc.Tasks {
		fmt.Printf("Task: %6q  Priority: %d\n", t.Name, t.Priority)
	}

	// Output:
	// Task:  "eat"  Priority: 0
	// Task: "code"  Priority: 1
	// Task: "read"  Priority: 2
}

func ExampleAdd_BoardColumn() {
	t1 := &Task{Name: "code", Priority: 0}
	t2 := &Task{Name: "read", Priority: 1}
	t3 := &Task{Name: "eat", Priority: 2}

	task1 := BoardTask{Task: t1}
	task2 := BoardTask{Task: t2}
	task3 := BoardTask{Task: t3}

	bc := BoardColumn{Tasks: []BoardTask{task1, task2, task3}}

	task4 := BoardTask{Task: &Task{
		Name:     "buy groceries",
		Priority: 4},
	}
	bc.Add(&task4)

	for _, t := range bc.Tasks {
		fmt.Printf("Task: %q  Priority: %d\n", t.Name, t.Priority)
	}

	// Output:
	// Task: "code"  Priority: 0
	// Task: "read"  Priority: 1
	// Task: "eat"  Priority: 2
	// Task: "buy groceries"  Priority: 4
}

func ExampleInsertTask() {
	t1 := &Task{Name: "code", Priority: 0}
	t2 := &Task{Name: "read", Priority: 1}
	t3 := &Task{Name: "eat", Priority: 2}

	task1 := BoardTask{Task: t1}
	task2 := BoardTask{Task: t2}
	task3 := BoardTask{Task: t3}

	bc := BoardColumn{Tasks: []BoardTask{task1, task2, task3}}

	task4 := BoardTask{Task: &Task{
		Name:     "buy groceries",
		Priority: 0}}
	bc.InsertTask(&task4, 0)

	for _, t := range bc.Tasks {
		fmt.Printf("Task: %q  Priority: %d\n", t.Name, t.Priority)
	}

	// Output:
	// Task: "buy groceries"  Priority: 0
	// Task: "code"  Priority: 0
	// Task: "read"  Priority: 1
	// Task: "eat"  Priority: 2
}

func ExampleInsertTask_noTasks() {
	bc := BoardColumn{Tasks: []BoardTask{}}

	task4 := BoardTask{Task: &Task{
		Name:     "buy groceries",
		Priority: 0}}
	bc.InsertTask(&task4, 1)

	for _, t := range bc.Tasks {
		fmt.Printf("Task: %q  Priority: %d\n", t.Name, t.Priority)
	}

	// Output:
	// Task: "buy groceries"  Priority: 0
}

func ExampleBounds_BoardColumn() {
	task1 := BoardTask{Task: &Task{Name: "code"}}
	bc := BoardColumn{Tasks: []BoardTask{task1}}
	fmt.Println(bc.Bounds(1))

	// Output:
	// index 1 out of range.
}

func ExampleSetTask() {
	bt := BoardTask{}
	bt.SetTask(&Task{Name: "code"})
	fmt.Printf("Task Name: %q\n", bt.Name)

	// Output:
	// Task Name: "code"
}

func ExampleGetChild() {
	c := Board{ID: 1}
	bt := BoardTask{ChildID: c.ID}
	fmt.Println("Board Task Child ID:", bt.GetChildID())

	// Output:
	// Board Task Child ID: 1
}

func ExampleSetChild() {
	bt := BoardTask{}
	c := Board{ID: 1}

	bt.SetChildID(c.ID)
	fmt.Println("Board Task Child ID:", bt.ChildID)

	// Output:
	// Board Task Child ID: 1
}
