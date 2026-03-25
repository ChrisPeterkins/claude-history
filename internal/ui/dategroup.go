package ui

import (
	"fmt"
	"time"

	"github.com/chrispeterkins/claude-history/internal/data"
)

// DateGroup groups sessions under a label like "Today" or "This Week".
type DateGroup struct {
	Label    string
	Sessions []indexedSession
}

// indexedSession wraps a session with its original index for cursor tracking.
type indexedSession struct {
	data.Session
	OriginalIndex int
}

// GroupSessionsByDate groups sessions into time-based buckets.
func GroupSessionsByDate(sessions []data.Session) []DateGroup {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	groups := map[string]*DateGroup{
		"Today":      {Label: "Today"},
		"Yesterday":  {Label: "Yesterday"},
		"This Week":  {Label: "This Week"},
		"This Month": {Label: "This Month"},
		"Older":      {Label: "Older"},
	}

	order := []string{"Today", "Yesterday", "This Week", "This Month", "Older"}

	for i, s := range sessions {
		is := indexedSession{Session: s, OriginalIndex: i}
		t := s.StartedAt

		switch {
		case t.After(today) || t.Equal(today):
			groups["Today"].Sessions = append(groups["Today"].Sessions, is)
		case t.After(yesterday) || t.Equal(yesterday):
			groups["Yesterday"].Sessions = append(groups["Yesterday"].Sessions, is)
		case t.After(weekAgo):
			groups["This Week"].Sessions = append(groups["This Week"].Sessions, is)
		case t.After(monthAgo):
			groups["This Month"].Sessions = append(groups["This Month"].Sessions, is)
		default:
			groups["Older"].Sessions = append(groups["Older"].Sessions, is)
		}
	}

	// Return only non-empty groups
	var result []DateGroup
	for _, label := range order {
		if g := groups[label]; len(g.Sessions) > 0 {
			result = append(result, *g)
		}
	}
	return result
}

// relativeTime returns a human-friendly relative timestamp.
func relativeTime(t time.Time) string {
	now := time.Now()
	d := now.Sub(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	case d < 14*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 8*7*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return t.Format("Jan 02")
	}
}
