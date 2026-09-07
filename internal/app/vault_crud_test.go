package app

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// blinkGrace is how long runVaultCmd waits for a command an open form issued
// before calling it a cursor timer and dropping it; the runtime's caller does
// not block on timers either. writeGrace is the wait for a submitted write,
// which shells out to the fake op.
const (
	blinkGrace = 50 * time.Millisecond
	writeGrace = 10 * time.Second
)

// vaultPane returns a sized model whose vault pane holds one selected vault.
func vaultPane(t *testing.T, client op.Client) Model {
	t.Helper()

	model := loadedWithVaults(t, op.Vault{ID: "vault-a", Name: "Personal", Items: 4})
	model.client = client
	return model
}

// runVaultCmd executes cmd the way the runtime would, flattening batches and
// dropping any command that has produced no message within commandGrace. Only
// timers behave that way, and no vault write is a timer.
func runVaultCmd(cmd tea.Cmd, grace time.Duration) []tea.Msg {
	if cmd == nil {
		return nil
	}

	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()

	var msg tea.Msg
	select {
	case msg = <-done:
	case <-time.After(grace):
		return nil
	}

	if batch, ok := msg.(tea.BatchMsg); ok {
		var msgs []tea.Msg
		for _, inner := range batch {
			msgs = append(msgs, runVaultCmd(inner, grace)...)
		}
		return msgs
	}
	if msg == nil {
		return nil
	}
	return []tea.Msg{msg}
}

// driveOverlay feeds keys to the open overlay and routes the commands huh
// returns back into it, then reports every message the submit produced.
func driveOverlay(t *testing.T, model Model, keys ...tea.KeyPressMsg) (Model, []tea.Msg) {
	t.Helper()

	var produced []tea.Msg
	for _, keyPress := range keys {
		if model.overlay == nil {
			break
		}
		next, cmd := model.updateOverlay(keyPress)
		model, produced = pumpOverlay(t, asModel(t, next), cmd, produced)
	}
	return model, produced
}

// pumpOverlay routes a command's messages back to the overlay while one is
// open, and collects them once it has closed.
func pumpOverlay(t *testing.T, model Model, cmd tea.Cmd, produced []tea.Msg) (Model, []tea.Msg) {
	t.Helper()

	grace := blinkGrace
	if model.overlay == nil {
		grace = writeGrace
	}

	for _, msg := range runVaultCmd(cmd, grace) {
		if model.overlay == nil {
			produced = append(produced, msg)
			continue
		}
		next, inner := model.updateOverlay(msg)
		model, produced = pumpOverlay(t, asModel(t, next), inner, produced)
	}
	return model, produced
}

func asModel(t *testing.T, model tea.Model) Model {
	t.Helper()

	typed, ok := model.(Model)
	if !ok {
		t.Fatalf("update returned %T, want Model", model)
	}
	return typed
}

// openVaultForm presses one of the CRUD keys on the vault pane and returns the
// model the key produced.
func openVaultForm(t *testing.T, model Model, keyPress rune) Model {
	t.Helper()

	next, _, handled := model.handleVaultKey(tea.KeyPressMsg{Code: keyPress, Text: string(keyPress)})
	if !handled {
		t.Fatalf("the vault pane did not consume %q", keyPress)
	}
	return asModel(t, next)
}

// typeVaultKeys spells out text as key presses and ends with enter.
func typeVaultKeys(text string) []tea.KeyPressMsg {
	presses := make([]tea.KeyPressMsg, 0, len(text)+1)
	for _, char := range text {
		presses = append(presses, tea.KeyPressMsg{Code: char, Text: string(char)})
	}
	return append(presses, tea.KeyPressMsg{Code: tea.KeyEnter})
}

// vaultWriteStatus reports the status carried by the write in msgs, and "" when
// no write happened.
func vaultWriteStatus(t *testing.T, msgs []tea.Msg) string {
	t.Helper()

	status := ""
	for _, msg := range msgs {
		switch typed := msg.(type) {
		case writeSucceededMsg:
			if typed.Reload != VaultPane {
				t.Errorf("write reloads %v, want VaultPane", typed.Reload)
			}
			status = typed.Status
		case opFailedMsg:
			t.Fatalf("write failed: %v", typed.Err)
		}
	}
	return status
}

func TestCreateKeyOpensTheVaultForm(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'a')

	if model.overlay == nil {
		t.Fatal("the create key opened no overlay")
	}
	if !strings.Contains(model.View().Content, theme.VaultNamePrompt) {
		t.Errorf("the overlay does not show %q", theme.VaultNamePrompt)
	}
}

func TestCreateKeyWorksOnAnEmptyVaultPane(t *testing.T) {
	model := loadedWithVaults(t)
	model.client = unreachableClient(t)

	if openVaultForm(t, model, 'a').overlay == nil {
		t.Error("the create key must work with no vault selected")
	}
}

func TestSubmittingTheCreateFormCreatesTheVault(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, fakeOpClient(t, `{"id":"v-new","name":"Archive"}`, 0)), 'a')

	after, msgs := driveOverlay(t, model, typeVaultKeys("Archive")...)

	if after.overlay != nil {
		t.Error("the create form is still open after it completed")
	}
	if got, want := vaultWriteStatus(t, msgs), theme.CreatedStatus; got != want {
		t.Errorf("status = %q, want %q", got, want)
	}
}

func TestTheCreateFormRejectsAnEmptyName(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'a')

	after, msgs := driveOverlay(t, model, tea.KeyPressMsg{Code: tea.KeyEnter})

	if after.overlay == nil {
		t.Error("the create form accepted an empty vault name")
	}
	if len(msgs) != 0 {
		t.Errorf("an empty name produced %v, want no write", msgs)
	}
}

func TestARejectedCreateReportsTheOpError(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, fakeOpClient(t, "", 1)), 'a')

	_, msgs := driveOverlay(t, model, typeVaultKeys("Personal")...)

	if len(msgs) != 1 {
		t.Fatalf("a rejected duplicate name produced %v, want one failure", msgs)
	}
	failure, ok := msgs[0].(opFailedMsg)
	if !ok {
		t.Fatalf("create produced %T, want opFailedMsg", msgs[0])
	}
	if !strings.Contains(failure.Err.Error(), "create vault:") {
		t.Errorf("err = %q, want it to name the operation", failure.Err)
	}
}

func TestEditKeySeedsTheSelectedVaultsName(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'e')

	if !strings.Contains(model.View().Content, "Personal") {
		t.Error("the edit form is not seeded with the selected vault's name")
	}
}

func TestEditKeyWithNoVaultSelectedOpensNothing(t *testing.T) {
	model := loadedWithVaults(t)
	model.client = unreachableClient(t)

	next, cmd, handled := model.handleVaultKey(tea.KeyPressMsg{Code: 'e', Text: "e"})

	if !handled {
		t.Error("the edit key must stay consumed by the vault pane")
	}
	if cmd != nil {
		t.Error("editing nothing issued a command")
	}
	if asModel(t, next).overlay != nil {
		t.Error("editing nothing opened a form")
	}
}

func TestSubmittingTheEditFormRenamesTheVault(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, fakeOpClient(t, "", 0)), 'e')

	after, msgs := driveOverlay(t, model, typeVaultKeys("2")...)

	if after.overlay != nil {
		t.Error("the edit form is still open after it completed")
	}
	if got, want := vaultWriteStatus(t, msgs), theme.SavedStatus; got != want {
		t.Errorf("status = %q, want %q", got, want)
	}
}

func TestSubmittingTheEditFormUnchangedWritesNothing(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'e')

	after, msgs := driveOverlay(t, model, tea.KeyPressMsg{Code: tea.KeyEnter})

	if after.overlay != nil {
		t.Error("the edit form did not close on submit")
	}
	if len(msgs) != 0 {
		t.Errorf("an unchanged name produced %v, want no write", msgs)
	}
}

func TestAbortingTheEditFormWritesNothing(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'e')

	typed, _ := driveOverlay(t, model, tea.KeyPressMsg{Code: 'X', Text: "X"})
	aborted, cmd := typed.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if asModel(t, aborted).overlay != nil {
		t.Error("esc did not dismiss the edit form")
	}
	if msgs := runVaultCmd(cmd, writeGrace); len(msgs) != 0 {
		t.Errorf("aborting the form produced %v, want no write", msgs)
	}
}

func TestDeleteKeyAsksForTheVaultNameBack(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'd')

	if !strings.Contains(model.View().Content, theme.ConfirmVaultPrompt) {
		t.Errorf("the delete form does not ask for %q", theme.ConfirmVaultPrompt)
	}
}

func TestDeleteKeyWithNoVaultSelectedOpensNothing(t *testing.T) {
	model := loadedWithVaults(t)
	model.client = unreachableClient(t)

	next, cmd, handled := model.handleVaultKey(tea.KeyPressMsg{Code: 'd', Text: "d"})

	if !handled {
		t.Error("the delete key must stay consumed by the vault pane")
	}
	if cmd != nil || asModel(t, next).overlay != nil {
		t.Error("deleting nothing opened a form or issued a command")
	}
}

func TestDeleteRefusesAMistypedVaultName(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'd')

	keys := append(typeVaultKeys("Persona"), tea.KeyPressMsg{Code: 'y', Text: "y"})
	after, msgs := driveOverlay(t, model, keys...)

	if after.overlay == nil {
		t.Fatal("the delete form completed on a mistyped vault name")
	}
	if len(msgs) != 0 {
		t.Errorf("a mistyped vault name produced %v, want no delete", msgs)
	}
}

func TestDeleteCancelledAtTheConfirmationDeletesNothing(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'd')

	keys := append(typeVaultKeys("Personal"), tea.KeyPressMsg{Code: 'n', Text: "n"})
	after, msgs := driveOverlay(t, model, keys...)

	if after.overlay != nil {
		t.Error("the delete form is still open after the confirmation")
	}
	if len(msgs) != 0 {
		t.Errorf("answering No produced %v, want no delete", msgs)
	}
}

func TestDeleteAcceptingTheConfirmationDefaultDeletesNothing(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, unreachableClient(t)), 'd')

	keys := append(typeVaultKeys("Personal"), tea.KeyPressMsg{Code: tea.KeyEnter})
	_, msgs := driveOverlay(t, model, keys...)

	if len(msgs) != 0 {
		t.Errorf("the confirmation defaults to Yes: %v", msgs)
	}
}

func TestDeleteConfirmedDeletesTheVault(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, fakeOpClient(t, "", 0)), 'd')

	keys := append(typeVaultKeys("Personal"), tea.KeyPressMsg{Code: 'y', Text: "y"})
	after, msgs := driveOverlay(t, model, keys...)

	if after.overlay != nil {
		t.Error("the delete form is still open after it completed")
	}
	if got, want := vaultWriteStatus(t, msgs), theme.DeletedStatus; got != want {
		t.Errorf("status = %q, want %q", got, want)
	}
}

func TestAFailingDeleteReportsTheOpError(t *testing.T) {
	model := openVaultForm(t, vaultPane(t, fakeOpClient(t, "", 1)), 'd')

	keys := append(typeVaultKeys("Personal"), tea.KeyPressMsg{Code: 'y', Text: "y"})
	_, msgs := driveOverlay(t, model, keys...)

	if len(msgs) != 1 {
		t.Fatalf("a failed delete produced %v, want one failure", msgs)
	}
	if _, ok := msgs[0].(opFailedMsg); !ok {
		t.Fatalf("delete produced %T, want opFailedMsg", msgs[0])
	}
}

func TestUnboundKeysAreNotConsumedByTheVaultPane(t *testing.T) {
	model := vaultPane(t, unreachableClient(t))

	for _, keyPress := range []tea.KeyPressMsg{{Code: 'r', Text: "r"}, {Code: 'y', Text: "y"}} {
		if _, _, handled := model.handleVaultKey(keyPress); handled {
			t.Errorf("the vault pane consumed %v", keyPress)
		}
	}
}
