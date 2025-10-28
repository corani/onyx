package input

import (
	"errors"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type noteInputModel struct {
	textarea textarea.Model
	done     bool
	canceled bool
}

func (m noteInputModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m noteInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
	ta := textarea.New()
	ta.Placeholder = prompt + " (Ctrl+S to save, Esc to cancel)"
	ta.Focus()
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(10)

	m := noteInputModel{textarea: ta}
	p := tea.NewProgram(m)

	result, err := p.Run()
	if err != nil {
		return "", err
	}

	fm := result.(noteInputModel)
	if fm.canceled {
		return "", errors.New("input canceled")
	}

	return fm.textarea.Value(), nil
}
