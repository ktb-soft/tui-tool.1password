package app

import (
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
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

// detailModel seeds an item pane holding one item whose fields have loaded,
// which is the state the detail pane is entered from.
func detailModel(t *testing.T, item op.Item) Model {
	t.Helper()

	model := New(op.Client{})
	model.resizePanes(120, 40)
	model.items.SetItems([]list.Item{item})
	model.focus = ItemPane
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
}

func TestRevealKeyUnmasksAndItemChangeResets(t *testing.T) {
	model := detailModel(t, loginItem())

	model, _ = press(t, model, "r")
	if !strings.Contains(model.render(), secretValue) {
		t.Fatal("reveal key did not unmask the concealed field")
	}

	changed, _ := model.Update(itemLoadedMsg(noteItem()))
	model, ok := changed.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want app.Model", changed)
	}
	if strings.Contains(model.render(), secretValue) {
		t.Fatal("concealed value survived a change of item")
	}
}

func TestRevealKeyIsIgnoredWhenTheVaultPaneHasFocus(t *testing.T) {
	model := detailModel(t, loginItem())
	model.focus = VaultPane
	model.applyFocus()

	model, _ = press(t, model, "r")
	if strings.Contains(model.render(), secretValue) {
		t.Fatal("reveal fired while the vault pane held focus")
	}
}

// TestEditEntersTheDetailPaneRatherThanAnOverlay is the redesign's headline:
// e moves focus into the pane's form instead of opening a form over the panes.
func TestEditEntersTheDetailPaneRatherThanAnOverlay(t *testing.T) {
	model, _ := press(t, detailModel(t, loginItem()), "e")

	if model.overlay != nil {
		t.Fatal("e opened an overlay")
	}
	if model.focus != DetailPane {
		t.Fatalf("focus = %d, want DetailPane", model.focus)
	}
	if !model.detail.Focused() {
		t.Fatal("the detail pane was not focused")
	}
	if strings.Contains(model.render(), secretValue) {
		t.Fatal("entering the form rendered the concealed value")
	}
}

func TestEditWaitsForTheItemsFieldsToLoad(t *testing.T) {
	model := detailModel(t, loginItem())
	model.detail.Clear()

	entered, cmd := press(t, model, "e")

	if entered.focus != ItemPane {
		t.Fatal("e entered the detail pane before the item's fields were loaded")
	}
	if cmd == nil {
		t.Fatal("e reported nothing while the item was still loading")
	}
}

// TestRightArrowNoLongerEntersTheDetailPane guards the binding the redesign
// removed: e is the only way in.
func TestRightArrowNoLongerEntersTheDetailPane(t *testing.T) {
	model := detailModel(t, loginItem())

	for _, k := range []string{"right", "l"} {
		moved, _ := press(t, model, k)
		if moved.focus != ItemPane {
			t.Errorf("%q moved focus off the item pane to %d", k, moved.focus)
		}
	}
}

// TestEscapeLeavesTheFormWithoutSaving is the discard path, driven the way the
// runtime drives it: focus the form, type, press esc, and check that neither
// the item nor a write survived.
func TestEscapeLeavesTheFormWithoutSaving(t *testing.T) {
	model := detailModel(t, loginItem())
	model.client = fakeOpClient(t, "", 0)

	entered, cmd := press(t, model, "e")
	entered = pumpUpdate(t, entered, cmd, settleDepth)

	typed := typeThroughUpdate(t, entered, "ZZZ")
	if !strings.Contains(typed.render(), "ZZZ") {
		t.Fatal("typing never reached the form, so the discard is untested")
	}

	escaped, cmd := press(t, typed, "esc")
	if cmd != nil {
		t.Fatalf("esc ran %v instead of discarding", cmd())
	}
	if escaped.focus != ItemPane {
		t.Fatal("esc did not return focus to the item pane")
	}
	if strings.Contains(escaped.render(), "ZZZ") {
		t.Fatal("the edit survived esc")
	}
	if got := escaped.detail.EditedItem().Name; got != loginItem().Name {
		t.Fatalf("the form kept the edited title %q after esc", got)
	}
}

// TestCompletingTheFormSavesThroughEditItem drives the form to completion
// through Update, the way huh's own messages arrive at runtime, and asserts the
// write actually reached op rather than that a command was returned.
func TestCompletingTheFormSavesThroughEditItem(t *testing.T) {
	model := detailModel(t, loginItem())
	model.client = fakeOpClient(t, `{"id":"1111","title":"Example Login","category":"LOGIN"}`, 0)

	entered, cmd := press(t, model, "e")
	var produced []tea.Msg
	model = collect(t, entered, cmd, settleDepth, &produced)

	for range len(loginItem().Fields) + 2 {
		if model.focus == ItemPane {
			break
		}
		submitted, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		model = collect(t, asModel(t, submitted), cmd, settleDepth, &produced)
	}

	if model.focus != ItemPane {
		t.Fatal("completing the form did not return focus to the item pane")
	}
	if !holdsWrite(produced) {
		t.Fatalf("completing the form issued no write; saw %d messages", len(produced))
	}
}

// formGrace is long enough for a write to spawn the fake op, which blinkGrace
// is not; a cursor timer resolving inside it is just another ignored message.
const formGrace = 400 * time.Millisecond

// collect drives a command's messages back through Update, recording every one
// so a write buried in a follow-up command is still visible to the test.
func collect(t *testing.T, model Model, cmd tea.Cmd, depth int, produced *[]tea.Msg) Model {
	t.Helper()

	if cmd == nil || depth == 0 {
		return model
	}
	for _, msg := range runVaultCmd(cmd, formGrace) {
		*produced = append(*produced, msg)
		next, follow := model.Update(msg)
		model = collect(t, asModel(t, next), follow, depth-1, produced)
	}
	return model
}

func holdsWrite(msgs []tea.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case writeSucceededMsg, opFailedMsg:
			return true
		}
	}
	return false
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
