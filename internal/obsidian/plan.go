package obsidian

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

type Plan struct {
	Doc *Document
	Sec *Section
}

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

// List returns the Day Planner section markdown, including the header.
func (p *Plan) List() string {
	// Compose full section text
	body := p.Sec.Body()

	if body != "" {
		body = "\n" + body + "\n"
	}

	return p.Sec.Header + "\n" + body
}

// Add inserts a new planner entry, keeping entries in chronological order.
func (p *Plan) Add(startStr, endStr, text string) error {
	if text == "" {
		return fmt.Errorf("%w: %s", ErrEmpty, text)
	}

	startTime, endTime, err := ParseTimes(startStr, endStr)
	if err != nil {
		return fmt.Errorf("parse time(s): %w", err)
	}

	entries := p.Sec.BodyLines()
	insertAt := findInsertionIndex(entries, startTime, startStr, endStr, text)

	cb := NewCheckboxWithTime(false, text, 0, &startTime, endTime)

	updated := InsertLineAt(entries, insertAt, cb.String())

	p.Sec.SetBody(strings.Join(updated, "\n"))

	return p.Doc.Save(false)
}

func findInsertionIndex(entries []string, startTime time.Time, startStr, endStr, text string) int {
	insertAt := len(entries)
	newStartMinutes := startTime.Hour()*60 + startTime.Minute() //nolint:mnd

	for index, line := range entries {
		checkbox, ok := ParseCheckbox(line)
		if !ok || checkbox.Start == nil {
			continue
		}

		existingMinutes := checkbox.Start.Hour()*60 + checkbox.Start.Minute() //nolint:mnd
		if newStartMinutes < existingMinutes {
			insertAt = index

			break
		}

		if checkbox.End != nil { // overlap check
			endMinutes := checkbox.End.Hour()*60 + checkbox.End.Minute() //nolint:mnd
			if newStartMinutes >= existingMinutes && newStartMinutes < endMinutes {
				log.Warn("Overlapping planner entry",
					"existing", line,
					"new", formatEntry(startStr, endStr, text))
			}
		}
	}

	return insertAt
}

// SetStatus toggles the checkbox for the entry matching token (exact time or range string).
func (p *Plan) SetStatus(token string, done bool) error {
	token = strings.ToLower(strings.TrimSpace(token))
	if token == "" {
		return fmt.Errorf("%w: %s", ErrEmpty, token)
	}

	body := p.Sec.BodyLines()

	// reconstruct lines slice for matching: mimic previous entries (no header, potential blank? already normalized)
	// The previous code included the header line and worked over full file indices; now we work within body slice.
	// We'll search in body for matching token.
	lineIndex, _, err := FindSingleCheckboxMatch(body, token)
	if err != nil {
		return err
	}

	body[lineIndex] = ToggleLineCheckbox(body[lineIndex], done)
	p.Sec.SetBody(strings.Join(body, "\n"))

	return p.Doc.Save(false)
}

func formatEntry(startStr, endStr, text string) string {
	timeToken := startStr
	if endStr != "" {
		timeToken += "-" + endStr
	}

	return timeToken + " " + text
}
