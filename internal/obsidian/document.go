package obsidian

import (
	"fmt"
	"os"
	"strings"
)

// Document represents a loaded markdown daily note.
// Lines is the file split on '\n' (the final newline is excluded when present).
type Document struct {
	Path  string
	Mode  os.FileMode
	Lines []string
	dirty bool
}

// Section is a view over a header section in a Document.
// Sections follow the invariant: header, a blank line, body (0+ lines), optional trailing blank.
// Start is the header line index; End is exclusive (one past last line of section).
//
// Start points to header line index; End is exclusive (one past last line of section).
type Section struct {
	Doc    *Document
	Header string
	Start  int
	End    int
}

// findSection returns the start and end indices of the section whose header matches exactly header.
// Mirrors previous implementation semantics: start==0 indicates not found.
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

// OpenDocument loads an existing file path into a Document.
func OpenDocument(path string) (*Document, error) {
	data, err := os.ReadFile(path) // #nosec G304 // file inclusion via variable
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	return &Document{
		Path:  path,
		Mode:  info.Mode(),
		Lines: strings.Split(string(data), "\n"),
		dirty: false,
	}, nil
}

// Save writes the document back if dirty or if force is true.
func (d *Document) Save(force bool) error {
	if !d.dirty && !force {
		return nil
	}

	content := strings.Join(d.Lines, "\n")
	if err := os.WriteFile(d.Path, []byte(content), d.Mode); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	d.dirty = false

	return nil
}

// Section returns the section with given header (e.g. "## Todo").
func (d *Document) Section(header string) (*Section, error) {
	start, end := findSection(d.Lines, header)
	if start == 0 { // not found (per existing convention start==0 means missing)
		return nil, fmt.Errorf("%w: %s", ErrSectionNotFound, strings.TrimPrefix(header, "## "))
	}

	return &Section{
		Doc:    d,
		Header: header,
		Start:  start,
		End:    end,
	}, nil
}

// BodyLines returns the body lines excluding header and mandatory blank line after header.
func (s *Section) BodyLines() []string {
	// Allow zero or more blank lines after header when reading
	bodyStart := s.Start + 1

	for bodyStart < s.End && strings.TrimSpace(s.Doc.Lines[bodyStart]) == "" {
		bodyStart++
	}

	if bodyStart >= s.End {
		return []string{}
	}

	// If last line is blank and body has at least one line, drop trailing blank.
	lines := s.Doc.Lines[bodyStart:s.End]
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// Body returns the body as a single string joined by newlines.
func (s *Section) Body() string {
	return strings.Join(s.BodyLines(), "\n")
}

// SetBody replaces the body ensuring section newline invariants.
func (s *Section) SetBody(body string) {
	rawLines := []string{}
	if strings.TrimSpace(body) != "" {
		rawLines = strings.Split(body, "\n")
	}

	// Build new section lines: header, blank, body..., optional trailing blank if body non-empty.
	newSec := []string{s.Header, ""}
	if len(rawLines) > 0 {
		newSec = append(newSec, rawLines...)
		newSec = append(newSec, "")
	}

	// Replace in document.
	before := s.Doc.Lines[:s.Start]
	after := s.Doc.Lines[s.End:]

	newLines := make([]string, 0, len(before)+len(newSec)+len(after))
	newLines = append(newLines, before...)
	newLines = append(newLines, newSec...)
	newLines = append(newLines, after...)

	s.End = s.Start + len(newSec)

	s.Doc.Lines = newLines
	s.Doc.dirty = true
}

// AppendLine appends a line to the body (creating body if empty) and maintains invariants.
func (s *Section) AppendLine(line string) {
	body := s.BodyLines()
	body = append(body, line)

	s.SetBody(strings.Join(body, "\n"))
}
