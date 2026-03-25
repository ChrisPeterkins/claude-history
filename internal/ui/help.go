package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelpOverlay renders a full keybinding reference as a centered modal.
func (m Model) renderHelpOverlay() string {
	sections := []struct {
		title string
		keys  []struct{ key, desc string }
	}{
		{
			title: "Navigation",
			keys: []struct{ key, desc string }{
				{"↑/↓ or j/k", "Move cursor / scroll"},
				{"g", "Jump to top"},
				{"G", "Jump to bottom"},
				{"PgUp/PgDn", "Page up / page down"},
				{"n/N", "Next / previous user message"},
			},
		},
		{
			title: "Panels",
			keys: []struct{ key, desc string }{
				{"Tab", "Next panel"},
				{"Shift+Tab", "Previous panel"},
				{"Enter", "Drill into next panel"},
				{"Esc", "Go back a panel"},
				{"f", "Toggle full-screen"},
			},
		},
		{
			title: "Conversation",
			keys: []struct{ key, desc string }{
				{"Space", "Expand/collapse tool calls"},
				{"m + a-z", "Set a bookmark"},
				{"' + a-z", "Jump to bookmark"},
			},
		},
		{
			title: "Search & Export",
			keys: []struct{ key, desc string }{
				{"/", "Search conversations"},
				{"y", "Copy conversation to clipboard"},
			},
		},
		{
			title: "Other",
			keys: []struct{ key, desc string }{
				{"t", "Cycle theme"},
				{"?", "Toggle this help"},
				{"q", "Quit"},
			},
		},
	}

	var lines []string
	lines = append(lines, helpOverlayTitleStyle.Render("Keyboard Shortcuts"))
	lines = append(lines, "")

	for _, section := range sections {
		lines = append(lines, helpSectionStyle.Render(section.title))
		for _, k := range section.keys {
			line := "  " + helpKeyStyle.Render(fmt.Sprintf("%-14s", k.key)) + " " + helpDescStyle.Render(k.desc)
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	lines = append(lines, timestampStyle.Render("Press ? or Esc to close"))

	content := strings.Join(lines, "\n")

	overlayW := 44
	overlayH := len(lines) + 2
	if overlayW > m.width-4 {
		overlayW = m.width - 4
	}

	box := helpOverlayStyle.
		Width(overlayW).
		Height(overlayH).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// renderHelp renders the bottom status/help bar.
func (m Model) renderHelp() string {
	allPairs := []struct{ key, desc string }{
		{"tab", "panel"},
		{"↑/↓", "navigate"},
		{"f", "full"}, {"/", "search"},
		{"space", "expand"},
		{"n/N", "jump"},
		{"y", "copy"},
		{"t", "theme"},
		{"?", "help"},
		{"q", "quit"},
	}

	// Show fewer keybindings on narrow terminals
	pairs := allPairs
	if m.width < 100 {
		pairs = []struct{ key, desc string }{
			{"f", "full"}, {"/", "search"}, {"?", "help"}, {"q", "quit"},
		}
	} else if m.width < 140 {
		pairs = []struct{ key, desc string }{
			{"tab", "panel"}, {"f", "full"}, {"/", "search"},
			{"space", "expand"}, {"?", "help"}, {"q", "quit"},
		}
	}

	var items []string
	for _, p := range pairs {
		items = append(items, helpKeyStyle.Render(p.key)+" "+helpDescStyle.Render(p.desc))
	}

	logo := logoStyle.Render("◈ Claude History")

	// Breadcrumb
	breadcrumb := m.renderBreadcrumb()

	left := logo
	if breadcrumb != "" && m.width >= 80 {
		left += statusBarStyle.Render("  │  ") + breadcrumb
	}

	// Status flash message
	if m.statusMessage != "" {
		left += statusBarStyle.Render("  ") + helpKeyStyle.Render(m.statusMessage)
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Center,
		left,
		statusBarStyle.Render("  │  "),
		strings.Join(items, statusBarStyle.Render("  ·  ")),
	)

	return statusBarStyle.Width(m.width).Render(bar)
}

// renderBreadcrumb shows current location: Project › Date › Msg N/M.
func (m Model) renderBreadcrumb() string {
	var parts []string

	if m.projectCursor < len(m.projects) {
		parts = append(parts, m.projects[m.projectCursor].Name)
	}

	if m.sessionCursor < len(m.sessions) {
		s := m.sessions[m.sessionCursor]
		parts = append(parts, s.StartedAt.Format("Jan 02"))
	}

	if m.focus == panelConversation && len(m.userMessageLines) > 0 {
		currentLine := m.viewport.YOffset
		msgIdx := 0
		for i, line := range m.userMessageLines {
			if line <= currentLine {
				msgIdx = i + 1
			}
		}
		if msgIdx > 0 {
			parts = append(parts, fmt.Sprintf("Msg %d/%d", msgIdx, len(m.userMessageLines)))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return timestampStyle.Render(strings.Join(parts, " › "))
}
