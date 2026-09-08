package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// The records the fake op serves. Neither carries a field value: a cache test
// asserts subprocess counts, not secrets.
const (
	countedVault = `{"id":"vault-a","name":"Personal"}`
	countedItem  = `{"id":"item1","title":"GitHub","category":"LOGIN"}`
)

// countingOp writes a fake op that appends one line per invocation to a
// counter file and answers every read with well-formed JSON, so a test can
// assert how many subprocesses a sequence ran.
func countingOp(t *testing.T) (op.Client, func() int) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "op")
	counter := filepath.Join(dir, "invocations")

	script := "#!/bin/sh\n" +
		"echo ran >> " + counter + "\n" +
		"case \"$1 $2\" in\n" +
		"'item get') printf '%s' '" + countedItem + "' ;;\n" +
		"'item list') printf '%s' '[" + countedItem + "]' ;;\n" +
		"*) printf '%s' '[" + countedVault + "]' ;;\n" +
		"esac\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake op: %v", err)
	}

	count := func() int {
		data, err := os.ReadFile(counter)
		if os.IsNotExist(err) {
			return 0
		}
		if err != nil {
			t.Fatalf("read invocations: %v", err)
		}
		return strings.Count(string(data), "\n")
	}
	return op.Client{Path: path}, count
}

// run executes a command the way the runtime would, failing on a nil command.
func run(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()

	if cmd == nil {
		t.Fatal("command is nil, want one that reaches op")
	}
	return cmd()
}

func TestSecondVaultListReadRunsNoSubprocess(t *testing.T) {
	client, invocations := countingOp(t)
	store := newCache()

	for range 3 {
		if _, ok := run(t, loadVaults(client, store)).(vaultsLoadedMsg); !ok {
			t.Fatal("loadVaults did not report the vault list")
		}
	}

	if got := invocations(); got != 1 {
		t.Errorf("op ran %d times for three vault list reads, want 1", got)
	}
}

func TestSecondItemListReadRunsNoSubprocess(t *testing.T) {
	client, invocations := countingOp(t)
	store := newCache()

	run(t, loadItems(client, store, personalVaultID))
	run(t, loadItems(client, store, personalVaultID))

	if got := invocations(); got != 2 {
		t.Errorf("op ran %d times for two reads of one vault, want 2 — list and bulk get", got)
	}
}

func TestEachVaultGetsItsOwnItemList(t *testing.T) {
	client, invocations := countingOp(t)
	store := newCache()

	run(t, loadItems(client, store, personalVaultID))
	run(t, loadItems(client, store, "vault-b"))
	run(t, loadItems(client, store, personalVaultID))

	if got := invocations(); got != 4 {
		t.Errorf("op ran %d times for two distinct vaults, want 4 — a list and a bulk get each", got)
	}
}

func TestSecondItemReadRunsNoSubprocess(t *testing.T) {
	client, invocations := countingOp(t)
	store := newCache()

	loaded, ok := run(t, loadItem(client, store, personalVaultID, "item1")).(itemLoadedMsg)
	if !ok {
		t.Fatal("loadItem did not report the item")
	}
	cached, ok := run(t, loadItem(client, store, personalVaultID, "item1")).(itemLoadedMsg)
	if !ok {
		t.Fatal("the cached read did not report the item")
	}

	if got := invocations(); got != 1 {
		t.Errorf("op ran %d times for two reads of one item, want 1", got)
	}
	if op.Item(cached).Name != op.Item(loaded).Name {
		t.Errorf("cached item = %q, want %q", op.Item(cached).Name, op.Item(loaded).Name)
	}
}

func TestAFailedReadIsNotCached(t *testing.T) {
	client := fakeOpClient(t, "", 1)
	store := newCache()

	if _, ok := run(t, loadVaults(client, store)).(opFailedMsg); !ok {
		t.Fatal("loadVaults did not report the failure")
	}
	if _, ok := run(t, loadVaults(client, store)).(opFailedMsg); !ok {
		t.Fatal("the retry was answered from the cache instead of reaching op")
	}
}

// itemWriteModel returns a model whose vault pane holds one vault and whose
// cache already holds that vault's items and one item's fields.
func itemWriteModel(t *testing.T, client op.Client) Model {
	t.Helper()

	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.client = client
	run(t, loadItems(client, model.cache, personalVaultID))
	run(t, loadItem(client, model.cache, personalVaultID, "item1"))
	return model
}

func TestAnItemWriteInvalidatesTheVaultsItemsAndValues(t *testing.T) {
	client, invocations := countingOp(t)
	model := itemWriteModel(t, client)
	before := invocations()

	dispatch(t, model, writeSucceededMsg{Reload: ItemPane})

	run(t, loadItems(client, model.cache, personalVaultID))
	run(t, loadItem(client, model.cache, personalVaultID, "item1"))

	if got := invocations() - before; got != 2 {
		t.Errorf("op ran %d times after an item write, want 2 — the list and its bulk get", got)
	}
}

func TestAVaultWriteInvalidatesTheVaultList(t *testing.T) {
	client, invocations := countingOp(t)
	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.client = client
	run(t, loadVaults(client, model.cache))

	dispatch(t, model, writeSucceededMsg{Reload: VaultPane})
	run(t, loadVaults(client, model.cache))

	if got := invocations(); got != 2 {
		t.Errorf("op ran %d times across a write, want 2", got)
	}
}

func TestAnItemWriteKeepsTheVaultList(t *testing.T) {
	client, invocations := countingOp(t)
	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.client = client
	run(t, loadVaults(client, model.cache))

	dispatch(t, model, writeSucceededMsg{Reload: ItemPane})
	run(t, loadVaults(client, model.cache))

	if got := invocations(); got != 1 {
		t.Errorf("op ran %d times, want 1 — an item write does not change the vault list", got)
	}
}

func TestRefreshRefetchesEverything(t *testing.T) {
	client, invocations := countingOp(t)
	model := itemWriteModel(t, client)
	run(t, loadVaults(client, model.cache))
	before := invocations()

	refreshed, cmd := press(t, model, "R")
	if _, ok := run(t, cmd).(vaultsLoadedMsg); !ok {
		t.Fatal("refresh did not reload the vault list")
	}

	run(t, loadItems(client, refreshed.cache, personalVaultID))
	run(t, loadItem(client, refreshed.cache, personalVaultID, "item1"))

	if got := invocations() - before; got != 3 {
		t.Errorf("op ran %d times after a refresh, want 3 — vaults, the item list, its bulk get", got)
	}
}

// TestRepeatedNavigationRefetchesNothing walks the cascade a user walks moving
// down a vault list and back up. Without a cache each read is one subprocess,
// so the nine reads below would be nine `op` processes. Each vault costs a
// list and a bulk get, and only an item the bulk read did not return costs
// anything more.
func TestRepeatedNavigationRefetchesNothing(t *testing.T) {
	client, invocations := countingOp(t)
	store := newCache()

	run(t, loadVaults(client, store))
	run(t, loadItems(client, store, personalVaultID))
	run(t, loadItem(client, store, personalVaultID, "item1"))
	run(t, loadItems(client, store, "vault-b"))
	run(t, loadItem(client, store, "vault-b", "item9"))
	run(t, loadItems(client, store, personalVaultID))
	run(t, loadItem(client, store, personalVaultID, "item1"))
	run(t, loadItem(client, store, personalVaultID, "item2"))
	run(t, loadItem(client, store, personalVaultID, "item1"))

	if got := invocations(); got != 7 {
		t.Errorf("op ran %d times walking the cascade twice, want 7", got)
	}
}
