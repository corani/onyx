package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/log/v2"
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
	Match string `arg:"" help:"Substring to match a single todo item." name:"match" optional:""`
}

type TodoUncheckCmd struct {
	Match string `arg:"" help:"Substring to match a single todo item." name:"match" optional:""`
}

func (cmd *TodoAddCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(todoCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
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

	todos, err := obsidian.NewTodo(doc)
	if err != nil {
		log.Error("Failed to open todo section", "err", err)

		return fmt.Errorf("new todo: %w", err)
	}

	if err := todos.Add(cmd.Text, cmd.Parent); err != nil {
		log.Error("Failed to add todo", "err", err)

		return fmt.Errorf("add todo: %w", err)
	}

	return nil
}

func (cmd *TodoListCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(todoCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	todos, err := obsidian.NewTodo(doc)
	if err != nil {
		log.Error("Failed to open todo section", "err", err)

		return fmt.Errorf("new todo: %w", err)
	}

	section := todos.List()

	if err := markdown.Render(os.Stdout, section); err != nil {
		log.Error("Failed to render markdown", "err", err)

		return fmt.Errorf("render markdown: %w", err)
	}

	return nil
}

func (cmd *TodoCheckCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(todoCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	todos, err := obsidian.NewTodo(doc)
	if err != nil {
		log.Error("Failed to open todo section", "err", err)

		return fmt.Errorf("new todo: %w", err)
	}

	if err := todos.SetStatus(cmd.Match, true); err != nil {
		log.Error("Failed to check todo", "err", err)

		return fmt.Errorf("check todo: %w", err)
	}

	return nil
}

func (cmd *TodoUncheckCmd) Run(cfg *config.Config, todoCmd *TodoCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(todoCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	todos, err := obsidian.NewTodo(doc)
	if err != nil {
		log.Error("Failed to open todo section", "err", err)

		return fmt.Errorf("new todo: %w", err)
	}

	if err := todos.SetStatus(cmd.Match, false); err != nil {
		log.Error("Failed to uncheck todo", "err", err)

		return fmt.Errorf("uncheck todo: %w", err)
	}

	return nil
}
