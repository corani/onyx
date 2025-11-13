package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Vault struct {
	Path string
}

func NewVault(path string) *Vault {
	return &Vault{
		Path: path,
	}
}

// OpenDaily loads the daily note as a Document for the given date ("YYYY-MM-DD").
// An empty date uses today. Returns an error if the file is missing.
func (v *Vault) OpenDaily(date string) (*Document, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	path := filepath.Join(v.Path, "Journal", "Daily", date+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("daily note does not exist: %w", err)
	}

	doc, err := OpenDocument(path)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
