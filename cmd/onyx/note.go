package main

import (
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/corani/onyx/internal/config"
	"github.com/corani/onyx/internal/input"
	"github.com/corani/onyx/internal/markdown"
	"github.com/corani/onyx/internal/obsidian"
)

type NoteCmd struct {
	Date string      `default:"" help:"Date for the daily note (YYYY-MM-DD)." short:"d"`
	Add  NoteAddCmd  `cmd:""     help:"Add a new note."`
	List NoteListCmd `cmd:""     help:"List all notes."`
}

type NoteAddCmd struct {
	Text string `arg:"" help:"The content of the note (if omitted, you will be prompted)." name:"text" optional:""`
}

type NoteListCmd struct{}

func (cmd *NoteAddCmd) Run(cfg *config.Config, noteCmd *NoteCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(noteCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	if cmd.Text == "" {
		noteText, err := input.PromptForText("Write your note")
		if err != nil {
			log.Error("Failed to get note text from user", "err", err)

			return fmt.Errorf("prompt for text: %w", err)
		}

		cmd.Text = noteText
	}

	note, err := obsidian.NewNote(doc)
	if err != nil {
		log.Error("Failed to open note", "err", err)

		return fmt.Errorf("new note: %w", err)
	}

	if err := note.Create(cmd.Text); err != nil {
		log.Error("Failed to create note", "err", err)

		return fmt.Errorf("create note: %w", err)
	}

	return nil
}

func (cmd *NoteListCmd) Run(cfg *config.Config, noteCmd *NoteCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(noteCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	note, err := obsidian.NewNote(doc)
	if err != nil {
		log.Error("Failed to open note", "err", err)

		return fmt.Errorf("new note: %w", err)
	}

	body := note.List()

	if err := markdown.Render(os.Stdout, body); err != nil {
		log.Error("Failed to render markdown", "err", err)

		return fmt.Errorf("render markdown: %w", err)
	}

	return nil
}
