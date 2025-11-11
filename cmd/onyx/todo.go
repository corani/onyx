package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/corani/onyx/internal/config"
	"github.com/corani/onyx/internal/input"
	"github.com/corani/onyx/internal/markdown"
	"github.com/corani/onyx/internal/obsidian"
)

//nolint:tagalign
type TodoCmd struct {
	Date    string         `default:"" help:"Date for the daily note (YYYY-MM-DD)." short:"d"`
	Add     TodoAddCmd     `cmd:""     help:"Add a new todo item."`
	List    TodoListCmd    `cmd:""     help:"List todo items."`
	Check   TodoCheckCmd   `cmd:""     help:"Mark a todo item as done (substring match)."`
	Uncheck TodoUncheckCmd `cmd:""     help:"Mark a todo item as not done (substring match)."`
}

//nolint:tagalign
type TodoAddCmd struct {
	Text   string `arg:"" help:"Todo text (if omitted you will be prompted)." name:"text" optional:""`
	Parent string `help:"Optional parent todo (substring match) for nesting; may be empty to add top-level." name:"parent"`
}

type TodoListCmd struct{}

type TodoCheckCmd struct {
	Match string `arg:"" help:"Substring to match a single todo item." name:"match"`
}

type TodoUncheckCmd struct {
	Match string `arg:"" help:"Substring to match a single todo item." name:"match"`
}

func (cmd *TodoAddCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	note, err := vault.GetDailyNote(todoCmd.Date)
	if err != nil {
		log.Error("Failed to get daily note", "err", err)

		return fmt.Errorf("get daily note: %w", err)
	}

	interactive := cmd.Text == ""

	if cmd.Text == "" {
		value, err := input.PromptForLine("Todo text")
		if err != nil {
			log.Error("Failed to get todo text", "err", err)

			return fmt.Errorf("prompt todo text: %w", err)
		}

		cmd.Text = strings.TrimSpace(value)
	}

	if interactive && cmd.Parent == "" { // optional parent prompt
		parent, err := input.PromptForLine("Parent (blank for top-level)")
		if err != nil {
			log.Error("Failed to prompt parent", "err", err)

			return fmt.Errorf("prompt parent: %w", err)
		}

		cmd.Parent = strings.TrimSpace(parent)
	}

	todos := obsidian.NewTodos(note)
	if err := todos.Add(cmd.Text, cmd.Parent); err != nil {
		log.Error("Failed to add todo", "err", err)

		return fmt.Errorf("add todo: %w", err)
	}

	return nil
}

func (cmd *TodoListCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	note, err := vault.GetDailyNote(todoCmd.Date)
	if err != nil {
		log.Error("Failed to get daily note", "err", err)

		return fmt.Errorf("get daily note: %w", err)
	}

	todos := obsidian.NewTodos(note)

	section, err := todos.List()
	if err != nil {
		log.Error("Failed to list todos", "err", err)

		return fmt.Errorf("list todos: %w", err)
	}

	if section == "" {
		section = "## Todo\n\n(empty)"
	}

	if err := markdown.Render(os.Stdout, section); err != nil {
		log.Error("Failed to render markdown", "err", err)

		return fmt.Errorf("render markdown: %w", err)
	}

	return nil
}

func (cmd *TodoCheckCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	note, err := vault.GetDailyNote(todoCmd.Date)
	if err != nil {
		log.Error("Failed to get daily note", "err", err)

		return fmt.Errorf("get daily note: %w", err)
	}

	todos := obsidian.NewTodos(note)
	if err := todos.SetStatus(cmd.Match, true); err != nil {
		log.Error("Failed to check todo", "err", err)

		return fmt.Errorf("check todo: %w", err)
	}

	return nil
}

func (cmd *TodoUncheckCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	note, err := vault.GetDailyNote(todoCmd.Date)
	if err != nil {
		log.Error("Failed to get daily note", "err", err)

		return fmt.Errorf("get daily note: %w", err)
	}

	todos := obsidian.NewTodos(note)
	if err := todos.SetStatus(cmd.Match, false); err != nil {
		log.Error("Failed to uncheck todo", "err", err)

		return fmt.Errorf("uncheck todo: %w", err)
	}

	return nil
}
