package main

import (
    "fmt"
    "os"

    "github.com/charmbracelet/log"
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

    note, err := vault.GetDailyNote(noteCmd.Date)
    if err != nil {
        log.Error("Failed to get daily note", "err", err)

        return fmt.Errorf("get daily note: %w", err)
    }

    if cmd.Text == "" {
        noteText, err := input.PromptForText("Write your note")
        if err != nil {
            log.Error("Failed to get note text from user", "err", err)

            return fmt.Errorf("prompt for text: %w", err)
        }

        cmd.Text = noteText
    }

    if err := note.Create(cmd.Text); err != nil {
        log.Error("Failed to create note", "err", err)

        return fmt.Errorf("create note: %w", err)
    }

    return nil
}

func (cmd *NoteListCmd) Run(cfg *config.Config, noteCmd *NoteCmd) error {
    vault := obsidian.NewVault(cfg.Vault)

    note, err := vault.GetDailyNote(noteCmd.Date)
    if err != nil {
        log.Error("Failed to get daily note", "err", err)

        return fmt.Errorf("get daily note: %w", err)
    }

    notesMarkdown, err := note.List()
    if err != nil {
        log.Error("Failed to list notes", "err", err)

        return fmt.Errorf("list notes: %w", err)
    }

    if len(notesMarkdown) == 0 {
        notesMarkdown = "## Notes\n\n(empty)"
    }

    if err := markdown.Render(os.Stdout, notesMarkdown); err != nil {
        log.Error("Failed to render markdown", "err", err)

        return fmt.Errorf("render markdown: %w", err)
    }

    return nil
}
