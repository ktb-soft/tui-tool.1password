package op

import (
	"encoding/json"
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

// secretValue is the concealed value every write test round-trips. No argv the
// package builds may ever contain it.
const secretValue = "hunter2"

func loginTemplate() Item {
	return Item{
		Name:     "Example Login",
		Category: "LOGIN",
		Vault:    VaultRef{ID: "aaaaaaaaaaaaaaaaaaaaaaaaaa"},
		Fields: []Field{
			{Label: "username", Type: FieldTypeString, Value: "user@example.com"},
			{Label: "password", Type: FieldTypeConcealed, Value: secretValue},
		},
	}
}

func TestCreateItemPipesTheTemplateAndKeepsValuesOutOfArgv(t *testing.T) {
	path, argvFile, stdinFile := fakeOp(t, fixtureJSON(t, "item-login.json"), 0)

	created, err := Client{Path: path}.CreateItem(loginTemplate())
	if err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	want := []string{
		"item", "create", vaultFlag, "aaaaaaaaaaaaaaaaaaaaaaaaaa", stdinArg, formatFlag,
	}
	argv := readLines(t, argvFile)
	if !slices.Equal(argv, want) {
		t.Errorf("argv = %v, want %v", argv, want)
	}

	var piped Item
	if err := json.Unmarshal([]byte(readFile(t, stdinFile)), &piped); err != nil {
		t.Fatalf("decode piped template: %v", err)
	}
	if got, want := piped.Fields[1].Value, secretValue; got != want {
		t.Errorf("piped concealed value = %q, want %q", got, want)
	}
	if got, want := piped.Fields[1].Type, FieldTypeConcealed; got != want {
		t.Errorf("piped concealed type = %q, want %q", got, want)
	}
	if piped.ID != "" {
		t.Errorf("piped id = %q, want it omitted on create", piped.ID)
	}
	if got, want := created.ID, "1111111111111111111111111a"; got != want {
		t.Errorf("created ID = %q, want %q", got, want)
	}
}

func TestEditItemPipesTheTemplateAddressedByID(t *testing.T) {
	path, argvFile, stdinFile := fakeOp(t, fixtureJSON(t, "item-login.json"), 0)

	item := loginTemplate()
	item.ID = "1111111111111111111111111a"

	if _, err := (Client{Path: path}).EditItem(item); err != nil {
		t.Fatalf("EditItem: %v", err)
	}

	want := []string{"item", "edit", "1111111111111111111111111a", stdinArg, formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if !strings.Contains(readFile(t, stdinFile), secretValue) {
		t.Error("the concealed value did not reach stdin")
	}
}

// TestItemWritesNeverPutAValueInArgv is the regression that matters most: a
// value in argv is readable by every other process on the machine. See
// docs/adr/04-00-00-secrets-never-in-argv.md.
func TestItemWritesNeverPutAValueInArgv(t *testing.T) {
	item := loginTemplate()
	item.ID = "1111111111111111111111111a"
	item.Tags = []string{"work"}
	item.Fields = append(item.Fields, Field{
		Label: "note", Type: FieldTypeString, Value: "renewal 2027-01",
	})

	writes := map[string]func(Client) error{
		"create": func(c Client) error { _, err := c.CreateItem(item); return err },
		"edit":   func(c Client) error { _, err := c.EditItem(item); return err },
		"delete": func(c Client) error { return c.DeleteItem(item.Vault.ID, item.ID) },
	}

	for name, write := range writes {
		t.Run(name, func(t *testing.T) {
			path, argvFile, _ := fakeOp(t, fixtureJSON(t, "item-login.json"), 0)
			if err := write(Client{Path: path, Account: "example"}); err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			argv := readLines(t, argvFile)
			for _, field := range item.Fields {
				for _, arg := range argv {
					if strings.Contains(arg, field.Value) {
						t.Fatalf("argv %v carries the value of field %q", argv, field.Label)
					}
				}
			}
			for _, arg := range argv {
				if !strings.HasPrefix(arg, "-") && strings.Contains(arg, "=") {
					t.Errorf("argv %v carries an assignment statement %q", argv, arg)
				}
			}
		})
	}
}

func TestCreateItemSendsAnItemWithNoFields(t *testing.T) {
	path, _, stdinFile := fakeOp(t, fixtureJSON(t, "item-empty.json"), 0)

	empty := Item{Name: "Example Empty Item", Category: "SECURE_NOTE", Vault: VaultRef{ID: "c"}}
	if _, err := (Client{Path: path}).CreateItem(empty); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	var piped Item
	if err := json.Unmarshal([]byte(readFile(t, stdinFile)), &piped); err != nil {
		t.Fatalf("decode piped template: %v", err)
	}
	if len(piped.Fields) != 0 {
		t.Errorf("piped fields = %d, want 0", len(piped.Fields))
	}
	if !piped.UpdatedAt.IsZero() {
		t.Error("piped template carries a timestamp, want it omitted")
	}
}

func TestCreateItemSendsAFieldWithNoValue(t *testing.T) {
	path, _, stdinFile := fakeOp(t, fixtureJSON(t, "item-login.json"), 0)

	item := Item{Name: "Example", Category: "LOGIN", Vault: VaultRef{ID: "a"}}
	item.Fields = []Field{{Label: "username", Type: FieldTypeString}}
	if _, err := (Client{Path: path}).CreateItem(item); err != nil {
		t.Fatalf("CreateItem: %v", err)
	}

	var piped Item
	if err := json.Unmarshal([]byte(readFile(t, stdinFile)), &piped); err != nil {
		t.Fatalf("decode piped template: %v", err)
	}
	if got, want := len(piped.Fields), 1; got != want {
		t.Fatalf("piped fields = %d, want %d", got, want)
	}
	if got := piped.Fields[0].Value; got != "" {
		t.Errorf("piped value = %q, want empty", got)
	}
}

func TestItemWritesSurfaceTheOpError(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)
	client := Client{Path: path}

	if _, err := client.CreateItem(loginTemplate()); !isOpError(t, err) {
		t.Error("CreateItem did not surface the op error")
	}
	if _, err := client.EditItem(loginTemplate()); !isOpError(t, err) {
		t.Error("EditItem did not surface the op error")
	}
	if err := client.DeleteItem("a", "b"); !isOpError(t, err) {
		t.Error("DeleteItem did not surface the op error")
	}
}

func TestDeleteItemNamesTheItemAndTheVault(t *testing.T) {
	path, argvFile, stdinFile := fakeOp(t, "", 0)

	if err := (Client{Path: path}).DeleteItem("aaaaaaaaaaaaaaaaaaaaaaaaaa", "1111"); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}

	want := []string{"item", "delete", "1111", vaultFlag, "aaaaaaaaaaaaaaaaaaaaaaaaaa", formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if got := readFile(t, stdinFile); got != "" {
		t.Errorf("stdin = %q, want empty", got)
	}
}

func TestEditItemReportsMalformedJSON(t *testing.T) {
	path, _, _ := fakeOp(t, "{oops", 0)

	item := loginTemplate()
	item.ID = "1111"
	if _, err := (Client{Path: path}).EditItem(item); err == nil {
		t.Fatal("EditItem: want a decode error, got nil")
	} else if strings.Contains(err.Error(), secretValue) {
		t.Errorf("err = %q, it leaks the concealed value", err)
	}
}

func isOpError(t *testing.T, err error) bool {
	t.Helper()
	var opErr *OpError
	return errors.As(err, &opErr)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
