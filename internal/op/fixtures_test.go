package op

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"charm.land/bubbles/v2/list"
)

const fixtureDir = "../../testdata"

func decodeFixture[T any](t *testing.T, name string) T {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return value
}

func TestVaultListFixtureDecodes(t *testing.T) {
	vaults := decodeFixture[[]Vault](t, "vault-list.json")

	if len(vaults) != 3 {
		t.Fatalf("len = %d, want 3", len(vaults))
	}
	if got, want := vaults[0].Name, "Personal"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := vaults[0].Items, 4; got != want {
		t.Errorf("Items = %d, want %d", got, want)
	}
	if got, want := vaults[1].Note, "example team vault"; got != want {
		t.Errorf("Note = %q, want %q", got, want)
	}
}

func TestEmptyListFixtureDecodes(t *testing.T) {
	if vaults := decodeFixture[[]Vault](t, "empty-list.json"); len(vaults) != 0 {
		t.Errorf("len = %d, want 0", len(vaults))
	}
	if items := decodeFixture[[]Item](t, "empty-list.json"); len(items) != 0 {
		t.Errorf("len = %d, want 0", len(items))
	}
}

func TestItemListFixtureDecodes(t *testing.T) {
	items := decodeFixture[[]Item](t, "item-list.json")

	if len(items) != 4 {
		t.Fatalf("len = %d, want 4", len(items))
	}
	if got, want := items[0].Name, "Example Login"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := items[0].Vault.Name, "Personal"; got != want {
		t.Errorf("Vault.Name = %q, want %q", got, want)
	}
	if items[0].UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero, want the parsed timestamp")
	}
}

func TestLoginFixtureCarriesConcealedField(t *testing.T) {
	item := decodeFixture[Item](t, "item-login.json")

	if got, want := len(item.Fields), 4; got != want {
		t.Fatalf("len(Fields) = %d, want %d", got, want)
	}

	secrets := 0
	for _, field := range item.Fields {
		if field.IsSecret() {
			secrets++
		}
	}
	if secrets != 1 {
		t.Errorf("concealed fields = %d, want 1", secrets)
	}
	if got, want := item.URLs[0].HRef, "https://example.com"; got != want {
		t.Errorf("URL = %q, want %q", got, want)
	}
}

func TestSectionsFixtureGroupsFields(t *testing.T) {
	item := decodeFixture[Item](t, "item-sections.json")

	if got, want := len(item.Sections), 2; got != want {
		t.Fatalf("len(Sections) = %d, want %d", got, want)
	}

	sectionless := 0
	for _, field := range item.Fields {
		if field.Section == nil {
			sectionless++
		}
	}
	if sectionless != 2 {
		t.Errorf("sectionless fields = %d, want 2", sectionless)
	}
	if got, want := item.Fields[2].Section.Label, "Admin"; got != want {
		t.Errorf("Section.Label = %q, want %q", got, want)
	}
}

func TestItemWithNoFieldsDecodes(t *testing.T) {
	item := decodeFixture[Item](t, "item-empty.json")

	if len(item.Fields) != 0 {
		t.Errorf("len(Fields) = %d, want 0", len(item.Fields))
	}
	if item.ID == "" {
		t.Error("ID is empty, want the fixture id")
	}
}

func TestDomainTypesSatisfyListItem(t *testing.T) {
	var _ list.DefaultItem = Vault{}
	var _ list.DefaultItem = Item{}

	vault := Vault{Name: "Personal", Items: 4}
	if got, want := vault.Title(), "Personal"; got != want {
		t.Errorf("Vault.Title() = %q, want %q", got, want)
	}
	if got, want := vault.Description(), "4 items"; got != want {
		t.Errorf("Vault.Description() = %q, want %q", got, want)
	}
	if got, want := vault.FilterValue(), "Personal"; got != want {
		t.Errorf("Vault.FilterValue() = %q, want %q", got, want)
	}

	item := Item{Name: "Example Login", Category: "LOGIN"}
	if got, want := item.Title(), "Example Login"; got != want {
		t.Errorf("Item.Title() = %q, want %q", got, want)
	}
	if got, want := item.Description(), "LOGIN"; got != want {
		t.Errorf("Item.Description() = %q, want %q", got, want)
	}
	if got, want := item.FilterValue(), "Example Login"; got != want {
		t.Errorf("Item.FilterValue() = %q, want %q", got, want)
	}
}
