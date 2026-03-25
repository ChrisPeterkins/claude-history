package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/chrispeterkins/claude-history/internal/data"
)

type renderResult struct {
	content          string
	userLines        []int
	collapsibleLines map[string]int
}

func (m Model) renderConversation() renderResult {
	if len(m.messages) == 0 {
		empty := "\n\n\n" +
			emptyLogoStyle.Render("◈") + "\n\n" +
			emptyStyle.Render("Select a session\nto view the conversation") + "\n\n" +
			timestampStyle.Render("↑/↓ navigate · enter to select")
		return renderResult{content: empty}
	}

	w := m.conversationWidth() - conversationPadding
	var parts []string
	var userLines []int
	collapsibleLines := make(map[string]int)
	lineCount := 0

	// Compute and render stats header
	statsHeader := m.renderConversationStats(w)
	if statsHeader != "" {
		parts = append(parts, statsHeader)
		lineCount += strings.Count(statsHeader, "\n") + 2
	}

	hasRendered := false
	for _, msg := range m.messages {
		var rendered string

		switch msg.Type {
		case "user":
			if msg.RawText != "" {
				if hasRendered {
					divider := turnDividerStyle.Render(strings.Repeat("─", w))
					parts = append(parts, divider)
					lineCount += 2
				}
				userLines = append(userLines, lineCount)
				rendered = m.renderUserMessage(msg, w)
			}
		case "assistant":
			var blockPositions map[string]int
			rendered, blockPositions = m.renderAssistantMessage(msg, w, lineCount)
			for k, v := range blockPositions {
				collapsibleLines[k] = v
			}
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

	return renderResult{
		content:          strings.Join(parts, "\n"),
		userLines:        userLines,
		collapsibleLines: collapsibleLines,
	}
}

func (m Model) renderUserMessage(msg data.Message, w int) string {
	ts := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
	avatar := avatarUserStyle.Render("●")
	label := " " + avatar + " " + userLabelStyle.Render("You") + " " + ts

	text := msg.RawText
	if text == "" {
		return ""
	}

	// Render through glamour for inline code, bold, etc.
	mdRendered, err := m.renderer.Render(text)
	if err != nil || strings.TrimSpace(mdRendered) == "" {
		mdRendered = text
	}

	body := userBubbleStyle.Width(w).Render(mdRendered)
	return label + "\n" + body
}

func (m Model) renderAssistantMessage(msg data.Message, w int, baseLineCount int) (string, map[string]int) {
	ts := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
	avatar := avatarAssistantStyle.Render("◆")
	label := " " + avatar + " " + assistantLabelStyle.Render("Claude") + " " + ts

	var sections []string
	sections = append(sections, label)
	positions := make(map[string]int)

	// Track line count within this message to record exact positions
	localLines := strings.Count(label, "\n") + 1

	for _, block := range msg.ContentBlocks {
		var rendered string
		var key string

		switch block.Type {
		case "text":
			if block.Text == "" {
				continue
			}
			mdRendered, err := m.renderer.Render(block.Text)
			if err != nil || strings.TrimSpace(mdRendered) == "" {
				mdRendered = block.Text
			}
			rendered = assistantBubbleStyle.Width(w).Render(mdRendered)

		case "thinking":
			key = "thinking:" + msg.UUID
			rendered = m.renderThinkingBlock(block, msg.UUID, w)

		case "tool_use":
			key = "tool:" + block.ToolID
			rendered = m.renderToolCall(block, msg, w)
		}

		if rendered == "" {
			continue
		}

		// Record the absolute line position of this collapsible block
		if key != "" {
			positions[key] = baseLineCount + localLines
		}

		sections = append(sections, rendered)
		localLines += strings.Count(rendered, "\n") + 1
	}

	if msg.Usage.OutputTokens > 0 {
		sections = append(sections, m.renderTokenInfo(msg))
	}

	return strings.Join(sections, "\n"), positions
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
	if len(text) > maxThinkingLen {
		text = text[:maxThinkingLen] + "\n... (truncated)"
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
				if len(result) > maxToolResultLen {
					result = result[:maxToolResultLen] + "\n... (truncated)"
				}
				result = hardWrap(result, w-8)
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

// renderConversationStats builds a stats summary line for the conversation header.
func (m Model) renderConversationStats(w int) string {
	if len(m.messages) == 0 {
		return ""
	}

	var msgCount, toolCount, totalTokens, totalDurationMs int
	for _, msg := range m.messages {
		if msg.Type == "user" || msg.Type == "assistant" {
			msgCount++
		}
		toolCount += len(msg.ToolPairs)
		totalTokens += msg.Usage.OutputTokens
		if msg.Type == "system" && msg.Subtype == "turn_duration" {
			totalDurationMs += msg.DurationMs
		}
	}

	var statParts []string
	if msgCount > 0 {
		statParts = append(statParts, fmt.Sprintf("%d messages", msgCount))
	}
	if toolCount > 0 {
		statParts = append(statParts, fmt.Sprintf("%d tool calls", toolCount))
	}
	if totalTokens > 0 {
		statParts = append(statParts, formatTokenCount(totalTokens)+" tokens")
	}
	if totalDurationMs > 0 {
		dur := time.Duration(totalDurationMs) * time.Millisecond
		if dur >= time.Minute {
			statParts = append(statParts, fmt.Sprintf("%dm %ds", int(dur.Minutes()), int(dur.Seconds())%60))
		} else {
			statParts = append(statParts, fmt.Sprintf("%.0fs", dur.Seconds()))
		}
	}

	if len(statParts) == 0 {
		return ""
	}

	return tokenStyle.Render("  " + strings.Join(statParts, " · "))
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
			if len(cmd) > maxCommandSummaryLen {
				cmd = cmd[:maxCommandSummaryLen-3] + "..."
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
			summary := shortPath(path)
			// Count lines changed
			if old, ok := block.Input["old_string"].(string); ok {
				if new, ok := block.Input["new_string"].(string); ok {
					oldN := strings.Count(old, "\n") + 1
					newN := strings.Count(new, "\n") + 1
					summary += fmt.Sprintf(" (-%d/+%d)", oldN, newN)
				}
			}
			return summary
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
			lines := strings.Split(cmd, "\n")
			if len(lines) == 1 {
				return "$ " + cmd
			}
			var parts []string
			for i, l := range lines {
				if i == 0 {
					parts = append(parts, "$ "+l)
				} else {
					parts = append(parts, "> "+l)
				}
			}
			return strings.Join(parts, "\n")
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
			header := fmt.Sprintf("File: %s", shortPath(path))
			if content, ok := block.Input["content"].(string); ok {
				lineCount := strings.Count(content, "\n") + 1
				header = fmt.Sprintf("File: %s (%d lines)", shortPath(path), lineCount)
				preview := content
				if len(preview) > maxToolResultLen {
					preview = preview[:maxToolResultLen] + "\n... (truncated)"
				}
				// Try syntax highlighting
				highlighted := highlightCode(preview, path)
				if highlighted != "" {
					return header + "\n" + highlighted
				}
				// Fallback: render as added lines
				var diffLines []string
				for _, l := range strings.Split(preview, "\n") {
					diffLines = append(diffLines, diffAddStyle.Render("+ "+l))
				}
				return header + "\n" + strings.Join(diffLines, "\n")
			}
			return header
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

// renderDiff renders old/new strings as a unified diff, showing only changed lines
// with a few lines of context around each change.
func renderDiff(old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	// Find common prefix and suffix to show only the changed region
	prefixLen := 0
	minLen := len(oldLines)
	if len(newLines) < minLen {
		minLen = len(newLines)
	}
	for prefixLen < minLen && oldLines[prefixLen] == newLines[prefixLen] {
		prefixLen++
	}

	suffixLen := 0
	for suffixLen < minLen-prefixLen &&
		oldLines[len(oldLines)-1-suffixLen] == newLines[len(newLines)-1-suffixLen] {
		suffixLen++
	}

	// If nothing changed (shouldn't happen but defensive)
	if prefixLen+suffixLen >= len(oldLines) && prefixLen+suffixLen >= len(newLines) {
		return diffHeaderStyle.Render("(no changes)")
	}

	var lines []string

	// Show up to 2 lines of context before the change
	contextStart := prefixLen - 2
	if contextStart < 0 {
		contextStart = 0
	}
	for i := contextStart; i < prefixLen; i++ {
		lines = append(lines, timestampStyle.Render("  "+oldLines[i]))
	}

	// Show removed lines
	removedEnd := len(oldLines) - suffixLen
	for i := prefixLen; i < removedEnd; i++ {
		lines = append(lines, diffRemoveStyle.Render("- "+oldLines[i]))
	}

	// Show added lines
	addedEnd := len(newLines) - suffixLen
	for i := prefixLen; i < addedEnd; i++ {
		lines = append(lines, diffAddStyle.Render("+ "+newLines[i]))
	}

	// Show up to 2 lines of context after the change
	contextEnd := len(oldLines) - suffixLen + 2
	if contextEnd > len(oldLines) {
		contextEnd = len(oldLines)
	}
	for i := len(oldLines) - suffixLen; i < contextEnd; i++ {
		lines = append(lines, timestampStyle.Render("  "+oldLines[i]))
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

// highlightCode applies syntax highlighting to code based on filename extension.
// Returns empty string if highlighting fails or isn't applicable.
func highlightCode(code, filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ""
	}

	lexer := lexers.Match(filename)
	if lexer == nil {
		return ""
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		return ""
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return ""
	}

	return buf.String()
}

// hardWrap wraps lines that exceed maxWidth by inserting line breaks.
func hardWrap(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	var result []string
	for _, line := range strings.Split(s, "\n") {
		for len(line) > maxWidth {
			result = append(result, line[:maxWidth])
			line = line[maxWidth:]
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}
