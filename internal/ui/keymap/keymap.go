// Package keymap holds every key binding in the program. The footer help and
// the ? overlay are generated from it, so a binding and its documentation
// cannot drift apart.
package keymap

import "charm.land/bubbles/v2/key"

// KeyMap is every binding the program responds to. It implements help.KeyMap.
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	PaneLeft  key.Binding
	PaneRight key.Binding
	Create    key.Binding
	Edit      key.Binding
	Delete    key.Binding
	Reveal    key.Binding
	Copy      key.Binding
	Filter    key.Binding
	Help      key.Binding
	Confirm   key.Binding
	Cancel    key.Binding
	Quit      key.Binding
}

// Default returns the bindings documented in
// docs/components/04-00-00-navigation-keymap.md.
func Default() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PaneLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "pane left"),
		),
		PaneRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "pane right"),
		),
		Create: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Reveal: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reveal"),
		),
		Copy: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp is the footer row.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.PaneLeft, k.Up, k.Create, k.Edit, k.Delete, k.Reveal, k.Help,
	}
}

// FullHelp is the ? overlay, grouped by column.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PaneLeft, k.PaneRight},
		{k.Create, k.Edit, k.Delete},
		{k.Reveal, k.Copy, k.Filter},
		{k.Confirm, k.Cancel, k.Help, k.Quit},
	}
}
