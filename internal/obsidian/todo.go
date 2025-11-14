package obsidian

import (
	"fmt"
	"strings"

	"github.com/corani/onyx/internal/input"
)

// Todo manages the `## Todo` section of a daily note.
type Todo struct {
	Doc *Document
	Sec *Section
}

func NewTodo(doc *Document) (*Todo, error) {
	sec, err := doc.Section("## Todo")
	if err != nil {
		return nil, err
	}

	return &Todo{
		Doc: doc,
		Sec: sec,
	}, nil
}

// List returns the Todo section markdown, including the header.
func (t *Todo) List() string {
	body := t.Sec.Body()
	if body != "" {
		body = "\n" + body + "\n"
	}

	return t.Sec.Header + "\n" + body
}

// Add inserts a new todo, optionally under a parent matched by substring (case-insensitive).
// Parent may be blank to insert at top-level.
func (t *Todo) Add(text, parentSubstring string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("%w: %s", ErrEmpty, text)
	}

	bodyLines := t.Sec.BodyLines()

	insertIdx, newLine, err := t.computeInsertionBody(bodyLines, text, parentSubstring)
	if err != nil {
		return err
	}

	updated := append(append(append([]string{}, bodyLines[:insertIdx]...), newLine), bodyLines[insertIdx:]...)
	t.Sec.SetBody(strings.Join(updated, "\n"))

	return t.Doc.Save(false)
}

// SetStatus toggles the checkbox for a single matched todo (substring, case-insensitive).
func (t *Todo) SetStatus(substring string, done bool) error {
	substring = strings.TrimSpace(substring)
	body := t.Sec.BodyLines()

	// If substring empty, show full list to let user pick.
	matchesIdx, matches := FindCheckboxMatches(body, substring)

	if len(matchesIdx) == 0 {
		return ErrNotFound
	}

	var chosen int
	if len(matchesIdx) == 1 {
		chosen = 0
	} else {
		// present interactive selector
		choices := make([]string, len(matches))
		for i, cb := range matches {
			choices[i] = MatchSummary(cb)
		}

		sel, err := input.PromptForSelection("Choose todo to toggle:", choices)
		if err != nil {
			return fmt.Errorf("select todo: %w", err)
		}

		chosen = sel
	}

	if chosen < 0 || chosen >= len(matchesIdx) {
		return fmt.Errorf("%w: invalid selection", ErrNotFound)
	}

	targetIdx := matchesIdx[chosen]
	body[targetIdx] = ToggleLineCheckbox(body[targetIdx], done)

	t.Sec.SetBody(strings.Join(body, "\n"))

	return t.Doc.Save(false)
}

func (t *Todo) computeInsertionBody(body []string, text, parentSubstring string) (int, string, error) {
	insertIdx := len(body)
	item := NewCheckbox(false, text, 0)

	if parentSubstring == "" {
		return insertIdx, item.String(), nil
	}

	parentIdx, parentItem, err := FindSingleCheckboxMatch(body, parentSubstring)
	if err != nil {
		return 0, "", err
	}

	blockEnd := FindBlockEnd(body, parentIdx)

	item.Indent = parentItem.Indent + 1

	return blockEnd, item.String(), nil
}
