package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKey dispatches keyboard input to focused handlers.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay intercepts all keys
	if m.showHelp {
		switch msg.String() {
		case "?", "esc", "q":
			m.showHelp = false
		}
		return m, nil
	}

	// Mark input mode (waiting for a-z after m or ')
	if m.awaitingMark != markNone {
		return m.handleMarkKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.showHelp = true
		return m, nil
	}

	// Try action keys, then nav keys, then panel keys
	if model, cmd, handled := m.handleActionKeys(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := m.handleNavKeys(msg); handled {
		return model, cmd
	}
	if model, cmd, handled := m.handlePanelKeys(msg); handled {
		return model, cmd
	}

	// Forward unhandled keys to viewport when in conversation
	if m.focus == panelConversation {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleActionKeys handles global action keybindings (search, theme, copy, marks, etc.).
func (m Model) handleActionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "m":
		m.awaitingMark = markSet
		m.statusMessage = "Set mark: a-z"
		return m, nil, true

	case "'":
		m.awaitingMark = markJump
		m.statusMessage = "Jump to mark: a-z"
		return m, nil, true

	case "/":
		m.searchMode = true
		m.searchInput.Focus()
		m.searchResults = nil
		m.searchCursor = 0
		return m, textinput.Blink, true

	case "t":
		m.themeIndex = (m.themeIndex + 1) % len(themes)
		applyTheme(themes[m.themeIndex])
		m.statusMessage = "Theme: " + themes[m.themeIndex].Name
		m.rebuildRenderer()
		if len(m.messages) > 0 {
			m.updateConversationContent()
		}
		return m, clearStatusAfter(2 * time.Second), true

	case "f":
		m.fullScreen = !m.fullScreen
		if m.fullScreen {
			m.focus = panelConversation
		}
		m.rebuildRenderer()
		m.viewport.Width = m.conversationWidth() - 4
		m.viewport.Height = m.contentHeight() - 3
		if len(m.messages) > 0 {
			m.updateConversationContent()
		}
		return m, nil, true

	case "y":
		if len(m.messages) > 0 {
			return m, m.copyConversationCmd(), true
		}
		return m, nil, true

	case "n":
		if m.focus == panelConversation {
			m.jumpToNextUserMessage(1)
		}
		return m, nil, true

	case "N":
		if m.focus == panelConversation {
			m.jumpToNextUserMessage(-1)
		}
		return m, nil, true

	case " ":
		if m.focus == panelConversation {
			m.toggleCollapsibleAtCursor()
			m.updateConversationContent()
			return m, nil, true
		}

	case "F":
		m.sessionFilter = (m.sessionFilter + 1) % len(sessionFilterTypes)
		m.sessionCursor = 0
		m.statusMessage = "Filter: " + sessionFilterTypes[m.sessionFilter].label
		return m, clearStatusAfter(2 * time.Second), true

	case "a":
		if m.focus == panelConversation {
			m.expandAll()
			m.updateConversationContent()
			return m, nil, true
		}

	case "A":
		if m.focus == panelConversation {
			m.collapseAll()
			m.updateConversationContent()
			return m, nil, true
		}
	}

	return m, nil, false
}

// handleNavKeys handles cursor movement and scrolling within panels.
func (m Model) handleNavKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "up", "k":
		switch m.focus {
		case panelProjects:
			if m.projectCursor > 0 {
				m.projectCursor--
				m.sessionCursor = 0
				return m, m.loadSessionsCmd(), true
			}
		case panelSessions:
			if m.sessionCursor > 0 {
				m.sessionCursor--
				return m, m.loadMessagesWithSpinner(), true
			}
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd, true
		}

	case "down", "j":
		switch m.focus {
		case panelProjects:
			if m.projectCursor < len(m.projects)-1 {
				m.projectCursor++
				m.sessionCursor = 0
				return m, m.loadSessionsCmd(), true
			}
		case panelSessions:
			if m.sessionCursor < len(m.sessions)-1 {
				m.sessionCursor++
				return m, m.loadMessagesWithSpinner(), true
			}
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd, true
		}

	case "g":
		switch m.focus {
		case panelProjects:
			if m.projectCursor != 0 {
				m.projectCursor = 0
				m.sessionCursor = 0
				return m, m.loadSessionsCmd(), true
			}
		case panelSessions:
			if m.sessionCursor != 0 {
				m.sessionCursor = 0
				return m, m.loadMessagesWithSpinner(), true
			}
		case panelConversation:
			m.viewport.GotoTop()
		}
		return m, nil, true

	case "G":
		switch m.focus {
		case panelProjects:
			last := len(m.projects) - 1
			if last >= 0 && m.projectCursor != last {
				m.projectCursor = last
				m.sessionCursor = 0
				return m, m.loadSessionsCmd(), true
			}
		case panelSessions:
			last := len(m.sessions) - 1
			if last >= 0 && m.sessionCursor != last {
				m.sessionCursor = last
				return m, m.loadMessagesWithSpinner(), true
			}
		case panelConversation:
			m.viewport.GotoBottom()
		}
		return m, nil, true

	case "pgup":
		switch m.focus {
		case panelProjects:
			m.projectCursor = clamp(m.projectCursor-m.contentHeight()/2, 0, max(0, len(m.projects)-1))
			m.sessionCursor = 0
			return m, m.loadSessionsCmd(), true
		case panelSessions:
			m.sessionCursor = clamp(m.sessionCursor-m.contentHeight()/2, 0, max(0, len(m.sessions)-1))
			return m, m.loadMessagesWithSpinner(), true
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd, true
		}

	case "pgdown":
		switch m.focus {
		case panelProjects:
			m.projectCursor = clamp(m.projectCursor+m.contentHeight()/2, 0, max(0, len(m.projects)-1))
			m.sessionCursor = 0
			return m, m.loadSessionsCmd(), true
		case panelSessions:
			m.sessionCursor = clamp(m.sessionCursor+m.contentHeight()/2, 0, max(0, len(m.sessions)-1))
			return m, m.loadMessagesWithSpinner(), true
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd, true
		}
	}

	return m, nil, false
}

// handlePanelKeys handles panel focus switching and sliding.
func (m Model) handlePanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		if m.fullScreen {
			return m, nil, true
		}
		switch m.visiblePanelCount() {
		case 1:
			// Single panel: tab does nothing
		case 2:
			if m.focus == panelProjects {
				m.focus = panelSessions
			} else if m.focus == panelSessions {
				if m.width < breakpointMedium && m.isShowingProjectsSessions() {
					m.focus = panelProjects
				} else {
					m.focus = panelConversation
				}
			} else {
				m.focus = panelSessions
			}
		default:
			m.focus = (m.focus + 1) % 3
		}
		return m, nil, true

	case "shift+tab":
		if m.fullScreen {
			return m, nil, true
		}
		switch m.visiblePanelCount() {
		case 1:
			// Single panel: shift+tab does nothing
		case 2:
			if m.focus == panelConversation {
				m.focus = panelSessions
			} else if m.focus == panelSessions {
				if m.width < breakpointMedium && !m.isShowingProjectsSessions() {
					m.focus = panelConversation
				} else {
					m.focus = panelProjects
				}
			} else {
				m.focus = panelSessions
			}
		default:
			m.focus = (m.focus + 2) % 3
		}
		return m, nil, true

	case "enter":
		if m.focus < panelConversation {
			m.focus++
			m.rebuildRendererIfNeeded()
		}
		return m, nil, true

	case "esc":
		if m.fullScreen {
			m.fullScreen = false
			m.rebuildRenderer()
			m.viewport.Width = m.conversationWidth() - 4
			m.viewport.Height = m.contentHeight() - 3
			if len(m.messages) > 0 {
				m.updateConversationContent()
			}
			return m, nil, true
		}
		if m.focus > panelProjects {
			m.focus--
			m.rebuildRendererIfNeeded()
		}
		return m, nil, true
	}

	return m, nil, false
}

// --- Navigation helpers ---

// jumpToNextUserMessage scrolls viewport to the next (dir=1) or previous (dir=-1) user message.
func (m *Model) jumpToNextUserMessage(dir int) {
	if len(m.userMessageLines) == 0 {
		return
	}

	currentLine := m.viewport.YOffset

	if dir > 0 {
		for _, line := range m.userMessageLines {
			if line > currentLine+1 {
				m.viewport.SetYOffset(line)
				return
			}
		}
	} else {
		for i := len(m.userMessageLines) - 1; i >= 0; i-- {
			if m.userMessageLines[i] < currentLine-1 {
				m.viewport.SetYOffset(m.userMessageLines[i])
				return
			}
		}
	}
}

// toggleCollapsibleAtCursor toggles the collapsible section nearest to the viewport position.
func (m *Model) toggleCollapsibleAtCursor() {
	key := m.nearestCollapsibleKey()
	if key != "" {
		m.collapsed[key] = !m.isCollapsed(key)
	}
}

// nearestCollapsibleKey returns the key of the collapsible section closest to the
// viewport center. Only considers sections currently visible in the viewport.
func (m *Model) nearestCollapsibleKey() string {
	if len(m.collapsibleLines) == 0 {
		return ""
	}

	viewTop := m.viewport.YOffset
	viewBottom := viewTop + m.viewport.Height
	viewCenter := viewTop + m.viewport.Height/2

	bestKey := ""
	bestDist := int(^uint(0) >> 1)

	for key, line := range m.collapsibleLines {
		// Only consider sections visible in the viewport
		if line < viewTop-5 || line > viewBottom+5 {
			continue
		}
		dist := viewCenter - line
		if dist < 0 {
			dist = -dist
		}
		if dist < bestDist {
			bestDist = dist
			bestKey = key
		}
	}

	return bestKey
}

// expandAll expands all collapsible sections in the current conversation.
func (m *Model) expandAll() {
	for k := range m.collapsed {
		m.collapsed[k] = false
	}
}

// collapseAll collapses all collapsible sections in the current conversation.
func (m *Model) collapseAll() {
	for _, msg := range m.messages {
		for _, block := range msg.ContentBlocks {
			switch block.Type {
			case "thinking":
				m.collapsed["thinking:"+msg.UUID] = true
			case "tool_use":
				m.collapsed["tool:"+block.ToolID] = true
			}
		}
	}
}

// --- Layout helpers ---

// visiblePanelCount returns how many panels are shown at the current width.
func (m Model) visiblePanelCount() int {
	if m.fullScreen || m.width < breakpointNarrow {
		return 1
	}
	if m.width < breakpointMedium {
		return 2
	}
	return 3
}

// isShowingProjectsSessions returns true if the 2-panel view is showing
// projects+sessions (as opposed to sessions+conversation).
func (m Model) isShowingProjectsSessions() bool {
	return m.focus == panelProjects
}

// rebuildRendererIfNeeded rebuilds the glamour renderer when the conversation
// panel width has changed (e.g. because focus shifted which panels are visible).
func (m *Model) rebuildRendererIfNeeded() {
	if m.visiblePanelCount() < 3 {
		m.rebuildRenderer()
		m.viewport.Width = m.conversationWidth() - 4
		m.viewport.Height = m.contentHeight() - 3
		if len(m.messages) > 0 {
			m.updateConversationContent()
		}
	}
}

// triggerTransition sets a brief highlight effect on the newly focused panel.
func (m *Model) triggerTransition() tea.Cmd {
	m.transitionUntil = time.Now().Add(time.Duration(transitionDurationMs) * time.Millisecond)
	return tea.Tick(time.Duration(transitionDurationMs)*time.Millisecond, func(time.Time) tea.Msg {
		return transitionDoneMsg{}
	})
}

// loadMessagesWithSpinner saves scroll position, sets loading state, and loads messages.
func (m *Model) loadMessagesWithSpinner() tea.Cmd {
	if m.sessionCursor < len(m.sessions) {
		m.scrollPositions[m.sessions[m.sessionCursor].ID] = m.viewport.YOffset
	}
	m.loading = true
	return tea.Batch(m.loadMessagesCmd(), m.spinner.Tick)
}
