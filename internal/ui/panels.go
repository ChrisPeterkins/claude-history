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
		title = panelTitleActiveStyle.Render("Projects")
	}

	var items []string
	items = append(items, title)

	visibleStart, visibleEnd := m.visibleRange(m.projectCursor, len(m.projects), h-2)
	for i := visibleStart; i < visibleEnd; i++ {
		p := m.projects[i]
		name := truncateStr(p.Name, w-6)
		// Dim indicator for history-only projects
		suffix := ""
		if p.HistoryOnly {
			suffix = " ○"
		}
		if i == m.projectCursor {
			items = append(items, selectedItemStyle.Width(w-4).Render("▸ "+name+suffix))
		} else {
			items = append(items, itemStyle.Width(w-4).Render("  "+name+suffix))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	style := panelStyle
	if m.focus == panelProjects {
		style = activePanelStyle
	}

	return style.Width(w).Height(h).Render(content)
}

func (m Model) renderSessionsPanel() string {
	w := m.sessionsWidth()
	h := m.contentHeight()

	title := panelTitleStyle.Render("Sessions")
	if m.focus == panelSessions {
		title = panelTitleActiveStyle.Render("Sessions")
	}

	var items []string
	items = append(items, title)

	if len(m.sessions) == 0 {
		items = append(items, emptyStyle.Width(w-4).Render("No sessions"))
	} else {
		groups := GroupSessionsByDate(m.sessions)
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
			date := s.StartedAt.Format("Jan 02 15:04")
			preview := truncateStr(s.Preview, w-6)
			if preview == "" {
				preview = "(empty session)"
			}
			stats := sessionStatsLine(s)

			if item.origIdx == m.sessionCursor {
				line1 := selectedItemStyle.Width(w - 4).Render("▸ " + date + "  " + stats)
				line2 := selectedItemDescStyle.Width(w - 4).Render("  " + preview)
				items = append(items, line1, line2)
			} else {
				line1 := itemStyle.Width(w - 4).Render("  " + date + "  " + stats)
				line2 := itemDescStyle.Width(w - 4).Render("  " + preview)
				items = append(items, line1, line2)
			}
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	style := panelStyle
	if m.focus == panelSessions {
		style = activePanelStyle
	}

	return style.Width(w).Height(h).Render(content)
}

func (m Model) renderConversationPanel() string {
	w := m.conversationWidth()
	h := m.contentHeight()

	title := panelTitleStyle.Render("Conversation")
	if m.focus == panelConversation {
		title = panelTitleActiveStyle.Render("Conversation")
	}

	scrollInfo := ""
	if m.viewport.TotalLineCount() > 0 {
		pct := int(m.viewport.ScrollPercent() * 100)
		scrollInfo = tokenStyle.Render(fmt.Sprintf(" %d%%", pct))
	}

	header := lipgloss.JoinHorizontal(lipgloss.Center, title, scrollInfo)

	m.viewport.Width = w - 4
	m.viewport.Height = h - 3

	content := lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View())

	style := panelStyle
	if m.focus == panelConversation {
		style = activePanelStyle
	}

	return style.Width(w).Height(h).Render(content)
}

// --- Layout math ---

func (m Model) projectsWidth() int {
	// Single panel mode: take full width
	if (m.fullScreen || m.width < 60) && m.focus == panelProjects {
		return m.width
	}
	if m.fullScreen || m.width < 60 {
		return 0
	}
	// Two-panel mode at medium width
	if m.width < 100 {
		if m.focus == panelProjects {
			return max(20, m.width*2/5)
		}
		return 0 // hidden when focus is sessions or conversation
	}
	return max(20, m.width/5)
}

func (m Model) sessionsWidth() int {
	// Single panel mode: take full width
	if (m.fullScreen || m.width < 60) && m.focus == panelSessions {
		return m.width
	}
	if m.fullScreen || m.width < 60 {
		return 0
	}
	// Two-panel mode at medium width
	if m.width < 100 {
		if m.focus == panelProjects {
			return m.width - m.projectsWidth()
		}
		return max(24, m.width/3)
	}
	return max(30, m.width*3/10)
}

func (m Model) conversationWidth() int {
	// Single panel mode: take full width
	if (m.fullScreen || m.width < 60) && m.focus == panelConversation {
		return m.width
	}
	if m.fullScreen || m.width < 60 {
		return 0
	}
	w := m.width - m.projectsWidth() - m.sessionsWidth()
	if w < 30 {
		return 30
	}
	return w
}

func (m Model) contentHeight() int {
	h := m.height - 2
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
