package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleConvSearchKey handles keyboard input during in-conversation search.
func (m Model) handleConvSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.convSearchMode = false
		m.convSearchInput.Blur()
		m.convSearchInput.SetValue("")
		m.convSearchMatches = nil
		return m, nil

	case "enter", "ctrl+n":
		// Jump to next match
		if len(m.convSearchMatches) > 0 {
			m.convSearchIdx = (m.convSearchIdx + 1) % len(m.convSearchMatches)
			m.viewport.SetYOffset(m.convSearchMatches[m.convSearchIdx])
		}
		return m, nil

	case "ctrl+p":
		// Jump to previous match
		if len(m.convSearchMatches) > 0 {
			m.convSearchIdx--
			if m.convSearchIdx < 0 {
				m.convSearchIdx = len(m.convSearchMatches) - 1
			}
			m.viewport.SetYOffset(m.convSearchMatches[m.convSearchIdx])
		}
		return m, nil
	}

	// Update text input
	var cmd tea.Cmd
	m.convSearchInput, cmd = m.convSearchInput.Update(msg)

	// Recompute matches on input change
	query := strings.ToLower(m.convSearchInput.Value())
	m.convSearchMatches = nil
	m.convSearchIdx = 0

	if query != "" {
		lines := strings.Split(m.viewport.View(), "\n")
		// Search in the full content, not just visible lines
		contentLines := strings.Split(m.viewportContent(), "\n")
		for i, line := range contentLines {
			if strings.Contains(strings.ToLower(line), query) {
				m.convSearchMatches = append(m.convSearchMatches, i)
			}
		}
		_ = lines // viewport lines used for display only
		// Jump to first match
		if len(m.convSearchMatches) > 0 {
			m.viewport.SetYOffset(m.convSearchMatches[0])
		}
	}

	return m, cmd
}

// viewportContent returns the full content string set on the viewport.
// We store it separately since viewport.View() only returns visible lines.
func (m Model) viewportContent() string {
	// Re-render to get full content (this is only called during search input changes)
	result := m.renderConversation()
	return result.content
}
