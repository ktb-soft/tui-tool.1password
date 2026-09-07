package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// loadItems fetches the item list for one vault.
//
//nolint:unused // frozen contract seam, wired up in wave 1 (ADR 07)
func loadItems(client op.Client, vaultID string) tea.Cmd {
	return func() tea.Msg {
		items, err := client.ListItems(vaultID)
		if err != nil {
			return opFailedMsg{err}
		}
		return itemsLoadedMsg(items)
	}
}

// handleItemKey handles keys while the item pane is focused. The bool reports
// whether the key was consumed.
func (m Model) handleItemKey(_ tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	return m, nil, false
}

// updateItems handles every message belonging to the item vertical.
func (m Model) updateItems(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
