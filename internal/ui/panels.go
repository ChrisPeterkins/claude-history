package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/chrispeterkins/claude-history/internal/data"
)

// sessionListItem represents either a date group header or a session entry.
type sessionListItem struct {
	isGroupHeader bool
	groupLabel    string
	session       data.Session
	origIdx       int
}

// --- Panel rendering ---

func (m Model) renderProjectsPanel() string {
	w := m.projectsWidth()
	h := m.contentHeight()

	title := panelTitleStyle.Render("Projects")
	if m.focus == panelProjects {
		title = panelTitleActiveStyle.Render(" Projects ")
	}

	var items []string
	items = append(items, title)

	visibleStart, visibleEnd := m.visibleRange(m.projectCursor, len(m.projects), h-2)
	for i := visibleStart; i < visibleEnd; i++ {
		p := m.projects[i]
		// Build suffix with activity dot, session count, and history indicator
		dot := activityDot(p.LastActive)
		suffix := ""
		if dot != "" {
			suffix = " " + dot
		}
		if p.SessionCount > 0 {
			suffix += tokenStyle.Render(fmt.Sprintf(" (%d)", p.SessionCount))
		}
		if p.HistoryOnly {
			suffix += " ○"
		}
		name := truncateStr(p.Name, w-8)
		focused := m.focus == panelProjects
		if i == m.projectCursor {
			style := selectedItemStyle
			if !focused {
				style = dimSelectedItemStyle
			}
			items = append(items, style.Width(w-4).Render(name+suffix))
		} else {
			style := itemStyle
			if !focused {
				style = dimItemStyle
			}
			items = append(items, style.Width(w-4).Render(name+suffix))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	return m.panelStyleFor(panelProjects).Width(w).Height(h).Render(content)
}

func (m Model) renderSessionsPanel() string {
	w := m.sessionsWidth()
	h := m.contentHeight()

	filterLabel := ""
	if m.sessionFilter > 0 {
		filterLabel = " (" + sessionFilterTypes[m.sessionFilter].label + ")"
	}
	title := panelTitleStyle.Render("Sessions" + filterLabel)
	if m.focus == panelSessions {
		title = panelTitleActiveStyle.Render(" Sessions" + filterLabel + " ")
	}

	var items []string
	items = append(items, title)

	if len(m.sessions) == 0 {
		items = append(items, "\n"+emptyLogoStyle.Width(w-4).Render("◈")+"\n"+emptyStyle.Width(w-4).Render("No sessions"))
	} else {
		groups := GroupSessionsByDate(m.filterSessions(m.sessions))
		// Flatten groups into displayable items with cursor tracking
		var flatItems []sessionListItem
		for _, g := range groups {
			flatItems = append(flatItems, sessionListItem{isGroupHeader: true, groupLabel: g.Label})
			for _, is := range g.Sessions {
				flatItems = append(flatItems, sessionListItem{
					session:  is.Session,
					origIdx:  is.OriginalIndex,
				})
			}
		}

		// Find which flat index corresponds to the cursor
		cursorFlat := 0
		for i, item := range flatItems {
			if !item.isGroupHeader && item.origIdx == m.sessionCursor {
				cursorFlat = i
				break
			}
		}

		visibleStart, visibleEnd := m.visibleRange(cursorFlat, len(flatItems), h-2)
		for i := visibleStart; i < visibleEnd; i++ {
			item := flatItems[i]
			if item.isGroupHeader {
				items = append(items, dateGroupStyle.Width(w-4).Render("  "+item.groupLabel))
				continue
			}

			s := item.session
			date := relativeTime(s.StartedAt)
			preview := truncateStr(s.Preview, w-6)
			if preview == "" {
				preview = "(empty session)"
			}
			stats := sessionStatsLine(s)

			focused := m.focus == panelSessions
			if item.origIdx == m.sessionCursor {
				s1 := selectedItemStyle
				if !focused {
					s1 = dimSelectedItemStyle
				}
				line1 := s1.Width(w - 4).Render(date + "  " + stats)
				// Preview elevated to normal brightness as pseudo-title
				line2 := s1.Width(w - 4).Render(preview)
				items = append(items, line1, line2)
			} else {
				s1, s2 := itemStyle, itemDescStyle
				if !focused {
					s1, s2 = dimItemStyle, dimItemDescStyle
				}
				line1 := s1.Width(w - 4).Render(date + "  " + stats)
				line2 := s2.Width(w - 4).Render(preview)
				items = append(items, line1, line2)
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	return m.panelStyleFor(panelSessions).Width(w).Height(h).Render(content)
}

func (m Model) renderConversationPanel() string {
	w := m.conversationWidth()
	h := m.contentHeight()

	var title string
	if m.convSearchMode {
		matchInfo := ""
		if len(m.convSearchMatches) > 0 {
			matchInfo = fmt.Sprintf(" %d/%d", m.convSearchIdx+1, len(m.convSearchMatches))
		}
		title = panelTitleActiveStyle.Render(" Find"+matchInfo+" ") + " " + m.convSearchInput.View()
	} else if m.focus == panelConversation {
		title = panelTitleActiveStyle.Render(" Conversation ")
	} else {
		title = panelTitleStyle.Render("Conversation")
	}

	scrollInfo := ""
	if m.viewport.TotalLineCount() > 0 {
		pct := int(m.viewport.ScrollPercent() * 100)
		scrollInfo = tokenStyle.Render(fmt.Sprintf(" %d%%", pct))
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, title, scrollInfo)

	m.viewport.Width = w - 6 // leave room for scroll indicator
	m.viewport.Height = h - 3

	var body string
	if m.loading {
		body = "\n\n" + emptyStyle.Width(w-6).Render(m.spinner.View()+" Loading session...")
	} else if m.focus != panelConversation && len(m.messages) == 0 && m.sessionCursor < len(m.sessions) {
		// Session peek: show preview when browsing sessions
		s := m.sessions[m.sessionCursor]
		peek := "\n\n" + timestampStyle.Render("  Preview") + "\n\n"
		if s.Preview != "" {
			preview := s.Preview
			if len(preview) > 300 {
				preview = preview[:297] + "..."
			}
			peek += emptyStyle.Width(w - 6).Render(preview)
		}
		peek += "\n\n" + tokenStyle.Render("  "+sessionStatsLine(s))
		peek += "\n\n" + timestampStyle.Render("  press enter to load full conversation")
		body = peek
	} else {
		body = m.applyLineHighlight(m.viewport.View(), w-6)
	}

	// Add scroll indicator
	scrollbar := m.renderScrollbar(h - 3)
	bodyWithScroll := lipgloss.JoinHorizontal(lipgloss.Top, body, " ", scrollbar)

	content := lipgloss.JoinVertical(lipgloss.Left, header, bodyWithScroll)

	return m.panelStyleFor(panelConversation).Width(w).Height(h).Render(content)
}

// --- Layout math ---

func (m Model) projectsWidth() int {
	// Single panel mode: take full width
	if (m.fullScreen || m.width < breakpointNarrow) && m.focus == panelProjects {
		return m.width
	}
	if m.fullScreen || m.width < breakpointNarrow {
		return 0
	}
	// Two-panel mode at medium width
	if m.width < breakpointMedium {
		if m.focus == panelProjects {
			return max(20, m.width*2/5)
		}
		return 0 // hidden when focus is sessions or conversation
	}
	return max(20, m.width/5)
}

func (m Model) sessionsWidth() int {
	// Single panel mode: take full width
	if (m.fullScreen || m.width < breakpointNarrow) && m.focus == panelSessions {
		return m.width
	}
	if m.fullScreen || m.width < breakpointNarrow {
		return 0
	}
	// Two-panel mode at medium width
	if m.width < breakpointMedium {
		if m.focus == panelProjects {
			return m.width - m.projectsWidth()
		}
		return max(24, m.width/3)
	}
	return max(30, m.width*3/10)
}

func (m Model) conversationWidth() int {
	// Single panel mode: take full width
	if (m.fullScreen || m.width < breakpointNarrow) && m.focus == panelConversation {
		return m.width
	}
	if m.fullScreen || m.width < breakpointNarrow {
		return 0
	}
	w := m.width - m.projectsWidth() - m.sessionsWidth()
	if w < 30 {
		return 30
	}
	return w
}

func (m Model) contentHeight() int {
	h := m.height - 3 // header + help bar + padding
	if h < 5 {
		return 5
	}
	return h
}

func (m Model) visibleRange(cursor, total, height int) (int, int) {
	if total <= height {
		return 0, total
	}
	start := cursor - height/2
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > total {
		end = total
		start = end - height
	}
	return start, end
}

// sessionStatsLine builds a compact stats string for a session.
func sessionStatsLine(s data.Session) string {
	var parts []string

	if s.MessageCount > 0 {
		parts = append(parts, fmt.Sprintf("%d msgs", s.MessageCount))
	}
	if s.TotalTokensOut > 0 {
		parts = append(parts, formatTokenCount(s.TotalTokensOut)+" tok")
	}
	if s.TotalDurationMs > 0 {
		dur := time.Duration(s.TotalDurationMs) * time.Millisecond
		if dur >= time.Minute {
			parts = append(parts, fmt.Sprintf("%dm", int(dur.Minutes())))
		} else {
			parts = append(parts, fmt.Sprintf("%ds", int(dur.Seconds())))
		}
	}

	if len(parts) == 0 {
		return formatSize(s.FileSize)
	}
	return strings.Join(parts, " · ")
}

// panelStyleFor returns the appropriate border style for a panel.
func (m Model) panelStyleFor(panel int) lipgloss.Style {
	if m.focus == panel && time.Now().Before(m.transitionUntil) {
		return transitionPanelStyle
	}
	if m.focus == panel {
		return activePanelStyle
	}
	return panelStyle
}

// renderScrollbar renders a vertical scroll indicator for the conversation viewport.
func (m Model) renderScrollbar(height int) string {
	if height <= 0 || m.viewport.TotalLineCount() <= m.viewport.Height {
		// No scrollbar needed — content fits
		track := strings.Repeat(" \n", height)
		return track
	}

	// Calculate thumb position
	pct := m.viewport.ScrollPercent()
	thumbPos := int(pct * float64(height-1))

	var lines []string
	for i := 0; i < height; i++ {
		if i == thumbPos {
			lines = append(lines, scrollThumbStyle.Render("█"))
		} else {
			lines = append(lines, scrollTrackStyle.Render("│"))
		}
	}
	return strings.Join(lines, "\n")
}

// renderHeader renders the top header bar with logo and breadcrumb.
func (m Model) renderHeader() string {
	logo := headerStyle.Render(" ◈ Claude History")
	breadcrumb := m.renderBreadcrumb()

	if breadcrumb != "" {
		breadcrumb = headerBreadcrumbStyle.Render(breadcrumb) + " "
	}

	// Fill the middle with ─
	logoLen := lipgloss.Width(logo)
	bcLen := lipgloss.Width(breadcrumb)
	fillLen := m.width - logoLen - bcLen - 2
	if fillLen < 3 {
		fillLen = 3
	}
	fill := headerLineStyle.Render(" " + strings.Repeat("─", fillLen) + " ")

	return logo + fill + breadcrumb
}

// applyLineHighlight highlights the line of the nearest collapsible section
// in the viewport output. Finds which section space-bar would target, then
// highlights its actual line position — not just the visual center.
func (m Model) applyLineHighlight(viewOutput string, maxWidth int) string {
	if m.focus != panelConversation || viewOutput == "" {
		return viewOutput
	}

	lines := strings.Split(viewOutput, "\n")

	// Highlight search matches if in conv search mode
	if m.convSearchMode && len(m.convSearchMatches) > 0 {
		matchSet := make(map[int]bool)
		for _, absLine := range m.convSearchMatches {
			rel := absLine - m.viewport.YOffset
			matchSet[rel] = true
		}
		// Highlight current match differently
		currentRel := -1
		if m.convSearchIdx < len(m.convSearchMatches) {
			currentRel = m.convSearchMatches[m.convSearchIdx] - m.viewport.YOffset
		}
		for i := range lines {
			if i == currentRel {
				lines[i] = selectedItemStyle.Width(maxWidth).Render(strings.TrimRight(lines[i], " "))
			} else if matchSet[i] {
				lines[i] = dimSelectedItemStyle.Width(maxWidth).Render(strings.TrimRight(lines[i], " "))
			}
		}
		return strings.Join(lines, "\n")
	}

	// Normal mode: highlight nearest collapsible section
	key := m.nearestCollapsibleKey()
	if key == "" {
		return viewOutput
	}

	absLine, ok := m.collapsibleLines[key]
	if !ok {
		return viewOutput
	}
	relativeLine := absLine - m.viewport.YOffset

	if relativeLine < 0 || relativeLine >= len(lines) {
		return viewOutput
	}

	lines[relativeLine] = selectedItemStyle.Width(maxWidth).Render(
		strings.TrimRight(lines[relativeLine], " "),
	)

	return strings.Join(lines, "\n")
}

// activityDot returns a colored dot indicating project recency.
func activityDot(lastActive time.Time) string {
	if lastActive.IsZero() {
		return ""
	}
	age := time.Since(lastActive)
	switch {
	case age < 7*24*time.Hour:
		return lipgloss.NewStyle().Foreground(colorAccent).Render("●") // green - active this week
	case age < 30*24*time.Hour:
		return lipgloss.NewStyle().Foreground(colorWarm).Render("●") // yellow - active this month
	default:
		return lipgloss.NewStyle().Foreground(colorSubtle).Render("○") // dim - older
	}
}

// filterSessions applies the current session filter.
func (m Model) filterSessions(sessions []data.Session) []data.Session {
	if m.sessionFilter == 0 {
		return sessions // "all" — no filtering
	}

	filterName := sessionFilterTypes[m.sessionFilter].name
	now := time.Now()

	var filtered []data.Session
	for _, s := range sessions {
		switch filterName {
		case "code":
			// Sessions with significant activity (proxy: >10 messages means tool use likely)
			if s.MessageCount > 10 {
				filtered = append(filtered, s)
			}
		case "long":
			if s.MessageCount >= filterLongMinMessages {
				filtered = append(filtered, s)
			}
		case "recent":
			if now.Sub(s.StartedAt).Hours() < float64(filterRecentDays*24) {
				filtered = append(filtered, s)
			}
		}
	}
	return filtered
}
