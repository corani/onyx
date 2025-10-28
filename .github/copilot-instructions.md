# Copilot Instructions for Onyx

## Project Overview

- **Onyx** is a CLI tool for managing daily notes in an [Obsidian](https://obsidian.md/) vault,
  written in Go.
- The main entry point is `cmd/onyx/main.go`, which wires together commands for adding and listing
  notes.
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
- Logging uses [charmbracelet/log](https://github.com/charmbracelet/log).
- All configuration is via environment variables or `.env` files; no CLI flags for config paths.

## Integration Points

- Expects an existing Obsidian vault structure. Does not create vaults or daily note files if missing.
- Relies on the file system for note storage and retrieval.
- No network or external service dependencies.

## Examples

- To add a note for today: `ONYX_VAULT=~/vault ./onyx note add "My note text"`
- To list notes for a date: `ONYX_VAULT=~/vault ./onyx note list -d 2025-10-28`

## Key Files

- `cmd/onyx/main.go`: CLI entry and command wiring
- `internal/obsidian/vault.go`, `note.go`: Vault and note logic
- `internal/input/input.go`: TUI input
- `internal/markdown/markdown.go`: Markdown rendering
- `internal/config/config.go`: Config loading

---

If you add new commands or change note file structure, update this file to keep AI agents productive.
