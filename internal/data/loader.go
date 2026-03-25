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

var claudeDir string

func init() {
	home, _ := os.UserHomeDir()
	claudeDir = filepath.Join(home, ".claude")
}

// LoadProjects discovers all projects from ~/.claude/projects/.
func LoadProjects() ([]Project, error) {
	projectsDir := filepath.Join(claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()

		// Try to get real path from a session file's cwd field
		realPath := readProjectPath(filepath.Join(projectsDir, dirName))
		name := decodeProjectName(dirName)
		if realPath != "" {
			name = projectNameFromPath(realPath)
		} else {
			realPath = decodeDirToPath(dirName)
		}

		p := Project{
			Name:    name,
			Path:    realPath,
			DirName: dirName,
		}
		projects = append(projects, p)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// readProjectPath reads the cwd from the first user message in any session file.
func readProjectPath(projectDir string) string {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		f, err := os.Open(filepath.Join(projectDir, entry.Name()))
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			var raw struct {
				Type string `json:"type"`
				Cwd  string `json:"cwd"`
			}
			if err := json.Unmarshal(scanner.Bytes(), &raw); err == nil && raw.Cwd != "" {
				f.Close()
				return raw.Cwd
			}
		}
		f.Close()
	}
	return ""
}

// projectNameFromPath extracts a friendly project name from a full path.
func projectNameFromPath(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	for i, p := range parts {
		if p == "Projects" && i+1 < len(parts) {
			return strings.Join(parts[i+1:], " ")
		}
	}
	// Fallback to last segment
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// LoadSessions loads session metadata for a project (without full messages).
func LoadSessions(project *Project) ([]Session, error) {
	projectDir := filepath.Join(claudeDir, "projects", project.DirName)
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		filePath := filepath.Join(projectDir, entry.Name())
		info, _ := entry.Info()
		var fileSize int64
		if info != nil {
			fileSize = info.Size()
		}

		// Read just enough to get the first user message and timestamp
		preview, startedAt := peekSession(filePath)

		sessions = append(sessions, Session{
			ID:        sessionID,
			Project:   project,
			StartedAt: startedAt,
			Preview:   preview,
			FilePath:  filePath,
			FileSize:  fileSize,
		})
	}

	// Sort by most recent first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	return sessions, nil
}

// LoadMessages loads all messages from a session JSONL file.
func LoadMessages(session *Session) ([]Message, error) {
	f, err := os.Open(session.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var messages []Message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line

	for scanner.Scan() {
		msg := parseMessage(scanner.Bytes())
		if msg != nil && (msg.Type == "user" || msg.Type == "assistant") {
			messages = append(messages, *msg)
		}
	}

	return messages, scanner.Err()
}

// peekSession reads the first few lines of a session file to extract a preview.
func peekSession(path string) (preview string, startedAt time.Time) {
	f, err := os.Open(path)
	if err != nil {
		return "", time.Time{}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		var msgType string
		if t, ok := raw["type"]; ok {
			json.Unmarshal(t, &msgType)
		}

		if msgType == "user" {
			// Extract timestamp
			if ts, ok := raw["timestamp"]; ok {
				var tsStr string
				if err := json.Unmarshal(ts, &tsStr); err == nil {
					startedAt, _ = time.Parse(time.RFC3339Nano, tsStr)
				}
			}

			// Extract user message text
			if msgData, ok := raw["message"]; ok {
				var msg struct {
					Content json.RawMessage `json:"content"`
				}
				if err := json.Unmarshal(msgData, &msg); err == nil {
					// Content can be a string or array of blocks
					var contentStr string
					if err := json.Unmarshal(msg.Content, &contentStr); err == nil {
						preview = truncate(contentStr, 120)
					} else {
						// Try array of content blocks
						var blocks []struct {
							Type string `json:"type"`
							Text string `json:"text"`
						}
						if err := json.Unmarshal(msg.Content, &blocks); err == nil {
							for _, b := range blocks {
								if b.Type == "text" && b.Text != "" {
									preview = truncate(b.Text, 120)
									break
								}
							}
						}
					}
				}
			}
			return
		}
	}
	return "", time.Time{}
}

func parseMessage(line []byte) *Message {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil
	}

	var msgType string
	if t, ok := raw["type"]; ok {
		json.Unmarshal(t, &msgType)
	}

	if msgType != "user" && msgType != "assistant" {
		return nil
	}

	msg := &Message{Type: msgType}

	if u, ok := raw["uuid"]; ok {
		json.Unmarshal(u, &msg.UUID)
	}
	if p, ok := raw["parentUuid"]; ok {
		json.Unmarshal(p, &msg.ParentUUID)
	}
	if ts, ok := raw["timestamp"]; ok {
		var tsStr string
		if err := json.Unmarshal(ts, &tsStr); err == nil {
			msg.Timestamp, _ = time.Parse(time.RFC3339Nano, tsStr)
		}
	}

	if msgData, ok := raw["message"]; ok {
		var inner struct {
			Role    string          `json:"role"`
			Model   string          `json:"model"`
			Content json.RawMessage `json:"content"`
			Usage   *struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(msgData, &inner); err == nil {
			msg.Role = inner.Role
			msg.Model = inner.Model
			if inner.Usage != nil {
				msg.TokensIn = inner.Usage.InputTokens
				msg.TokensOut = inner.Usage.OutputTokens
			}

			// Parse content - can be string or array of blocks
			var contentStr string
			if err := json.Unmarshal(inner.Content, &contentStr); err == nil {
				msg.RawText = contentStr
			} else {
				var blocks []struct {
					Type     string `json:"type"`
					Text     string `json:"text"`
					Thinking string `json:"thinking"`
					ID       string `json:"id"`
					Name     string `json:"name"`
					Input    json.RawMessage `json:"input"`
				}
				if err := json.Unmarshal(inner.Content, &blocks); err == nil {
					var textParts []string
					for _, b := range blocks {
						switch b.Type {
						case "text":
							textParts = append(textParts, b.Text)
						case "tool_use":
							msg.ToolUses = append(msg.ToolUses, ToolUse{
								Name:  b.Name,
								Input: string(b.Input),
							})
						}
					}
					msg.RawText = strings.Join(textParts, "\n")
				}
			}
		}
	}

	return msg
}

// decodeProjectName extracts a friendly name from the encoded directory name.
// Uses the path segments after "Projects" if available, otherwise last 2 segments.
func decodeProjectName(dirName string) string {
	fullPath := decodeDirToPath(dirName)
	parts := strings.Split(fullPath, "/")

	// Find "Projects" directory and take everything after it
	for i, p := range parts {
		if p == "Projects" && i+1 < len(parts) {
			remaining := parts[i+1:]
			return strings.Join(remaining, " ")
		}
	}

	// Fallback: last non-empty segment
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return dirName
}

// decodeDirToPath converts the encoded dir name back to the original path.
func decodeDirToPath(dirName string) string {
	// The encoding replaces "/" with "-" and removes the leading slash
	// e.g., "-Users-chris-Desktop-Projects-MyProject" -> "/Users/chris/Desktop/Projects/MyProject"
	if strings.HasPrefix(dirName, "-") {
		return strings.ReplaceAll(dirName, "-", "/")
	}
	return dirName
}

func truncate(s string, maxLen int) string {
	// Remove newlines for preview
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
