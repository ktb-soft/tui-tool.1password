// Package pane renders a bordered, focusable list container. It never calls
// op and never decides what data to fetch.
package pane

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// Pane is a bubbles list inside a border whose color tracks focus.
type Pane struct {
	Title   string
	list    list.Model
	focused bool
	width   int
	height  int
}

// New builds a pane titled title, whose status bar counts singular/plural.
func New(title, singular, plural string) Pane {
	model := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	model.Title = title
	model.SetShowHelp(false)
	model.SetStatusBarItemName(singular, plural)
	model.DisableQuitKeybindings()
	return Pane{Title: title, list: model}
}

// List exposes the underlying list so the controller can drive it.
func (p *Pane) List() *list.Model { return &p.list }

// SetItems replaces the pane's contents.
func (p *Pane) SetItems(items []list.Item) tea.Cmd { return p.list.SetItems(items) }

// SelectedItem returns the highlighted item, or nil when the pane is empty.
func (p Pane) SelectedItem() list.Item { return p.list.SelectedItem() }

// StartSpinner shows the pane's loading indicator. The returned command
// animates it.
func (p *Pane) StartSpinner() tea.Cmd { return p.list.StartSpinner() }

// StopSpinner hides the loading indicator.
func (p *Pane) StopSpinner() { p.list.StopSpinner() }

// Focus marks the pane focused, which selects the accent border.
func (p *Pane) Focus() { p.focused = true }

// Blur marks the pane unfocused.
func (p *Pane) Blur() { p.focused = false }

// Focused reports whether the pane holds focus.
func (p Pane) Focused() bool { return p.focused }

// SetSize sets the outer dimensions, border included.
func (p *Pane) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.list.SetSize(max(width-theme.BorderWidth, 0), max(height-theme.BorderWidth, 0))
}

// Update forwards a message to the list.
func (p Pane) Update(msg tea.Msg) (Pane, tea.Cmd) {
	model, cmd := p.list.Update(msg)
	p.list = model
	return p, cmd
}

// View renders the list inside its border.
func (p Pane) View() string {
	return theme.Border(p.focused).
		Width(p.width).
		Height(p.height).
		Render(p.list.View())
}
