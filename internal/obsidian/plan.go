package obsidian

import (
    "errors"
    "fmt"
    "os"
    "regexp"
    "sort"
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
    timePattern                 = regexp.MustCompile(
        `^(?P<start>[0-2][0-9]:[0-5][0-9])(?:-(?P<end>[0-2][0-9]:[0-5][0-9]))?$`,
    )
)

type Planner struct {
    Note *Note
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

    targetIdx := findPlannerLineIndex(lines, sectionStart, sectionEnd, token)
    if targetIdx == -1 {
        return fmt.Errorf("%w: %s", ErrPlannerEntryNotFound, token)
    }

    line := lines[targetIdx]
    if done && strings.HasPrefix(line, "- [ ] ") {
        lines[targetIdx] = strings.Replace(line, "- [ ] ", "- [x] ", 1)
    } else if !done && strings.HasPrefix(line, "- [x] ") {
        lines[targetIdx] = strings.Replace(line, "- [x] ", "- [ ] ", 1)
    }

    if err := os.WriteFile(p.Note.Path, []byte(strings.Join(lines, "\n")), info.Mode()); err != nil {
        return fmt.Errorf("write note file: %w", err)
    }

    return nil
}

func findPlannerLineIndex(lines []string, sectionStart, sectionEnd int, token string) int {
    for idx := sectionStart + 1; idx < sectionEnd; idx++ {
        line := strings.TrimSpace(lines[idx])
        if !strings.HasPrefix(line, "- [ ] ") && !strings.HasPrefix(line, "- [x] ") {
            continue
        }

        rest := line[6:]
        
        fields := strings.Fields(rest)
        if len(fields) == 0 {
            continue
        }

        timeToken := fields[0]
        if timeToken == token {
            return idx
        }
    }

    return -1
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
