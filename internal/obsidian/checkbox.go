package obsidian

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

var (
	timePattern = regexp.MustCompile(`^(?P<start>[0-2][0-9]:[0-5][0-9])(?:-(?P<end>[0-2][0-9]:[0-5][0-9]))?$`)
)

// Checkbox represents a markdown checkbox list item with nesting and optional times.
//
//nolint:recvcheck
type Checkbox struct {
	Checked bool       // true if [x], false if [ ]
	Text    string     // the item text (excluding checkbox marker and time token)
	Indent  int        // number of leading tabs (nesting)
	Start   *time.Time // optional start time parsed from leading token
	End     *time.Time // optional end time parsed from leading token
}

// NewCheckbox constructs a Checkbox without times.
func NewCheckbox(checked bool, text string, indent int) Checkbox {
	return Checkbox{
		Checked: checked,
		Text:    text,
		Indent:  indent,
		Start:   nil,
		End:     nil,
	}
}

// NewCheckboxWithTime constructs a Checkbox with start and optional end times.
func NewCheckboxWithTime(checked bool, text string, indent int, start *time.Time, end *time.Time) Checkbox {
	return Checkbox{
		Checked: checked,
		Text:    text,
		Indent:  indent,
		Start:   start,
		End:     end,
	}
}

// ParseTimeToken parses a token like "09:30" or "09:30-10:00" and returns
// parsed start and optional end times; the boolean indicates success.
func ParseTimeToken(token string) (*time.Time, *time.Time, bool) {
	matches := timePattern.FindStringSubmatch(token)
	if matches == nil {
		return nil, nil, false
	}

	startVal := matches[1]
	endVal := matches[2]

	startParsed, err := time.Parse("15:04", startVal)
	if err != nil {
		return nil, nil, false
	}

	var endParsedPtr *time.Time

	if endVal != "" {
		parsedEnd, err := time.Parse("15:04", endVal)
		if err == nil {
			endParsedPtr = &parsedEnd
		}
	}

	return &startParsed, endParsedPtr, true
}

// ParseTimes validates and parses start and end strings, returning parsed start and optional end.
func ParseTimes(startStr, endStr string) (time.Time, *time.Time, error) {
	if !timePattern.MatchString(startStr) && !regexp.MustCompile(`^[0-2][0-9]:[0-5][0-9]$`).MatchString(startStr) {
		return time.Time{}, nil, fmt.Errorf("%w: %s", ErrInvalidTimeFormat, startStr)
	}

	startParsed, err := time.Parse("15:04", startStr)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("parse start: %w", err)
	}

	if endStr == "" {
		return startParsed, nil, nil
	}

	if !regexp.MustCompile(`^[0-2][0-9]:[0-5][0-9]$`).MatchString(endStr) {
		return time.Time{}, nil, fmt.Errorf("%w: %s", ErrInvalidTimeFormat, endStr)
	}

	endParsed, err := time.Parse("15:04", endStr)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("parse end: %w", err)
	}

	if !endParsed.After(startParsed) {
		return time.Time{}, nil, fmt.Errorf("%w: %s", ErrInvalidTimeRange, endStr)
	}

	return startParsed, &endParsed, nil
}

// ParseCheckbox parses a markdown checkbox line into a Checkbox.
// Recognizes leading tab indentation, the checkbox marker, and an optional leading time token.
func ParseCheckbox(line string) (Checkbox, bool) {
	indent := 0
	for strings.HasPrefix(line, "\t") {
		indent++
		line = line[1:]
	}

	checked := false
	if strings.HasPrefix(line, "- [x] ") {
		checked = true
	} else if !strings.HasPrefix(line, "- [ ] ") {
		return NewCheckbox(false, "", 0), false
	}

	text := strings.TrimSpace(line[6:])

	// Attempt to parse a leading time token and remove it from the text.
	var (
		startPtr *time.Time
		endPtr   *time.Time
	)

	fields := strings.Fields(text)
	if len(fields) > 0 {
		if s, e, ok := ParseTimeToken(fields[0]); ok {
			startPtr = s
			endPtr = e
			text = strings.TrimSpace(strings.Join(fields[1:], " "))
		}
	}

	return NewCheckboxWithTime(checked, text, indent, startPtr, endPtr), true
}

// String formats a Checkbox back to a markdown line.
func (cb Checkbox) String() string {
	indent := strings.Repeat("\t", cb.Indent)

	box := "- [ ] "
	if cb.Checked {
		box = "- [x] "
	}

	if cb.Start != nil {
		timeToken := cb.Start.Format("15:04")
		if cb.End != nil {
			timeToken += "-" + cb.End.Format("15:04")
		}

		if cb.Text != "" {
			return indent + box + timeToken + " " + cb.Text
		}

		return indent + box + timeToken
	}

	return indent + box + cb.Text
}

// SetChecked updates the checked state.
func (cb *Checkbox) SetChecked(checked bool) {
	cb.Checked = checked
}

// FindSingleCheckboxMatch finds exactly one checkbox whose text contains needle (case-insensitive).
func FindSingleCheckboxMatch(lines []string, needle string) (int, Checkbox, error) {
	needle = strings.TrimSpace(strings.ToLower(needle))
	if needle == "" {
		return 0, Checkbox{}, ErrNotFound
	}

	var matches []int

	for index, line := range lines {
		checkbox, ok := ParseCheckbox(line)
		if !ok {
			continue
		}

		// Match against the rendered checkbox line (includes time token if present).
		if strings.Contains(strings.ToLower(checkbox.String()), needle) {
			matches = append(matches, index)
		}
	}

	if len(matches) == 0 {
		return 0, Checkbox{}, ErrNotFound
	}

	if len(matches) > 1 {
		logMultipleMatches(lines, matches, "checkbox")

		return 0, Checkbox{}, ErrMultipleMatches
	}

	idx := matches[0]
	c, _ := ParseCheckbox(lines[idx])

	return idx, c, nil
}

// ToggleLineCheckbox toggles the checkbox state on a single line string.
func ToggleLineCheckbox(line string, done bool) string {
	checkbox, ok := ParseCheckbox(line)
	if !ok {
		return line
	}

	checkbox.SetChecked(done)

	return checkbox.String()
}

// FindBlockEnd returns the index one past the end of the child block starting at startIdx.
func FindBlockEnd(lines []string, startIdx int) int {
	if startIdx < 0 || startIdx >= len(lines) {
		return startIdx
	}

	first, ok := ParseCheckbox(lines[startIdx])
	if !ok {
		return startIdx + 1
	}

	end := startIdx + 1
	for end < len(lines) {
		checkbox, ok := ParseCheckbox(lines[end])
		if !ok {
			end++

			continue
		}

		if checkbox.Indent <= first.Indent {
			break
		}

		end++
	}

	return end
}

// InsertLineAt inserts a line at idx into the slice.
func InsertLineAt(lines []string, idx int, line string) []string {
	if idx < 0 {
		idx = 0
	}

	if idx > len(lines) {
		idx = len(lines)
	}

	res := make([]string, 0, len(lines)+1)
	res = append(res, lines[:idx]...)
	res = append(res, line)
	res = append(res, lines[idx:]...)

	return res
}

// logMultipleMatches logs multiple matches for debugging.
func logMultipleMatches(lines []string, matches []int, ctx string) {
	const maxLen = 160

	for index, lineIndex := range matches {
		cb, _ := ParseCheckbox(lines[lineIndex])

		text := cb.Text
		if len(text) > maxLen {
			text = text[:maxLen-3] + "..."
		}

		log.Error(ctx+" multiple match", "choice", strconv.Itoa(index+1), "text", text)
	}
}
