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
	"github.com/charmbracelet/lipgloss"

	"github.com/chrispeterkins/claude-history/internal/data"
)

type renderResult struct {
	content          string
	userLines        []int
	collapsibleLines map[string]int
}

func (m Model) renderConversation() renderResult {
	if len(m.messages) == 0 {
		w := m.conversationWidth() - conversationPadding
		boxContent := logoStyle.Render("◈  Claude History") + "\n\n" +
			emptyStyle.Render("Browse your Claude\nCode conversations")
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 3).
			Align(lipgloss.Center).
			Width(min(36, w)).
			Render(boxContent)
		hints := helpKeyStyle.Render("↑/↓") + " " + helpDescStyle.Render("navigate") +
			"  " + helpKeyStyle.Render("enter") + " " + helpDescStyle.Render("select") +
			"  " + helpKeyStyle.Render("?") + " " + helpDescStyle.Render("help")
		empty := "\n\n" + lipgloss.PlaceHorizontal(w, lipgloss.Center, box) +
			"\n\n" + lipgloss.PlaceHorizontal(w, lipgloss.Center, hints)
		return renderResult{content: empty}
	}

	w := m.conversationWidth() - conversationPadding
	var parts []string
	var userLines []int
	lineCount := 0

	// Compute and render stats header
	statsHeader := m.renderConversationStats(w)
	if statsHeader != "" {
		parts = append(parts, statsHeader)
		lineCount += strings.Count(statsHeader, "\n") + 2
	}

	hasRendered := false
	var lastTimestamp time.Time
	for _, msg := range m.messages {
		var rendered string

		// Insert time gap indicator if significant time passed
		if hasRendered && !msg.Timestamp.IsZero() && !lastTimestamp.IsZero() {
			gap := msg.Timestamp.Sub(lastTimestamp)
			if gapStr := formatTimeGap(gap, w); gapStr != "" {
				parts = append(parts, gapStr)
				lineCount += 2
			}
		}

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
			if !msg.Timestamp.IsZero() {
				lastTimestamp = msg.Timestamp
			}
		}
	}

	content := strings.Join(parts, "\n")

	// Scan the final output to find exact line positions of collapsible sections.
	// This is the only reliable way since glamour output has unpredictable height.
	collapsibleLines := scanCollapsibleLines(content, m.messages)

	return renderResult{
		content:          content,
		userLines:        userLines,
		collapsibleLines: collapsibleLines,
	}
}

// scanCollapsibleLines finds the line numbers of collapsible section headers
// in the final rendered content by matching the arrow markers (▸/▾).
// Returns a map of sequential index → line number for use by the highlight.
func scanCollapsibleLines(content string, messages []data.Message) map[string]int {
	result := make(map[string]int)
	lines := strings.Split(content, "\n")

	// Build ordered list of collapsible keys from messages
	var keys []string
	for _, msg := range messages {
		if msg.Type != "assistant" {
			continue
		}
		for _, block := range msg.ContentBlocks {
			switch block.Type {
			case "thinking":
				keys = append(keys, "thinking:"+msg.UUID)
			case "tool_use":
				keys = append(keys, "tool:"+block.ToolID)
			}
		}
	}

	// Find lines containing ▸ or ▾ (collapsible headers) and match to keys in order
	keyIdx := 0
	for lineNum, line := range lines {
		if keyIdx >= len(keys) {
			break
		}
		if strings.Contains(line, "▸") || strings.Contains(line, "▾") {
			result[keys[keyIdx]] = lineNum
			keyIdx++
		}
	}

	return result
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

func (m Model) renderAssistantMessage(msg data.Message, w int) string {
	ts := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))
	avatar := avatarAssistantStyle.Render("◆")
	label := " " + avatar + " " + assistantLabelStyle.Render("Claude") + " " + ts

	var sections []string
	sections = append(sections, label)

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

	header := toolHeaderStyle.Render(fmt.Sprintf("%s %s", arrow, toolBadge(block.ToolName))) +
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

	result := tokenStyle.Render("  " + strings.Join(statParts, " · "))

	// Add file change summary
	fileNames := m.collectChangedFiles()
	if len(fileNames) > 0 {
		var fileStr string
		if len(fileNames) <= 3 {
			fileStr = strings.Join(fileNames, ", ")
		} else {
			fileStr = strings.Join(fileNames[:3], ", ") + fmt.Sprintf(" (+%d more)", len(fileNames)-3)
		}
		result += "\n" + tokenStyle.Render("  "+fileStr)
	}

	return result
}

// collectChangedFiles extracts unique file names from Edit/Write tool calls.
func (m Model) collectChangedFiles() []string {
	seen := make(map[string]bool)
	var names []string
	for _, msg := range m.messages {
		for _, pair := range msg.ToolPairs {
			if pair.Name != "Edit" && pair.Name != "Write" {
				continue
			}
			if path, ok := pair.Use.Input["file_path"].(string); ok {
				// Use just the filename
				parts := strings.Split(path, "/")
				name := parts[len(parts)-1]
				if !seen[name] {
					seen[name] = true
					names = append(names, name)
				}
			}
		}
	}
	return names
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

// toolBadge renders a tool name badge with a color specific to the tool type.
func toolBadge(name string) string {
	bg, ok := toolBadgeColors[name]
	if !ok {
		bg = colorWarm
	}
	style := lipgloss.NewStyle().
		Foreground(colorBg).
		Background(bg).
		Bold(true).
		Padding(0, 1)
	return style.Render(name)
}

// formatTimeGap returns a styled time gap indicator if the gap is significant.
func formatTimeGap(gap time.Duration, w int) string {
	if gap < 5*time.Minute {
		return ""
	}

	var label string
	switch {
	case gap < time.Hour:
		label = fmt.Sprintf("%dm", int(gap.Minutes()))
	case gap < 24*time.Hour:
		h := int(gap.Hours())
		m := int(gap.Minutes()) % 60
		if m > 0 {
			label = fmt.Sprintf("%dh %dm", h, m)
		} else {
			label = fmt.Sprintf("%dh", h)
		}
	default:
		d := int(gap.Hours() / 24)
		h := int(gap.Hours()) % 24
		if h > 0 {
			label = fmt.Sprintf("%dd %dh", d, h)
		} else {
			label = fmt.Sprintf("%dd", d)
		}
	}

	if gap < time.Hour {
		return systemMessageStyle.Width(w).Render("··· " + label + " ···")
	}
	return systemMessageStyle.Width(w).Render("── " + label + " ──")
}

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
