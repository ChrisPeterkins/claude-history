package data

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// historyProject groups history entries by project path.
type historyProject struct {
	path     string
	name     string
	sessions map[string][]HistoryEntry // sessionID -> entries
}

// LoadHistory reads history.jsonl and returns projects that have no session files.
// These are merged into the project list so older conversations are browsable.
func LoadHistory() ([]Project, error) {
	historyPath := filepath.Join(claudeDir, "history.jsonl")
	f, err := os.Open(historyPath)
	if err != nil {
		return nil, nil // no history file is fine
	}
	defer f.Close()

	// Collect all projects that have session directories
	projectsWithSessions := make(map[string]bool)
	projectsDir := filepath.Join(claudeDir, "projects")
	if entries, err := os.ReadDir(projectsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				// Read the real path from session files
				realPath := readProjectPath(filepath.Join(projectsDir, entry.Name()))
				if realPath != "" {
					projectsWithSessions[realPath] = true
				}
			}
		}
	}

	// Parse history.jsonl and group by project
	projectMap := make(map[string]*historyProject)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var entry HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		// Skip entries that are just slash commands
		display := strings.TrimSpace(entry.Display)
		if display == "" || strings.HasPrefix(display, "/") {
			continue
		}

		// Skip projects that already have full session files
		if projectsWithSessions[entry.Project] {
			continue
		}

		if entry.Project == "" {
			continue
		}

		hp, ok := projectMap[entry.Project]
		if !ok {
			hp = &historyProject{
				path:     entry.Project,
				name:     projectNameFromPath(entry.Project),
				sessions: make(map[string][]HistoryEntry),
			}
			projectMap[entry.Project] = hp
		}

		// Use sessionId if available, otherwise group by date
		groupKey := entry.SessionID
		if groupKey == "" {
			ts := time.UnixMilli(entry.Timestamp)
			groupKey = "day:" + ts.Format("2006-01-02")
		}

		hp.sessions[groupKey] = append(hp.sessions[groupKey], entry)
	}

	// Convert to Project/Session types
	var projects []Project
	for _, hp := range projectMap {
		p := Project{
			Name:        hp.name,
			Path:        hp.path,
			HistoryOnly: true,
		}

		var sessions []Session
		for sessionID, entries := range hp.sessions {
			if len(entries) == 0 {
				continue
			}

			// Sort entries by timestamp
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Timestamp < entries[j].Timestamp
			})

			firstEntry := entries[0]
			preview := strings.TrimSpace(firstEntry.Display)
			if len(preview) > 120 {
				preview = preview[:117] + "..."
			}

			sessions = append(sessions, Session{
				ID:             sessionID,
				Project:        &p,
				StartedAt:      time.UnixMilli(firstEntry.Timestamp),
				Preview:        preview,
				HistoryOnly:    true,
				HistoryEntries: entries,
				MessageCount:   len(entries),
			})
		}

		// Sort sessions by most recent first
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].StartedAt.After(sessions[j].StartedAt)
		})

		p.Sessions = sessions
		p.SessionCount = len(sessions)
		projects = append(projects, p)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// LoadHistoryMessages converts history entries into Messages for display.
func LoadHistoryMessages(session *Session) ([]Message, error) {
	if !session.HistoryOnly || len(session.HistoryEntries) == 0 {
		return nil, nil
	}

	var messages []Message
	for _, entry := range session.HistoryEntries {
		display := strings.TrimSpace(entry.Display)
		if display == "" || strings.HasPrefix(display, "/") {
			continue
		}

		messages = append(messages, Message{
			Type:      "user",
			Role:      "user",
			RawText:   display,
			Timestamp: time.UnixMilli(entry.Timestamp),
			SessionID: entry.SessionID,
			ContentBlocks: []ContentBlock{
				{Type: "text", Text: display},
			},
		})
	}

	return messages, nil
}
