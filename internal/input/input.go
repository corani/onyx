package input

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	// ErrInputCanceled is returned when the user cancels the input.
	ErrInputCanceled = errors.New("input canceled")
	// ErrUnexpectedModelType is returned when the model type is not what was expected.
	ErrUnexpectedModelType = errors.New("unexpected model type")
)

type noteInputModel struct {
	textarea textarea.Model
	done     bool
	canceled bool
}

func (m noteInputModel) Init() tea.Cmd {
	return textarea.Blink
}

//nolint:ireturn
func (m noteInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Only one type in switch, so use if statement
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+s":
			m.done = true

			return m, tea.Quit
		case "esc":
			m.canceled = true

			return m, tea.Quit
		}
	}

	var cmd tea.Cmd

	m.textarea, cmd = m.textarea.Update(msg)

	return m, cmd
}

func (m noteInputModel) View() string {
	return m.textarea.View() + "\n(Press Ctrl+S to save, Esc to cancel)"
}

func PromptForText(prompt string) (string, error) {
	const (
		defaultWidth  = 80
		defaultHeight = 10
	)

	textareaModel := textarea.New()
	textareaModel.Placeholder = prompt + " (Ctrl+S to save, Esc to cancel)"
	textareaModel.Focus()
	textareaModel.CharLimit = 0
	textareaModel.SetWidth(defaultWidth)
	textareaModel.SetHeight(defaultHeight)

	m := noteInputModel{textarea: textareaModel, done: false, canceled: false}
	program := tea.NewProgram(m)

	result, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("input prompt failed: %w", err)
	}

	//nolint:varnamelen
	fm, ok := result.(noteInputModel)
	if !ok {
		return "", fmt.Errorf("%w: %T", ErrUnexpectedModelType, result)
	}

	if fm.canceled {
		return "", ErrInputCanceled
	}

	return fm.textarea.Value(), nil
}
