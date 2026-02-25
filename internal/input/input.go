package input

import (
	"errors"
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

var (
	// ErrInputCanceled is returned when the user cancels the input.
	ErrInputCanceled = errors.New("input canceled")

	// ErrUnexpectedModelType is returned when the model type is not what was expected.
	ErrUnexpectedModelType = errors.New("unexpected model type")

	// ErrNoChoices is returned when no choices were provided to the selector.
	ErrNoChoices = errors.New("no choices provided")

	// ErrUnexpectedSelectionModel indicates the selection UI returned an unexpected model.
	ErrUnexpectedSelectionModel = errors.New("unexpected selection model result")
)

type noteInputModel struct {
	textarea textarea.Model
	done     bool
	canceled bool
}

func (m noteInputModel) Init() tea.Cmd {
	return nil
}

//nolint:ireturn
func (m noteInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Only one type in switch, so use if statement
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		//nolint:exhaustive
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

func (m noteInputModel) View() tea.View {
	return tea.NewView(m.textarea.View() + "\n(Press Ctrl+S to save, Esc to cancel)")
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

// PromptForTextWithInitial shows a multiline textarea pre-filled with the provided initial value.
// Behavior (save / cancel) matches PromptForText.
func PromptForTextWithInitial(prompt, initial string) (string, error) {
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
	textareaModel.SetValue(initial)

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

// -------------------- Single-line input --------------------

type lineInputModel struct {
	input    textinput.Model
	done     bool
	canceled bool
}

func (m lineInputModel) Init() tea.Cmd {
	return nil
}

//nolint:ireturn
func (m lineInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		//nolint:exhaustive
		switch keyMsg.String() {
		case "enter":
			m.done = true

			return m, tea.Quit
		case "esc":
			m.canceled = true

			return m, tea.Quit
		}
	}

	var cmd tea.Cmd

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m lineInputModel) View() tea.View {
	return tea.NewView(m.input.View() + "\n(Enter to submit, Esc to cancel)")
}

// PromptForLine shows a single line input using Bubbles textinput.
func PromptForLine(prompt string) (string, error) {
	const defaultWidth = 60

	ti := textinput.New()
	ti.Placeholder = prompt
	ti.Focus()
	ti.CharLimit = 0
	ti.SetWidth(defaultWidth)
	m := lineInputModel{input: ti, done: false, canceled: false}
	program := tea.NewProgram(m)

	result, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("input prompt failed: %w", err)
	}

	finalModel, ok := result.(lineInputModel)
	if !ok {
		return "", fmt.Errorf("%w: %T", ErrUnexpectedModelType, result)
	}

	if finalModel.canceled {
		return "", ErrInputCanceled
	}

	return finalModel.input.Value(), nil
}

// -------------------- Selection list --------------------

type listItem string

func (i listItem) Title() string       { return string(i) }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return string(i) }

// PromptForSelection presents an interactive filterable list and returns the
// selected index from the provided choices slice. If the user cancels, returns ErrInputCanceled.
func PromptForSelection(prompt string, choices []string) (int, error) {
	if len(choices) == 0 {
		return -1, ErrNoChoices
	}

	items := make([]list.Item, len(choices))
	for i, c := range choices {
		items[i] = listItem(c)
	}

	const defaultWidth = 80

	height := len(choices)

	// Compute a taller height to show more items. Clamp between 10 and 24.
	//nolint:mnd
	{
		height += 6

		if height < 10 {
			height = 10
		}

		if height > 24 {
			height = 24
		}
	}

	// use a compact delegate so each item occupies a single line
	picker := list.New(items, compactDelegate{}, defaultWidth, height)
	picker.Title = prompt
	picker.SetShowStatusBar(false)
	picker.SetShowHelp(false)
	picker.SetFilteringEnabled(true)
	picker.SetShowPagination(false)
	m := selModel{l: picker, canceled: false, chosen: -1}

	program := tea.NewProgram(&m)

	result, err := program.Run()
	if err != nil {
		return -1, fmt.Errorf("selection prompt failed: %w", err)
	}

	// The program may return our selModel or, depending on how the list handles
	// certain keys, it might return the underlying list.Model. Handle both.
	if final, ok := result.(*selModel); ok {
		if final.canceled || final.chosen < 0 {
			return -1, ErrInputCanceled
		}

		return final.chosen, nil
	}

	return -1, fmt.Errorf("%w: %T", ErrUnexpectedSelectionModel, result)
}

type selModel struct {
	l        list.Model
	canceled bool
	chosen   int
}

func (m *selModel) Init() tea.Cmd { return nil }

//nolint:ireturn
func (m *selModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		//nolint:exhaustive
		switch keyMsg.String() {
		case "enter":
			m.chosen = m.l.Index()

			return m, tea.Quit
		case "esc":
			m.canceled = true

			return m, tea.Quit
		}
	}

	lm, cmd := m.l.Update(msg)
	m.l = lm

	return m, cmd
}

func (m *selModel) View() tea.View { return tea.NewView(m.l.View()) }

// compactDelegate renders each item on a single line with minimal spacing.
type compactDelegate struct{}

func (compactDelegate) Height() int { return 1 }

func (compactDelegate) Spacing() int { return 0 }

func (compactDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (compactDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	var option string

	if li, ok := item.(listItem); ok {
		option = string(li)
	} else {
		option = item.FilterValue()
	}

	if index == model.Index() {
		_, _ = fmt.Fprint(writer, "> ")
	} else {
		_, _ = fmt.Fprint(writer, "  ")
	}

	_, _ = fmt.Fprint(writer, option)
}
