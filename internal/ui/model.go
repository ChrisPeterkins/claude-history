package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/chrispeterkins/claude-history/internal/config"
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
	collapsed        map[string]bool
	collapsibleLines map[string]int // key → line number (populated during render)

	// Help overlay
	showHelp bool

	// Loading spinner
	spinner spinner.Model
	loading bool

	// Scroll position memory (sessionID → YOffset)
	scrollPositions map[string]int

	// Message jumping: line numbers where user messages start
	userMessageLines []int

	// Search
	searchMode    bool
	searchInput   textinput.Model
	searchResults []SearchResult
	searchCursor  int

	// Vim marks
	marks            map[rune]markPosition
	awaitingMark     markMode
	pendingMarkOffset int // offset to restore after cross-session mark jump, -1 = none

	// In-conversation search
	convSearchMode    bool
	convSearchInput   textinput.Model
	convSearchMatches []int    // line numbers with matches
	convSearchContent []string // cached content lines (set on search entry)
	convSearchIdx     int   // current match index

	// Session filter
	sessionFilter int // index into sessionFilterTypes

	// Status flash message
	statusMessage string
	statusExpiry  time.Time

	// Transition effect
	transitionUntil time.Time

	// Version
	version       string
	updateAvail   string // non-empty if a newer version exists

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

// NewModel creates and returns an initialized Model with default settings.
func NewModel(version string) Model {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	ti := textinput.New()
	ti.Placeholder = "Search conversations..."
	ti.CharLimit = searchCharLimit

	csi := textinput.New()
	csi.Placeholder = "Find in conversation..."
	csi.CharLimit = searchCharLimit

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0"))

	// Load saved theme preference
	themeIdx := 0
	cfg := config.Get()
	if strings.EqualFold(cfg.Theme, "custom") && cfg.CustomTheme != nil {
		// Apply custom theme from config
		custom := buildCustomTheme(cfg.CustomTheme)
		themes = append(themes, custom)
		themeIdx = len(themes) - 1
		applyTheme(custom)
	} else if cfg.Theme != "" {
		for i, t := range themes {
			if strings.EqualFold(t.Name, cfg.Theme) {
				themeIdx = i
				applyTheme(themes[i])
				break
			}
		}
	}

	// Load saved filter preference
	filterIdx := 0
	filterName := config.DefaultFilterName()
	for i, ft := range sessionFilterTypes {
		if strings.EqualFold(ft.name, filterName) {
			filterIdx = i
			break
		}
	}

	return Model{
		renderer:          r,
		collapsed:         make(map[string]bool),
		searchInput:       ti,
		spinner:           s,
		scrollPositions:   make(map[string]int),
		marks:             make(map[rune]markPosition),
		convSearchInput:   csi,
		pendingMarkOffset: -1,
		themeIndex:        themeIdx,
		sessionFilter:     filterIdx,
		version:           version,
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

// updateConversationContent re-renders the conversation and updates the viewport and line tracking.
func (m *Model) updateConversationContent() {
	result := m.renderConversation()
	m.viewport.SetContent(result.content)
	m.userMessageLines = result.userLines
	m.collapsibleLines = result.collapsibleLines
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadProjects, checkForUpdate(m.version))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleKey(msg)

	case tea.MouseMsg:
		if !m.searchMode && !m.showHelp {
			return m.handleMouse(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.rebuildRenderer()
		m.viewport = viewport.New(m.conversationWidth(), m.contentHeight())
		m.viewport.Style = lipgloss.NewStyle()
		if len(m.messages) > 0 {
			m.updateConversationContent()
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
			m.loading = true
			return m, tea.Batch(m.loadMessagesCmd(), m.spinner.Tick)
		}
		m.messages = nil
		m.viewport.SetContent(emptyStyle.Render("No sessions found"))
		return m, nil

	case messagesLoaded:
		m.messages = msg.messages
		m.loading = false
		m.collapsed = make(map[string]bool)
		m.updateConversationContent()
		// Pending mark jump takes priority
		if m.pendingMarkOffset >= 0 {
			m.viewport.SetYOffset(m.pendingMarkOffset)
			m.pendingMarkOffset = -1
		} else if m.sessionCursor < len(m.sessions) {
			// Restore scroll position if we've been here before
			if offset, ok := m.scrollPositions[m.sessions[m.sessionCursor].ID]; ok {
				m.viewport.SetYOffset(offset)
			} else {
				m.viewport.GotoTop()
			}
		} else {
			m.viewport.GotoTop()
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case updateAvailableMsg:
		m.updateAvail = msg.version
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

	case transitionDoneMsg:
		// Forces a re-render to clear the transition highlight
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

	if m.showHelp {
		return m.renderHelpOverlay()
	}

	if m.searchMode {
		return m.renderSearchView()
	}

	var main string
	if m.fullScreen || m.width < breakpointNarrow {
		// Full-screen or very narrow: show whichever panel is focused
		switch m.focus {
		case panelProjects:
			main = m.renderProjectsPanel()
		case panelSessions:
			main = m.renderSessionsPanel()
		default:
			main = m.renderConversationPanel()
		}
	} else if m.width < breakpointMedium {
		// Medium width: show 2 panels — the focused one and its neighbor
		switch m.focus {
		case panelProjects:
			main = lipgloss.JoinHorizontal(lipgloss.Top,
				m.renderProjectsPanel(), m.renderSessionsPanel())
		default:
			main = lipgloss.JoinHorizontal(lipgloss.Top,
				m.renderSessionsPanel(), m.renderConversationPanel())
		}
	} else {
		// Wide: full three-panel layout
		projectsPanel := m.renderProjectsPanel()
		sessionsPanel := m.renderSessionsPanel()
		convoPanel := m.renderConversationPanel()
		main = lipgloss.JoinHorizontal(lipgloss.Top, projectsPanel, sessionsPanel, convoPanel)
	}

	header := m.renderHeader()
	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left, header, main, help)
}

// statusClearMsg clears the flash message.
type statusClearMsg struct{}

// transitionDoneMsg signals that a panel transition animation has completed.
type transitionDoneMsg struct{}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}
