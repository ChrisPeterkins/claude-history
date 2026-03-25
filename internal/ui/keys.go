package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help overlay intercepts all keys
	if m.showHelp {
		switch msg.String() {
		case "?", "esc", "q":
			m.showHelp = false
		}
		return m, nil
	}

	// Handle mark input mode (waiting for a-z after m or ')
	if m.awaitingMark != "" {
		return m.handleMarkKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "?":
		m.showHelp = true
		return m, nil

	case "m":
		m.awaitingMark = "set"
		m.statusMessage = "Set mark: a-z"
		return m, nil

	case "'":
		m.awaitingMark = "jump"
		m.statusMessage = "Jump to mark: a-z"
		return m, nil

	case "/":
		m.searchMode = true
		m.searchInput.Focus()
		m.searchResults = nil
		m.searchCursor = 0
		return m, textinput.Blink

	case "t":
		m.themeIndex = (m.themeIndex + 1) % len(themes)
		applyTheme(themes[m.themeIndex])
		m.statusMessage = "Theme: " + themes[m.themeIndex].Name
		m.rebuildRenderer()
		if len(m.messages) > 0 {
			m.viewport.SetContent(m.renderConversation())
		}
		return m, clearStatusAfter(2 * time.Second)

	case "f":
		m.fullScreen = !m.fullScreen
		if m.fullScreen {
			m.focus = panelConversation
		}
		// Rebuild renderer for new width, resize viewport
		m.rebuildRenderer()
		m.viewport.Width = m.conversationWidth() - 4
		m.viewport.Height = m.contentHeight() - 3
		if len(m.messages) > 0 {
			m.viewport.SetContent(m.renderConversation())
		}
		return m, nil

	case "tab":
		if m.fullScreen {
			return m, nil
		}
		switch m.visiblePanelCount() {
		case 1:
			// Single panel: tab does nothing, use enter/esc to slide
		case 2:
			// Two panels: tab toggles between the two visible panels
			if m.focus == panelProjects {
				m.focus = panelSessions
			} else if m.focus == panelSessions {
				// Determine which pair is visible
				if m.width < 100 && m.isShowingProjectsSessions() {
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
		return m, nil

	case "shift+tab":
		if m.fullScreen {
			return m, nil
		}
		switch m.visiblePanelCount() {
		case 1:
			// Single panel: shift+tab does nothing
		case 2:
			if m.focus == panelConversation {
				m.focus = panelSessions
			} else if m.focus == panelSessions {
				if m.width < 100 && !m.isShowingProjectsSessions() {
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
				return m, m.loadMessagesWithSpinner()
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
				return m, m.loadMessagesWithSpinner()
			}
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case "y":
		// Copy conversation to clipboard
		if len(m.messages) > 0 {
			return m, m.copyConversationCmd()
		}
		return m, nil

	case "g":
		// Jump to top
		switch m.focus {
		case panelProjects:
			if m.projectCursor != 0 {
				m.projectCursor = 0
				m.sessionCursor = 0
				return m, m.loadSessionsCmd()
			}
		case panelSessions:
			if m.sessionCursor != 0 {
				m.sessionCursor = 0
				return m, m.loadMessagesWithSpinner()
			}
		case panelConversation:
			m.viewport.GotoTop()
		}
		return m, nil

	case "G":
		// Jump to bottom
		switch m.focus {
		case panelProjects:
			last := len(m.projects) - 1
			if last >= 0 && m.projectCursor != last {
				m.projectCursor = last
				m.sessionCursor = 0
				return m, m.loadSessionsCmd()
			}
		case panelSessions:
			last := len(m.sessions) - 1
			if last >= 0 && m.sessionCursor != last {
				m.sessionCursor = last
				return m, m.loadMessagesWithSpinner()
			}
		case panelConversation:
			m.viewport.GotoBottom()
		}
		return m, nil

	case "pgup":
		switch m.focus {
		case panelProjects:
			m.projectCursor = clamp(m.projectCursor-m.contentHeight()/2, 0, max(0, len(m.projects)-1))
			m.sessionCursor = 0
			return m, m.loadSessionsCmd()
		case panelSessions:
			m.sessionCursor = clamp(m.sessionCursor-m.contentHeight()/2, 0, max(0, len(m.sessions)-1))
			return m, m.loadMessagesWithSpinner()
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case "pgdown":
		switch m.focus {
		case panelProjects:
			m.projectCursor = clamp(m.projectCursor+m.contentHeight()/2, 0, max(0, len(m.projects)-1))
			m.sessionCursor = 0
			return m, m.loadSessionsCmd()
		case panelSessions:
			m.sessionCursor = clamp(m.sessionCursor+m.contentHeight()/2, 0, max(0, len(m.sessions)-1))
			return m, m.loadMessagesWithSpinner()
		case panelConversation:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case "n":
		// Jump to next user message
		if m.focus == panelConversation {
			m.jumpToNextUserMessage(1)
		}
		return m, nil

	case "N":
		// Jump to previous user message
		if m.focus == panelConversation {
			m.jumpToNextUserMessage(-1)
		}
		return m, nil

	case " ":
		// Toggle collapsible section (space bar)
		if m.focus == panelConversation {
			m.toggleCollapsibleAtCursor()
			m.viewport.SetContent(m.renderConversation())
			return m, nil
		}

	case "enter":
		// Enter slides focus forward (and shifts visible panels in constrained layouts)
		if m.focus < panelConversation {
			m.focus++
			m.rebuildRendererIfNeeded()
		}
		return m, nil

	case "esc":
		if m.fullScreen {
			m.fullScreen = false
			m.rebuildRenderer()
			m.viewport.Width = m.conversationWidth() - 4
			m.viewport.Height = m.contentHeight() - 3
			if len(m.messages) > 0 {
				m.viewport.SetContent(m.renderConversation())
			}
			return m, nil
		}
		// Esc slides focus backward (and shifts visible panels in constrained layouts)
		if m.focus > panelProjects {
			m.focus--
			m.rebuildRendererIfNeeded()
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

// jumpToNextUserMessage scrolls viewport to the next (dir=1) or previous (dir=-1) user message.
func (m *Model) jumpToNextUserMessage(dir int) {
	if len(m.userMessageLines) == 0 {
		return
	}

	currentLine := m.viewport.YOffset

	if dir > 0 {
		// Find next user message line after current position
		for _, line := range m.userMessageLines {
			if line > currentLine+1 {
				m.viewport.SetYOffset(line)
				return
			}
		}
	} else {
		// Find previous user message line before current position
		for i := len(m.userMessageLines) - 1; i >= 0; i-- {
			if m.userMessageLines[i] < currentLine-1 {
				m.viewport.SetYOffset(m.userMessageLines[i])
				return
			}
		}
	}
}

// toggleCollapsibleAtCursor toggles the collapsible section nearest to the viewport cursor.
func (m *Model) toggleCollapsibleAtCursor() {
	// Find collapsible keys near the current viewport position
	// We use a simple heuristic: look for tool/thinking blocks in messages
	// and toggle the first one that's approximately at the viewport position
	currentLine := m.viewport.YOffset + m.viewport.Height/3

	lineCount := 0
	for _, msg := range m.messages {
		if msg.Type != "assistant" {
			// Rough estimate: user messages are ~3 lines
			lineCount += 3
			continue
		}

		for _, block := range msg.ContentBlocks {
			var key string
			switch block.Type {
			case "thinking":
				key = "thinking:" + msg.UUID
			case "tool_use":
				key = "tool:" + block.ToolID
			default:
				lineCount += 2
				continue
			}

			if key != "" && lineCount >= currentLine-5 && lineCount <= currentLine+5 {
				// Toggle this section
				m.collapsed[key] = !m.isCollapsed(key)
				return
			}
			lineCount += 3
		}
		lineCount += 2
	}

	// Fallback: if we couldn't find a specific one, toggle the nearest tool
	// just expand all or collapse all
	allExpanded := true
	for _, v := range m.collapsed {
		if v {
			allExpanded = false
			break
		}
	}

	if allExpanded || len(m.collapsed) == 0 {
		// Collapse all
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
	} else {
		// Expand all
		for k := range m.collapsed {
			m.collapsed[k] = false
		}
	}
}

// visiblePanelCount returns how many panels are shown at the current width.
func (m Model) visiblePanelCount() int {
	if m.fullScreen || m.width < 60 {
		return 1
	}
	if m.width < 100 {
		return 2
	}
	return 3
}

// isShowingProjectsSessions returns true if the 2-panel view is showing projects+sessions
// (as opposed to sessions+conversation).
func (m Model) isShowingProjectsSessions() bool {
	return m.focus == panelProjects
}

// rebuildRendererIfNeeded rebuilds the glamour renderer when the conversation panel
// width has changed (e.g. because focus shifted which panels are visible).
func (m *Model) rebuildRendererIfNeeded() {
	if m.visiblePanelCount() < 3 {
		m.rebuildRenderer()
		m.viewport.Width = m.conversationWidth() - 4
		m.viewport.Height = m.contentHeight() - 3
		if len(m.messages) > 0 {
			m.viewport.SetContent(m.renderConversation())
		}
	}
}

// triggerTransition sets a brief highlight effect on the newly focused panel.
func (m *Model) triggerTransition() tea.Cmd {
	m.transitionUntil = time.Now().Add(150 * time.Millisecond)
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
		return transitionDoneMsg{}
	})
}

// loadMessagesWithSpinner saves scroll position, sets loading state, and loads messages.
func (m *Model) loadMessagesWithSpinner() tea.Cmd {
	// Save current scroll position before switching
	if m.sessionCursor < len(m.sessions) {
		m.scrollPositions[m.sessions[m.sessionCursor].ID] = m.viewport.YOffset
	}
	m.loading = true
	return tea.Batch(m.loadMessagesCmd(), m.spinner.Tick)
}

// ensure textinput import is used
var _ textinput.Model
