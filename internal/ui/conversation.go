package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chrispeterkins/claude-history/internal/data"
)

func (m *Model) renderConversation() string {
	if len(m.messages) == 0 {
		return emptyStyle.Render("\n\n  No messages to display")
	}

	w := m.conversationWidth() - 8
	var parts []string
	m.userMessageLines = nil
	lineCount := 0

	hasRendered := false
	for _, msg := range m.messages {
		var rendered string

		switch msg.Type {
		case "user":
			if msg.RawText != "" {
				// Add turn divider before user messages (except the first)
				if hasRendered {
					divider := turnDividerStyle.Render(strings.Repeat("─", w))
					parts = append(parts, divider)
					lineCount += 2
				}
				m.userMessageLines = append(m.userMessageLines, lineCount)
				rendered = m.renderUserMessage(msg, w)
			}
		case "assistant":
			rendered = m.renderAssistantMessage(msg, w)
		case "system":
			if msg.Subtype == "turn_duration" && msg.DurationMs > 0 {
				rendered = m.renderSystemMessage(msg, w)
			}
		}

		if rendered != "" {
			parts = append(parts, rendered)
			lineCount += strings.Count(rendered, "\n") + 2
			hasRendered = true
		}
	}

	return strings.Join(parts, "\n")
}

func (m Model) renderUserMessage(msg data.Message, w int) string {
	ts := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
	label := userLabelStyle.Render("  You") + " " + ts

	text := msg.RawText
	if text == "" {
		// Check if this is a tool result only message (no text)
		return ""
	}

	body := userBubbleStyle.Width(w).Render(text)
	return label + "\n" + body
}

func (m Model) renderAssistantMessage(msg data.Message, w int) string {
	ts := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
	label := assistantLabelStyle.Render("  Claude") + " " + ts

	var sections []string
	sections = append(sections, label)

	// Render each content block in order
	for _, block := range msg.ContentBlocks {
		switch block.Type {
		case "text":
			if block.Text == "" {
				continue
			}
			mdRendered, err := m.renderer.Render(block.Text)
			if err != nil || strings.TrimSpace(mdRendered) == "" {
				mdRendered = block.Text
			}
			sections = append(sections, assistantBubbleStyle.Width(w).Render(mdRendered))

		case "thinking":
			sections = append(sections, m.renderThinkingBlock(block, msg.UUID, w))

		case "tool_use":
			sections = append(sections, m.renderToolCall(block, msg, w))
		}
	}

	// Token info
	if msg.Usage.OutputTokens > 0 {
		sections = append(sections, m.renderTokenInfo(msg))
	}

	return strings.Join(sections, "\n")
}

func (m Model) renderThinkingBlock(block data.ContentBlock, msgUUID string, w int) string {
	key := "thinking:" + msgUUID
	collapsed := m.isCollapsed(key)

	if collapsed {
		return thinkingGutterStyle.Render(thinkingHeaderStyle.Render("▸ Thinking..."))
	}

	header := thinkingHeaderStyle.Render("▾ Thinking")
	text := block.Thinking
	if text == "" {
		text = "(redacted)"
	}
	if len(text) > 2000 {
		text = text[:2000] + "\n... (truncated)"
	}
	body := thinkingBodyStyle.Width(w - 6).Render(text)
	return thinkingGutterStyle.Render(header + "\n" + body)
}

func (m Model) renderToolCall(block data.ContentBlock, msg data.Message, w int) string {
	key := "tool:" + block.ToolID
	collapsed := m.isCollapsed(key)

	// Build header with tool name and a brief summary
	summary := toolCallSummary(block)
	arrow := "▸"
	if !collapsed {
		arrow = "▾"
	}

	header := toolHeaderStyle.Render(fmt.Sprintf("%s %s", arrow, toolBadgeStyle.Render(block.ToolName))) +
		" " + toolHeaderStyle.Render(summary)

	if collapsed {
		return toolGutterCollapsedStyle.Render(header)
	}

	var bodyParts []string

	// Show input
	inputStr := formatToolInput(block)
	if inputStr != "" {
		bodyParts = append(bodyParts, toolBodyStyle.Width(w-6).Render(inputStr))
	}

	// Find and show the paired result
	for _, pair := range msg.ToolPairs {
		if pair.Use.ToolID == block.ToolID {
			result := pair.Result.Content
			if result != "" {
				if len(result) > 3000 {
					result = result[:3000] + "\n... (truncated)"
				}
				if pair.Result.IsError {
					bodyParts = append(bodyParts, toolErrorStyle.Width(w-6).Render("Error: "+result))
				} else {
					bodyParts = append(bodyParts, toolBodyStyle.Width(w-6).Render(result))
				}
			}
			break
		}
	}

	content := header
	if len(bodyParts) > 0 {
		content += "\n" + strings.Join(bodyParts, "\n")
	}

	return toolGutterExpandedStyle.Render(content)
}

func (m Model) renderSystemMessage(msg data.Message, w int) string {
	dur := time.Duration(msg.DurationMs) * time.Millisecond
	var durStr string
	if dur >= time.Minute {
		durStr = fmt.Sprintf("%dm %ds", int(dur.Minutes()), int(dur.Seconds())%60)
	} else {
		durStr = fmt.Sprintf("%.1fs", dur.Seconds())
	}
	line := fmt.Sprintf("── turn completed in %s ──", durStr)
	return systemMessageStyle.Width(w).Render(line)
}

func (m Model) renderTokenInfo(msg data.Message) string {
	u := msg.Usage
	parts := []string{msg.Model}

	if u.OutputTokens > 0 {
		parts = append(parts, formatTokenCount(u.OutputTokens)+" out")
	}
	if u.CacheCreationInputTokens > 0 {
		parts = append(parts, formatTokenCount(u.CacheCreationInputTokens)+" cache create")
	}
	if u.CacheReadInputTokens > 0 {
		parts = append(parts, formatTokenCount(u.CacheReadInputTokens)+" cache read")
	}

	return tokenStyle.Render("  " + strings.Join(parts, " · "))
}

func (m Model) isCollapsed(key string) bool {
	collapsed, exists := m.collapsed[key]
	if !exists {
		return true // default: collapsed
	}
	return collapsed
}

// --- Helpers ---

// toolCallSummary returns a brief description of what a tool call does.
func toolCallSummary(block data.ContentBlock) string {
	if block.Input == nil {
		return ""
	}

	switch block.ToolName {
	case "Bash":
		if cmd, ok := block.Input["command"].(string); ok {
			cmd = strings.ReplaceAll(cmd, "\n", " ")
			if len(cmd) > 60 {
				cmd = cmd[:57] + "..."
			}
			return cmd
		}
	case "Read":
		if path, ok := block.Input["file_path"].(string); ok {
			return shortPath(path)
		}
	case "Write":
		if path, ok := block.Input["file_path"].(string); ok {
			return shortPath(path)
		}
	case "Edit":
		if path, ok := block.Input["file_path"].(string); ok {
			return shortPath(path)
		}
	case "Glob":
		if pattern, ok := block.Input["pattern"].(string); ok {
			return pattern
		}
	case "Grep":
		if pattern, ok := block.Input["pattern"].(string); ok {
			return pattern
		}
	case "Agent":
		if desc, ok := block.Input["description"].(string); ok {
			return desc
		}
	}
	return ""
}

// formatToolInput renders tool input as a readable string.
func formatToolInput(block data.ContentBlock) string {
	if block.Input == nil {
		return ""
	}

	switch block.ToolName {
	case "Bash":
		if cmd, ok := block.Input["command"].(string); ok {
			return "$ " + cmd
		}
	case "Edit":
		parts := []string{}
		if path, ok := block.Input["file_path"].(string); ok {
			parts = append(parts, "File: "+shortPath(path))
		}
		if old, ok := block.Input["old_string"].(string); ok {
			if new, ok := block.Input["new_string"].(string); ok {
				parts = append(parts, renderDiff(old, new))
			}
		}
		return strings.Join(parts, "\n")
	case "Write":
		if path, ok := block.Input["file_path"].(string); ok {
			if content, ok := block.Input["content"].(string); ok {
				lines := strings.Count(content, "\n") + 1
				return fmt.Sprintf("File: %s (%d lines)", shortPath(path), lines)
			}
			return "File: " + shortPath(path)
		}
	case "Read":
		if path, ok := block.Input["file_path"].(string); ok {
			return "File: " + shortPath(path)
		}
	default:
		// Generic: marshal to indented JSON
		b, err := json.MarshalIndent(block.Input, "", "  ")
		if err == nil && len(b) < 500 {
			return string(b)
		}
	}
	return ""
}

// renderDiff renders old/new strings as a simple diff.
func renderDiff(old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var lines []string
	for _, l := range oldLines {
		lines = append(lines, diffRemoveStyle.Render("- "+l))
	}
	for _, l := range newLines {
		lines = append(lines, diffAddStyle.Render("+ "+l))
	}
	return strings.Join(lines, "\n")
}

// shortPath returns the last 2-3 segments of a file path.
func shortPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return ".../" + strings.Join(parts[len(parts)-3:], "/")
	}
	return path
}

func truncateStr(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.0fKB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func formatTokenCount(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(val, lo, hi int) int {
	if val < lo {
		return lo
	}
	if val > hi {
		return hi
	}
	return val
}
