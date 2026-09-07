package app

import (
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

// handleVaultKey handles keys while the vault pane is focused. The bool
// reports whether the key was consumed.
func (m Model) handleVaultKey(_ tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	return m, nil, false
}

// updateVaults handles every message belonging to the vault vertical.
func (m Model) updateVaults(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
