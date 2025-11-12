package obsidian

import (
    "errors"
    "fmt"
    "os"
    "strings"
)

var (
    // ErrDiarySectionNotFound is returned when the daily note lacks the required diary section.
    ErrDiarySectionNotFound = errors.New("one line section not found in note")
)

// Diary manages the `## One Line` section of a daily note.
type Diary struct {
    Note *Note
}

func NewDiary(note *Note) *Diary { //nolint:revive
    return &Diary{Note: note}
}

// Get returns the body of the diary section (without the header). Leading and trailing
// blank lines are stripped. A whitespace-only body is returned as an empty string.
func (d *Diary) Get() (string, error) {
    data, err := os.ReadFile(d.Note.Path)
    if err != nil {
        return "", fmt.Errorf("read note file: %w", err)
    }

    lines := strings.Split(string(data), "\n")
	
    start, end := findSection(lines, "## One Line")
    if start == 0 { // section missing
        return "", fmt.Errorf("%w", ErrDiarySectionNotFound)
    }

    bodyLines := make([]string, 0)
    if start+1 < end { // there is at least one line after the header
        bodyLines = append(bodyLines, lines[start+1:end]...)
    }

    // Strip leading/trailing blank lines
    for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[0]) == "" {
        bodyLines = bodyLines[1:]
    }

    for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[len(bodyLines)-1]) == "" {
        bodyLines = bodyLines[:len(bodyLines)-1]
    }

    if len(bodyLines) == 0 { // empty or whitespace-only
        return "", nil
    }

    return strings.Join(bodyLines, "\n"), nil
}

// Set replaces the entire diary section with the provided body. The section is written as:
//   ## One Line
//   <blank line>
//   <body>
//   <blank line> (only if body is non-empty)
// If body is empty (or whitespace), only a single blank line after the header is written.
// The user's body is preserved verbatim (no wrapping or trimming).
func (d *Diary) Set(body string) error {
    data, err := os.ReadFile(d.Note.Path)
    if err != nil {
        return fmt.Errorf("read note file: %w", err)
    }

    info, err := os.Stat(d.Note.Path)
    if err != nil {
        return fmt.Errorf("stat note file: %w", err)
    }

    lines := strings.Split(string(data), "\n")

    start, end := findSection(lines, "## One Line")
    if start == 0 { // section missing
        return fmt.Errorf("%w", ErrDiarySectionNotFound)
    }

    trimmed := strings.TrimSpace(body)

    var newSection []string

    newSection = append(newSection, "## One Line")
    newSection = append(newSection, "") // single blank line after header always

    if trimmed != "" { // non-empty body; preserve verbatim lines
        bodyLines := strings.Split(body, "\n")
        newSection = append(newSection, bodyLines...)
        newSection = append(newSection, "") // trailing blank line after body
    }

    // Reconstruct file: replace original section block with newSection
    out := make([]string, 0, len(lines)-((end-start)))
    out = append(out, lines[:start]...)
    out = append(out, newSection...)
    out = append(out, lines[end:]...)

    if err := os.WriteFile(d.Note.Path, []byte(strings.Join(out, "\n")), info.Mode()); err != nil {
        return fmt.Errorf("write note file: %w", err)
    }

    return nil
}
