package main

import (
	"github.com/alecthomas/kong"
	"github.com/charmbracelet/log"
	"github.com/corani/onyx/internal/config"
)


type CLI struct {
	Vault string  `help:"Path to the Obsidian vault." short:"v"`
	Note  NoteCmd `cmd:""                             help:"Manage notes."`
	Plan  PlanCmd `cmd:""                             help:"Manage day planner."`
	Todo  TodoCmd `cmd:""                             help:"Manage todos."`
}

func main() {
	// Try environment and config files
	cfg, err := config.Load()
	if err != nil {
		log.Warn("Could not parse environment variables", "err", err)
	}

	// Parse CLI flags (flags override env)
	var cli CLI

	ctx := kong.Parse(&cli)

	if cli.Vault != "" {
		cfg.Vault = cli.Vault
	}

	// Run the selected command
	ctx.FatalIfErrorf(
		ctx.Run(cfg))
}
