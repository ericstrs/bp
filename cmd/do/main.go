package main

import (
	"log"
	"os"

	s "github.com/iuiq/do/internal/storage"
	t "github.com/iuiq/do/internal/tasks"
	"github.com/iuiq/do/internal/ui"
)

func run() {
	switch len(os.Args) {
	case 2:
		switch os.Args[1] {
		case "form":
			ui.InitForm()
			return
		}
	}

	yamlPath := os.Getenv("DO_DATA_PATH")
	if len(yamlPath) == 0 {
		log.Fatal("Environment variable DO_DATA_PATH must be set")
	}

	store := s.YAMLStorage{Filename: yamlPath}

	list := new(t.TodoList)
	list.SetTitle("Daily TODOs")
	if err := store.Load("list", &list); err != nil {
		log.Fatalf("Error loading list:", err)
	}
	tree := new(t.BoardTree)
	if err := store.Load("boards", &tree); err != nil {
		log.Fatalf("Error loading boards:", err)
	}

	tui := new(ui.TUI)
	tui.Init(list, tree)

	if err := store.Save("list", list); err != nil {
		log.Fatalf("Error saving list:", err)
	}
	if err := store.Save("boards", tree); err != nil {
		log.Fatalf("Error saving board:", err)
	}
}

func main() {
	run()
}
