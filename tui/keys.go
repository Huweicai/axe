package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up             key.Binding
	Down           key.Binding
	Enter          key.Binding
	AltEnter       key.Binding
	Delete         key.Binding
	Archive        key.Binding
	Undo           key.Binding
	NoteEdit       key.Binding
	Star           key.Binding
	DeepSearch     key.Binding
	TabFilter      key.Binding
	Group          key.Binding
	ToggleArchived key.Binding
	ToggleTrash    key.Binding
	Quit           key.Binding
	Escape         key.Binding
	Backspace      key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	AltEnter: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "alt tool"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Archive: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "archive"),
	),
	Undo: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "restore"),
	),
	NoteEdit: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "note"),
	),
	Star: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "star"),
	),
	DeepSearch: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	TabFilter: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "filter"),
	),
	Group: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "group"),
	),
	ToggleArchived: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("^a", "show archived"),
	),
	ToggleTrash: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("^t", "show trash"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "clear/quit"),
	),
	Backspace: key.NewBinding(
		key.WithKeys("backspace"),
	),
}
