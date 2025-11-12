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

// DiaryCmd manages the single daily diary entry under `## One Line`.
type DiaryCmd struct {
	Date string       `default:"" help:"Date for the daily note (YYYY-MM-DD)." short:"d"`
	Show DiaryShowCmd `cmd:""     help:"Show the diary entry."`
	Edit DiaryEditCmd `cmd:""     help:"Edit the diary entry."`
}

type DiaryShowCmd struct{}

type DiaryEditCmd struct{}

func (cmd *DiaryShowCmd) Run(cfg *config.Config, diaryCmd *DiaryCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(diaryCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	diary, err := obsidian.NewDiary(doc)
	if err != nil {
		log.Error("Failed to read diary entry", "err", err)

		return fmt.Errorf("get diary: %w", err)
	}

	body := diary.Get()

	// Compose markdown for rendering.
	content := []string{"## One Line", ""}
	if strings.TrimSpace(body) == "" { // empty
		content = append(content, "(empty)")
	} else {
		content = append(content, body)
	}

	if err := markdown.Render(os.Stdout, strings.Join(content, "\n")); err != nil {
		log.Error("Failed to render markdown", "err", err)

		return fmt.Errorf("render markdown: %w", err)
	}

	return nil
}

func (cmd *DiaryEditCmd) Run(cfg *config.Config, diaryCmd *DiaryCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(diaryCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	diary, err := obsidian.NewDiary(doc)
	if err != nil {
		log.Error("Failed to read diary entry", "err", err)

		return fmt.Errorf("get diary: %w", err)
	}

	existing := diary.Get()

	updated, err := input.PromptForTextWithInitial("Edit diary entry", existing)
	if err != nil {
		log.Error("Failed during diary edit input", "err", err)

		return fmt.Errorf("input diary: %w", err)
	}

	if err := diary.Set(updated); err != nil {
		log.Error("Failed to write diary entry", "err", err)

		return fmt.Errorf("set diary: %w", err)
	}

	return nil
}
