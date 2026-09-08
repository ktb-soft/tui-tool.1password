package app

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/form"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// loadItems fetches the item list for one vault, or returns the cached one.
func loadItems(client op.Client, store *cache, vaultID string) tea.Cmd {
	return func() tea.Msg {
		items, err := store.getItems(vaultID, func() ([]op.Item, error) {
			return client.ListItems(vaultID)
		})
		if err != nil {
			return opFailedMsg{err}
		}
		return itemsLoadedMsg(items)
	}
}

// saveItem creates or edits an item, choosing by whether it already has an ID.
// The template goes to `op` on stdin, so no field value reaches argv.
func saveItem(client op.Client, item op.Item) tea.Cmd {
	return func() tea.Msg {
		status := theme.SavedStatus
		var err error
		if item.ID == "" {
			status = theme.CreatedStatus
			_, err = client.CreateItem(item)
		} else {
			_, err = client.EditItem(item)
		}
		if err != nil {
			return opFailedMsg{err}
		}
		return writeSucceededMsg{Reload: ItemPane, Status: status}
	}
}

// deleteItem removes one item from the given vault.
func deleteItem(client op.Client, vaultID, itemID string) tea.Cmd {
	return func() tea.Msg {
		if err := client.DeleteItem(vaultID, itemID); err != nil {
			return opFailedMsg{err}
		}
		return writeSucceededMsg{Reload: ItemPane, Status: theme.DeletedStatus, Removed: true}
	}
}

// handleItemKey handles keys while the item pane is focused. The bool reports
// whether the key was consumed.
func (m Model) handleItemKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Create):
		return m.openCreateItem()
	case key.Matches(msg, m.keys.Delete):
		return m.openDeleteItem()
	}
	return m, nil, false
}

// openCreateItem opens an empty item form scoped to the selected vault.
func (m Model) openCreateItem() (tea.Model, tea.Cmd, bool) {
	vaultID := m.selectedVaultID()
	if vaultID == "" {
		return m, nil, false
	}
	return m.openItemForm(op.Item{Vault: op.VaultRef{ID: vaultID}})
}

func (m Model) openItemForm(item op.Item) (tea.Model, tea.Cmd, bool) {
	itemForm, draft := form.NewItem(item)
	cmd := m.openOverlay(itemForm, func(model Model) (tea.Model, tea.Cmd) {
		return model, saveItem(model.client, draft.Item())
	})
	return m, cmd, true
}

// openDeleteItem asks for confirmation, defaulted to No, before deleting.
func (m Model) openDeleteItem() (tea.Model, tea.Cmd, bool) {
	selected, ok := m.items.SelectedItem().(op.Item)
	vaultID := m.selectedVaultID()
	if !ok || vaultID == "" {
		return m, nil, false
	}

	confirmation, confirmed := form.NewDeleteItem(selected.Name)
	cmd := m.openOverlay(confirmation, func(model Model) (tea.Model, tea.Cmd) {
		if !*confirmed {
			return model, nil
		}
		return model, deleteItem(model.client, vaultID, selected.ID)
	})
	return m, cmd, true
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

// showItems replaces the pane's contents and loads whatever row the list
// highlights once filled. An empty vault leaves an empty list, whose own empty
// state names it, and schedules nothing.
func (m Model) showItems(loaded itemsLoadedMsg) (tea.Model, tea.Cmd) {
	rows := make([]list.Item, len(loaded))
	for i, item := range loaded {
		rows[i] = item
	}
	populate := m.items.SetItems(rows)
	return m, tea.Batch(populate, m.scheduleLoad(ItemPane))
}

// reloadItems refetches the item list for the vault currently selected.
func (m Model) reloadItems() tea.Cmd {
	vaultID := m.selectedVaultID()
	if vaultID == "" {
		return nil
	}
	return loadItems(m.client, m.cache, vaultID)
}

// loadSelectedItem fetches the fields of the item the cursor has settled on.
func (m Model) loadSelectedItem(itemID string) tea.Cmd {
	vaultID := m.selectedVaultID()
	if vaultID == "" || itemID == "" {
		return nil
	}
	return loadItem(m.client, m.cache, vaultID, itemID)
}

// selectedVaultID returns the highlighted vault's ID, empty when none is.
func (m Model) selectedVaultID() string {
	vault, ok := m.vaults.SelectedItem().(op.Vault)
	if !ok {
		return ""
	}
	return vault.ID
}
