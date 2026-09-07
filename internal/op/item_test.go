package op

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func fixtureJSON(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(fixtureDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

func TestListItemsRequestsTheVaultAndDecodesTheList(t *testing.T) {
	path, argvFile, _ := fakeOp(t, fixtureJSON(t, "item-list.json"), 0)

	items, err := Client{Path: path}.ListItems("aaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}

	want := []string{"item", "list", vaultFlag, "aaaaaaaaaaaaaaaaaaaaaaaaaa", formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if len(items) != 4 {
		t.Fatalf("len = %d, want 4", len(items))
	}
	if got, want := items[0].Name, "Example Login"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := items[2].Category, "WIRELESS_ROUTER"; got != want {
		t.Errorf("Category = %q, want %q", got, want)
	}
}

func TestListItemsReturnsNoItemsForAnEmptyVault(t *testing.T) {
	path, _, _ := fakeOp(t, fixtureJSON(t, "empty-list.json"), 0)

	items, err := Client{Path: path}.ListItems("cccccccccccccccccccccccccc")
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("len = %d, want 0", len(items))
	}
}

func TestListItemsSurfacesTheOpErrorVerbatim(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	_, err := Client{Path: path}.ListItems("aaaaaaaaaaaaaaaaaaaaaaaaaa")

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %v, want *OpError", err)
	}
	if !strings.Contains(opErr.Error(), "not signed in") {
		t.Errorf("Error() = %q, want the stderr text", opErr.Error())
	}
}

func TestListItemsReportsAMissingBinary(t *testing.T) {
	_, err := Client{Path: filepath.Join(t.TempDir(), "absent")}.ListItems("a")

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %v, want *OpError", err)
	}
}

func TestListItemsReportsMalformedJSON(t *testing.T) {
	path, _, _ := fakeOp(t, "not json", 0)

	_, err := Client{Path: path}.ListItems("aaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err == nil {
		t.Fatal("ListItems: want a decode error, got nil")
	}
	if !strings.Contains(err.Error(), "aaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Errorf("err = %q, want it to name the vault", err)
	}
}

func TestGetItemRequestsTheItemAndDecodesItsFields(t *testing.T) {
	path, argvFile, stdinFile := fakeOp(t, fixtureJSON(t, "item-login.json"), 0)

	item, err := Client{Path: path}.GetItem("aaaaaaaaaaaaaaaaaaaaaaaaaa", "1111111111111111111111111a")
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}

	want := []string{
		"item", "get", "1111111111111111111111111a",
		vaultFlag, "aaaaaaaaaaaaaaaaaaaaaaaaaa", formatFlag,
	}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if stdin, err := os.ReadFile(stdinFile); err != nil || len(stdin) != 0 {
		t.Errorf("stdin = %q (err %v), want empty", stdin, err)
	}

	if got, want := len(item.Fields), 4; got != want {
		t.Fatalf("len(Fields) = %d, want %d", got, want)
	}
	if got, want := item.Category, "LOGIN"; got != want {
		t.Errorf("Category = %q, want %q", got, want)
	}
	if got, want := item.Fields[1].Value, "hunter2"; got != want {
		t.Errorf("password value = %q, want %q", got, want)
	}
	if !item.Fields[1].IsSecret() {
		t.Error("password field is not concealed")
	}
	if item.Fields[0].IsSecret() {
		t.Error("username field is concealed, want plain")
	}
	if got, want := item.Fields[3].Type, FieldTypeOTP; got != want {
		t.Errorf("otp field type = %q, want %q", got, want)
	}
	if got, want := item.URLs[0].HRef, "https://example.com"; got != want {
		t.Errorf("URL = %q, want %q", got, want)
	}
	if !item.URLs[0].Primary {
		t.Error("URL is not marked primary")
	}
	if item.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero, want the parsed timestamp")
	}
}

func TestGetItemCarriesSectionsAndTheirFields(t *testing.T) {
	path, _, _ := fakeOp(t, fixtureJSON(t, "item-sections.json"), 0)

	item, err := Client{Path: path}.GetItem("aaaaaaaaaaaaaaaaaaaaaaaaaa", "3333333333333333333333333c")
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}

	if got, want := len(item.Sections), 2; got != want {
		t.Fatalf("len(Sections) = %d, want %d", got, want)
	}
	if got, want := item.Sections[1].Label, "Guest network"; got != want {
		t.Errorf("Sections[1].Label = %q, want %q", got, want)
	}

	sectioned := map[string]int{}
	for _, field := range item.Fields {
		if field.Section == nil {
			sectioned[""]++
			continue
		}
		sectioned[field.Section.ID]++
	}
	if got, want := sectioned[""], 2; got != want {
		t.Errorf("sectionless fields = %d, want %d", got, want)
	}
	if got, want := sectioned["sec_admin"], 2; got != want {
		t.Errorf("Admin fields = %d, want %d", got, want)
	}
	if got, want := item.Tags, []string{"home", "network"}; !slices.Equal(got, want) {
		t.Errorf("Tags = %v, want %v", got, want)
	}
}

func TestGetItemDecodesAnItemWithNoFields(t *testing.T) {
	path, _, _ := fakeOp(t, fixtureJSON(t, "item-empty.json"), 0)

	item, err := Client{Path: path}.GetItem("cccccccccccccccccccccccccc", "5555555555555555555555555e")
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}
	if len(item.Fields) != 0 {
		t.Errorf("len(Fields) = %d, want 0", len(item.Fields))
	}
	if got, want := item.Name, "Example Empty Item"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
}

func TestGetItemSurfacesTheOpError(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	_, err := Client{Path: path}.GetItem("a", "b")

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %v, want *OpError", err)
	}
	if !strings.HasPrefix(opErr.Error(), "op item get b") {
		t.Errorf("Error() = %q, want it to name the command", opErr.Error())
	}
}

func TestGetItemReportsMalformedJSON(t *testing.T) {
	path, _, _ := fakeOp(t, "{oops", 0)

	_, err := Client{Path: path}.GetItem("a", "1111111111111111111111111a")
	if err == nil {
		t.Fatal("GetItem: want a decode error, got nil")
	}
	if !strings.Contains(err.Error(), "1111111111111111111111111a") {
		t.Errorf("err = %q, want it to name the item", err)
	}
}

func TestItemWritesRemainUnimplemented(t *testing.T) {
	client := Client{Path: filepath.Join(t.TempDir(), "absent")}

	if _, err := client.CreateItem(Item{}); !errors.Is(err, errNotImplemented) {
		t.Errorf("CreateItem err = %v, want errNotImplemented", err)
	}
	if _, err := client.EditItem(Item{}); !errors.Is(err, errNotImplemented) {
		t.Errorf("EditItem err = %v, want errNotImplemented", err)
	}
	if err := client.DeleteItem("a", "b"); !errors.Is(err, errNotImplemented) {
		t.Errorf("DeleteItem err = %v, want errNotImplemented", err)
	}
}
