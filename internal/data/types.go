package data

import "time"

// Project represents a Claude Code project directory.
type Project struct {
	Name        string // human-readable name (last path segment)
	Path        string // original project path (e.g., /Users/.../MyProject)
	DirName     string // encoded directory name in ~/.claude/projects/
	Sessions     []Session
	SessionCount int       // number of sessions (for display in project list)
	LastActive   time.Time // most recent session activity
	HistoryOnly  bool      // true if this project only has history.jsonl entries (no full sessions)
}

// Session represents a single conversation session.
type Session struct {
	ID              string
	Project         *Project
	StartedAt       time.Time
	Preview         string // first user message as a preview
	Messages        []Message
	FilePath        string // path to the JSONL file
	FileSize        int64
	MessageCount    int
	TotalTokensIn   int
	TotalTokensOut  int
	TotalDurationMs int
	HistoryOnly     bool            // true if sourced from history.jsonl only
	HistoryEntries  []HistoryEntry  // populated for history-only sessions
}

// Message represents a single entry in a conversation.
type Message struct {
	UUID       string
	ParentUUID string
	Type       string // "user", "assistant", "system"
	Role       string
	RawText    string // pre-rendered text content
	Timestamp  time.Time
	Model      string
	Usage      TokenUsage
	SessionID  string

	// Content blocks from the message
	ContentBlocks []ContentBlock

	// Paired tool interactions (populated by PairToolInteractions)
	ToolPairs []ToolInteraction

	// System message fields
	Subtype    string // "turn_duration", etc.
	DurationMs int
}

// TokenUsage holds full token accounting for a message.
type TokenUsage struct {
	InputTokens              int
	OutputTokens             int
	CacheCreationInputTokens int
	CacheReadInputTokens     int
	ServiceTier              string
}

// ContentBlock represents a piece of message content.
type ContentBlock struct {
	Type string // "text", "tool_use", "tool_result", "thinking"

	// For text blocks
	Text string

	// For tool_use blocks
	ToolID   string
	ToolName string
	Input    map[string]interface{} // parsed tool input

	// For tool_result blocks
	ToolUseID string // links to the tool_use this is a result for
	Content   string // tool result output text
	IsError   bool

	// For thinking blocks
	Thinking string
}

// ToolInteraction pairs a tool_use with its tool_result.
type ToolInteraction struct {
	Use    ContentBlock // the tool_use block (from assistant)
	Result ContentBlock // the tool_result block (from following user message)
	Name   string
}

// HistoryEntry represents a line from history.jsonl.
type HistoryEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
	SessionID string `json:"sessionId"`
}
