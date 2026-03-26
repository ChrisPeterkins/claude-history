package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/chrispeterkins/claude-history/internal/config"
)

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
		Primary:    lipgloss.Color("#B48EAD"), // purple
		Secondary:  lipgloss.Color("#88C0D0"), // frost blue
		Accent:     lipgloss.Color("#5E81AC"), // steel blue (distinct from green)
		Warm:       lipgloss.Color("#EBCB8B"), // yellow
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
		Primary:    lipgloss.Color("#BD93F9"), // purple
		Secondary:  lipgloss.Color("#8BE9FD"), // cyan
		Accent:     lipgloss.Color("#FFB86C"), // orange (distinct from green)
		Warm:       lipgloss.Color("#F1FA8C"), // yellow
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
		Primary:    lipgloss.Color("#CBA6F7"), // mauve
		Secondary:  lipgloss.Color("#89DCEB"), // sky
		Accent:     lipgloss.Color("#F5C2E7"), // pink (distinct from green)
		Warm:       lipgloss.Color("#F9E2AF"), // yellow
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
		Primary:    lipgloss.Color("#7C3AED"), // violet
		Secondary:  lipgloss.Color("#0369A1"), // darker blue for contrast
		Accent:     lipgloss.Color("#9333EA"), // purple (distinct from green)
		Warm:       lipgloss.Color("#B45309"), // darker amber for contrast
		Subtle:     lipgloss.Color("#9CA3AF"), // darker gray for visibility
		Fg:         lipgloss.Color("#111827"), // near-black text
		FgDim:      lipgloss.Color("#4B5563"), // darker dim text
		Bg:         lipgloss.Color("#FFFFFF"),
		BgSelected: lipgloss.Color("#E5E7EB"), // stronger selection contrast
		Border:     lipgloss.Color("#9CA3AF"), // darker border
		Red:        lipgloss.Color("#DC2626"),
		Green:      lipgloss.Color("#15803D"), // darker green for contrast on white
	},
	{
		Name:       "Solarized",
		Primary:    lipgloss.Color("#268BD2"), // blue
		Secondary:  lipgloss.Color("#2AA198"), // cyan
		Accent:     lipgloss.Color("#6C71C4"), // violet
		Warm:       lipgloss.Color("#B58900"), // yellow
		Subtle:     lipgloss.Color("#073642"), // base02
		Fg:         lipgloss.Color("#839496"), // base0
		FgDim:      lipgloss.Color("#586E75"), // base01
		Bg:         lipgloss.Color("#002B36"), // base03
		BgSelected: lipgloss.Color("#073642"), // base02
		Border:     lipgloss.Color("#586E75"), // base01
		Red:        lipgloss.Color("#DC322F"),
		Green:      lipgloss.Color("#859900"),
	},
	{
		Name:       "Gruvbox",
		Primary:    lipgloss.Color("#D3869B"), // purple
		Secondary:  lipgloss.Color("#83A598"), // aqua
		Accent:     lipgloss.Color("#FE8019"), // orange
		Warm:       lipgloss.Color("#FABD2F"), // yellow
		Subtle:     lipgloss.Color("#3C3836"), // bg1
		Fg:         lipgloss.Color("#EBDBB2"), // fg
		FgDim:      lipgloss.Color("#928374"), // gray
		Bg:         lipgloss.Color("#282828"), // bg
		BgSelected: lipgloss.Color("#3C3836"), // bg1
		Border:     lipgloss.Color("#504945"), // bg2
		Red:        lipgloss.Color("#FB4934"),
		Green:      lipgloss.Color("#B8BB26"),
	},
	{
		Name:       "Tokyo Night",
		Primary:    lipgloss.Color("#BB9AF7"), // purple
		Secondary:  lipgloss.Color("#7AA2F7"), // blue
		Accent:     lipgloss.Color("#FF9E64"), // orange
		Warm:       lipgloss.Color("#E0AF68"), // yellow
		Subtle:     lipgloss.Color("#292E42"), // bg highlight
		Fg:         lipgloss.Color("#C0CAF5"), // fg
		FgDim:      lipgloss.Color("#565F89"), // comment
		Bg:         lipgloss.Color("#1A1B26"), // bg
		BgSelected: lipgloss.Color("#292E42"), // bg highlight
		Border:     lipgloss.Color("#3B4261"), // border
		Red:        lipgloss.Color("#F7768E"),
		Green:      lipgloss.Color("#9ECE6A"),
	},
	{
		Name:       "High Contrast",
		Primary:    lipgloss.Color("#FFFF00"), // bright yellow
		Secondary:  lipgloss.Color("#00FFFF"), // bright cyan
		Accent:     lipgloss.Color("#FF00FF"), // magenta
		Warm:       lipgloss.Color("#FFAA00"), // orange
		Subtle:     lipgloss.Color("#333333"),
		Fg:         lipgloss.Color("#FFFFFF"), // pure white
		FgDim:      lipgloss.Color("#AAAAAA"),
		Bg:         lipgloss.Color("#000000"), // pure black
		BgSelected: lipgloss.Color("#333333"),
		Border:     lipgloss.Color("#666666"),
		Red:        lipgloss.Color("#FF0000"),
		Green:      lipgloss.Color("#00FF00"),
	},
}

func init() {
	applyTheme(themes[0]) // Default to Nord theme
}

// buildCustomTheme creates a Theme from user-provided config hex colors.
// Missing colors fall back to Nord defaults.
func buildCustomTheme(ct *config.CustomTheme) Theme {
	nord := themes[0] // Nord as fallback
	t := Theme{Name: "Custom"}

	t.Primary = colorOrDefault(ct.Primary, nord.Primary)
	t.Secondary = colorOrDefault(ct.Secondary, nord.Secondary)
	t.Accent = colorOrDefault(ct.Accent, nord.Accent)
	t.Warm = colorOrDefault(ct.Warm, nord.Warm)
	t.Fg = colorOrDefault(ct.Fg, nord.Fg)
	t.FgDim = colorOrDefault(ct.FgDim, nord.FgDim)
	t.Bg = colorOrDefault(ct.Bg, nord.Bg)
	t.BgSelected = colorOrDefault(ct.BgSelected, nord.BgSelected)
	t.Border = colorOrDefault(ct.Border, nord.Border)
	t.Red = colorOrDefault(ct.Red, nord.Red)
	t.Green = colorOrDefault(ct.Green, nord.Green)
	// Subtle is derived from Border if not in the custom palette
	t.Subtle = t.Border

	return t
}

func colorOrDefault(hex string, fallback lipgloss.Color) lipgloss.Color {
	if hex != "" {
		return lipgloss.Color(hex)
	}
	return fallback
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
	colorGreen = t.Green

	rebuildStyles()
}

// panelBorderStyle creates a panel style with the given border type and color.
func panelBorderStyle(border lipgloss.Border, color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Border(border).BorderForeground(color).Padding(0, 1)
}

// leftBorderStyle creates a left-only border style.
func leftBorderStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(color).
		PaddingLeft(1)
}

// rebuildStyles reconstructs all lipgloss styles from the current color vars.
func rebuildStyles() {
	// --- Panels ---
	panelStyle = panelBorderStyle(lipgloss.NormalBorder(), colorSubtle)
	activePanelStyle = panelBorderStyle(lipgloss.RoundedBorder(), colorPrimary)
	panelTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Padding(0, 1)
	panelTitleActiveStyle = lipgloss.NewStyle().Bold(true).Foreground(colorBg).Background(colorPrimary).Padding(0, 1)

	// --- List items ---
	itemStyle = lipgloss.NewStyle().Foreground(colorFg).PaddingLeft(2)
	itemDescStyle = lipgloss.NewStyle().Foreground(colorFgDim).PaddingLeft(2)
	selectedItemStyle = leftBorderStyle(colorSecondary).Foreground(colorFg).Background(colorBgSelected).Bold(true)
	selectedItemDescStyle = lipgloss.NewStyle().Foreground(colorSecondary).Background(colorBgSelected).PaddingLeft(2)

	// --- Dimmed list items (unfocused panels) ---
	dimItemStyle = lipgloss.NewStyle().Foreground(colorSubtle).PaddingLeft(2)
	dimItemDescStyle = lipgloss.NewStyle().Foreground(colorSubtle).PaddingLeft(2)
	dimSelectedItemStyle = leftBorderStyle(colorSubtle).Foreground(colorFgDim)

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

	// Per-tool badge colors
	toolBadgeColors = map[string]lipgloss.Color{
		"Bash":  colorGreen,
		"Edit":  colorWarm,
		"Write": colorSecondary,
		"Read":  colorFgDim,
		"Grep":  colorPrimary,
		"Glob":  colorPrimary,
		"Agent": colorRed,
	}

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

	toolGutterCollapsedStyle = leftBorderStyle(colorSubtle)
	toolGutterExpandedStyle = leftBorderStyle(colorWarm)

	// --- Diffs ---
	diffAddStyle = lipgloss.NewStyle().
		Foreground(colorAccent)

	diffRemoveStyle = lipgloss.NewStyle().
		Foreground(colorRed)

	// --- Thinking ---
	thinkingHeaderStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Italic(true)

	thinkingBodyStyle = lipgloss.NewStyle().
		Foreground(colorFgDim).
		Padding(0, 2)

	thinkingGutterStyle = leftBorderStyle(colorFgDim)

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
