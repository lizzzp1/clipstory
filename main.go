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
	Content    string
	WorkingDir string
	Timestamp  time.Time
}

type History struct {
	Entries []Entry
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: whatdidido {add|list}")
		return
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: whatdidido add <text>")
			return
		}
		addEntry(os.Args[2])
	case "list":
		listEntries()
	case "summary":
		flat := len(os.Args) > 2 && os.Args[2] == "--flat"
		summarizeToday(flat)
	default:
		fmt.Println("Unknown command. Use: list, or get")
	}
}

func getHistoryPath() string {
	var clipDir string

	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "whatdidido")
	}

	home, _ := os.UserHomeDir()
	clipDir = filepath.Join(home, ".local", "share", "whatdidido")

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
	dir, _ := os.Getwd()

	entry := Entry{
		Content:    content,
		WorkingDir: dir,
		Timestamp:  time.Now(),
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
		fmt.Println("No command history")
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
		fmt.Printf("\033[32m%-40s\033[0m \033[2m[%s]\033[0m \033[36m(%s)\033[0m\n",
			entry.Content,
			entry.Timestamp.Format("15:04"),
			entry.WorkingDir)
	}
}

func summarizeToday(flat bool) {
	history, err := loadHistory()
	if err != nil {
		fmt.Println("Error loading history")
		return
	}

	today := time.Now().Format("2006-01-02")
	var todayEntries []Entry

	for _, entry := range history.Entries {
		if entry.Timestamp.Format("2006-01-02") == today {
			todayEntries = append(todayEntries, entry)
		}
	}

	if flat {
		for _, entry := range todayEntries {
			fmt.Printf("%-50s [%s]\n", entry.Content, entry.Timestamp.Format("15:04:05"))
		}
		return
	}

	cmds := unique(todayEntries)

	if len(cmds) == 0 {
		fmt.Println("No activity logged today.")
		return
	}

	fmt.Println("Commands:")
	for _, e := range cmds {
		fmt.Println("  -", e.Content)
	}
}

func unique(entries []Entry) []Entry {
	seen := make(map[string]struct{})
	var out []Entry
	for _, entry := range entries {
		if _, ok := seen[entry.Content]; !ok {
			seen[entry.Content] = struct{}{}
			out = append(out, entry)
		}
	}
	return out
}
