// Package detail renders the selected item's fields as a scrolling document.
package detail

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// Detail is the field pane: a viewport over per-section field tables, with
// per-field reveal state that resets when the item changes.
type Detail struct {
	item     op.Item
	revealed map[string]bool
	viewport viewport.Model
	focused  bool
	width    int
	height   int
}

// New builds an empty detail pane.
func New() Detail {
	pane := Detail{revealed: map[string]bool{}, viewport: viewport.New()}
	pane.viewport.SetContent(pane.content())
	return pane
}

// Item returns the item currently displayed.
func (d Detail) Item() op.Item { return d.item }

// SetItem replaces the displayed item and drops all reveal state.
func (d *Detail) SetItem(item op.Item) {
	d.item = item
	d.revealed = map[string]bool{}
	d.viewport.SetContent(d.content())
}

// Clear empties the pane.
func (d *Detail) Clear() { d.SetItem(op.Item{}) }

// ToggleReveal flips the masking of every concealed field on the item.
func (d *Detail) ToggleReveal() {
	for _, field := range d.item.Fields {
		if field.IsSecret() {
			d.revealed[field.ID] = !d.revealed[field.ID]
		}
	}
	d.viewport.SetContent(d.content())
}

// Revealed reports whether the given field is currently shown in plaintext.
func (d Detail) Revealed(fieldID string) bool { return d.revealed[fieldID] }

// Focus marks the pane focused.
func (d *Detail) Focus() { d.focused = true }

// Blur marks the pane unfocused.
func (d *Detail) Blur() { d.focused = false }

// Focused reports whether the pane holds focus.
func (d Detail) Focused() bool { return d.focused }

// SetSize sets the outer dimensions, border included.
func (d *Detail) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.viewport.SetWidth(max(width-theme.BorderWidth, 0))
	d.viewport.SetHeight(max(height-theme.BorderWidth, 0))
}

// Update forwards a message to the viewport.
func (d Detail) Update(msg tea.Msg) (Detail, tea.Cmd) {
	model, cmd := d.viewport.Update(msg)
	d.viewport = model
	return d, cmd
}

// View renders the fields inside the pane's border.
func (d Detail) View() string {
	return theme.Border(d.focused).
		Width(d.width).
		Height(d.height).
		Render(d.viewport.View())
}

func (d Detail) content() string {
	if d.item.ID == "" {
		return theme.Empty.Render(theme.NoSelection)
	}
	return d.item.Name
}
