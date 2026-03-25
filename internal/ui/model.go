package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
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
	fullScreen    bool

	// Dimensions
	width  int
	height int

	// Markdown renderer
	renderer *glamour.TermRenderer

	// Collapsible sections: key -> collapsed (true = collapsed)
	collapsed map[string]bool

	// Message jumping: line numbers where user messages start
	userMessageLines []int

	// Search
	searchMode    bool
	searchInput   textinput.Model
	searchResults []SearchResult
	searchCursor  int

	// Status flash message
	statusMessage string
	statusExpiry  time.Time

	// Theme
	themeIndex int
}

// SearchResult represents a match from search.
type SearchResult struct {
	ProjectIdx int
	SessionIdx int
	Preview    string
	Project    string
	Date       string
}

func NewModel() Model {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	ti := textinput.New()
	ti.Placeholder = "Search conversations..."
	ti.CharLimit = 100

	return Model{
		renderer:  r,
		collapsed: make(map[string]bool),
		searchInput: ti,
	}
}

// rebuildRenderer creates a new glamour renderer sized to the current conversation width.
func (m *Model) rebuildRenderer() {
	wrapWidth := m.conversationWidth() - 12
	if wrapWidth < 40 {
		wrapWidth = 40
	}

	style := "dark"
	if m.themeIndex < len(themes) && themes[m.themeIndex].Name == "Light" {
		style = "light"
	}

	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(wrapWidth),
	)
}

func (m Model) Init() tea.Cmd {
	return loadProjects
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.rebuildRenderer()
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
		m.collapsed = make(map[string]bool) // reset collapsed state for new session
		m.viewport.SetContent(m.renderConversation())
		m.viewport.GotoTop()
		return m, nil

	case searchResultsMsg:
		m.searchResults = msg.results
		m.searchCursor = 0
		return m, nil

	case clipboardCopiedMsg:
		if msg.err != nil {
			m.statusMessage = "Copy failed: " + msg.err.Error()
		} else {
			m.statusMessage = "Copied to clipboard!"
		}
		return m, clearStatusAfter(2 * time.Second)

	case statusClearMsg:
		m.statusMessage = ""
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

func (m Model) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	if m.searchMode {
		return m.renderSearchView()
	}

	var main string
	if m.fullScreen || m.width < 60 {
		// Full-screen or very narrow: only conversation panel
		main = m.renderConversationPanel()
	} else if m.width < 100 {
		// Medium width: hide projects, show sessions + conversation
		sessionsPanel := m.renderSessionsPanel()
		convoPanel := m.renderConversationPanel()
		main = lipgloss.JoinHorizontal(lipgloss.Top, sessionsPanel, convoPanel)
	} else {
		// Wide: full three-panel layout
		projectsPanel := m.renderProjectsPanel()
		sessionsPanel := m.renderSessionsPanel()
		convoPanel := m.renderConversationPanel()
		main = lipgloss.JoinHorizontal(lipgloss.Top, projectsPanel, sessionsPanel, convoPanel)
	}

	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, main, help)
}

// statusClearMsg clears the flash message.
type statusClearMsg struct{}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}
