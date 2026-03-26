package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chrispeterkins/claude-history/internal/data"
)

// exportMarkdown renders the current conversation as clean markdown.
func exportMarkdown(messages []data.Message) string {
	var sb strings.Builder

	for _, msg := range messages {
		switch msg.Type {
		case "user":
			if msg.RawText == "" {
				continue
			}
			sb.WriteString("## You\n")
			sb.WriteString(fmt.Sprintf("*%s*\n\n", msg.Timestamp.Format("Jan 02 15:04:05")))
			sb.WriteString(msg.RawText)
			sb.WriteString("\n\n---\n\n")

		case "assistant":
			sb.WriteString("## Claude\n")
			sb.WriteString(fmt.Sprintf("*%s*", msg.Timestamp.Format("Jan 02 15:04:05")))
			if msg.Model != "" {
				sb.WriteString(fmt.Sprintf(" · *%s*", msg.Model))
			}
			sb.WriteString("\n\n")

			// Tool calls
			for _, pair := range msg.ToolPairs {
				sb.WriteString(fmt.Sprintf("**[%s]**", pair.Name))
				if pair.Use.Input != nil {
					if cmd, ok := pair.Use.Input["command"].(string); ok {
						sb.WriteString(fmt.Sprintf(" `%s`", cmd))
					} else if path, ok := pair.Use.Input["file_path"].(string); ok {
						sb.WriteString(fmt.Sprintf(" `%s`", path))
					}
				}
				sb.WriteString("\n")
				if pair.Result.Content != "" {
					content := pair.Result.Content
					if len(content) > maxExportContentLen {
						content = content[:maxExportContentLen] + "\n... (truncated)"
					}
					sb.WriteString("```\n")
					sb.WriteString(content)
					sb.WriteString("\n```\n")
				}
				sb.WriteString("\n")
			}

			// Text content
			sb.WriteString(msg.RawText)
			sb.WriteString("\n\n---\n\n")

		case "system":
			if msg.Subtype == "turn_duration" && msg.DurationMs > 0 {
				dur := time.Duration(msg.DurationMs) * time.Millisecond
				sb.WriteString(fmt.Sprintf("*Turn: %s*\n\n", dur.Round(time.Millisecond)))
			}
		}
	}

	return sb.String()
}

// copyToClipboard copies text to the system clipboard.
// Supports macOS (pbcopy), Linux (xclip/xsel), and Windows (clip.exe).
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, fall back to xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		}
	case "windows":
		cmd = exec.Command("clip.exe")
	default:
		return fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

type clipboardCopiedMsg struct {
	err error
}

func (m Model) copyConversationCmd() tea.Cmd {
	messages := m.messages
	return func() tea.Msg {
		md := exportMarkdown(messages)
		err := copyToClipboard(md)
		return clipboardCopiedMsg{err: err}
	}
}
