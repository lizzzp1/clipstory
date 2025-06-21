package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Content   string
	Timestamp time.Time
}

type History struct {
	Entries []Entry `json:"entries"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: clipstory {add|list}")
		return
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: clipstory add <text>")
			return
		}
		addEntry(os.Args[2])
	case "list":
		listEntries()
	}
}

func getHistoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".clipstory", "history.json")
}

func addEntry(content string) {
	fmt.Printf("Added: %s\n", content)
}

func listEntries() {
	fmt.Println("History:")
}
