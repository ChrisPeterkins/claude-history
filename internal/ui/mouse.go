package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	x := msg.X
	y := msg.Y

	// Determine which panel was clicked
	panel := m.panelAtX(x)

	switch {
	case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
		return m.handleMouseClick(panel, x, y)

	case msg.Button == tea.MouseButtonWheelUp:
		return m.handleMouseScroll(panel, -1)

	case msg.Button == tea.MouseButtonWheelDown:
		return m.handleMouseScroll(panel, 1)
	}

	return m, nil
}

func (m Model) panelAtX(x int) int {
	if m.fullScreen || m.width < 60 {
		return m.focus
	}
	if m.width < 100 {
		// Two-panel mode
		if m.focus == panelProjects {
			if x < m.projectsWidth() {
				return panelProjects
			}
			return panelSessions
		}
		if x < m.sessionsWidth() {
			return panelSessions
		}
		return panelConversation
	}
	// Three-panel mode
	if x < m.projectsWidth() {
		return panelProjects
	}
	if x < m.projectsWidth()+m.sessionsWidth() {
		return panelSessions
	}
	return panelConversation
}

func (m Model) handleMouseClick(panel, x, y int) (tea.Model, tea.Cmd) {
	m.focus = panel

	switch panel {
	case panelProjects:
		// Map y to project index (border=1, title=1, then items)
		itemY := y - 2
		if itemY < 0 {
			return m, nil
		}
		visibleStart, _ := m.visibleRange(m.projectCursor, len(m.projects), m.contentHeight()-2)
		idx := visibleStart + itemY
		if idx >= 0 && idx < len(m.projects) && idx != m.projectCursor {
			m.projectCursor = idx
			m.sessionCursor = 0
			return m, m.loadSessionsCmd()
		}

	case panelSessions:
		// Sessions have date headers (1 line) and items (2 lines each) — approximate
		itemY := y - 2
		if itemY < 0 {
			return m, nil
		}
		// Simple approximation: each session takes ~2 lines
		visibleStart, _ := m.visibleRange(m.sessionCursor, len(m.sessions), m.contentHeight()-2)
		idx := visibleStart + itemY/2
		if idx >= 0 && idx < len(m.sessions) && idx != m.sessionCursor {
			m.sessionCursor = idx
			return m, m.loadMessagesWithSpinner()
		}

	case panelConversation:
		// Clicks in conversation just focus the panel (scrolling handled by wheel)
	}

	return m, nil
}

func (m Model) handleMouseScroll(panel, dir int) (tea.Model, tea.Cmd) {
	switch panel {
	case panelProjects:
		newCursor := m.projectCursor + dir
		if newCursor >= 0 && newCursor < len(m.projects) {
			m.projectCursor = newCursor
			m.sessionCursor = 0
			return m, m.loadSessionsCmd()
		}

	case panelSessions:
		newCursor := m.sessionCursor + dir
		if newCursor >= 0 && newCursor < len(m.sessions) {
			m.sessionCursor = newCursor
			return m, m.loadMessagesWithSpinner()
		}

	case panelConversation:
		if dir < 0 {
			m.viewport.LineUp(3)
		} else {
			m.viewport.LineDown(3)
		}
	}

	return m, nil
}
