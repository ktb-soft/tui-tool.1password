package app

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

const personalVaultID = "aaaaaaaaaaaaaaaaaaaaaaaaaa"

// fakeOpClient returns a client pointing at a script that prints stdout and
// exits with the given code, so no test touches the real `op`.
func fakeOpClient(t *testing.T, stdout string, exitCode int) op.Client {
	t.Helper()

	path := filepath.Join(t.TempDir(), "op")
	script := "#!/bin/sh\nprintf '%s' '" + stdout + "'\nprintf 'not signed in\\n' >&2\nexit " +
		strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake op: %v", err)
	}
	return op.Client{Path: path}
}

// withVault seeds the vault pane so the item vertical has a vault to scope to.
func withVault(t *testing.T, model Model, vaultID string) Model {
	t.Helper()

	model.vaults.SetItems([]list.Item{op.Vault{ID: vaultID, Name: "Personal", Items: 2}})
	return model
}

func dispatch(t *testing.T, model Model, msg tea.Msg) (Model, tea.Cmd) {
	t.Helper()

	next, cmd := model.Update(msg)
	updated, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	return updated, cmd
}

func TestItemsLoadedPopulatesTheItemPane(t *testing.T) {
	loaded := itemsLoadedMsg{
		{ID: "1", Name: "Example Login", Category: "LOGIN"},
		{ID: "2", Name: "Example Note", Category: "SECURE_NOTE"},
	}

	model, _ := dispatch(t, sized(t, 120, 40), loaded)

	rows := model.items.List().Items()
	if len(rows) != 2 {
		t.Fatalf("len(Items()) = %d, want 2", len(rows))
	}
	item, ok := rows[0].(op.Item)
	if !ok {
		t.Fatalf("row is %T, want op.Item", rows[0])
	}
	if got, want := item.Name, "Example Login"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if !strings.Contains(model.items.View(), "Example Note") {
		t.Error("item pane does not render the loaded items")
	}
}

func TestItemsLoadedMakesTheItemPaneFocusable(t *testing.T) {
	model, _ := dispatch(t, sized(t, 120, 40), itemsLoadedMsg{{ID: "1", Name: "Example Login"}})

	moved, _ := press(t, model, "l")
	if moved.focus != ItemPane {
		t.Errorf("focus = %d after l, want ItemPane", moved.focus)
	}
}

func TestAnEmptyVaultLeavesTheItemPaneEmptyAndUnfocusable(t *testing.T) {
	loaded, _ := dispatch(t, sized(t, 120, 40), itemsLoadedMsg{{ID: "1", Name: "Example Login"}})
	model, _ := dispatch(t, loaded, itemsLoadedMsg{})

	if got := len(model.items.List().Items()); got != 0 {
		t.Errorf("len(Items()) = %d, want 0", got)
	}
	if moved, _ := press(t, model, "l"); moved.focus != VaultPane {
		t.Error("focus advanced onto an empty item pane")
	}
}

func TestItemsLoadedReplacesThePreviousVaultsItems(t *testing.T) {
	first, _ := dispatch(t, sized(t, 120, 40), itemsLoadedMsg{{ID: "1", Name: "Example Login"}})
	second, _ := dispatch(t, first, itemsLoadedMsg{{ID: "9", Name: "Example Server"}})

	rows := second.items.List().Items()
	if len(rows) != 1 {
		t.Fatalf("len(Items()) = %d, want 1", len(rows))
	}
	if strings.Contains(second.items.View(), "Example Login") {
		t.Error("item pane still shows the previous vault's items")
	}
}

func TestItemSelectionLoadsThatItemsFields(t *testing.T) {
	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.client = fakeOpClient(t, `{"id":"1111","title":"Example Login","category":"LOGIN"}`, 0)

	_, cmd := dispatch(t, model, selectionChangedMsg{Pane: ItemPane, ID: "1111"})
	if cmd == nil {
		t.Fatal("selectionChangedMsg produced no load command")
	}

	loaded, ok := cmd().(itemLoadedMsg)
	if !ok {
		t.Fatalf("command returned %T, want itemLoadedMsg", cmd())
	}
	if got, want := loaded.Name, "Example Login"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
}

func TestItemSelectionWithoutAVaultLoadsNothing(t *testing.T) {
	_, cmd := dispatch(t, sized(t, 120, 40), selectionChangedMsg{Pane: ItemPane, ID: "1111"})

	if cmd != nil {
		t.Errorf("command = %T, want none while no vault is selected", cmd())
	}
}

func TestAFailedItemLoadReportsTheOpErrorVerbatim(t *testing.T) {
	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.client = fakeOpClient(t, "", 1)

	_, cmd := dispatch(t, model, selectionChangedMsg{Pane: ItemPane, ID: "1111"})
	if cmd == nil {
		t.Fatal("selectionChangedMsg produced no load command")
	}

	failed, ok := cmd().(opFailedMsg)
	if !ok {
		t.Fatalf("command returned %T, want opFailedMsg", cmd())
	}
	if !strings.Contains(failed.Err.Error(), "not signed in") {
		t.Errorf("err = %q, want the stderr text", failed.Err)
	}
}

func TestAnItemWriteReloadsTheSelectedVaultsItems(t *testing.T) {
	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.client = fakeOpClient(t, `[{"id":"1111","title":"Example Login","category":"LOGIN"}]`, 0)

	_, cmd := dispatch(t, model, writeSucceededMsg{Reload: ItemPane, Status: theme.SavedStatus})
	if cmd == nil {
		t.Fatal("writeSucceededMsg produced no reload command")
	}

	reloaded := false
	for _, msg := range batched(cmd) {
		if items, ok := msg.(itemsLoadedMsg); ok && len(items) == 1 {
			reloaded = true
		}
	}
	if !reloaded {
		t.Error("the write did not reload the item list")
	}
}

func TestAnItemWriteWithoutAVaultReloadsNothing(t *testing.T) {
	model := sized(t, 120, 40)
	model.client = fakeOpClient(t, "[]", 0)

	_, cmd := dispatch(t, model, writeSucceededMsg{Reload: ItemPane, Status: theme.SavedStatus})

	for _, msg := range batched(cmd) {
		if _, ok := msg.(itemsLoadedMsg); ok {
			t.Error("the write reloaded items with no vault selected")
		}
	}
}

// batched runs cmd and flattens any tea.Batch it returns.
func batched(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return []tea.Msg{msg}
	}

	msgs := make([]tea.Msg, 0, len(batch))
	for _, inner := range batch {
		msgs = append(msgs, batched(inner)...)
	}
	return msgs
}
