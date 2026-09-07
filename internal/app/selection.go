package app

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// identify reports the op id of a highlighted row, or "" when the pane is
// empty.
func identify(row list.Item) string {
	switch record := row.(type) {
	case op.Vault:
		return record.ID
	case op.Item:
		return record.ID
	default:
		return ""
	}
}

// moveCursor forwards a movement key to the focused list, then cascades if the
// highlighted row actually changed. Repeating a movement at the end of a list
// changes nothing and must not issue a load.
func (m Model) moveCursor(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	before := m.selectedID()

	moved, cmd := m.forwardToFocused(msg)
	next, ok := moved.(Model)
	if !ok || next.selectedID() == before {
		return moved, cmd
	}

	next.clearBelow(next.focus)
	return next, tea.Batch(cmd, debounceSelection(next.focus, next.selectedID()))
}

// selectedID reports the identifier highlighted in the focused pane. The
// detail pane has no cursor, so it has no selection.
func (m Model) selectedID() string {
	switch m.focus {
	case VaultPane:
		return identify(m.vaults.SelectedItem())
	case ItemPane:
		return identify(m.items.SelectedItem())
	default:
		return ""
	}
}

// clearBelow empties the panes fed by the named one, so a stale item list or a
// previous item's fields never sit beside a new selection while the load is in
// flight, and a deleted record never lingers on screen.
func (m *Model) clearBelow(pane Focus) {
	switch pane {
	case VaultPane:
		m.items.SetItems(nil)
		m.detail.Clear()
	case ItemPane:
		m.detail.Clear()
	}
}
