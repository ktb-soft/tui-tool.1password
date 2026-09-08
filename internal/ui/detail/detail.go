// Package detail renders the selected item as an editable huh form.
package detail

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/form"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// Detail is the field pane: a huh form over the selected item, blurred until
// the item pane hands it focus. Blurring rebuilds the form from the item, so
// leaving the pane discards every edit made in it.
type Detail struct {
	item     op.Item
	revealed bool
	form     *huh.Form
	draft    *form.ItemDraft
	initCmd  tea.Cmd
	focused  bool
	width    int
	height   int
}

// New builds an empty detail pane.
func New() Detail {
	pane := Detail{}
	pane.rebuild()
	return pane
}

// Item returns the item currently displayed.
func (d Detail) Item() op.Item { return d.item }

// SetItem replaces the displayed item and drops all reveal state, so nothing
// stays revealed after navigating away.
func (d *Detail) SetItem(item op.Item) {
	d.item = item
	d.revealed = false
	d.rebuild()
}

// Clear empties the pane.
func (d *Detail) Clear() { d.SetItem(op.Item{}) }

// ToggleReveal unmasks the item's concealed fields, or masks them again.
func (d *Detail) ToggleReveal() {
	d.revealed = !d.revealed
	d.rebuild()
}

// Revealed reports whether concealed values are currently shown in plaintext.
func (d Detail) Revealed() bool { return d.revealed }

// Focus marks the pane focused and arms a fresh form for editing.
func (d *Detail) Focus() {
	d.focused = true
	d.rebuild()
}

// Blur marks the pane unfocused and rebuilds the form from the stored item,
// which is how leaving the pane discards uncommitted edits.
func (d *Detail) Blur() {
	d.focused = false
	d.rebuild()
}

// Focused reports whether the pane holds focus.
func (d Detail) Focused() bool { return d.focused }

// Init returns the command that starts the form the pane currently holds.
func (d Detail) Init() tea.Cmd { return d.initCmd }

// SetSize sets the outer dimensions, border included.
func (d *Detail) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.rebuild()
}

// Update forwards a message to the form, which only accepts messages while the
// pane holds focus.
func (d Detail) Update(msg tea.Msg) (Detail, tea.Cmd) {
	if !d.focused || d.form == nil {
		return d, nil
	}
	updated, cmd := d.form.Update(msg)
	if next, ok := updated.(*huh.Form); ok {
		d.form = next
	}
	return d, cmd
}

// Completed reports whether the form has been submitted and its draft is ready
// to write.
func (d Detail) Completed() bool {
	return d.form != nil && d.form.State == huh.StateCompleted
}

// EditedItem folds the form's draft back onto the item it was seeded from.
func (d Detail) EditedItem() op.Item {
	if d.draft == nil {
		return d.item
	}
	return d.draft.Item()
}

// View renders the form, or the empty state, inside the pane's border.
func (d Detail) View() string {
	return theme.Border(d.focused).
		Width(d.width).
		Height(d.height).
		Render(d.content())
}

func (d Detail) content() string {
	if d.form == nil {
		return theme.EmptyDetail.Render(theme.NoSelection)
	}
	return d.form.View()
}

// rebuild replaces the form with one bound to a fresh draft of the current
// item. huh fixes an input's echo mode at construction, so revealing a secret
// and resizing the pane both go through here.
func (d *Detail) rebuild() {
	if d.item.ID == "" {
		d.form, d.draft, d.initCmd = nil, nil, nil
		return
	}

	built, draft := form.NewInlineItem(d.item, d.revealed)
	d.form, d.draft = built.
		WithWidth(max(d.width-theme.BorderWidth, 0)).
		WithHeight(max(d.height-theme.BorderWidth, 0)).
		WithShowHelp(d.focused), draft
	d.initCmd = d.form.Init()
}
