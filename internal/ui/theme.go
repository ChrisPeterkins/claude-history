package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines a color palette for the TUI.
type Theme struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Warm       lipgloss.Color
	Subtle     lipgloss.Color
	Fg         lipgloss.Color
	FgDim      lipgloss.Color
	Bg         lipgloss.Color
	BgSelected lipgloss.Color
	Border     lipgloss.Color
	Red        lipgloss.Color
	Green      lipgloss.Color
}

var themes = []Theme{
	{
		Name:       "Nord",
		Primary:    lipgloss.Color("#B48EAD"),
		Secondary:  lipgloss.Color("#88C0D0"),
		Accent:     lipgloss.Color("#A3BE8C"),
		Warm:       lipgloss.Color("#EBCB8B"),
		Subtle:     lipgloss.Color("#4C566A"),
		Fg:         lipgloss.Color("#ECEFF4"),
		FgDim:      lipgloss.Color("#8891A5"),
		Bg:         lipgloss.Color("#2E3440"),
		BgSelected: lipgloss.Color("#3B4252"),
		Border:     lipgloss.Color("#434C5E"),
		Red:        lipgloss.Color("#BF616A"),
		Green:      lipgloss.Color("#A3BE8C"),
	},
	{
		Name:       "Dracula",
		Primary:    lipgloss.Color("#BD93F9"),
		Secondary:  lipgloss.Color("#8BE9FD"),
		Accent:     lipgloss.Color("#50FA7B"),
		Warm:       lipgloss.Color("#F1FA8C"),
		Subtle:     lipgloss.Color("#44475A"),
		Fg:         lipgloss.Color("#F8F8F2"),
		FgDim:      lipgloss.Color("#6272A4"),
		Bg:         lipgloss.Color("#282A36"),
		BgSelected: lipgloss.Color("#44475A"),
		Border:     lipgloss.Color("#6272A4"),
		Red:        lipgloss.Color("#FF5555"),
		Green:      lipgloss.Color("#50FA7B"),
	},
	{
		Name:       "Catppuccin",
		Primary:    lipgloss.Color("#CBA6F7"),
		Secondary:  lipgloss.Color("#89DCEB"),
		Accent:     lipgloss.Color("#A6E3A1"),
		Warm:       lipgloss.Color("#F9E2AF"),
		Subtle:     lipgloss.Color("#45475A"),
		Fg:         lipgloss.Color("#CDD6F4"),
		FgDim:      lipgloss.Color("#7F849C"),
		Bg:         lipgloss.Color("#1E1E2E"),
		BgSelected: lipgloss.Color("#313244"),
		Border:     lipgloss.Color("#585B70"),
		Red:        lipgloss.Color("#F38BA8"),
		Green:      lipgloss.Color("#A6E3A1"),
	},
	{
		Name:       "Light",
		Primary:    lipgloss.Color("#7C3AED"),
		Secondary:  lipgloss.Color("#0284C7"),
		Accent:     lipgloss.Color("#16A34A"),
		Warm:       lipgloss.Color("#CA8A04"),
		Subtle:     lipgloss.Color("#D1D5DB"),
		Fg:         lipgloss.Color("#1F2937"),
		FgDim:      lipgloss.Color("#6B7280"),
		Bg:         lipgloss.Color("#FFFFFF"),
		BgSelected: lipgloss.Color("#F3F4F6"),
		Border:     lipgloss.Color("#D1D5DB"),
		Red:        lipgloss.Color("#DC2626"),
		Green:      lipgloss.Color("#16A34A"),
	},
}

func init() {
	applyTheme(themes[0]) // Default to Nord theme
}

// applyTheme updates all style variables to use the given theme's colors.
func applyTheme(t Theme) {
	colorPrimary = t.Primary
	colorSecondary = t.Secondary
	colorAccent = t.Accent
	colorWarm = t.Warm
	colorSubtle = t.Subtle
	colorFg = t.Fg
	colorFgDim = t.FgDim
	colorBg = t.Bg
	colorBgSelected = t.BgSelected
	colorBorder = t.Border
	colorRed = t.Red

	rebuildStyles()
}

// rebuildStyles reconstructs all lipgloss styles from the current color vars.
func rebuildStyles() {
	// --- Panels ---
	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorSubtle).
		Padding(0, 1)

	activePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	transitionPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
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

	// --- List items (focused panel) ---
	itemStyle = lipgloss.NewStyle().
		Foreground(colorFg).
		PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
		Foreground(colorFg).
		Background(colorBgSelected).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorSecondary).
		PaddingLeft(1)

	itemDescStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		PaddingLeft(2)

	selectedItemDescStyle = lipgloss.NewStyle().
		Foreground(colorSecondary).
		Background(colorBgSelected).
		PaddingLeft(2)

	// --- List items (dimmed for unfocused panels) ---
	dimItemStyle = lipgloss.NewStyle().
		Foreground(colorSubtle).
		PaddingLeft(2)

	dimSelectedItemStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorSubtle).
		PaddingLeft(1)

	dimItemDescStyle = lipgloss.NewStyle().
		Foreground(colorSubtle).
		PaddingLeft(2)

	dimSelectedItemDescStyle = lipgloss.NewStyle().
		Foreground(colorSubtle).
		PaddingLeft(2)

	// --- Conversation ---
	userBubbleStyle = lipgloss.NewStyle().
		Foreground(colorFg).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorSecondary).
		PaddingLeft(1)

	userLabelStyle = lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true)

	assistantLabelStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	assistantBubbleStyle = lipgloss.NewStyle().
		Foreground(colorFg).
		Padding(0, 1).
		MarginTop(1).
		MarginBottom(1)

	avatarUserStyle = lipgloss.NewStyle().
		Foreground(colorSecondary).
		Bold(true)

	avatarAssistantStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

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

	// --- Header bar ---
	headerStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	headerLineStyle = lipgloss.NewStyle().
		Foreground(colorSubtle)

	headerBreadcrumbStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Italic(true)

	// --- Status bar ---
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

	// --- Tool calls ---
	toolHeaderStyle = lipgloss.NewStyle().
		Foreground(colorWarm).
		Bold(true)

	toolBodyStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Padding(0, 2)

	toolErrorStyle = lipgloss.NewStyle().
		Foreground(colorRed)

	toolGutterCollapsedStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorSubtle).
		PaddingLeft(1)

	toolGutterExpandedStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorWarm).
		PaddingLeft(1)

	// --- Diffs ---
	diffAddStyle = lipgloss.NewStyle().
		Foreground(colorAccent)

	diffRemoveStyle = lipgloss.NewStyle().
		Foreground(colorRed)

	diffHeaderStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Bold(true)

	// --- Thinking ---
	thinkingHeaderStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Italic(true)

	thinkingBodyStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Padding(0, 2)

	thinkingGutterStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(colorFgDim).
		PaddingLeft(1)

	// --- System ---
	systemMessageStyle = lipgloss.NewStyle().
		Foreground(colorSubtle).
		Italic(true).
		Align(lipgloss.Center)

	turnDividerStyle = lipgloss.NewStyle().
		Foreground(colorSubtle)

	dateGroupStyle = lipgloss.NewStyle().
		Foreground(colorWarm).
		Bold(true).
		Padding(0, 1).
		MarginTop(1)

	// --- Scroll indicator ---
	scrollTrackStyle = lipgloss.NewStyle().
		Foreground(colorSubtle)

	scrollThumbStyle = lipgloss.NewStyle().
		Foreground(colorPrimary)

	// --- Help overlay ---
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

	// --- Empty state ---
	emptyStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Italic(true).
		Align(lipgloss.Center)

	emptyLogoStyle = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Align(lipgloss.Center)
}
