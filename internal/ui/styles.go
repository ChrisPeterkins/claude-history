package ui

import "github.com/charmbracelet/lipgloss"

// Color palette — set by applyTheme(), default to Nord.
var (
	colorPrimary    lipgloss.Color
	colorSecondary  lipgloss.Color
	colorAccent     lipgloss.Color
	colorWarm       lipgloss.Color
	colorSubtle     lipgloss.Color
	colorFg         lipgloss.Color
	colorFgDim      lipgloss.Color
	colorBg         lipgloss.Color
	colorBgSelected lipgloss.Color
	colorBorder     lipgloss.Color
	colorRed        lipgloss.Color
	colorGreen      lipgloss.Color
)

// All styles are built by rebuildStyles() in theme.go.
// Declare as package-level vars so they're accessible everywhere.
var (
	// Panels
	panelStyle            lipgloss.Style
	activePanelStyle      lipgloss.Style
	panelTitleStyle       lipgloss.Style
	panelTitleActiveStyle lipgloss.Style

	// List items (focused panel)
	itemStyle             lipgloss.Style
	selectedItemStyle     lipgloss.Style
	itemDescStyle         lipgloss.Style
	selectedItemDescStyle lipgloss.Style

	// List items (dimmed for unfocused panels)
	dimItemStyle             lipgloss.Style
	dimSelectedItemStyle     lipgloss.Style
	dimItemDescStyle         lipgloss.Style
	dimSelectedItemDescStyle lipgloss.Style

	// Conversation
	userBubbleStyle      lipgloss.Style
	userLabelStyle       lipgloss.Style
	assistantLabelStyle  lipgloss.Style
	assistantBubbleStyle lipgloss.Style
	toolBadgeStyle       lipgloss.Style
	toolBadgeColors      map[string]lipgloss.Color
	timestampStyle       lipgloss.Style
	tokenStyle           lipgloss.Style

	// Message avatars
	avatarUserStyle      lipgloss.Style
	avatarAssistantStyle lipgloss.Style

	// Status bar
	statusBarStyle lipgloss.Style
	helpKeyStyle   lipgloss.Style
	helpDescStyle  lipgloss.Style
	logoStyle      lipgloss.Style

	// Header bar
	headerStyle           lipgloss.Style
	headerLineStyle       lipgloss.Style
	headerBreadcrumbStyle lipgloss.Style

	// Tool calls
	toolHeaderStyle lipgloss.Style
	toolBodyStyle   lipgloss.Style
	toolErrorStyle  lipgloss.Style

	// Diffs
	diffAddStyle    lipgloss.Style
	diffRemoveStyle lipgloss.Style
	diffHeaderStyle lipgloss.Style

	// Thinking blocks
	thinkingHeaderStyle lipgloss.Style
	thinkingBodyStyle   lipgloss.Style

	// System messages
	systemMessageStyle lipgloss.Style

	// Session stats and date groups
	statsStyle     lipgloss.Style
	dateGroupStyle lipgloss.Style

	// Transition highlight
	transitionPanelStyle lipgloss.Style

	// Tool call gutters
	toolGutterCollapsedStyle lipgloss.Style
	toolGutterExpandedStyle  lipgloss.Style
	thinkingGutterStyle      lipgloss.Style

	// Turn divider
	turnDividerStyle lipgloss.Style

	// Help overlay
	helpOverlayStyle      lipgloss.Style
	helpOverlayTitleStyle lipgloss.Style
	helpSectionStyle      lipgloss.Style

	// Scroll indicator
	scrollTrackStyle lipgloss.Style
	scrollThumbStyle lipgloss.Style

	// Empty state
	emptyStyle     lipgloss.Style
	emptyLogoStyle lipgloss.Style
)
