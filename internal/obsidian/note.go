package obsidian

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

var ErrNotesSectionNotFound = errors.New("notes section not found in note")

type Note struct {
	Vault *Vault
	Path  string
	Date  string
}

func (n *Note) List() (string, error) {
	data, err := os.ReadFile(n.Path)
	if err != nil {
		return "", fmt.Errorf("read note file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	start, end := findNotesSection(lines)
	if start == 0 {
		return "", nil
	}

	return strings.Join(lines[start:end], "\n"), nil
}

func (n *Note) Create(text string) error {
	data, err := os.ReadFile(n.Path)
	if err != nil {
		return fmt.Errorf("read note file: %w", err)
	}

	// Stat the file to get its original permissions
	info, err := os.Stat(n.Path)
	if err != nil {
		return fmt.Errorf("stat note file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	start, end := findNotesSection(lines)
	if start == 0 {
		return fmt.Errorf("%w", ErrNotesSectionNotFound)
	}

	// Insert the new text
	newLines := slices.Clone(lines[:end])
	newLines = append(newLines, text, "")
	newLines = append(newLines, lines[end:]...)

	// Write back to the file with the original permissions
	if err := os.WriteFile(n.Path, []byte(strings.Join(newLines, "\n")), info.Mode()); err != nil {
		return fmt.Errorf("write note file: %w", err)
	}

	return nil
}

// findNotesSection returns the start and end indices of the notes section in the lines slice.
func findNotesSection(lines []string) (int, int) {
	var start, end int

	for i, line := range lines {
		if strings.HasPrefix(line, "## Notes") {
			start = i
		} else if start > 0 && strings.HasPrefix(line, "## ") {
			end = i

			break
		}
	}

	if end == 0 {
		end = len(lines)
	}

	return start, end
}
