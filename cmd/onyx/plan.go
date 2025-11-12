package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/corani/onyx/internal/config"
	"github.com/corani/onyx/internal/input"
	"github.com/corani/onyx/internal/markdown"
	"github.com/corani/onyx/internal/obsidian"
)

type PlanCmd struct {
	Date    string         `default:"" help:"Date for the daily note (YYYY-MM-DD)." short:"d"`
	Add     PlanAddCmd     `cmd:""     help:"Add a new planner entry."`
	List    PlanListCmd    `cmd:""     help:"List planner entries."`
	Check   PlanCheckCmd   `cmd:""     help:"Mark an entry as done."`
	Uncheck PlanUncheckCmd `cmd:""     help:"Mark an entry as not done."`
}

type PlanAddCmd struct {
	TimeToken string `arg:"" help:"Start time or range (HH:MM or HH:MM-HH:MM)."  name:"time"`
	Text      string `arg:"" help:"Entry description (if omitted, interactive)." name:"text" optional:""`
}

type PlanListCmd struct{}

type PlanCheckCmd struct {
	Time string `arg:"" help:"Start time or time range (HH:MM or HH:MM-HH:MM)." name:"time"`
}

type PlanUncheckCmd struct {
	Time string `arg:"" help:"Start time or time range (HH:MM or HH:MM-HH:MM)." name:"time"`
}

func (cmd *PlanAddCmd) Run(cfg *config.Config, planCmd *PlanCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(planCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	if cmd.TimeToken == "" {
		token, err := input.PromptForLine("Time (HH:MM or HH:MM-HH:MM)")
		if err != nil {
			log.Error("Failed to get time token", "err", err)

			return fmt.Errorf("prompt time token: %w", err)
		}

		cmd.TimeToken = strings.TrimSpace(token)
	}

	start, end, err := splitTimeToken(cmd.TimeToken)
	if err != nil {
		return fmt.Errorf("invalid time token: %w", err)
	}

	if cmd.Text == "" {
		desc, err := input.PromptForLine("Planner entry description")
		if err != nil {
			log.Error("Failed to get planner description", "err", err)

			return fmt.Errorf("prompt description: %w", err)
		}

		cmd.Text = strings.TrimSpace(desc)
	}

	planner, err := obsidian.NewPlan(doc)
	if err != nil {
		log.Error("Failed to read planner", "err", err)

		return fmt.Errorf("read planner: %w", err)
	}

	if err := planner.Add(start, end, cmd.Text); err != nil {
		log.Error("Failed to add planner entry", "err", err)

		return fmt.Errorf("add planner entry: %w", err)
	}

	return nil
}

var timeTokenPattern = regexp.MustCompile(`^[0-2][0-9]:[0-5][0-9](?:-[0-2][0-9]:[0-5][0-9])?$`)

var (
	ErrBadTimeTokenFormat      = errors.New("bad time token format")
	ErrTimeTokenEndBeforeStart = errors.New("time token end must be after start")
)

func splitTimeToken(token string) (string, string, error) {
	token = strings.TrimSpace(token)
	if !timeTokenPattern.MatchString(token) {
		return "", "", fmt.Errorf("%w: %s", ErrBadTimeTokenFormat, token)
	}

	parts := strings.Split(token, "-")
	if len(parts) == 1 {
		return parts[0], "", nil
	}

	if parts[1] <= parts[0] {
		return "", "", ErrTimeTokenEndBeforeStart
	}

	return parts[0], parts[1], nil
}

func (cmd *PlanListCmd) Run(cfg *config.Config, planCmd *PlanCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(planCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	planner, err := obsidian.NewPlan(doc)
	if err != nil {
		log.Error("Failed to read planner", "err", err)

		return fmt.Errorf("read planner: %w", err)
	}

	body := planner.List()

	if err := markdown.Render(os.Stdout, body); err != nil {
		log.Error("Failed to render markdown", "err", err)

		return fmt.Errorf("render markdown: %w", err)
	}

	return nil
}

func (cmd *PlanCheckCmd) Run(cfg *config.Config, planCmd *PlanCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(planCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	planner, err := obsidian.NewPlan(doc)
	if err != nil {
		log.Error("Failed to read planner", "err", err)

		return fmt.Errorf("read planner: %w", err)
	}

	if err := planner.SetStatus(cmd.Time, true); err != nil {
		log.Error("Failed to check planner entry", "err", err)

		return fmt.Errorf("check planner entry: %w", err)
	}

	return nil
}

func (cmd *PlanUncheckCmd) Run(cfg *config.Config, planCmd *PlanCmd) error {
	vault := obsidian.NewVault(cfg.Vault)

	doc, err := vault.OpenDaily(planCmd.Date)
	if err != nil {
		log.Error("Failed to open daily document", "err", err)

		return fmt.Errorf("open daily: %w", err)
	}

	planner, err := obsidian.NewPlan(doc)
	if err != nil {
		log.Error("Failed to read planner", "err", err)

		return fmt.Errorf("read planner: %w", err)
	}

	if err := planner.SetStatus(cmd.Time, false); err != nil {
		log.Error("Failed to uncheck planner entry", "err", err)

		return fmt.Errorf("uncheck planner entry: %w", err)
	}

	return nil
}
