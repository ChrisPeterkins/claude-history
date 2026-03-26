package ui

const (
	// Responsive layout breakpoints (terminal column width)
	breakpointNarrow = 60
	breakpointMedium = 100
	breakpointWide   = 140

	// Content truncation limits (characters)
	maxThinkingLen       = 2000
	maxToolResultLen     = 3000
	maxCommandSummaryLen = 60
	maxExportContentLen  = 1000

	// Conversation rendering (padding accounts for panel border + gutter + scroll indicator)
	conversationPadding = 8

	// Search
	searchCharLimit   = 100
	maxSearchResults  = 20
	minSearchQueryLen = 2

	// Help overlay
	helpOverlayWidth = 44

	// Session filter thresholds
	filterLongMinMessages = 20
	filterRecentDays      = 7
)

// sessionFilterTypes defines the available session filter modes.
var sessionFilterTypes = []struct {
	name  string
	label string
}{
	{"all", "All"},
	{"code", "With Code"},
	{"long", "Long"},
	{"recent", "Recent"},
}
