package data

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chrispeterkins/claude-history/internal/config"
)

var claudeDir string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	claudeDir = filepath.Join(home, ".claude")
}

// newScanner creates a buffered scanner for reading JSONL files.
// Use large=true for session files that may have very long lines.
func newScanner(f *os.File, large bool) *bufio.Scanner {
	s := bufio.NewScanner(f)
	if large {
		s.Buffer(make([]byte, 0, scannerMaxBuf), scannerLargeBuf)
	} else {
		s.Buffer(make([]byte, 0, scannerInitBuf), scannerMaxBuf)
	}
	return s
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

		realPath := readProjectPath(filepath.Join(projectsDir, dirName))
		name := decodeProjectName(dirName)
		if realPath != "" {
			name = projectNameFromPath(realPath)
		} else {
			realPath = decodeDirToPath(dirName)
		}

		// Count session files and find most recent
		sessionCount := 0
		var lastActive time.Time
		subEntries, err := os.ReadDir(filepath.Join(projectsDir, dirName))
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if !se.IsDir() && strings.HasSuffix(se.Name(), ".jsonl") {
				sessionCount++
				if info, err := se.Info(); err == nil {
					if info.ModTime().After(lastActive) {
						lastActive = info.ModTime()
					}
				}
			}
		}

		p := Project{
			Name:         name,
			Path:         realPath,
			DirName:      dirName,
			SessionCount: sessionCount,
			LastActive:   lastActive,
		}
		projects = append(projects, p)
	}

	// Merge in history-only projects (older conversations without session files)
	historyProjects, _ := LoadHistory() // error is non-fatal; history-only projects are optional
	projects = append(projects, historyProjects...)

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

// readProjectPath reads the cwd from the first message in any session file.
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
		scanner := newScanner(f, false)
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
// Uses project root directories from config (or defaults) to find the
// boundary between the parent dir and the project name.
func projectNameFromPath(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	roots := config.ProjectRoots()
	for i, p := range parts {
		if roots[p] && i+1 < len(parts) {
			return strings.Join(parts[i+1:], " ")
		}
	}
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// LoadSessions loads session metadata for a project (without full messages).
func LoadSessions(project *Project) ([]Session, error) {
	// History-only projects already have sessions populated
	if project.HistoryOnly {
		return project.Sessions, nil
	}

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

		preview, startedAt, stats := peekSession(filePath)

		sessions = append(sessions, Session{
			ID:              sessionID,
			Project:         project,
			StartedAt:       startedAt,
			Preview:         preview,
			FilePath:        filePath,
			FileSize:        fileSize,
			MessageCount:    stats.messageCount,
			TotalTokensIn:   stats.tokensIn,
			TotalTokensOut:  stats.tokensOut,
			TotalDurationMs: stats.durationMs,
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	return sessions, nil
}

type sessionStats struct {
	messageCount int
	tokensIn     int
	tokensOut    int
	durationMs   int
}

// peekSession reads a session file to extract preview, timestamp, and stats.
func peekSession(path string) (string, time.Time, sessionStats) {
	var preview string
	var startedAt time.Time
	var stats sessionStats
	f, err := os.Open(path)
	if err != nil {
		return "", time.Time{}, stats
	}
	defer f.Close()

	scanner := newScanner(f, true)

	foundPreview := false

	for scanner.Scan() {
		var raw struct {
			Type    string `json:"type"`
			Subtype string `json:"subtype"`
			// Duration for system messages
			DurationMs int `json:"durationMs"`
			// Message content
			Message *struct {
				Content json.RawMessage `json:"content"`
				Usage   *struct {
					InputTokens              int `json:"input_tokens"`
					OutputTokens             int `json:"output_tokens"`
					CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
					CacheReadInputTokens     int `json:"cache_read_input_tokens"`
				} `json:"usage"`
			} `json:"message"`
			Timestamp json.RawMessage `json:"timestamp"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		// Count user and assistant messages
		if raw.Type == "user" || raw.Type == "assistant" {
			stats.messageCount++
		}

		// Accumulate tokens from assistant messages
		if raw.Message != nil && raw.Message.Usage != nil {
			stats.tokensIn += raw.Message.Usage.InputTokens +
				raw.Message.Usage.CacheCreationInputTokens +
				raw.Message.Usage.CacheReadInputTokens
			stats.tokensOut += raw.Message.Usage.OutputTokens
		}

		// Accumulate turn duration from system messages
		if raw.Type == "system" && raw.Subtype == "turn_duration" {
			stats.durationMs += raw.DurationMs
		}

		// Extract preview from first user message
		if !foundPreview && raw.Type == "user" {
			foundPreview = true

			// Parse timestamp
			if raw.Timestamp != nil {
				var tsStr string
				if err := json.Unmarshal(raw.Timestamp, &tsStr); err == nil {
					startedAt, _ = time.Parse(time.RFC3339Nano, tsStr)
				}
			}

			// Extract user message text
			if raw.Message != nil {
				var contentStr string
				if err := json.Unmarshal(raw.Message.Content, &contentStr); err == nil {
					preview = truncate(contentStr, maxPreviewLen)
				} else {
					var blocks []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}
					if err := json.Unmarshal(raw.Message.Content, &blocks); err == nil {
						for _, b := range blocks {
							if b.Type == "text" && b.Text != "" {
								preview = truncate(b.Text, maxPreviewLen)
								break
							}
						}
					}
				}
			}
		}
	}
	return preview, startedAt, stats
}

// LoadMessages loads all messages from a session JSONL file.
func LoadMessages(session *Session) ([]Message, error) {
	// History-only sessions have no JSONL file — build from history entries
	if session.HistoryOnly {
		return LoadHistoryMessages(session)
	}

	f, err := os.Open(session.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var messages []Message
	scanner := newScanner(f, true)

	for scanner.Scan() {
		msg := parseMessage(scanner.Bytes())
		if msg != nil {
			messages = append(messages, *msg)
		}
	}

	if err := scanner.Err(); err != nil {
		return messages, err
	}

	// Pair tool uses with their results
	PairToolInteractions(messages)

	return messages, nil
}

// parseMessage extracts a Message from a JSONL line. Returns nil for
// unrecognized types. Individual field unmarshals intentionally ignore
// errors — missing or malformed fields are left at zero values since
// JSONL entries have inconsistent schemas across Claude Code versions.
func parseMessage(line []byte) *Message {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil
	}

	var msgType string
	if t, ok := raw["type"]; ok {
		json.Unmarshal(t, &msgType)
	}

	switch msgType {
	case "user", "assistant":
		return parseConversationMessage(raw, msgType)
	case "system":
		return parseSystemMessage(raw)
	default:
		return nil
	}
}

func parseConversationMessage(raw map[string]json.RawMessage, msgType string) *Message {
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
				InputTokens              int    `json:"input_tokens"`
				OutputTokens             int    `json:"output_tokens"`
				CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
				CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
				ServiceTier              string `json:"service_tier"`
			} `json:"usage"`
		}
		if err := json.Unmarshal(msgData, &inner); err == nil {
			msg.Role = inner.Role
			msg.Model = inner.Model
			if inner.Usage != nil {
				msg.Usage = TokenUsage{
					InputTokens:              inner.Usage.InputTokens,
					OutputTokens:             inner.Usage.OutputTokens,
					CacheCreationInputTokens: inner.Usage.CacheCreationInputTokens,
					CacheReadInputTokens:     inner.Usage.CacheReadInputTokens,
					ServiceTier:              inner.Usage.ServiceTier,
				}
			}

			msg.ContentBlocks = parseContentBlocks(inner.Content)
			msg.RawText = extractText(msg.ContentBlocks)
		}
	}

	return msg
}

func parseContentBlocks(content json.RawMessage) []ContentBlock {
	if content == nil {
		return nil
	}

	// Try as plain string first
	var contentStr string
	if err := json.Unmarshal(content, &contentStr); err == nil {
		return []ContentBlock{{Type: "text", Text: contentStr}}
	}

	// Parse as array of blocks
	var rawBlocks []struct {
		Type      string          `json:"type"`
		Text      string          `json:"text"`
		Thinking  string          `json:"thinking"`
		ID        string          `json:"id"`
		Name      string          `json:"name"`
		Input     json.RawMessage `json:"input"`
		ToolUseID string          `json:"tool_use_id"`
		Content   json.RawMessage `json:"content"`
		IsError   bool            `json:"is_error"`
	}

	if err := json.Unmarshal(content, &rawBlocks); err != nil {
		return nil
	}

	var blocks []ContentBlock
	for _, b := range rawBlocks {
		block := ContentBlock{Type: b.Type}

		switch b.Type {
		case "text":
			block.Text = b.Text

		case "tool_use":
			block.ToolID = b.ID
			block.ToolName = b.Name
			if b.Input != nil {
				var inputMap map[string]interface{}
				if err := json.Unmarshal(b.Input, &inputMap); err == nil {
					block.Input = inputMap
				}
			}

		case "tool_result":
			block.ToolUseID = b.ToolUseID
			block.IsError = b.IsError
			// Content can be a string or an array of content blocks
			if b.Content != nil {
				var resultStr string
				if err := json.Unmarshal(b.Content, &resultStr); err == nil {
					block.Content = resultStr
				} else {
					// Try as array of nested content blocks
					var nested []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}
					if err := json.Unmarshal(b.Content, &nested); err == nil {
						var parts []string
						for _, n := range nested {
							if n.Type == "text" {
								parts = append(parts, n.Text)
							}
						}
						block.Content = strings.Join(parts, "\n")
					}
				}
			}

		case "thinking":
			block.Thinking = b.Thinking
		}

		blocks = append(blocks, block)
	}

	return blocks
}

// extractText pulls out all text content from content blocks.
func extractText(blocks []ContentBlock) string {
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func parseSystemMessage(raw map[string]json.RawMessage) *Message {
	msg := &Message{Type: "system"}

	if ts, ok := raw["timestamp"]; ok {
		var tsStr string
		if err := json.Unmarshal(ts, &tsStr); err == nil {
			msg.Timestamp, _ = time.Parse(time.RFC3339Nano, tsStr)
		}
	}

	if s, ok := raw["subtype"]; ok {
		json.Unmarshal(s, &msg.Subtype)
	}
	if d, ok := raw["durationMs"]; ok {
		json.Unmarshal(d, &msg.DurationMs)
	}

	return msg
}

// PairToolInteractions matches tool_use blocks in assistant messages
// with tool_result blocks in the following user messages.
func PairToolInteractions(messages []Message) {
	// Build index of tool_use blocks by their ID
	toolUseIndex := make(map[string]*Message) // tool_use ID -> assistant message
	toolUseBlocks := make(map[string]ContentBlock)

	for i := range messages {
		msg := &messages[i]
		for _, block := range msg.ContentBlocks {
			if block.Type == "tool_use" && block.ToolID != "" {
				toolUseIndex[block.ToolID] = msg
				toolUseBlocks[block.ToolID] = block
			}
		}
	}

	// Walk through messages looking for tool_result blocks
	for _, msg := range messages {
		for _, block := range msg.ContentBlocks {
			if block.Type == "tool_result" && block.ToolUseID != "" {
				if assistantMsg, ok := toolUseIndex[block.ToolUseID]; ok {
					useBlock := toolUseBlocks[block.ToolUseID]
					pair := ToolInteraction{
						Use:    useBlock,
						Result: block,
						Name:   useBlock.ToolName,
					}
					assistantMsg.ToolPairs = append(assistantMsg.ToolPairs, pair)
				}
			}
		}
	}
}

// --- Path helpers ---

func decodeProjectName(dirName string) string {
	return projectNameFromPath(decodeDirToPath(dirName))
}

func decodeDirToPath(dirName string) string {
	if strings.HasPrefix(dirName, "-") {
		return strings.ReplaceAll(dirName, "-", "/")
	}
	return dirName
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
