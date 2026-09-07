package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// loadItem fetches one item, field values included.
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
func (m Model) handleDetailKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	if key.Matches(msg, m.keys.Reveal) {
		m.detail.ToggleReveal()
		return m, nil, true
	}
	return m, nil, false
}

// updateDetail handles every message belonging to the detail vertical.
func (m Model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case itemLoadedMsg:
		m.detail.SetItem(op.Item(msg))
		return m, nil

	case writeSucceededMsg:
		return m, m.reloadItem()

	case selectionChangedMsg:
		return m, nil
	}

	return m.forwardToFocused(msg)
}

// reloadItem re-fetches the displayed item, so a field write shows its result.
func (m Model) reloadItem() tea.Cmd {
	item := m.detail.Item()
	if item.ID == "" {
		return nil
	}
	return loadItem(m.client, item.Vault.ID, item.ID)
}
