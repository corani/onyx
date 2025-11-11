package obsidian

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

var (
	ErrTodoSectionNotFound = errors.New("todo section not found in note")
	ErrTodoNotFound        = errors.New("no todo item matched substring")
	ErrTodoMultipleMatches = errors.New("multiple todo items matched substring")
	ErrEmptyTodoText       = errors.New("empty todo text")
	ErrEmptyTodoMatch      = errors.New("empty match substring")
)

// Todos manages the `## Todo` section of a daily note.
type Todos struct {
	Note *Note
}

func NewTodos(note *Note) *Todos { //nolint:revive
	return &Todos{Note: note}
}

// List returns the Todo section markdown (including header) or empty string if missing.
func (t *Todos) List() (string, error) {
	data, err := os.ReadFile(t.Note.Path)
	if err != nil {
		return "", fmt.Errorf("read note file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	
	start, end := findSection(lines, "## Todo")
	if start == 0 { // missing
		return "", nil
	}

	return strings.Join(lines[start:end], "\n"), nil
}

// Add inserts a new todo optionally under a parent matched by substring. Case-insensitive.
// Parent may be blank for top-level.
func (t *Todos) Add(text, parentSubstring string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return ErrEmptyTodoText
	}

	lines, info, sectionStart, sectionEnd, err := t.loadSection()
	if err != nil {
		return err
	}

	insertIdx, newLine, err := t.computeInsertion(lines, sectionStart, sectionEnd, text, parentSubstring)
	if err != nil {
		return err
	}

	updated := make([]string, 0, len(lines)+1)
	updated = append(updated, lines[:insertIdx]...)
	updated = append(updated, newLine)
	updated = append(updated, lines[insertIdx:]...)

	if err := os.WriteFile(t.Note.Path, []byte(strings.Join(updated, "\n")), info.Mode()); err != nil {
		return fmt.Errorf("write note file: %w", err)
	}

	return nil
}

// SetStatus toggles the checkbox for a single matched todo (substring match, case-insensitive).
func (t *Todos) SetStatus(substring string, done bool) error {
	substring = strings.TrimSpace(substring)
	if substring == "" {
		return ErrEmptyTodoMatch
	}

	lines, info, sectionStart, sectionEnd, err := t.loadSection()
	if err != nil {
		return err
	}

	matchIdx, _, err := findSingleMatch(lines, sectionStart+1, sectionEnd, substring)
	if err != nil {
		return err
	}

	line := lines[matchIdx]
	if done && strings.Contains(line, "- [ ] ") {
		lines[matchIdx] = replaceFirst(line, "- [ ] ", "- [x] ")
	} else if !done && strings.Contains(line, "- [x] ") {
		lines[matchIdx] = replaceFirst(line, "- [x] ", "- [ ] ")
	}

	if err := os.WriteFile(t.Note.Path, []byte(strings.Join(lines, "\n")), info.Mode()); err != nil {
		return fmt.Errorf("write note file: %w", err)
	}

	return nil
}

// loadSection reads file, splits to lines and locates todo section.
func (t *Todos) loadSection() ([]string, os.FileInfo, int, int, error) {
	data, err := os.ReadFile(t.Note.Path)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("read note file: %w", err)
	}

	info, err := os.Stat(t.Note.Path)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("stat note file: %w", err)
	}

	lines := strings.Split(string(data), "\n")

	sectionStart, sectionEnd := findSection(lines, "## Todo")
	if sectionStart == 0 {
		return nil, nil, 0, 0, fmt.Errorf("%w", ErrTodoSectionNotFound)
	}

	return lines, info, sectionStart, sectionEnd, nil
}

func (t *Todos) computeInsertion(
	lines []string,
	sectionStart, sectionEnd int,
	text, parentSubstring string,
) (int, string, error) {
	insertIdx := sectionEnd

	newLine := "- [ ] " + text
	if parentSubstring == "" {
		return insertIdx, newLine, nil
	}

	parentIdx, parentDepth, err := findSingleMatch(lines, sectionStart+1, sectionEnd, parentSubstring)
	if err != nil {
		return 0, "", err
	}

	blockEnd := parentIdx + 1
	for blockEnd < sectionEnd {
		line := lines[blockEnd]
		if !isTodoLine(line) { // allow blank or non-todo lines inside section
			blockEnd++

			continue
		}

		depth := countLeadingTabs(line)
		if depth <= parentDepth {
			break
		}

		blockEnd++
	}

	return blockEnd, strings.Repeat("\t", parentDepth+1) + "- [ ] " + text, nil
}

// findSingleMatch finds exactly one matching todo line by case-insensitive substring.
// Returns index and depth.
func findSingleMatch(lines []string, start, end int, needle string) (int, int, error) {
	needleLower := strings.ToLower(needle)

	var matches []int

	for lineIdx := start; lineIdx < end; lineIdx++ { // explicit name for clarity
		line := lines[lineIdx]
		if !isTodoLine(line) {
			continue
		}

		text := extractTodoText(line)
		if strings.Contains(strings.ToLower(text), needleLower) {
			matches = append(matches, lineIdx)
		}
	}

	if len(matches) == 0 {
		return 0, 0, fmt.Errorf("%w: %q", ErrTodoNotFound, needle)
	}

	if len(matches) > 1 {
		for choiceIdx, lineIdx := range matches {
			text := extractTodoText(lines[lineIdx])

			const maxLen = 160

			if len(text) > maxLen {
				text = text[:maxLen-3] + "..."
			}

			log.Error("todo multiple match", "choice", strconv.Itoa(choiceIdx+1), "text", text)
		}

		return 0, 0, ErrTodoMultipleMatches
	}

	idx := matches[0]
	depth := countLeadingTabs(lines[idx])

	return idx, depth, nil
}

func isTodoLine(line string) bool {
	line = strings.TrimRight(line, "\r")
	trimmed := strings.TrimLeft(line, "\t")

	return strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ")
}

func countLeadingTabs(line string) int {
	count := 0

	for _, ch := range line {
		if ch == '\t' {
			count++

			continue
		}

		break
	}

	return count
}

func extractTodoText(line string) string {
	trimmed := strings.TrimLeft(line, "\t")

	if rest, ok := strings.CutPrefix(trimmed, "- [ ] "); ok {
		return strings.TrimSpace(rest)
	}

	if rest, ok := strings.CutPrefix(trimmed, "- [x] "); ok {
		return strings.TrimSpace(rest)
	}

	return ""
}

func replaceFirst(subject, old, replacement string) string {
	idx := strings.Index(subject, old)
	if idx == -1 {
		return subject
	}

	return subject[:idx] + replacement + subject[idx+len(old):]
}
