package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// loadItem fetches one item, field values included.
//
//nolint:unused // frozen contract seam, wired up in wave 1 (ADR 07)
func loadItem(client op.Client, vaultID, itemID string) tea.Cmd {
	return func() tea.Msg {
		item, err := client.GetItem(vaultID, itemID)
		if err != nil {
			return opFailedMsg{err}
		}
		return itemLoadedMsg(item)
	}
}

// handleDetailKey handles keys while the detail pane is focused. The bool
// reports whether the key was consumed.
func (m Model) handleDetailKey(_ tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	return m, nil, false
}

// updateDetail handles every message belonging to the detail vertical.
func (m Model) updateDetail(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
