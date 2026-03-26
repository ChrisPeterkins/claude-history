package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chrispeterkins/claude-history/internal/ui"
)

// Set at build time via: go build -ldflags "-X main.version=v1.0.0"
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("claude-history " + version)
			return
		case "--help", "-h":
			printHelp()
			return
		}
	}

	// Graceful signal handling — ensure terminal is restored on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	p := tea.NewProgram(
		ui.NewModel(version),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithContext(ctx),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`claude-history — browse your Claude Code conversation history

Usage:
  claude-history              Launch the TUI
  claude-history --version    Print version
  claude-history --help       Show this help

Configuration:
  ~/.claude-history.json      User preferences (theme, filters, project roots)

Data source:
  ~/.claude/                  Claude Code conversation history (read-only)

Keybindings:
  Press ? inside the app for a full keybinding reference.

More info:
  https://github.com/ChrisPeterkins/claude-history
`)
}
