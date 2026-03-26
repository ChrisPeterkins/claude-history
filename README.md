# Claude History

A terminal UI for browsing your [Claude Code](https://claude.ai/claude-code) conversation history. Built with Go and the [Charm](https://charm.sh) ecosystem (Bubble Tea, Lip Gloss, Glamour).

## Features

- **Three-panel layout** — Projects, Sessions, and Conversation with responsive breakpoints
- **Full conversation rendering** — Markdown, syntax-highlighted code, inline diffs, tool call details
- **Collapsible tool calls** — Expand/collapse individual tool calls or all at once
- **4 color themes** — Nord, Dracula, Catppuccin, Light (persisted across sessions)
- **Full-text search** — `/` to search across all conversations, `Ctrl+F` to find within a conversation
- **Vim-style marks** — `m`+letter to bookmark, `'`+letter to jump back
- **Mouse support** — Click to select, scroll wheel in all panels
- **Session filtering** — Filter by recent, long, or code-heavy sessions
- **Clipboard export** — Copy any conversation as Markdown
- **Cross-platform** — macOS, Linux, Windows clipboard support

## Install

### From source (requires Go 1.21+)

```sh
go install github.com/chrispeterkins/claude-history@latest
```

### Build locally

```sh
git clone https://github.com/ChrisPeterkins/claude-history.git
cd claude-history
make install   # builds and copies to /usr/local/bin
```

## Usage

```sh
claude-history
```

The app reads conversation data from `~/.claude/` — the same location where Claude Code stores your session history. No setup required.

## Keybindings

### Navigation
| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Move cursor / scroll |
| `g` / `G` | Jump to top / bottom |
| `PgUp` / `PgDn` | Page up / page down |
| `n` / `N` | Next / previous user message |
| `Tab` / `Shift+Tab` | Switch panels |
| `Enter` | Drill into next panel |
| `Esc` | Go back |

### Actions
| Key | Action |
|-----|--------|
| `f` | Toggle full-screen conversation |
| `/` | Search across all conversations |
| `Ctrl+F` | Find text within current conversation |
| `Space` | Expand/collapse tool call at cursor |
| `a` / `A` | Expand all / collapse all |
| `m` + `a-z` | Set a bookmark |
| `'` + `a-z` | Jump to bookmark |
| `y` | Copy conversation as Markdown |
| `t` | Cycle color theme |
| `F` | Cycle session filter |
| `?` | Show help overlay |
| `q` | Quit |

## Configuration

Settings are stored in `~/.claude-history.json` and created automatically when you change a preference.

```json
{
  "projectRoots": ["Projects", "code", "work/clients"],
  "theme": "dracula",
  "defaultFilter": "all"
}
```

| Field | Description | Default |
|-------|-------------|---------|
| `projectRoots` | Directory names that contain your projects. Used to extract friendly project names from paths. | Common names: Projects, code, dev, src, repos, workspace |
| `theme` | Color theme on startup. One of: `nord`, `dracula`, `catppuccin`, `light`, `solarized`, `gruvbox`, `tokyo-night`, `high-contrast`, `custom` | `nord` |
| `customTheme` | Custom color palette (used when theme is `"custom"`). See below. | — |
| `defaultFilter` | Session filter on startup. One of: `all`, `code`, `long`, `recent` | `all` |

## How it works

Claude Code stores conversation history locally in `~/.claude/`:

- **`history.jsonl`** — Lightweight index of every query (display text, timestamp, project, session ID)
- **`projects/{encoded-path}/{sessionId}.jsonl`** — Full conversation data with messages, tool calls, thinking blocks, token usage

This app reads those files directly — no API calls, no authentication, fully offline.

## Themes

Switch themes with `t`. Your choice is saved automatically.

- **Nord** — Soft arctic blues and purples (default)
- **Dracula** — Dark with vibrant colors
- **Catppuccin** — Muted pastels on dark background
- **Light** — Light background for bright environments
- **Solarized** — Ethan Schoonover's precision dark palette
- **Gruvbox** — Retro warm tones
- **Tokyo Night** — Modern dark with vivid accents
- **High Contrast** — Maximum contrast for accessibility

### Custom Theme

Create your own palette in `~/.claude-history.json`:

```json
{
  "theme": "custom",
  "customTheme": {
    "primary": "#FF6B6B",
    "secondary": "#4ECDC4",
    "accent": "#45B7D1",
    "warm": "#F7DC6F",
    "fg": "#EEEEEE",
    "fgDim": "#888888",
    "bg": "#1A1A2E",
    "bgSelected": "#2A2A4E",
    "border": "#444466",
    "red": "#FF4444",
    "green": "#44FF44"
  }
}
```

Any missing colors fall back to Nord defaults.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Glamour](https://github.com/charmbracelet/glamour) — Markdown rendering
- [Chroma](https://github.com/alecthomas/chroma) — Syntax highlighting

## License

MIT
