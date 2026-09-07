package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// pumpUpdate routes a command's messages back through Model.Update — the real
// entry point — rather than reaching into updateOverlay. A form that only
// settles when driven through updateOverlay is not one the running program can
// ever submit.
func pumpUpdate(t *testing.T, model Model, cmd tea.Cmd, depth int) Model {
	t.Helper()

	if cmd == nil || depth == 0 {
		return model
	}
	for _, msg := range runVaultCmd(cmd, blinkGrace) {
		next, follow := model.Update(msg)
		model = pumpUpdate(t, asModel(t, next), follow, depth-1)
	}
	return model
}

func typeThroughUpdate(t *testing.T, model Model, text string) Model {
	t.Helper()

	for _, char := range text {
		next, cmd := model.Update(tea.KeyPressMsg{Code: char, Text: string(char)})
		model = pumpUpdate(t, asModel(t, next), cmd, settleDepth)
	}
	return model
}

const settleDepth = 8

func TestFormReachesCompletionThroughUpdate(t *testing.T) {
	model := loadedWithVaults(t, op.Vault{ID: personalVaultID, Name: "Personal", Items: 2})
	model.client = fakeOpClient(t, `{"id":"new","name":"Archive"}`, 0)

	opened, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = pumpUpdate(t, asModel(t, opened), cmd, settleDepth)
	if model.overlay == nil {
		t.Fatal("the create key opened no form")
	}

	model = typeThroughUpdate(t, model, "Archive")

	submitted, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model = pumpUpdate(t, asModel(t, submitted), cmd, settleDepth)

	if model.overlay != nil {
		t.Fatal("form never completed when driven through Update: huh's own messages " +
			"are not reaching the overlay")
	}
}
