package app

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// onItemPane seeds a vault and one item, focuses the item pane, and loads the
// item's fields into the detail pane, which is the state edit requires.
func onItemPane(t *testing.T, loadDetail bool) Model {
	t.Helper()

	model := withVault(t, sized(t, 120, 40), personalVaultID)
	model.items.SetItems([]list.Item{loginItem()})
	model.focus = ItemPane
	if loadDetail {
		model.detail.SetItem(loginItem())
	}
	return model
}

func TestCreateOpensAnItemFormOnTheItemPane(t *testing.T) {
	model, _ := press(t, onItemPane(t, false), "a")

	if model.overlay == nil {
		t.Fatal("a did not open a form")
	}
	view := model.View().Content
	if !strings.Contains(view, theme.ItemTitlePrompt) {
		t.Errorf("form is missing %q", theme.ItemTitlePrompt)
	}
	if !strings.Contains(view, theme.ItemCategoryPrompt) {
		t.Errorf("create form is missing %q", theme.ItemCategoryPrompt)
	}
}

func TestCreateWithNoVaultSelectedOpensNothing(t *testing.T) {
	model := sized(t, 120, 40)
	model.focus = ItemPane

	if opened, _ := press(t, model, "a"); opened.overlay != nil {
		t.Error("a opened a form with no vault selected")
	}
}

func TestEditOpensTheFormSeededWithTheLoadedItem(t *testing.T) {
	model, _ := press(t, onItemPane(t, true), "e")

	if model.overlay == nil {
		t.Fatal("e did not open a form")
	}
	view := model.View().Content
	if !strings.Contains(view, "Example Login") {
		t.Error("the edit form is not seeded with the item title")
	}
	if strings.Contains(view, "hunter2") {
		t.Error("the edit form shows a concealed value on its first page")
	}
}

func TestEditWaitsForTheItemsFieldsToLoad(t *testing.T) {
	model, cmd := press(t, onItemPane(t, false), "e")

	if model.overlay != nil {
		t.Error("e opened a form before the item's fields were loaded")
	}
	if cmd == nil {
		t.Fatal("e reported nothing while the item was still loading")
	}
}

func TestSubmittingTheItemFormCreatesTheItem(t *testing.T) {
	model := onItemPane(t, false)
	model.client = fakeOpClient(t, `{"id":"9999","title":"New Item","category":"LOGIN"}`, 0)

	opened, _ := press(t, model, "a")
	form, draft := opened.overlay, opened.overlaySubmit
	if form == nil || draft == nil {
		t.Fatal("create did not register a submit")
	}

	_, cmd := draft(opened)
	msg, ok := cmd().(writeSucceededMsg)
	if !ok {
		t.Fatalf("submit returned %T, want writeSucceededMsg", cmd())
	}
	if msg.Reload != ItemPane {
		t.Errorf("Reload = %d, want ItemPane", msg.Reload)
	}
	if got, want := msg.Status, theme.CreatedStatus; got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestSubmittingAnEditSavesRatherThanCreates(t *testing.T) {
	model := onItemPane(t, true)
	model.client = fakeOpClient(t, `{"id":"1111","title":"Example Login","category":"LOGIN"}`, 0)

	opened, _ := press(t, model, "e")
	if opened.overlaySubmit == nil {
		t.Fatal("edit did not register a submit")
	}

	_, cmd := opened.overlaySubmit(opened)
	msg, ok := cmd().(writeSucceededMsg)
	if !ok {
		t.Fatalf("submit returned %T, want writeSucceededMsg", cmd())
	}
	if got, want := msg.Status, theme.SavedStatus; got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestAFailedItemWriteReportsTheOpErrorVerbatim(t *testing.T) {
	model := onItemPane(t, false)
	model.client = fakeOpClient(t, "", 1)

	opened, _ := press(t, model, "a")
	_, cmd := opened.overlaySubmit(opened)

	failed, ok := cmd().(opFailedMsg)
	if !ok {
		t.Fatalf("submit returned %T, want opFailedMsg", cmd())
	}
	if !strings.Contains(failed.Err.Error(), "not signed in") {
		t.Errorf("err = %q, want the stderr text", failed.Err)
	}
}

func TestAbortingTheItemFormWritesNothing(t *testing.T) {
	opened, _ := press(t, onItemPane(t, true), "e")

	next, _ := opened.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	aborted, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	if aborted.overlay != nil {
		t.Error("esc did not dismiss the item form")
	}
	if aborted.overlaySubmit != nil {
		t.Error("esc left the submit armed")
	}
}

func TestDeleteAsksForConfirmationNamingTheItem(t *testing.T) {
	model, _ := press(t, onItemPane(t, true), "d")

	if model.overlay == nil {
		t.Fatal("d did not open a confirmation")
	}
	if !strings.Contains(model.View().Content, "Example Login") {
		t.Error("the confirmation does not name the item")
	}
}

func TestACancelledDeleteWritesNothing(t *testing.T) {
	model := onItemPane(t, true)
	model.client = fakeOpClient(t, "", 0)

	opened, _ := press(t, model, "d")
	if _, cmd := opened.overlaySubmit(opened); cmd != nil {
		t.Errorf("declining the confirmation still ran %v", cmd())
	}
}

func TestAConfirmedDeleteRemovesTheItem(t *testing.T) {
	model := onItemPane(t, true)
	model.client = fakeOpClient(t, "", 0)

	opened, _ := press(t, model, "d")

	confirmed, _ := opened.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	answered, ok := confirmed.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", confirmed)
	}

	_, cmd := answered.overlaySubmit(answered)
	if cmd == nil {
		t.Fatal("confirming the delete ran nothing")
	}
	msg, ok := cmd().(writeSucceededMsg)
	if !ok {
		t.Fatalf("delete returned %T, want writeSucceededMsg", cmd())
	}
	if got, want := msg.Status, theme.DeletedStatus; got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestAFailedDeleteReportsTheOpError(t *testing.T) {
	model := onItemPane(t, true)
	model.client = fakeOpClient(t, "", 1)

	opened, _ := press(t, model, "d")
	confirmed, _ := opened.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	answered, ok := confirmed.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", confirmed)
	}

	_, cmd := answered.overlaySubmit(answered)
	if _, ok := cmd().(opFailedMsg); !ok {
		t.Fatalf("delete returned %T, want opFailedMsg", cmd())
	}
}

// TestCrudKeysOpenNoItemFormOnTheVaultPane guards the context-sensitive verbs:
// a, e and d act on the focused pane, so on the vault pane they must reach the
// vault vertical. Any form they open there is a vault form, never an item one.
func TestCrudKeysOpenNoItemFormOnTheVaultPane(t *testing.T) {
	model := withVault(t, sized(t, 120, 40), personalVaultID)

	for _, k := range []string{"a", "e", "d"} {
		opened, _ := press(t, model, k)
		if opened.overlay == nil {
			continue
		}
		if strings.Contains(opened.overlay.View(), theme.ItemTitlePrompt) {
			t.Errorf("%q opened an item form while the vault pane had focus", k)
		}
	}
}
