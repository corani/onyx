package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"strings"
)

type Vault struct {
	Path string
}

func NewVault(path string) *Vault {
	return &Vault{
		Path: path,
	}
}

// GetDailyNote retrieves the daily note for the given date.
// If date is an empty string, it defaults to today's date.
func (v *Vault) GetDailyNote(date string) (*Note, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Check if note exists in `$vault/Journal/Daily/$date.md`
	notePath := filepath.Join(v.Path, "Journal", "Daily", date+".md")

	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("daily note does not exist: %w", err)
	}

	return &Note{
		Vault: v,
		Path:  notePath,
		Date:  date,
	}, nil
}

// findSection returns the start and end indices of the section whose header matches exactly header.
func findSection(lines []string, header string) (int, int) {
	var start, end int

	for idx, line := range lines {
		if strings.HasPrefix(line, header) {
			start = idx
		} else if start > 0 && strings.HasPrefix(line, "## ") && line != header {
			end = idx

			break
		}
	}

	if start > 0 && end == 0 {
		end = len(lines)
	}

	return start, end
}
