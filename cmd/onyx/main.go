package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/log"
	"github.com/corani/onyx/internal/config"
	"github.com/corani/onyx/internal/input"
	"github.com/corani/onyx/internal/markdown"
	"github.com/corani/onyx/internal/obsidian"
)

type NoteCmd struct {
	Date string      `help:"Date for the daily note (YYYY-MM-DD)." short:"d" default:""`
	Add  NoteAddCmd  `cmd:"" help:"Add a new note."`
	List NoteListCmd `cmd:"" help:"List all notes."`
}

type NoteAddCmd struct {
	Text string `arg:"" optional:"" name:"text" help:"The content of the note (if omitted, you will be prompted)."`
}

type NoteListCmd struct{}

func (cmd *NoteAddCmd) Run(config *config.Config, noteCmd *NoteCmd) error {
	vault := obsidian.NewVault(config.Vault)

	note, err := vault.GetDailyNote(noteCmd.Date)
	if err != nil {
		log.Error("Failed to get daily note", "err", err)
		return err
	}

	if cmd.Text == "" {
		t, err := input.PromptForText("Write your note")
		if err != nil {
			log.Error("Failed to get note text from user", "err", err)

			return err
		}

		cmd.Text = t
	}

	if err := note.Create(cmd.Text); err != nil {
		log.Error("Failed to create note", "err", err)
		return err
	}

	return nil
}

func (cmd *NoteListCmd) Run(config *config.Config, noteCmd *NoteCmd) error {
	vault := obsidian.NewVault(config.Vault)

	note, err := vault.GetDailyNote(noteCmd.Date)
	if err != nil {
		log.Error("Failed to get daily note", "err", err)

		return err
	}

	md, err := note.List()
	if err != nil {
		log.Error("Failed to list notes", "err", err)

		return err
	}

	if len(md) == 0 {
		md = "## Notes\n\n(empty)"
	}

	if err := markdown.Render(os.Stdout, md); err != nil {
		log.Error("Failed to render markdown", "err", err)

		return err
	}

	return nil
}

type CLI struct {
	Vault string  `help:"Path to the Obsidian vault." short:"v"`
	Note  NoteCmd `cmd:"" help:"Manage notes."`
}

func main() {
	// Try environment and config files
	config, err := config.Load()
	if err != nil {
		log.Warn("Could not parse environment variables", "err", err)
	}

	// Parse CLI flags (flags override env)
	var cli CLI

	ctx := kong.Parse(&cli)

	if cli.Vault != "" {
		config.Vault = cli.Vault
	}

	// Run the selected command
	ctx.FatalIfErrorf(
		ctx.Run(config))
}
