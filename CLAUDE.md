# Claude Code History

A terminal UI for browsing Claude Code conversation history, built with Go and the Charm ecosystem (Bubble Tea, Lip Gloss, Glamour).

## Project Structure

```
main.go                  # Entry point
internal/
  data/
    types.go             # Data types (Project, Session, Message, etc.)
    loader.go            # JSONL parsing and file I/O
  ui/
    model.go             # Bubble Tea model, Update/View logic
    styles.go            # Lip Gloss style definitions
```

## Go Style Rules

These rules apply to all Go code in this project. Follow them when writing or modifying code.

### Naming
- **Use camelCase** for local variables, **PascalCase** for exported names. No underscores in Go names (except test functions).
- **Short variable names** are idiomatic in small scopes: `f` for a file, `p` for a project, `i` for an index. Use descriptive names only when the scope is large or the meaning is ambiguous.
- **Receiver names** should be 1-2 letters matching the type: `(m Model)`, `(s Session)`. Never use `self` or `this`.
- **Acronyms** stay all-caps: `ID`, `URL`, `HTTP`, not `Id`, `Url`, `Http`.

### Error Handling
- **Always handle errors explicitly.** Never use `_` to discard an error unless you have a clear reason and add a comment explaining why.
- **Return errors, don't panic.** Reserve `panic` for truly unrecoverable programmer bugs.
- **Wrap errors with context** using `fmt.Errorf("loading session %s: %w", id, err)` so the caller knows what failed.

### Structure & Organization
- **Keep functions short.** If a function exceeds ~40 lines, look for a logical split.
- **Group related code** in the same file. Don't split across files until a file exceeds ~300 lines.
- **Use `internal/`** for packages that shouldn't be imported by external code (this is enforced by the Go compiler).
- **One package, one purpose.** The `data` package handles I/O and parsing. The `ui` package handles rendering and input. Don't mix concerns.

### Bubble Tea Patterns
- **Cmd pattern**: Side effects (file I/O, network) happen in `tea.Cmd` functions, never in `Update` or `View`.
- **Messages are typed**: Define a struct for each message type (`sessionsLoaded`, `messagesLoaded`). Use the type switch in `Update`.
- **View is pure**: `View()` should only read model state, never mutate it.
- **Styles in styles.go**: Keep all `lipgloss.NewStyle()` definitions in `styles.go`. Reference them by name in `model.go`.

### General Go Idioms
- **Accept interfaces, return structs.** Don't over-abstract — only extract an interface when you have 2+ implementations.
- **Zero values are useful.** Design types so their zero value is ready to use when possible.
- **Prefer composition over inheritance** (Go doesn't have inheritance). Embed types to reuse behavior.
- **`defer` for cleanup**: Always `defer f.Close()` immediately after opening a file/resource.
- **Don't stutter**: A type in package `data` should be `data.Session`, not `data.DataSession`.
- **Comments on exported names** should start with the name: `// LoadProjects discovers all projects from ~/.claude/projects/.`

### Performance
- **Stream large files** with `bufio.Scanner` instead of reading into memory with `os.ReadFile`.
- **Lazy load**: Don't parse all messages upfront — load them when the user selects a session.
