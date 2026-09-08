package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// vaultOp writes a fake op serving a vault of itemCount items, counting its
// invocations the way countingOp does. `item get -` answers with every item,
// or exits 1 when bulkFails, so a test can make the bulk read fail while the
// per-item read still works.
func vaultOp(t *testing.T, itemCount int, bulkFails bool) (op.Client, func() int, []string) {
	t.Helper()

	ids, records := vaultRecords(itemCount)
	all := "[" + strings.Join(records, ",") + "]"
	one := "{}"
	if itemCount > 0 {
		one = records[0]
	}

	bulk := "  printf '%s' '" + all + "'\n  exit 0\n"
	if bulkFails {
		bulk = "  echo 'bulk read failed' >&2\n  exit 1\n"
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "op")
	counter := filepath.Join(dir, "invocations")

	script := "#!/bin/sh\n" +
		"echo ran >> " + counter + "\n" +
		"if [ \"$1 $2 $3\" = 'item get -' ]; then\n" +
		bulk +
		"fi\n" +
		"case \"$1 $2\" in\n" +
		"'item get') printf '%s' '" + one + "' ;;\n" +
		"'item list') printf '%s' '" + all + "' ;;\n" +
		"*) printf '%s' '[" + countedVault + "]' ;;\n" +
		"esac\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake op: %v", err)
	}

	return op.Client{Path: path}, invocationCounter(t, counter), ids
}

// vaultRecords builds itemCount item records, each carrying one concealed
// field, so a test can tell a bulk-read hit from a per-item fetch.
func vaultRecords(itemCount int) (ids, records []string) {
	ids = make([]string, itemCount)
	records = make([]string, itemCount)
	for i := range ids {
		ids[i] = "item" + strconv.Itoa(i)
		records[i] = fmt.Sprintf(
			`{"id":%q,"title":"Item %d","category":"LOGIN",`+
				`"fields":[{"id":"password","label":"password","type":"CONCEALED","value":"hunter%d"}]}`,
			ids[i], i, i)
	}
	return ids, records
}

func invocationCounter(t *testing.T, counter string) func() int {
	t.Helper()

	return func() int {
		data, err := os.ReadFile(counter)
		if os.IsNotExist(err) {
			return 0
		}
		if err != nil {
			t.Fatalf("read invocations: %v", err)
		}
		return strings.Count(string(data), "\n")
	}
}

// TestBrowsingAWholeVaultCostsTwoSubprocesses is the point of the bulk read:
// the cursor visiting every row of a vault must not cost one `op` per row.
func TestBrowsingAWholeVaultCostsTwoSubprocesses(t *testing.T) {
	client, invocations, ids := vaultOp(t, 40, false)
	store := newCache()

	run(t, loadItems(client, store, personalVaultID))
	for _, id := range ids {
		if _, ok := run(t, loadItem(client, store, personalVaultID, id)).(itemLoadedMsg); !ok {
			t.Fatalf("item %s did not load", id)
		}
	}

	if got := invocations(); got != 2 {
		t.Errorf("op ran %d times to browse %d items, want 2", got, len(ids))
	}
}

func TestTheBulkReadCachesTheFieldValues(t *testing.T) {
	client, _, ids := vaultOp(t, 3, false)
	store := newCache()

	run(t, loadItems(client, store, personalVaultID))
	loaded, ok := run(t, loadItem(client, store, personalVaultID, ids[2])).(itemLoadedMsg)
	if !ok {
		t.Fatal("the cached item did not load")
	}
	fields := op.Item(loaded).Fields
	if len(fields) != 1 {
		t.Fatalf("cached item has %d fields, want the bulk read's 1", len(fields))
	}
	if got, want := fields[0].Value, "hunter2"; got != want {
		t.Errorf("cached value = %q, want %q, the bulk read's value for that item", got, want)
	}
}

func TestAnEmptyVaultRunsNoBulkRead(t *testing.T) {
	client, invocations, _ := vaultOp(t, 0, true)
	store := newCache()

	loaded, ok := run(t, loadItems(client, store, personalVaultID)).(itemsLoadedMsg)
	if !ok {
		t.Fatal("the empty vault did not load")
	}
	if len(loaded) != 0 {
		t.Fatalf("loaded %d items, want none", len(loaded))
	}
	if got := invocations(); got != 1 {
		t.Errorf("op ran %d times for an empty vault, want 1, the list alone", got)
	}
}

// TestAFailedBulkReadFallsBackToPerItemFetches keeps a vault usable when the
// bulk read fails: the list still fills and each selection fetches itself.
func TestAFailedBulkReadFallsBackToPerItemFetches(t *testing.T) {
	client, invocations, ids := vaultOp(t, 3, true)
	store := newCache()

	loaded, ok := run(t, loadItems(client, store, personalVaultID)).(itemsLoadedMsg)
	if !ok {
		t.Fatal("a failed bulk read emptied the item pane")
	}
	if len(loaded) != len(ids) {
		t.Fatalf("loaded %d items, want %d", len(loaded), len(ids))
	}

	for _, id := range ids {
		if _, ok := run(t, loadItem(client, store, personalVaultID, id)).(itemLoadedMsg); !ok {
			t.Fatalf("item %s did not load after the failed bulk read", id)
		}
	}

	if got, want := invocations(), 2+len(ids); got != want {
		t.Errorf("op ran %d times, want %d: list, failed bulk, one per item", got, want)
	}
}
