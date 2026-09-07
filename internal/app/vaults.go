package app

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/form"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
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

// createVault runs `op vault create` and reports the reload the vault pane owes.
func createVault(client op.Client, name string) tea.Cmd {
	return func() tea.Msg {
		if _, err := client.CreateVault(name); err != nil {
			return opFailedMsg{err}
		}
		return writeSucceededMsg{Reload: VaultPane, Status: theme.CreatedStatus}
	}
}

// editVault renames a vault.
func editVault(client op.Client, id, name string) tea.Cmd {
	return func() tea.Msg {
		if _, err := client.EditVault(id, name); err != nil {
			return opFailedMsg{err}
		}
		return writeSucceededMsg{Reload: VaultPane, Status: theme.SavedStatus}
	}
}

// deleteVault deletes a vault and every item in it.
func deleteVault(client op.Client, id string) tea.Cmd {
	return func() tea.Msg {
		if err := client.DeleteVault(id); err != nil {
			return opFailedMsg{err}
		}
		return writeSucceededMsg{Reload: VaultPane, Status: theme.DeletedStatus, Removed: true}
	}
}

// handleVaultKey handles keys while the vault pane is focused. The bool
// reports whether the key was consumed.
func (m Model) handleVaultKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Create):
		return m, m.openVaultCreate(), true
	case key.Matches(msg, m.keys.Edit):
		return m, m.openVaultEdit(), true
	case key.Matches(msg, m.keys.Delete):
		return m, m.openVaultDelete(), true
	}
	return m, nil, false
}

// selectedVault reports the highlighted vault, and false when the pane is empty.
func (m Model) selectedVault() (op.Vault, bool) {
	vault, ok := m.vaults.SelectedItem().(op.Vault)
	return vault, ok
}

// openVaultCreate opens an empty name form that creates on submit.
func (m *Model) openVaultCreate() tea.Cmd {
	name := new(string)
	return m.openOverlay(form.NewVault(name), func(submitted Model) (tea.Model, tea.Cmd) {
		return submitted, createVault(submitted.client, *name)
	})
}

// openVaultEdit opens the name form seeded with the selected vault's name.
func (m *Model) openVaultEdit() tea.Cmd {
	vault, ok := m.selectedVault()
	if !ok {
		return nil
	}

	name := new(string)
	*name = vault.Name
	return m.openOverlay(form.NewVault(name), func(submitted Model) (tea.Model, tea.Cmd) {
		if *name == vault.Name {
			return submitted, nil
		}
		return submitted, editVault(submitted.client, vault.ID, *name)
	})
}

// openVaultDelete opens the confirmation, which deletes only when the vault
// name was typed back and the confirm was answered yes.
func (m *Model) openVaultDelete() tea.Cmd {
	vault, ok := m.selectedVault()
	if !ok {
		return nil
	}

	typed, confirmed := new(string), new(bool)
	return m.openOverlay(form.NewVaultDelete(vault.Name, typed, confirmed),
		func(submitted Model) (tea.Model, tea.Cmd) {
			if !form.IsVaultDeleteConfirmed(vault.Name, *typed, *confirmed) {
				return submitted, nil
			}
			return submitted, deleteVault(submitted.client, vault.ID)
		})
}

// updateVaults handles every message belonging to the vault vertical.
func (m Model) updateVaults(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case vaultsLoadedMsg:
		populate := m.vaults.SetItems(vaultListItems(msg))
		return m, tea.Batch(populate, m.scheduleLoad(VaultPane))
	case selectionChangedMsg:
		return m, loadItems(m.client, msg.ID)
	case writeSucceededMsg:
		return m, loadVaults(m.client)
	}
	return m, nil
}
