package markdown

import (
	"io"

	"github.com/charmbracelet/glamour"
)

func Render(w io.Writer, md string) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return err
	}

	out, err := renderer.Render(md)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(out))

	return err
}
