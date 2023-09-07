package ui

import "fmt"

func ExampleWordWrap() {
	s := "The quick brown fox jumps over the lazy dog."
	lines := WordWrap(s, 15)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// The quick brown
	// fox jumps over
	// the lazy dog.
}

func ExampleWordWrap_zeroWidth() {
	s := "The quick brown fox jumps over the lazy dog."
	lines := WordWrap(s, 0)
	for _, line := range lines {
		fmt.Println(line)
	}

	// Output:
	// The
	// quick
	// brown
	// fox
	// jumps
	// over
	// the
	// lazy
	// dog.
}
