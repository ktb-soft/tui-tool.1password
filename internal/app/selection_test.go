package app

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

const secondVaultID = "bbbbbbbbbbbbbbbbbbbbbbbbbb"

// withVaults seeds the vault pane with two vaults so the cursor has somewhere
// to move.
func withVaults(t *testing.T, model Model) Model {
	t.Helper()

	model.vaults.SetItems([]list.Item{
		op.Vault{ID: personalVaultID, Name: "Personal", Items: 2},
		op.Vault{ID: secondVaultID, Name: "Shared", Items: 1},
	})
	return model
}

// settle returns the first selection change a command schedules, or nil when
// it schedules none.
func settle(cmd tea.Cmd) tea.Msg {
	scheduled := collectSelections(cmd)
	if len(scheduled) == 0 {
		return nil
	}
	return scheduled[0]
}

func TestMovingTheVaultCursorSchedulesAnItemLoad(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))

	moved, cmd := press(t, model, "j")

	if got := identify(moved.vaults.SelectedItem()); got != secondVaultID {
		t.Fatalf("cursor on %q, want %q", got, secondVaultID)
	}

	changed, ok := settle(cmd).(selectionChangedMsg)
	if !ok {
		t.Fatal("moving the cursor scheduled no selection change")
	}
	if changed.Pane != VaultPane || changed.ID != secondVaultID {
		t.Fatalf("scheduled %+v, want vault pane and %q", changed, secondVaultID)
	}
}

func TestCursorAtTheEndOfTheListDoesNotReload(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))

	atTop, _ := press(t, model, "k")

	_, cmd := press(t, atTop, "k")
	if changed := settle(cmd); changed != nil {
		t.Fatalf("cursor could not move but scheduled %+v", changed)
	}
}

func TestMovingTheVaultCursorClearsTheDownstreamPanes(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))
	model.items.SetItems([]list.Item{op.Item{ID: "stale", Name: "Stale Item"}})
	model.detail.SetItem(op.Item{ID: "stale", Name: "Stale Item"})

	moved, _ := press(t, model, "j")

	if got := len(moved.items.List().Items()); got != 0 {
		t.Fatalf("item pane kept %d rows from the previous vault", got)
	}
	if got := moved.detail.Item().ID; got != "" {
		t.Fatalf("detail pane still shows %q", got)
	}
}

func TestDeletingAnItemClearsTheDetailPane(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))
	model.detail.SetItem(op.Item{
		ID:   "deleted",
		Name: "Router Admin",
		Fields: []op.Field{
			{ID: "password", Label: "password", Type: op.FieldTypeConcealed, Value: "hunter2"},
		},
	})
	model.detail.ToggleReveal()

	if !strings.Contains(model.detail.View(), "hunter2") {
		t.Fatal("test setup did not reveal the secret it means to check")
	}

	after, _ := model.Update(writeSucceededMsg{Reload: ItemPane, Status: "deleted", Removed: true})
	cleared, ok := after.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", after)
	}

	if got := cleared.detail.Item().ID; got != "" {
		t.Fatalf("detail pane still shows the deleted item %q", got)
	}
	if strings.Contains(cleared.detail.View(), "hunter2") {
		t.Fatal("a revealed secret survived the delete that removed it")
	}
}

func TestSelectionChangeLoadsTheVaultsItems(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))
	model.client = fakeOpClient(t, `[{"id":"item1","title":"GitHub","category":"LOGIN"}]`, 0)

	_, cmd := model.Update(selectionChangedMsg{Pane: VaultPane, ID: personalVaultID})
	if cmd == nil {
		t.Fatal("selection change issued no load")
	}

	var loaded itemsLoadedMsg
	for _, msg := range batched(cmd) {
		if items, ok := msg.(itemsLoadedMsg); ok {
			loaded = items
		}
	}
	if loaded == nil {
		t.Fatal("the selection change loaded no items")
	}
	if len(loaded) != 1 || loaded[0].Name != "GitHub" {
		t.Fatalf("loaded %+v, want the one item the fake op printed", loaded)
	}
}
