package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/chrispeterkins/claude-history/internal/data"
)

// Focus panels
const (
	panelProjects = iota
	panelSessions
	panelConversation
)

// Model is the top-level Bubble Tea model.
type Model struct {
	// Data
	projects []data.Project
	sessions []data.Session
	messages []data.Message

	// UI state
	focus         int // which panel is active
	projectCursor int
	sessionCursor int
	viewport      viewport.Model
	ready         bool

	// Dimensions
	width  int
	height int

	// Markdown renderer
	renderer *glamour.TermRenderer
}

// sessionsLoaded is a message indicating sessions have been loaded.
type sessionsLoaded struct {
	sessions []data.Session
}

// messagesLoaded is a message indicating messages have been loaded.
type messagesLoaded struct {
	messages []data.Message
}

func NewModel() Model {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	return Model{
		renderer: r,
	}
}

func (m Model) Init() tea.Cmd {
	return loadProjects
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.viewport = viewport.New(m.conversationWidth(), m.contentHeight())
		m.viewport.Style = lipgloss.NewStyle()
		if len(m.messages) > 0 {
			m.viewport.SetContent(m.renderConversation())
		}
		return m, nil

	case projectsLoaded:
		m.projects = msg.projects
		if len(m.projects) > 0 {
			return m, m.loadSessionsCmd()
		}
		return m, nil

	case sessionsLoaded:
		m.sessions = msg.sessions
		m.sessionCursor = 0
		if len(m.sessions) > 0 {
			return m, m.loadMessagesCmd()
		}
		m.messages = nil
		m.viewport.SetContent(emptyStyle.Render("No sessions found"))
		return m, nil

	case messagesLoaded:
		m.messages = msg.messages
		m.viewport.SetContent(m.renderConversation())
		m.viewport.GotoTop()
		return m, nil
	}

	// Update viewport if in conversation panel
	if m.focus == panelConversation {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		m.focus = (m.focus + 1) % 3
		return m, nil

	case "shift+tab":
		m.focus = (m.focus + 2) % 3
		return m, nil

	case "up", "k":
		switch m.focus {
		case panelProjects:
			if m.projectCursor > 0 {
				m.projectCursor--
				m.sessionCursor = 0
				return m, m.loadSessionsCmd()
			}
		case panelSessions:
			if m.sessionCursor > 0 {
				m.sessionCursor--
				return m, m.loadMessagesCmd()
			}
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case "down", "j":
		switch m.focus {
		case panelProjects:
			if m.projectCursor < len(m.projects)-1 {
				m.projectCursor++
				m.sessionCursor = 0
				return m, m.loadSessionsCmd()
			}
		case panelSessions:
			if m.sessionCursor < len(m.sessions)-1 {
				m.sessionCursor++
				return m, m.loadMessagesCmd()
			}
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case "enter":
		if m.focus < panelConversation {
			m.focus++
		}
		return m, nil

	case "esc":
		if m.focus > panelProjects {
			m.focus--
		}
		return m, nil

	default:
		if m.focus == panelConversation {
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	// Layout: [Projects | Sessions | Conversation]
	projectsPanel := m.renderProjectsPanel()
	sessionsPanel := m.renderSessionsPanel()
	convoPanel := m.renderConversationPanel()

	main := lipgloss.JoinHorizontal(lipgloss.Top, projectsPanel, sessionsPanel, convoPanel)

	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, main, help)
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
		name := truncateStr(p.Name, w-4)
		if i == m.projectCursor {
			items = append(items, selectedItemStyle.Width(w-4).Render("▸ "+name))
		} else {
			items = append(items, itemStyle.Width(w-4).Render("  "+name))
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
		visibleStart, visibleEnd := m.visibleRange(m.sessionCursor, len(m.sessions), h-2)
		for i := visibleStart; i < visibleEnd; i++ {
			s := m.sessions[i]
			date := s.StartedAt.Format("Jan 02 15:04")
			preview := truncateStr(s.Preview, w-6)
			if preview == "" {
				preview = "(empty session)"
			}
			size := formatSize(s.FileSize)

			if i == m.sessionCursor {
				line1 := selectedItemStyle.Width(w - 4).Render("▸ " + date + "  " + size)
				line2 := selectedItemDescStyle.Width(w - 4).Render("  " + preview)
				items = append(items, line1, line2)
			} else {
				line1 := itemStyle.Width(w - 4).Render("  " + date + "  " + size)
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

func (m Model) renderConversation() string {
	if len(m.messages) == 0 {
		return emptyStyle.Render("\n\n  No messages to display")
	}

	w := m.conversationWidth() - 8
	var parts []string

	for _, msg := range m.messages {
		var rendered string

		ts := timestampStyle.Render(msg.Timestamp.Format("15:04:05"))

		if msg.Type == "user" {
			label := userLabelStyle.Render("  You") + " " + ts
			text := msg.RawText
			if len(text) > 2000 {
				text = text[:2000] + "\n\n... (truncated)"
			}
			body := userBubbleStyle.Width(w).Render(text)
			rendered = label + "\n" + body
		} else {
			label := assistantLabelStyle.Render("  Claude") + " " + ts

			// Show tool badges
			if len(msg.ToolUses) > 0 {
				var badges []string
				seen := make(map[string]bool)
				for _, t := range msg.ToolUses {
					if !seen[t.Name] {
						badges = append(badges, toolBadgeStyle.Render(t.Name))
						seen[t.Name] = true
					}
				}
				if len(badges) > 5 {
					badges = badges[:5]
					badges = append(badges, toolBadgeStyle.Render(fmt.Sprintf("+%d more", len(msg.ToolUses)-5)))
				}
				label += "\n  " + strings.Join(badges, " ")
			}

			text := msg.RawText
			if len(text) > 3000 {
				text = text[:3000] + "\n\n... (truncated)"
			}

			// Try markdown rendering
			mdRendered, err := m.renderer.Render(text)
			if err != nil || strings.TrimSpace(mdRendered) == "" {
				mdRendered = text
			}

			body := assistantBubbleStyle.Width(w).Render(mdRendered)

			// Token info
			if msg.TokensOut > 0 {
				tokens := tokenStyle.Render(fmt.Sprintf("  %s · %d tokens out", msg.Model, msg.TokensOut))
				rendered = label + "\n" + body + "\n" + tokens
			} else {
				rendered = label + "\n" + body
			}
		}

		parts = append(parts, rendered)
	}

	return strings.Join(parts, "\n")
}

func (m Model) renderHelp() string {
	pairs := []struct{ key, desc string }{
		{"tab/shift+tab", "switch panel"},
		{"↑/↓", "navigate"},
		{"enter", "drill in"},
		{"esc", "back"},
		{"q", "quit"},
	}

	var items []string
	for _, p := range pairs {
		items = append(items, helpKeyStyle.Render(p.key)+" "+helpDescStyle.Render(p.desc))
	}

	logo := logoStyle.Render("◈ Claude History")

	bar := lipgloss.JoinHorizontal(lipgloss.Center,
		logo,
		statusBarStyle.Render("  │  "),
		strings.Join(items, statusBarStyle.Render("  ·  ")),
	)

	return statusBarStyle.Width(m.width).Render(bar)
}

// --- Layout math ---

func (m Model) projectsWidth() int {
	return max(20, m.width/5)
}

func (m Model) sessionsWidth() int {
	return max(30, m.width*3/10)
}

func (m Model) conversationWidth() int {
	return m.width - m.projectsWidth() - m.sessionsWidth()
}

func (m Model) contentHeight() int {
	return m.height - 2 // room for help bar
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

// --- Commands ---

type projectsLoaded struct {
	projects []data.Project
}

func loadProjects() tea.Msg {
	projects, err := data.LoadProjects()
	if err != nil {
		return projectsLoaded{}
	}
	return projectsLoaded{projects: projects}
}

func (m Model) loadSessionsCmd() tea.Cmd {
	if m.projectCursor >= len(m.projects) {
		return nil
	}
	p := &m.projects[m.projectCursor]
	return func() tea.Msg {
		sessions, err := data.LoadSessions(p)
		if err != nil {
			return sessionsLoaded{}
		}
		return sessionsLoaded{sessions: sessions}
	}
}

func (m Model) loadMessagesCmd() tea.Cmd {
	if m.sessionCursor >= len(m.sessions) {
		return nil
	}
	s := &m.sessions[m.sessionCursor]
	return func() tea.Msg {
		messages, err := data.LoadMessages(s)
		if err != nil {
			return messagesLoaded{}
		}
		return messagesLoaded{messages: messages}
	}
}

// --- Helpers ---

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
