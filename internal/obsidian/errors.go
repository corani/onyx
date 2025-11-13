package obsidian

import "errors"

var (
	ErrSectionNotFound   = errors.New("section not found")
	ErrNotFound          = errors.New("item not found")
	ErrMultipleMatches   = errors.New("multiple items matched")
	ErrEmpty             = errors.New("empty value")
	ErrInvalidTimeFormat = errors.New("invalid time format")
	ErrInvalidTimeRange  = errors.New("invalid time range")
	ErrInvalidNotePath   = errors.New("invalid note path")
)
