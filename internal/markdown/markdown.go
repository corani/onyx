package markdown

import (
	"fmt"
	"io"

	"github.com/charmbracelet/glamour"
)

const defaultWordWrap = 80

func Render(writer io.Writer, markdown string) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(defaultWordWrap),
	)
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}

	rendered, err := renderer.Render(markdown)
	if err != nil {
		return fmt.Errorf("render markdown: %w", err)
	}

	if _, err := writer.Write([]byte(rendered)); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
