package obsidian

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

var (
	ErrTodoNotFound        = errors.New("no todo item matched substring")
	ErrTodoMultipleMatches = errors.New("multiple todo items matched substring")
	ErrEmptyTodoText       = errors.New("empty todo text")
	ErrEmptyTodoMatch      = errors.New("empty match substring")
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

// List returns the Todo section markdown (including header).
func (t *Todo) List() string {
	body := t.Sec.Body()
	if body != "" {
		body = "\n" + body + "\n"
	}

	return t.Sec.Header + "\n" + body
}

// Add inserts a new todo optionally under a parent matched by substring. Case-insensitive.
// Parent may be blank for top-level.
func (t *Todo) Add(text, parentSubstring string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return ErrEmptyTodoText
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

// SetStatus toggles the checkbox for a single matched todo (substring match, case-insensitive).
func (t *Todo) SetStatus(substring string, done bool) error {
	substring = strings.TrimSpace(substring)
	if substring == "" {
		return ErrEmptyTodoMatch
	}

	sec, err := t.Doc.Section("## Todo")
	if err != nil {
		return err
	}

	body := sec.BodyLines()

	matchIdx, _, err := findSingleMatchBody(body, substring)
	if err != nil {
		return err
	}

	line := body[matchIdx]

	if done && strings.Contains(line, "- [ ] ") {
		body[matchIdx] = replaceFirst(line, "- [ ] ", "- [x] ")
	} else if !done && strings.Contains(line, "- [x] ") {
		body[matchIdx] = replaceFirst(line, "- [x] ", "- [ ] ")
	}

	sec.SetBody(strings.Join(body, "\n"))

	return t.Doc.Save(false)
}

// loadSection reads file, splits to lines and locates todo section.
// legacy loadSection removed; Section abstraction used instead

func (t *Todo) computeInsertionBody(body []string, text, parentSubstring string) (int, string, error) {
	insertIdx := len(body)
	newLine := "- [ ] " + text

	if parentSubstring == "" {
		return insertIdx, newLine, nil
	}

	parentIdx, parentDepth, err := findSingleMatchBody(body, parentSubstring)
	if err != nil {
		return 0, "", err
	}

	blockEnd := parentIdx + 1
	for blockEnd < len(body) {
		line := body[blockEnd]

		if !isTodoLine(line) {
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
func findSingleMatchBody(body []string, needle string) (int, int, error) {
	needleLower := strings.ToLower(needle)

	var matches []int

	for index, line := range body {
		if !isTodoLine(line) {
			continue
		}

		text := extractTodoText(line)
		if strings.Contains(strings.ToLower(text), needleLower) {
			matches = append(matches, index)
		}
	}

	if len(matches) == 0 {
		return 0, 0, fmt.Errorf("%w: %q", ErrTodoNotFound, needle)
	}

	const maxLen = 160

	if len(matches) > 1 {
		for choiceIdx, idx := range matches {
			text := extractTodoText(body[idx])
			if len(text) > maxLen {
				text = text[:maxLen-3] + "..."
			}

			log.Error("todo multiple match", "choice", strconv.Itoa(choiceIdx+1), "text", text)
		}

		return 0, 0, ErrTodoMultipleMatches
	}

	idx := matches[0]
	depth := countLeadingTabs(body[idx])

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
