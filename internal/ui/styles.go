package ui

import "github.com/charmbracelet/lipgloss"

// Color palette — warm, modern, easy on the eyes
var (
	// Core palette
	colorPrimary    = lipgloss.Color("#B48EAD") // soft purple
	colorSecondary  = lipgloss.Color("#88C0D0") // frost blue
	colorAccent     = lipgloss.Color("#A3BE8C") // sage green
	colorWarm       = lipgloss.Color("#EBCB8B") // warm yellow
	colorSubtle     = lipgloss.Color("#4C566A") // muted gray
	colorFg         = lipgloss.Color("#ECEFF4") // bright text
	colorFgDim      = lipgloss.Color("#8891A5") // dim text
	colorBg         = lipgloss.Color("#2E3440") // dark background
	colorBgSelected = lipgloss.Color("#3B4252") // slightly lighter
	colorBorder     = lipgloss.Color("#434C5E") // border gray
	colorRed        = lipgloss.Color("#BF616A") // soft red
)

// Panel styles
var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	panelTitleActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBg).
				Background(colorPrimary).
				Padding(0, 1)
)

// List item styles
var (
	itemStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(colorFg).
				Background(colorBgSelected).
				Bold(true).
				Padding(0, 1)

	itemDescStyle = lipgloss.NewStyle().
			Foreground(colorFgDim).
			Padding(0, 1)

	selectedItemDescStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Background(colorBgSelected).
				Padding(0, 1)
)

// Conversation view styles
var (
	userBubbleStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(0, 1).
			MarginTop(1)

	userLabelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	assistantLabelStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	assistantBubbleStyle = lipgloss.NewStyle().
				Foreground(colorFg).
				Padding(0, 1).
				MarginTop(1)

	toolBadgeStyle = lipgloss.NewStyle().
			Foreground(colorBg).
			Background(colorWarm).
			Bold(true).
			Padding(0, 1).
			MarginRight(1)

	timestampStyle = lipgloss.NewStyle().
			Foreground(colorFgDim).
			Italic(true)

	tokenStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)
)

// Status bar
var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorFgDim).
			Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorFgDim)

	logoStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)
)

// Tool call styles
var (
	toolHeaderStyle = lipgloss.NewStyle().
			Foreground(colorWarm).
			Bold(true)

	toolBodyStyle = lipgloss.NewStyle().
			Foreground(colorFgDim).
			Padding(0, 2)

	toolErrorStyle = lipgloss.NewStyle().
			Foreground(colorRed)
)

// Diff styles
var (
	diffAddStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	diffRemoveStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	diffHeaderStyle = lipgloss.NewStyle().
			Foreground(colorFgDim).
			Bold(true)
)

// Thinking block styles
var (
	thinkingHeaderStyle = lipgloss.NewStyle().
				Foreground(colorFgDim).
				Italic(true)

	thinkingBodyStyle = lipgloss.NewStyle().
				Foreground(colorFgDim).
				Padding(0, 2)
)

// System message styles
var (
	systemMessageStyle = lipgloss.NewStyle().
				Foreground(colorSubtle).
				Italic(true).
				Align(lipgloss.Center)
)

// Session stats
var (
	statsStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	dateGroupStyle = lipgloss.NewStyle().
			Foreground(colorWarm).
			Bold(true).
			Padding(0, 1).
			MarginTop(1)
)

// Transition highlight
var (
	transitionPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1)
)

// Tool call gutter styles
var (
	toolGutterCollapsedStyle = lipgloss.NewStyle().
					Border(lipgloss.NormalBorder(), false, false, false, true).
					BorderForeground(colorSubtle).
					PaddingLeft(1)

	toolGutterExpandedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorWarm).
				PaddingLeft(1)

	thinkingGutterStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(colorFgDim).
				PaddingLeft(1)
)

// Turn divider
var (
	turnDividerStyle = lipgloss.NewStyle().
				Foreground(colorSubtle)
)

// Help overlay
var (
	helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(1, 2)

	helpOverlayTitleStyle = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				Align(lipgloss.Center)

	helpSectionStyle = lipgloss.NewStyle().
				Foreground(colorWarm).
				Bold(true)
)

// Empty state
var (
	emptyStyle = lipgloss.NewStyle().
			Foreground(colorFgDim).
			Italic(true).
			Align(lipgloss.Center)
)
