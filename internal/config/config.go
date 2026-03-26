package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config holds user preferences for claude-history.
// Stored at ~/.claude-history.json.
type Config struct {
	// ProjectRoots are directory names that contain projects.
	// Paths after these segments become the project display name.
	// Example: ["Projects", "code", "work/clients"]
	ProjectRoots []string `json:"projectRoots,omitempty"`

	// Theme is the name of the color theme to use on startup.
	// Built-in: nord, dracula, catppuccin, light, solarized, gruvbox, tokyo-night, high-contrast
	Theme string `json:"theme,omitempty"`

	// DefaultFilter is the session filter applied on startup.
	// One of: all, code, long, recent
	DefaultFilter string `json:"defaultFilter,omitempty"`

	// CustomTheme defines a user-created color theme.
	// If set and Theme is "custom", this palette is used.
	CustomTheme *CustomTheme `json:"customTheme,omitempty"`
}

// CustomTheme defines a user-created color palette via hex color strings.
type CustomTheme struct {
	Primary    string `json:"primary,omitempty"`    // Main accent (selected panels, titles)
	Secondary  string `json:"secondary,omitempty"`  // Interactive elements (user messages, links)
	Accent     string `json:"accent,omitempty"`     // Highlights (transitions, search matches)
	Warm       string `json:"warm,omitempty"`       // Warnings, tool badges default
	Fg         string `json:"fg,omitempty"`         // Main text
	FgDim      string `json:"fgDim,omitempty"`      // Dimmed text, timestamps
	Bg         string `json:"bg,omitempty"`          // Background
	BgSelected string `json:"bgSelected,omitempty"` // Selected item background
	Border     string `json:"border,omitempty"`      // Panel borders
	Red        string `json:"red,omitempty"`         // Errors, diff removed
	Green      string `json:"green,omitempty"`       // Success, diff added
}

// DefaultProjectRoots are used when no config file exists or projectRoots is empty.
var DefaultProjectRoots = []string{
	"Projects", "projects",
	"Code", "code",
	"Dev", "dev",
	"src", "repos",
	"Workspace", "workspace",
}

var configPath string
var current Config

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	configPath = filepath.Join(home, ".claude-history.json")
	current = load()
}

// Get returns the current configuration.
func Get() Config {
	return current
}

// ProjectRoots returns the effective project root names, expanding any
// multi-segment roots (like "work/clients") into individual segments to match.
func ProjectRoots() map[string]bool {
	roots := current.ProjectRoots
	if len(roots) == 0 {
		roots = DefaultProjectRoots
	}

	result := make(map[string]bool)
	for _, root := range roots {
		// Support multi-segment roots like "work/clients"
		// by matching on the last segment
		parts := strings.Split(root, "/")
		result[parts[len(parts)-1]] = true
	}
	return result
}

// Save writes the current config to disk.
func Save(c Config) error {
	current = c
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// DefaultFilterName returns the configured default filter, or "all".
func DefaultFilterName() string {
	if current.DefaultFilter != "" {
		return current.DefaultFilter
	}
	return "all"
}

func load() Config {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{} // no config file is fine
	}
	var c Config
	json.Unmarshal(data, &c) // ignore parse errors, use defaults
	return c
}
