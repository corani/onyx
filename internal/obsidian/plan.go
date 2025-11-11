
package obsidian

import (
    "errors"
    "fmt"
    "os"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/charmbracelet/log"
)

var (
    ErrPlannerSectionNotFound   = errors.New("day planner section not found in note")
    ErrEmptyPlannerEntry        = errors.New("empty planner entry text")
    ErrPlannerEntryNotFound     = errors.New("planner entry not found for token")
    ErrInvalidStartTimeFormat   = errors.New("invalid start time format")
    ErrInvalidEndTimeFormat     = errors.New("invalid end time format")
    ErrEndTimeMustBeAfterStart  = errors.New("end time must be after start time")
    ErrEmptyPlannerMatchToken   = errors.New("empty planner match token")
    ErrMultiplePlannerEntriesMatch = errors.New("multiple planner entries match")
    ErrInvalidNotePath          = errors.New("invalid note path")
)

var (
    timePattern = regexp.MustCompile(
        `^(?P<start>[0-2][0-9]:[0-5][0-9])(?:-(?P<end>[0-2][0-9]:[0-5][0-9]))?$`)
)

type Planner struct {
    Note *Note
}

// readPlannerSection reads the planner section lines from the note file.
func readPlannerSection(notePath string) ([]string, int, int, os.FileInfo, error) {
    // gosec: validate notePath is not empty and is absolute
    if notePath == "" || !strings.HasPrefix(notePath, "/") {
        return nil, 0, 0, nil, ErrInvalidNotePath
    }

    // #nosec G304
    data, err := os.ReadFile(notePath)
    if err != nil {
        return nil, 0, 0, nil, fmt.Errorf("read note file: %w", err)
    }

    info, err := os.Stat(notePath)
    if err != nil {
        return nil, 0, 0, nil, fmt.Errorf("stat note file: %w", err)
    }

    lines := strings.Split(string(data), "\n")
    sectionStart, sectionEnd := findSection(lines, "## Day Planner")

    return lines, sectionStart, sectionEnd, info, nil
}

func NewPlanner(note *Note) *Planner { //nolint:revive
    return &Planner{Note: note}
}

// List returns the Day Planner section markdown (including header) or empty string if missing.
func (p *Planner) List() (string, error) {
    data, err := os.ReadFile(p.Note.Path)
    if err != nil {
        return "", fmt.Errorf("read note file: %w", err)
    }

    lines := strings.Split(string(data), "\n")
    start, end := findSection(lines, "## Day Planner")

    if start == 0 { // missing
        return "", nil
    }

    return strings.Join(lines[start:end], "\n"), nil
}

// Add inserts a new planner entry maintaining chronological order.
func (p *Planner) Add(startStr, endStr, text string) error {
    if text == "" {
        return ErrEmptyPlannerEntry
    }

    startTime, _, err := parseTimes(startStr, endStr)
    if err != nil {
        return fmt.Errorf("parse time(s): %w", err)
    }

    data, err := os.ReadFile(p.Note.Path)
    if err != nil {
        return fmt.Errorf("read note file: %w", err)
    }

    info, err := os.Stat(p.Note.Path)
    if err != nil {
        return fmt.Errorf("stat note file: %w", err)
    }

    lines := strings.Split(string(data), "\n")
    sectionStart, sectionEnd := findSection(lines, "## Day Planner")

    if sectionStart == 0 {
        return fmt.Errorf("%w", ErrPlannerSectionNotFound)
    }

    entries := lines[sectionStart+1 : sectionEnd]
    insertAt := findInsertionIndex(entries, startTime, startStr, endStr, text)
    newLine := "- [ ] " + formatEntry(startStr, endStr, text)

    updatedEntries := buildUpdatedEntries(entries, insertAt, newLine)

    out := make([]string, 0, len(lines)+1)
    out = append(out, lines[:sectionStart+1]...)
    out = append(out, updatedEntries...)
    out = append(out, lines[sectionEnd:]...)

    if err := os.WriteFile(p.Note.Path, []byte(strings.Join(out, "\n")), info.Mode()); err != nil {
        return fmt.Errorf("write note file: %w", err)
    }

    return nil
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
func (p *Planner) SetStatus(token string, done bool) error {
    token = strings.ToLower(strings.TrimSpace(token))
    if token == "" {
        return ErrEmptyPlannerMatchToken
    }

    lines, sectionStart, sectionEnd, info, err := readPlannerSection(p.Note.Path)
    if err != nil {
        return err
    }

    if sectionStart == 0 {
        return fmt.Errorf("%w", ErrPlannerSectionNotFound)
    }

    targetIdx, err := findPlannerLineIndexCI(lines, sectionStart, sectionEnd, token)
    if err != nil {
        return err
    }

    lines[targetIdx] = togglePlannerCheckbox(lines[targetIdx], done)

    if err := os.WriteFile(p.Note.Path, []byte(strings.Join(lines, "\n")), info.Mode()); err != nil {
        return fmt.Errorf("write note file: %w", err)
    }

    return nil
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

// findPlannerLineIndexCI finds exactly one matching planner entry by case-insensitive substring.
// Returns index or error if not found or multiple matches.
func findPlannerLineIndexCI(lines []string, sectionStart, sectionEnd int, needle string) (int, error) {
    const minFields = 2

    needleLower := strings.ToLower(needle)

    var matches []int

    for idx := sectionStart + 1; idx < sectionEnd; idx++ {
        entryText, valid := extractPlannerEntryText(lines[idx], minFields)
        if !valid {
            continue
        }

        if strings.Contains(strings.ToLower(entryText), needleLower) {
            matches = append(matches, idx)
        }
    }

    if len(matches) == 0 {
        return 0, fmt.Errorf("%w: %q", ErrPlannerEntryNotFound, needle)
    }

    if len(matches) > 1 {
        logPlannerMultipleMatches(lines, matches)

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
func logPlannerMultipleMatches(lines []string, matches []int) {
    const (
        maxLen = 160
     minFields = 2
    )

    for choiceIdx, idx := range matches {
        entryText, _ := extractPlannerEntryText(lines[idx], minFields)
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

// For any future operations that need sorting of arbitrary collected entries.
func sortByStart(lines []string) []string { //nolint:unused
    type entry struct {
        line  string
        start int
    }

    entriesSlice := make([]entry, 0, len(lines))
    for _, lineVal := range lines {
        parsed, ok := parsePlannerLine(lineVal)
        if !ok {
            continue
        }

        minutes := parsed.start.Hour()*minutesPerHour + parsed.start.Minute()
        entriesSlice = append(entriesSlice, entry{line: lineVal, start: minutes})
    }

    sort.SliceStable(entriesSlice, func(i, j int) bool { return entriesSlice[i].start < entriesSlice[j].start })

    out := make([]string, 0, len(entriesSlice))
    for _, e := range entriesSlice {
        out = append(out, e.line)
    }

    return out
}
