package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// collectSelections walks a command, following batches, and returns every
// selectionChangedMsg it produces.
func collectSelections(cmd tea.Cmd) []selectionChangedMsg {
	if cmd == nil {
		return nil
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		var found []selectionChangedMsg
		for _, batched := range msg {
			found = append(found, collectSelections(batched)...)
		}
		return found
	case selectionChangedMsg:
		return []selectionChangedMsg{msg}
	default:
		return nil
	}
}

func mustCmd(t *testing.T, model Model, msg tea.Msg) tea.Cmd {
	t.Helper()

	_, cmd := dispatch(t, model, msg)
	return cmd
}

func TestArrivingVaultsLoadTheHighlightedVault(t *testing.T) {
	loaded := vaultsLoadedMsg{
		{ID: personalVaultID, Name: "Personal", Items: 2},
		{ID: secondVaultID, Name: "Shared", Items: 1},
	}

	scheduled := collectSelections(mustCmd(t, sized(t, 120, 40), loaded))

	if len(scheduled) != 1 {
		t.Fatalf("the vault list arrival scheduled %d loads, want 1", len(scheduled))
	}
	if scheduled[0].Pane != VaultPane || scheduled[0].ID != personalVaultID {
		t.Fatalf("scheduled %+v, want the vault pane and %q", scheduled[0], personalVaultID)
	}
}

func TestArrivingItemsLoadTheHighlightedItem(t *testing.T) {
	model := withVault(t, sized(t, 120, 40), personalVaultID)
	loaded := itemsLoadedMsg{
		{ID: "item1", Name: "GitHub", Category: "LOGIN"},
		{ID: "item2", Name: "Router", Category: "LOGIN"},
	}

	scheduled := collectSelections(mustCmd(t, model, loaded))

	if len(scheduled) != 1 {
		t.Fatalf("the item list arrival scheduled %d loads, want 1", len(scheduled))
	}
	if scheduled[0].Pane != ItemPane || scheduled[0].ID != "item1" {
		t.Fatalf("scheduled %+v, want the item pane and %q", scheduled[0], "item1")
	}
}

func TestAnEmptyListSchedulesNoLoad(t *testing.T) {
	model := withVault(t, sized(t, 120, 40), personalVaultID)

	if scheduled := collectSelections(mustCmd(t, model, vaultsLoadedMsg(nil))); scheduled != nil {
		t.Fatalf("an empty vault list scheduled %+v", scheduled)
	}
	if scheduled := collectSelections(mustCmd(t, model, itemsLoadedMsg(nil))); scheduled != nil {
		t.Fatalf("an empty item list scheduled %+v", scheduled)
	}
}

// TestLoadingAnItemDoesNotRepopulateTheItemPane guards the cascade against a
// loop: the load an item list schedules must end at the detail pane, never
// produce another item list that would schedule the same load again.
func TestLoadingAnItemDoesNotRepopulateTheItemPane(t *testing.T) {
	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.client = fakeOpClient(t, `{"id":"item1","title":"GitHub","category":"LOGIN"}`, 0)

	populated, _ := dispatch(t, model, itemsLoadedMsg{{ID: "item1", Name: "GitHub", Category: "LOGIN"}})
	cmd := mustCmd(t, populated, selectionChangedMsg{Pane: ItemPane, ID: "item1"})

	if cmd == nil {
		t.Fatal("the settled item selection issued no load")
	}
	if produced := cmd(); !isItemLoaded(produced) {
		t.Fatalf("the settled item selection produced %T, want itemLoadedMsg", produced)
	}
	if scheduled := collectSelections(cmd); scheduled != nil {
		t.Fatalf("loading an item scheduled another load: %+v", scheduled)
	}
}

func isItemLoaded(msg tea.Msg) bool {
	_, ok := msg.(itemLoadedMsg)
	return ok
}
