package main

import (
	"fmt"
	"runtime/debug"

	"github.com/corani/onyx/internal/config"
	"github.com/corani/onyx/internal/version"
)

type VersionCmd struct{}

//nolint:unparam // Kong command signature requires cfg param; Run currently always returns nil.
func (cmd *VersionCmd) Run(cfg *config.Config) error {
	_ = cfg // Kong injects config; keep param to match command signature.

	info, ok := debug.ReadBuildInfo()
	if ok && info != nil {
		printBuildInfo(info)

		return nil
	}

	printFallback()

	return nil
}

func printBuildInfo(info *debug.BuildInfo) {
	fmt.Printf("Version: %s\n", info.Main.Version)

	var vcsRev, vcsTime string

	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			vcsRev = s.Value
		case "vcs.time":
			vcsTime = s.Value
		}
	}

	if vcsRev != "" {
		fmt.Printf("Commit: %s\n", vcsRev)
	} else if version.Commit != "" {
		fmt.Printf("Commit: %s\n", version.Commit)
	}

	if vcsTime != "" {
		fmt.Printf("Built: %s\n", vcsTime)
	} else if version.BuildDate != "" {
		fmt.Printf("Built: %s\n", version.BuildDate)
	}
}

func printFallback() {
	fmt.Printf("Version: %s\n", version.Version)

	if version.Commit != "" {
		fmt.Printf("Commit: %s\n", version.Commit)
	}

	if version.BuildDate != "" {
		fmt.Printf("Built: %s\n", version.BuildDate)
	}
}
