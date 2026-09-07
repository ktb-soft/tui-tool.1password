package form

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

func loginItem() op.Item {
	return op.Item{
		ID:       "1111",
		Name:     "Example Login",
		Category: "LOGIN",
		Vault:    op.VaultRef{ID: "aaaa"},
		Fields: []op.Field{
			{ID: "username", Label: "username", Type: op.FieldTypeString,
				Purpose: "USERNAME", Value: "user@example.com"},
			{ID: "password", Label: "password", Type: op.FieldTypeConcealed,
				Purpose: "PASSWORD", Value: "hunter2"},
		},
	}
}

// send delivers one message and then the messages the form's own commands
// produce, which is what the tea runtime does and what advances form state.
func send(form *huh.Form, msg tea.Msg) {
	_, cmd := form.Update(msg)
	drain(form, cmd, 16)
}

func drain(form *huh.Form, cmd tea.Cmd, budget int) {
	if cmd == nil || budget == 0 {
		return
	}

	switch msg := cmd().(type) {
	case nil:
	case tea.BatchMsg:
		for _, inner := range msg {
			drain(form, inner, budget-1)
		}
	default:
		_, next := form.Update(msg)
		drain(form, next, budget-1)
	}
}

func TestAnEmptyItemFormOffersACategorySelect(t *testing.T) {
	form, draft := NewItem(op.Item{Vault: op.VaultRef{ID: "aaaa"}})
	form.Init()

	if got, want := draft.Category, op.Categories[0]; got != want {
		t.Errorf("Category = %q, want %q", got, want)
	}
	if !strings.Contains(form.View(), theme.ItemCategoryPrompt) {
		t.Error("the create form does not offer a category")
	}
	if !strings.Contains(form.View(), op.Categories[1]) {
		t.Errorf("the create form does not list %q as an option", op.Categories[1])
	}
}

func TestEditingAnItemShowsTheCategoryReadOnly(t *testing.T) {
	form, _ := NewItem(loginItem())
	form.Init()

	view := form.View()
	if !strings.Contains(view, "LOGIN") {
		t.Error("the edit form does not name the item's category")
	}
	for _, other := range op.Categories[1:] {
		if strings.Contains(view, other) {
			t.Fatalf("the edit form offers %q as an alternative category", other)
		}
	}
}

func TestAnEmptyTitleBlocksTheForm(t *testing.T) {
	form, _ := NewItem(op.Item{Vault: op.VaultRef{ID: "aaaa"}})
	form.Init()

	for range 40 {
		send(form, tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	if form.State == huh.StateCompleted {
		t.Error("the form completed with an empty title")
	}
	if !strings.Contains(form.View(), theme.RequiredValidation) {
		t.Error("the form does not report the empty title")
	}
}

func TestTheDraftSeedsFromTheItemAndFoldsBack(t *testing.T) {
	_, draft := NewItem(loginItem())

	if got, want := len(draft.Fields), 3; got != want {
		t.Fatalf("draft fields = %d, want %d (two seeded plus a blank row)", got, want)
	}
	if got, want := draft.Fields[1].Value, "hunter2"; got != want {
		t.Errorf("seeded concealed value = %q, want %q", got, want)
	}

	item := draft.Item()
	if got, want := len(item.Fields), 2; got != want {
		t.Fatalf("folded fields = %d, want %d", got, want)
	}
	if got, want := item.Fields[1].Purpose, "PASSWORD"; got != want {
		t.Errorf("Purpose = %q, want %q (op-assigned purposes must round-trip)", got, want)
	}
	if got, want := item.ID, "1111"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	if got, want := item.Vault.ID, "aaaa"; got != want {
		t.Errorf("Vault.ID = %q, want %q", got, want)
	}
}

func TestAddingAFieldFillsTheBlankRow(t *testing.T) {
	_, draft := NewItem(loginItem())

	draft.Fields[2].Label = "recovery code"
	draft.Fields[2].Type = op.FieldTypeConcealed
	draft.Fields[2].Value = "abcd-efgh"

	fields := draft.Item().Fields
	if got, want := len(fields), 3; got != want {
		t.Fatalf("fields = %d, want %d", got, want)
	}
	if got, want := fields[2].Label, "recovery code"; got != want {
		t.Errorf("Label = %q, want %q", got, want)
	}
	if fields[2].ID != "" {
		t.Errorf("added field carries id %q, want op to assign one", fields[2].ID)
	}
}

func TestClearingALabelDeletesTheField(t *testing.T) {
	_, draft := NewItem(loginItem())

	draft.Fields[0].Label = "   "

	fields := draft.Item().Fields
	if got, want := len(fields), 1; got != want {
		t.Fatalf("fields = %d, want %d", got, want)
	}
	if got, want := fields[0].Label, "password"; got != want {
		t.Errorf("surviving field = %q, want %q", got, want)
	}
}

func TestAnItemWithNoFieldsProducesNone(t *testing.T) {
	_, draft := NewItem(op.Item{Name: "Example", Vault: op.VaultRef{ID: "aaaa"}})

	if fields := draft.Item().Fields; len(fields) != 0 {
		t.Errorf("fields = %d, want 0 for an untouched blank row", len(fields))
	}
}

func TestAFieldWithAnEmptyValueSurvives(t *testing.T) {
	_, draft := NewItem(op.Item{Name: "Example", Vault: op.VaultRef{ID: "aaaa"}})

	draft.Fields[0].Label = "username"

	fields := draft.Item().Fields
	if got, want := len(fields), 1; got != want {
		t.Fatalf("fields = %d, want %d", got, want)
	}
	if fields[0].Value != "" {
		t.Errorf("Value = %q, want empty", fields[0].Value)
	}
}

// valueView pages the form forward to the first value input and returns what
// it renders.
func valueView(t *testing.T, form *huh.Form) string {
	t.Helper()

	for range 12 {
		if strings.Contains(form.View(), theme.FieldValuePrompt) {
			return form.View()
		}
		form.NextGroup()
	}
	t.Fatal("the form never reached a value input")
	return ""
}

func oneFieldItem(fieldType, value string) op.Item {
	return op.Item{
		ID: "1111", Name: "Example", Category: "LOGIN", Vault: op.VaultRef{ID: "aaaa"},
		Fields: []op.Field{{ID: "f", Label: "password", Type: fieldType, Value: value}},
	}
}

func TestAConcealedValueIsMasked(t *testing.T) {
	form, _ := NewItem(oneFieldItem(op.FieldTypeConcealed, "swordfish"))
	form.Init()

	if view := valueView(t, form); strings.Contains(view, "swordfish") {
		t.Error("the concealed value is echoed on screen")
	}
}

func TestAPlainValueIsNotMasked(t *testing.T) {
	form, _ := NewItem(oneFieldItem(op.FieldTypeString, "user@example.com"))
	form.Init()

	if view := valueView(t, form); !strings.Contains(view, "user@example.com") {
		t.Error("a plain value is masked, want it echoed")
	}
}

func TestDeleteConfirmationNamesTheItemAndDefaultsToNo(t *testing.T) {
	form, confirmed := NewDeleteItem("Example Login")
	form.Init()

	if *confirmed {
		t.Error("the delete confirmation defaults to yes")
	}
	if !strings.Contains(form.View(), "Example Login") {
		t.Error("the delete confirmation does not name the item")
	}
}

func TestDeleteConfirmationRecordsAYes(t *testing.T) {
	form, confirmed := NewDeleteItem("Example Login")
	form.Init()

	send(form, tea.KeyPressMsg{Code: tea.KeyLeft})
	send(form, tea.KeyPressMsg{Code: tea.KeyEnter})

	if !*confirmed {
		t.Error("selecting yes did not set the confirmation")
	}
	if form.State != huh.StateCompleted {
		t.Errorf("State = %v, want completed", form.State)
	}
}
