package op

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestGetItemsSendsTheIDsOnStdinAndDecodesTheArray(t *testing.T) {
	path, argvFile, stdinFile := fakeOp(t, "["+fixtureJSON(t, "item-login.json")+"]", 0)

	listed := decodeFixture[[]Item](t, "item-list.json")
	items, err := Client{Path: path}.GetItems(listed)
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}

	want := []string{"item", "get", stdinArg, formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}

	var sent []map[string]string
	if err := json.Unmarshal([]byte(readFile(t, stdinFile)), &sent); err != nil {
		t.Fatalf("stdin is not a JSON array: %v", err)
	}
	if len(sent) != len(listed) {
		t.Fatalf("stdin carries %d specifiers, want %d", len(sent), len(listed))
	}
	for i, specifier := range sent {
		if got, want := specifier["id"], listed[i].ID; got != want {
			t.Errorf("stdin[%d].id = %q, want %q", i, got, want)
		}
		if len(specifier) != 1 {
			t.Errorf("stdin[%d] = %v, want the id alone", i, specifier)
		}
	}

	if len(items) != 1 || len(items[0].Fields) != 4 {
		t.Fatalf("decoded %+v, want the one item the fake op printed, with its fields", items)
	}
}

// TestGetItemsDecodesAStreamOfObjects covers the shape `op item get -` returns
// for a single specifier: a bare object rather than an array.
func TestGetItemsDecodesAStreamOfObjects(t *testing.T) {
	one := fixtureJSON(t, "item-login.json")
	path, _, _ := fakeOp(t, one+"\n"+one, 0)

	items, err := Client{Path: path}.GetItems([]Item{{ID: "a"}, {ID: "b"}})
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("decoded %d items, want 2", len(items))
	}
}

func TestGetItemsRunsNothingForAnEmptyVault(t *testing.T) {
	path, argvFile, _ := fakeOp(t, "[]", 0)

	items, err := Client{Path: path}.GetItems(nil)
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}
	if items != nil {
		t.Errorf("items = %v, want nil", items)
	}
	if _, err := os.Stat(argvFile); !os.IsNotExist(err) {
		t.Errorf("op ran with argv %v, want no subprocess at all", readLines(t, argvFile))
	}
}

func TestGetItemsReportsTheOpErrorVerbatim(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	_, err := Client{Path: path}.GetItems([]Item{{ID: "a"}})
	if err == nil {
		t.Fatal("GetItems succeeded, want the op failure")
	}
	if !strings.Contains(err.Error(), "op item get") {
		t.Errorf("err = %q, want the failing command named", err)
	}
}

// TestGetItemsNeverPutsAValueInArgv extends the argv rule to the bulk read:
// only ids leave the process, and they go on stdin. See
// docs/adr/04-00-00-secrets-never-in-argv.md.
func TestGetItemsNeverPutsAValueInArgv(t *testing.T) {
	path, argvFile, _ := fakeOp(t, "[]", 0)

	item := loginTemplate()
	item.ID = "1111111111111111111111111a"
	if _, err := (Client{Path: path}).GetItems([]Item{item}); err != nil {
		t.Fatalf("GetItems: %v", err)
	}

	argv := readLines(t, argvFile)
	for _, field := range item.Fields {
		for _, arg := range argv {
			if field.Value != "" && strings.Contains(arg, field.Value) {
				t.Fatalf("argv %v carries the value of field %q", argv, field.Label)
			}
		}
	}
}
