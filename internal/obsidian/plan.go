package obsidian

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

var (
	ErrEmptyPlannerEntry           = errors.New("empty planner entry text")
	ErrPlannerEntryNotFound        = errors.New("planner entry not found for token")
	ErrInvalidStartTimeFormat      = errors.New("invalid start time format")
	ErrInvalidEndTimeFormat        = errors.New("invalid end time format")
	ErrEndTimeMustBeAfterStart     = errors.New("end time must be after start time")
	ErrEmptyPlannerMatchToken      = errors.New("empty planner match token")
	ErrMultiplePlannerEntriesMatch = errors.New("multiple planner entries match")
	ErrInvalidNotePath             = errors.New("invalid note path")
)

var (
	timePattern = regexp.MustCompile(
		`^(?P<start>[0-2][0-9]:[0-5][0-9])(?:-(?P<end>[0-2][0-9]:[0-5][0-9]))?$`)
)

type Plan struct {
	Doc *Document
	Sec *Section
}

// readPlannerSection reads the planner section lines from the note file.
// Deprecated direct section reader removed; use Document + Section instead.

func NewPlan(doc *Document) (*Plan, error) {
	sec, err := doc.Section("## Day Planner")
	if err != nil {
		return nil, err
	}

	return &Plan{
		Doc: doc,
		Sec: sec,
	}, nil
}

// List returns the Day Planner section markdown (including header).
func (p *Plan) List() string {
	// Compose full section text
	body := p.Sec.Body()

	if body != "" {
		body = "\n" + body + "\n"
	}

	return p.Sec.Header + "\n" + body
}

// Add inserts a new planner entry maintaining chronological order.
func (p *Plan) Add(startStr, endStr, text string) error {
	if text == "" {
		return ErrEmptyPlannerEntry
	}

	startTime, _, err := parseTimes(startStr, endStr)
	if err != nil {
		return fmt.Errorf("parse time(s): %w", err)
	}

	entries := p.Sec.BodyLines()
	insertAt := findInsertionIndex(entries, startTime, startStr, endStr, text)
	newLine := "- [ ] " + formatEntry(startStr, endStr, text)
	updatedEntries := buildUpdatedEntries(entries, insertAt, newLine)

	p.Sec.SetBody(strings.Join(updatedEntries, "\n"))

	return p.Doc.Save(false)
}

const minutesPerHour = 60

func findInsertionIndex(entries []string, startTime time.Time, startStr, endStr, text string) int {
	insertAt := len(entries)
	newStartMinutes := startTime.Hour()*minutesPerHour + startTime.Minute()

	for idx, lineVal := range entries {
		parsed, ok := parsePlannerLine(lineVal)
		if !ok {
			continue
		}

		existingMinutes := parsed.start.Hour()*minutesPerHour + parsed.start.Minute()
		if newStartMinutes < existingMinutes {
			insertAt = idx

			break
		}

		if parsed.end != nil { // overlap check
			endMinutes := parsed.end.Hour()*minutesPerHour + parsed.end.Minute()
			if newStartMinutes >= existingMinutes && newStartMinutes < endMinutes {
				log.Warn(
					"Overlapping planner entry",
					"existing", lineVal,
					"new", formatEntry(startStr, endStr, text),
				)
			}
		}
	}

	return insertAt
}

func buildUpdatedEntries(entries []string, insertAt int, newLine string) []string {
	result := make([]string, 0, len(entries)+1)
	result = append(result, entries[:insertAt]...)
	result = append(result, newLine)
	result = append(result, entries[insertAt:]...)

	return result
}

// SetStatus toggles the checkbox for the entry matching token (exact time or time range string).
func (p *Plan) SetStatus(token string, done bool) error {
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "" {
		return ErrEmptyPlannerMatchToken
	}

	sec, err := p.Doc.Section("## Day Planner")
	if err != nil {
		return err
	}

	body := sec.BodyLines()

	// reconstruct lines slice for matching: mimic previous entries (no header, potential blank? already normalized)
	// The previous code included the header line and worked over full file indices; now we work within body slice.
	// We'll search in body for matching token.
	idx, err := findPlannerLineIndexFromBody(body, token)
	if err != nil {
		return err
	}

	body[idx] = togglePlannerCheckbox(body[idx], done)
	sec.SetBody(strings.Join(body, "\n"))

	return p.Doc.Save(false)
}

// togglePlannerCheckbox toggles the checkbox for a planner line.
func togglePlannerCheckbox(line string, done bool) string {
	if done && strings.HasPrefix(line, "- [ ] ") {
		return strings.Replace(line, "- [ ] ", "- [x] ", 1)
	}

	if !done && strings.HasPrefix(line, "- [x] ") {
		return strings.Replace(line, "- [x] ", "- [ ] ", 1)
	}

	return line
}

// findPlannerLineIndexFromBody finds exactly one matching planner entry by case-insensitive substring.
// Returns index or error if not found or multiple matches.
func findPlannerLineIndexFromBody(body []string, needle string) (int, error) {
	const minFields = 2

	lower := strings.ToLower(needle)

	var matches []int

	for index, line := range body {
		entryText, valid := extractPlannerEntryText(line, minFields)
		if !valid {
			continue
		}

		if strings.Contains(strings.ToLower(entryText), lower) {
			matches = append(matches, index)
		}
	}

	if len(matches) == 0 {
		return 0, fmt.Errorf("%w: %q", ErrPlannerEntryNotFound, needle)
	}

	if len(matches) > 1 {
		logPlannerMultipleMatchesBody(body, matches)

		return 0, ErrMultiplePlannerEntriesMatch
	}

	return matches[0], nil
}

// extractPlannerEntryText extracts the entry text from a planner line.
func extractPlannerEntryText(line string, minFields int) (string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "- [ ] ") && !strings.HasPrefix(line, "- [x] ") {
		return "", false
	}

	rest := line[6:]

	fields := strings.Fields(rest)
	if len(fields) < minFields {
		return "", false
	}

	return strings.Join(fields[1:], " "), true
}

// logPlannerMultipleMatches logs multiple planner matches for debugging.
func logPlannerMultipleMatchesBody(body []string, matches []int) {
	const (
		maxLen    = 160
		minFields = 2
	)

	for choiceIdx, idx := range matches {
		entryText, _ := extractPlannerEntryText(body[idx], minFields)
		if len(entryText) > maxLen {
			entryText = entryText[:maxLen-3] + "..."
		}

		log.Error("planner multiple match", "choice", strconv.Itoa(choiceIdx+1), "text", entryText)
	}
}

// Helper structures.
type parsedPlannerLine struct {
	start time.Time
	end   *time.Time
}

func parsePlannerLine(line string) (*parsedPlannerLine, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "- [") {
		return nil, false
	}

	idx := strings.Index(line, "]")
	if idx == -1 || idx+2 >= len(line) { // ] + space
		return nil, false
	}

	rest := strings.TrimSpace(line[idx+1:])

	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return nil, false
	}

	timeToken := fields[0]

	matches := timePattern.FindStringSubmatch(timeToken)
	if matches == nil {
		return nil, false
	}

	startVal := matches[1]
	endVal := matches[2]
	startParsed, _ := time.Parse("15:04", startVal)

	var endParsedPtr *time.Time

	if endVal != "" {
		parsedEnd, _ := time.Parse("15:04", endVal)
		endParsedPtr = &parsedEnd
	}

	return &parsedPlannerLine{start: startParsed, end: endParsedPtr}, true
}

func parseTimes(startStr, endStr string) (time.Time, *time.Time, error) {
	if !timePattern.MatchString(startStr) && !regexp.MustCompile(`^[0-2][0-9]:[0-5][0-9]$`).MatchString(startStr) {
		return time.Time{}, nil, fmt.Errorf("%w: %s", ErrInvalidStartTimeFormat, startStr)
	}

	startParsed, err := time.Parse("15:04", startStr)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("parse start: %w", err)
	}

	if endStr == "" {
		return startParsed, nil, nil
	}

	if !regexp.MustCompile(`^[0-2][0-9]:[0-5][0-9]$`).MatchString(endStr) {
		return time.Time{}, nil, fmt.Errorf("%w: %s", ErrInvalidEndTimeFormat, endStr)
	}

	endParsed, err := time.Parse("15:04", endStr)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("parse end: %w", err)
	}

	if !endParsed.After(startParsed) {
		return time.Time{}, nil, ErrEndTimeMustBeAfterStart
	}

	return startParsed, &endParsed, nil
}

func formatEntry(startStr, endStr, text string) string {
	timeToken := startStr
	if endStr != "" {
		timeToken += "-" + endStr
	}

	return timeToken + " " + text
}
