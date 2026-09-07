package app

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// loadVaults fetches the vault list.
func loadVaults(client op.Client) tea.Cmd {
	return func() tea.Msg {
		vaults, err := client.ListVaults()
		if err != nil {
			return opFailedMsg{err}
		}
		return vaultsLoadedMsg(vaults)
	}
}

// vaultListItems adapts the domain slice to the slice bubbles/list wants.
func vaultListItems(vaults []op.Vault) []list.Item {
	items := make([]list.Item, len(vaults))
	for i, vault := range vaults {
		items[i] = vault
	}
	return items
}

// handleVaultKey handles keys while the vault pane is focused. The bool
// reports whether the key was consumed.
func (m Model) handleVaultKey(_ tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	return m, nil, false
}

// updateVaults handles every message belonging to the vault vertical.
func (m Model) updateVaults(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case vaultsLoadedMsg:
		return m, m.vaults.SetItems(vaultListItems(msg))
	case writeSucceededMsg:
		return m, loadVaults(m.client)
	}
	return m, nil
}
