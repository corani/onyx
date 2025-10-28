package config

import (
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

type Config struct {
	Vault string `env:"ONYX_VAULT"`
}

func Load() (*Config, error) {
	// Prepare list of env file paths to try, in order
	var paths []string

	paths = append(paths, ".env")

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "onyx", "config"))
	}

	if home := os.Getenv("HOME"); home != "" {
		paths = append(paths, filepath.Join(home, ".config", "onyx", "config"))
	}

	// Try loading each file in order
	for i, path := range paths {
		var err error

		if i == 0 {
			err = godotenv.Load(path)
		} else {
			err = godotenv.Overload(path)
		}

		if err != nil {
			log.Debug("Could not load env file", "path", path, "err", err)
		}
	}

	var config Config

	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
