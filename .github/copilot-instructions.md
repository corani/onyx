# Copilot Instructions for Onyx

## Project Overview

- **Onyx** is a CLI tool for managing daily notes in an [Obsidian](https://obsidian.md/) vault,
  written in Go.
- The main entry point is `cmd/onyx/main.go`, which wires together commands for adding and listing
  notes, managing the day planner, and todos.
- Notes are stored as markdown files under `$VAULT/Journal/Daily/YYYY-MM-DD.md`.
- The project is structured with clear separation: `internal/config` (configuration), 
  `internal/input` (user input), `internal/markdown` (rendering), and `internal/obsidian` 
  (vault/note logic).

## Key Components

- **Config**: Loads environment variables from `.env`, `$XDG_CONFIG_HOME/onyx/config`, 
  or `$HOME/.config/onyx/config`. The main variable is `ONYX_VAULT` (path to the Obsidian vault).
- **Obsidian Vault**: `internal/obsidian/vault.go` and `note.go` handle locating and manipulating
  daily note files.
- **Input**: Uses [Bubbletea](https://github.com/charmbracelet/bubbletea) and 
  [Bubbles](https://github.com/charmbracelet/bubbles) for interactive TUI input.
- **Markdown Rendering**: Uses [Glamour](https://github.com/charmbracelet/glamour) for terminal
  markdown rendering.

## Developer Workflows

- **Build**: Standard Go build: `go build ./cmd/onyx`
- **Run**: `ONYX_VAULT=/path/to/vault ./onyx note add` or `note list`
- **Test**: Standard Go test: `go test ./...`
- **Debug**: No custom debug scripts; use standard Go tools.
- **Dependencies**: Managed via Go modules (`go.mod`).

## Project Conventions

- All user-facing text input is handled via the TUI prompt in `internal/input`.
- Notes must have a `## Notes` section header in the markdown file for correct parsing/appending.
- Planner entries live under `## Day Planner`.
- Todos live under `## Todo`.
- Logging uses [charmbracelet/log](https://github.com/charmbracelet/log).
- All configuration is via environment variables or `.env` files; no CLI flags for config paths.

## Integration Points

- Expects an existing Obsidian vault structure. Does not create vaults or daily note files if missing.
- Relies on the file system for note storage and retrieval.
- No network or external service dependencies.

## Examples

- To add a note for today: `ONYX_VAULT=~/vault ./onyx note add "My note text"`
- To list notes for a date: `ONYX_VAULT=~/vault ./onyx note list -d 2025-10-28`
- To add a planner entry: `ONYX_VAULT=~/vault ./onyx plan add 09:30 "Start work"`
- To add a todo: `ONYX_VAULT=~/vault ./onyx todo add "Inbox Zero"`
- To nest a todo under a parent (substring match): `ONYX_VAULT=~/vault ./onyx todo add --parent Inbox "Archive old mail"`
- To check a todo by substring: `ONYX_VAULT=~/vault ./onyx todo check "Inbox Zero"`

## Todo Section

The daily note may contain a section:

```
## Todo
```

Todos are markdown checkbox list items. Nesting is represented by leading tab characters:

```
- [ ] Inbox Zero
- [ ] Architecture
  - [ ] REST Runtime
- [x] Add 'plan' command to Onyx
```

Adding with `--parent` performs a case-insensitive substring match against exactly one existing
todo (at any depth). The new item is inserted after the matched parent's descendant block with one
additional leading tab. If the substring matches 0 items an error is returned. If it matches more
than one item an error is returned listing all matched items.

Interactive add (no text argument) prompts for todo text; if `--parent` was not supplied, an
optional parent prompt is shown (blank = top-level).

Checking / unchecking also uses case-insensitive substring matching over the todo text (without the
checkbox marker or indentation). Same single-match rule applies.

The section must already exist; it is not created implicitly.

## Key Files

- `cmd/onyx/main.go`: CLI entry and command wiring
- `internal/obsidian/vault.go`, `note.go`: Vault and note logic
- `internal/input/input.go`: TUI input
- `internal/markdown/markdown.go`: Markdown rendering
- `internal/config/config.go`: Config loading

---

If you add new commands or change note file structure, update this file to keep AI agents productive.
