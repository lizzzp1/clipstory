package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/flock"
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
	default:
		fmt.Println("Unknown command. Use: add, list, or get")
	}
}

func getHistoryPath() string {
	var clipDir string

	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "clipstory")
	}

	home, _ := os.UserHomeDir()
	clipDir = filepath.Join(home, ".local", "share", "clipstory")

	os.MkdirAll(clipDir, os.FileMode(0700))

	return filepath.Join(clipDir, "history.json")
}

func getLockPath() string {
	return getHistoryPath() + ".lock"
}

func loadHistory() (*History, error) {
	lock := flock.New(getLockPath())
	err := lock.Lock()
	if err != nil {
		return nil, err
	}
	defer lock.Unlock()

	historyPath := getHistoryPath()
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return &History{Entries: []Entry{}}, nil
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, err
	}

	var history History

	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	return &history, nil
}

func saveHistory(history *History) error {
	lock := flock.New(getLockPath())
	err := lock.Lock()

	if err != nil {
		return err
	}

	defer lock.Unlock()

	// only last 100, todo -- respect whatever history settings are already present in bash
	if len(history.Entries) > 100 {
		history.Entries = history.Entries[len(history.Entries)-100:]
	}

	data, err := json.MarshalIndent(history, "", "")
	if err != nil {
		return err
	}
	tmpPath := getHistoryPath() + ".tmp"

	if err := os.WriteFile(tmpPath, data, os.FileMode(0600)); err != nil {
		return err
	}
	return os.Rename(tmpPath, getHistoryPath())
}

func addEntry(content string) {
	history, err := loadHistory()

	if err != nil {
		return
	}

	if len(history.Entries) > 0 && history.Entries[len(history.Entries)-1].Content == content {
		fmt.Println("Entry already exists as most recent -- SKIPPING")
		return
	}

	entry := Entry{
		Content:   content,
		Timestamp: time.Now(),
	}

	history.Entries = append(history.Entries, entry)

	if err := saveHistory(history); err != nil {
		fmt.Printf("Error saving history: %v\n", err)
		return
	}

	fmt.Printf("Added entry #%d\n", len(history.Entries))
}

func listEntries() {
	history, err := loadHistory()
	if err != nil {
		fmt.Printf("Error loading history: %v\n", err)
		return
	}

	if len(history.Entries) == 0 {
		fmt.Println("No clipboard history")
		return
	}

	start := 0
	if len(history.Entries) > 10 {
		start = len(history.Entries) - 10
	}

	for i := start; i < len(history.Entries); i++ {
		entry := history.Entries[i]
		content := entry.Content
		if len(content) > 60 {
			content = content[:57] + "..."
		}
		fmt.Printf("%d: %s (%s)\n", i+1, content, entry.Timestamp.Format("15:04:05"))
	}
}
