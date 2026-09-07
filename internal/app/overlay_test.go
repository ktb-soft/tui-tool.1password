package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

func withOverlay(t *testing.T) Model {
	t.Helper()

	var name string
	model := sized(t, 120, 40)
	model.openOverlay(huh.NewForm(huh.NewGroup(
		huh.NewInput().Title(theme.VaultNamePrompt).Value(&name),
	)), nil)
	return model
}

func TestOverlayRendersTheFormCenteredAtFullSize(t *testing.T) {
	content := withOverlay(t).View().Content

	if !strings.Contains(content, theme.VaultNamePrompt) {
		t.Errorf("view is missing the form prompt %q", theme.VaultNamePrompt)
	}
}

func TestOverlayKeepsTheLayoutSize(t *testing.T) {
	rows := strings.Split(withOverlay(t).View().Content, "\n")

	if len(rows) != 40 {
		t.Errorf("view is %d rows with an overlay open, want 40", len(rows))
	}
}

func TestAbortingTheOverlayClearsIt(t *testing.T) {
	model := withOverlay(t)

	next, _ := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	cleared, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	if cleared.overlay != nil {
		t.Error("esc did not clear the overlay")
	}
}

func TestKeysGoToTheOverlayNotThePanes(t *testing.T) {
	model := withOverlay(t)

	next, _ := model.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	after, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	if after.overlay == nil {
		t.Error("q reached the global quit binding while a form was open")
	}
}
