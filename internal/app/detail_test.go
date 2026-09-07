package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

const secretValue = "hunter2"

func loginItem() op.Item {
	return op.Item{
		ID:       "1111111111111111111111111a",
		Name:     "Example Login",
		Category: "LOGIN",
		Vault:    op.VaultRef{ID: "aaaaaaaaaaaaaaaaaaaaaaaaaa", Name: "Personal"},
		Fields: []op.Field{
			{ID: "username", Label: "username", Type: op.FieldTypeString, Value: "user@example.com"},
			{ID: "password", Label: "password", Type: op.FieldTypeConcealed, Value: secretValue},
		},
	}
}

func noteItem() op.Item {
	return op.Item{
		ID:       "2222222222222222222222222b",
		Name:     "Example Note",
		Category: "SECURE_NOTE",
		Fields:   []op.Field{{ID: "notesPlain", Label: "notesPlain", Type: op.FieldTypeString, Value: "plain"}},
	}
}

func detailModel(t *testing.T, item op.Item) Model {
	t.Helper()

	model := New(op.Client{})
	model.resizePanes(120, 40)
	model.focus = DetailPane
	model.applyFocus()

	updated, _ := model.Update(itemLoadedMsg(item))
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want app.Model", updated)
	}
	return next
}

func TestItemLoadedRendersMaskedFields(t *testing.T) {
	model := detailModel(t, loginItem())
	view := model.render()

	if !strings.Contains(view, "username") {
		t.Fatal("loaded item's fields are not rendered")
	}
	if strings.Contains(view, secretValue) {
		t.Fatal("concealed value rendered without a reveal")
	}
	if !strings.Contains(view, theme.Mask) {
		t.Fatal("mask not rendered for concealed field")
	}
}

func TestRevealKeyUnmasksAndItemChangeResets(t *testing.T) {
	model := detailModel(t, loginItem())

	revealed, _ := model.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	model, ok := revealed.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want app.Model", revealed)
	}
	if !strings.Contains(model.render(), secretValue) {
		t.Fatal("reveal key did not unmask the concealed field")
	}

	changed, _ := model.Update(itemLoadedMsg(noteItem()))
	model, ok = changed.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want app.Model", changed)
	}
	if strings.Contains(model.render(), secretValue) {
		t.Fatal("concealed value survived a change of item")
	}
}

func TestRevealKeyIsIgnoredWhenAnotherPaneHasFocus(t *testing.T) {
	model := detailModel(t, loginItem())
	model.focus = VaultPane
	model.applyFocus()

	updated, _ := model.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	model, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want app.Model", updated)
	}
	if strings.Contains(model.render(), secretValue) {
		t.Fatal("reveal fired while the vault pane held focus")
	}
}

func TestDetailShowsEmptyStateBeforeAnyItemLoads(t *testing.T) {
	model := New(op.Client{})
	model.resizePanes(120, 40)

	if !strings.Contains(model.render(), theme.NoSelection) {
		t.Fatalf("detail pane does not show %q before a load", theme.NoSelection)
	}
}

func TestWriteSucceededReloadsNothingWithoutAnItem(t *testing.T) {
	model := New(op.Client{})
	model.resizePanes(120, 40)

	_, cmd := model.updateDetail(writeSucceededMsg{Reload: DetailPane, Status: theme.SavedStatus})
	if cmd != nil {
		t.Fatal("a write reloaded the detail pane with no item displayed")
	}
}

func TestWriteSucceededReloadsTheDisplayedItem(t *testing.T) {
	model := detailModel(t, loginItem())

	_, cmd := model.updateDetail(writeSucceededMsg{Reload: DetailPane, Status: theme.SavedStatus})
	if cmd == nil {
		t.Fatal("a write did not reload the displayed item")
	}
	if _, ok := cmd().(opFailedMsg); !ok {
		t.Fatal("reload did not run GetItem")
	}
}

func TestTerminalTooSmallReplacesThePanes(t *testing.T) {
	model := detailModel(t, loginItem())
	model.resizePanes(theme.MinWidth-1, theme.MinHeight-1)

	view := model.render()
	if !strings.Contains(view, theme.TerminalTooSmall) {
		t.Fatal("a terminal below the minimum does not show the too-small notice")
	}
	if strings.Contains(view, secretValue) {
		t.Fatal("concealed value rendered in a too-small terminal")
	}
}
