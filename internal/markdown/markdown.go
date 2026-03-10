package markdown

import (
	"fmt"
	"io"
	"os"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

const defaultWordWrap = 80

func Render(writer io.Writer, markdown string) error {
	style := "dark"
	if !lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
		style = "light"
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(style),
		glamour.WithWordWrap(defaultWordWrap),
	)
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}

	rendered, err := renderer.Render(markdown)
	if err != nil {
		return fmt.Errorf("render markdown: %w", err)
	}

	if _, err := lipgloss.Fprint(writer, rendered); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
