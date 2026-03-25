package data

import "time"

// Project represents a Claude Code project directory.
type Project struct {
	Name     string // human-readable name (last path segment)
	Path     string // original project path (e.g., /Users/.../MyProject)
	DirName  string // encoded directory name in ~/.claude/projects/
	Sessions []Session
}

// Session represents a single conversation session.
type Session struct {
	ID        string
	Project   *Project
	StartedAt time.Time
	Preview   string // first user message as a preview
	Messages  []Message
	FilePath  string // path to the JSONL file
	FileSize  int64
}

// Message represents a single entry in a conversation.
type Message struct {
	UUID       string
	ParentUUID string
	Type       string // "user", "assistant", "progress", "file-history-snapshot"
	Role       string
	Content    []ContentBlock
	RawText    string // pre-rendered text content
	Timestamp  time.Time
	Model      string
	TokensIn   int
	TokensOut  int
	SessionID  string
	ToolUses   []ToolUse
}

// ContentBlock represents a piece of message content.
type ContentBlock struct {
	Type    string // "text", "tool_use", "tool_result", "thinking"
	Text    string
	ToolID  string
	ToolName string
	Input   string // JSON string of tool input
}

// ToolUse captures a tool invocation within a message.
type ToolUse struct {
	Name  string
	Input string
}

// HistoryEntry represents a line from history.jsonl.
type HistoryEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
	SessionID string `json:"sessionId"`
}
