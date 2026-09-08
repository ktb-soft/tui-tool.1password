package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// loadItem fetches one item, field values included, or returns the cached one.
func loadItem(client op.Client, store *cache, vaultID, itemID string) tea.Cmd {
	return func() tea.Msg {
		item, err := store.getItem(vaultID, itemID, func() (op.Item, error) {
			return client.GetItem(vaultID, itemID)
		})
		if err != nil {
			return opFailedMsg{err}
		}
		return itemLoadedMsg(item)
	}
}

// handleItemPaneDetailKey handles the keys the item pane aims at the detail
// pane: Edit moves focus into the form, Reveal unmasks the item's secrets. The
// bool reports whether the key was consumed.
func (m Model) handleItemPaneDetailKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Edit):
		return m.enterDetail()
	case key.Matches(msg, m.keys.Fields):
		return m.openFieldEditor()
	case key.Matches(msg, m.keys.Reveal):
		m.detail.ToggleReveal()
		return m, nil, true
	}
	return m, nil, false
}

// enterDetail hands focus to the detail pane's form, where the item's values
// and title are edited.
func (m Model) enterDetail() (tea.Model, tea.Cmd, bool) {
	_, notLoaded, ok := m.loadedSelection()
	if !ok {
		return m, notLoaded, notLoaded != nil
	}

	m.focus = DetailPane
	m.applyFocus()
	return m, m.detail.Init(), true
}

// openFieldEditor opens the item form over the loaded item, which is where a
// field is added, renamed, retyped, or removed. It is the same form a create
// fills in, so the rules for what a field row means are written once.
func (m Model) openFieldEditor() (tea.Model, tea.Cmd, bool) {
	item, notLoaded, ok := m.loadedSelection()
	if !ok {
		return m, notLoaded, notLoaded != nil
	}
	return m.openItemForm(item)
}

// loadedSelection returns the highlighted item once the detail pane holds its
// field values, since the item list carries none and saving from it would
// erase them. A selection whose fields have not arrived yet comes back with
// the command that says so.
func (m Model) loadedSelection() (op.Item, tea.Cmd, bool) {
	selected, ok := m.items.SelectedItem().(op.Item)
	if !ok {
		return op.Item{}, nil, false
	}
	if m.detail.Item().ID != selected.ID {
		return op.Item{}, m.statusCmd(theme.ItemNotLoadedStatus), false
	}
	return m.detail.Item(), nil, true
}

// handleDetailKey handles keys while the detail pane's form holds focus. Every
// key except the two below belongs to huh, so the form navigates itself.
func (m Model) handleDetailKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Interrupt):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Cancel):
		m.leaveDetail()
		return m, nil
	}
	return m.forwardToDetail(msg)
}

// leaveDetail returns focus to the item pane. Blurring the detail pane rebuilds
// its form from the loaded item, so nothing typed into it is kept.
func (m *Model) leaveDetail() {
	m.focus = ItemPane
	m.applyFocus()
}

// forwardToDetail hands a message to the form and, once it completes, saves the
// edited item and returns focus to the item pane. The pane keeps the values it
// submitted rather than the ones it was seeded with, so a saved edit does not
// blink back to the old value while the write is in flight.
func (m Model) forwardToDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	if !m.detail.Completed() {
		return m, cmd
	}

	edited := m.detail.EditedItem()
	m.detail.SetItem(edited)
	m.leaveDetail()
	return m, tea.Batch(cmd, saveItem(m.client, edited))
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

	return m.forwardToDetail(msg)
}

// reloadItem re-fetches the displayed item, so a field write shows its result.
func (m Model) reloadItem() tea.Cmd {
	item := m.detail.Item()
	if item.ID == "" {
		return nil
	}
	return loadItem(m.client, m.cache, item.Vault.ID, item.ID)
}
