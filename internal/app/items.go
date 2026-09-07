package app

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// loadItems fetches the item list for one vault.
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
func (m Model) updateItems(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case itemsLoadedMsg:
		return m.showItems(msg)
	case selectionChangedMsg:
		return m, m.loadSelectedItem(msg.ID)
	case writeSucceededMsg:
		return m, m.reloadItems()
	}
	return m, nil
}

// showItems replaces the pane's contents. An empty vault leaves an empty
// list, whose own empty state names it.
func (m Model) showItems(loaded itemsLoadedMsg) (tea.Model, tea.Cmd) {
	rows := make([]list.Item, len(loaded))
	for i, item := range loaded {
		rows[i] = item
	}
	return m, m.items.SetItems(rows)
}

// reloadItems refetches the item list for the vault currently selected.
func (m Model) reloadItems() tea.Cmd {
	vaultID := m.selectedVaultID()
	if vaultID == "" {
		return nil
	}
	return loadItems(m.client, vaultID)
}

// loadSelectedItem fetches the fields of the item the cursor has settled on.
func (m Model) loadSelectedItem(itemID string) tea.Cmd {
	vaultID := m.selectedVaultID()
	if vaultID == "" || itemID == "" {
		return nil
	}
	return loadItem(m.client, vaultID, itemID)
}

// selectedVaultID returns the highlighted vault's ID, empty when none is.
func (m Model) selectedVaultID() string {
	vault, ok := m.vaults.SelectedItem().(op.Vault)
	if !ok {
		return ""
	}
	return vault.ID
}
