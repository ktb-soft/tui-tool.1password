package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/keymap"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

func sized(t *testing.T, width, height int) Model {
	t.Helper()

	model, _ := New(op.Client{}).Update(tea.WindowSizeMsg{Width: width, Height: height})
	sizedModel, ok := model.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", model)
	}
	return sizedModel
}

func press(t *testing.T, model Model, keys string) (Model, tea.Cmd) {
	t.Helper()

	next, cmd := model.Update(tea.KeyPressMsg{Code: rune(keys[0]), Text: keys})
	pressed, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	return pressed, cmd
}

func TestViewRendersThreeBorderedPanes(t *testing.T) {
	content := sized(t, 120, 40).View().Content

	for _, title := range []string{theme.VaultPaneTitle, theme.ItemPaneTitle} {
		if !strings.Contains(content, title) {
			t.Errorf("view is missing the %q pane title", title)
		}
	}
	if !strings.Contains(content, "╭") || !strings.Contains(content, "╯") {
		t.Error("view has no rounded borders")
	}
	if !strings.Contains(content, theme.NoSelection) {
		t.Errorf("detail pane is missing %q", theme.NoSelection)
	}
}

func TestViewReportsATerminalTooSmall(t *testing.T) {
	content := sized(t, 20, 5).View().Content

	if !strings.Contains(content, theme.TerminalTooSmall) {
		t.Errorf("view = %q, want it to report %q", content, theme.TerminalTooSmall)
	}
}

func TestQuitKeyReturnsTheQuitCommand(t *testing.T) {
	_, cmd := press(t, sized(t, 120, 40), "q")

	if cmd == nil {
		t.Fatal("q returned no command, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("q returned %T, want tea.QuitMsg", cmd())
	}
}

func TestFocusStartsOnVaultsAndDoesNotWrapLeft(t *testing.T) {
	model := sized(t, 120, 40)

	if model.focus != VaultPane {
		t.Fatalf("focus = %d, want VaultPane", model.focus)
	}
	if !model.vaults.Focused() {
		t.Error("vault pane is not focused at startup")
	}

	moved, _ := press(t, model, "h")
	if moved.focus != VaultPane {
		t.Errorf("focus = %d after h in the leftmost pane, want VaultPane", moved.focus)
	}
}

func TestFocusDoesNotAdvanceOntoAnEmptyPane(t *testing.T) {
	moved, _ := press(t, sized(t, 120, 40), "l")

	if moved.focus != VaultPane {
		t.Errorf("focus = %d, want VaultPane while the item pane is empty", moved.focus)
	}
}

func TestHelpKeyTogglesTheFullHelp(t *testing.T) {
	toggled, _ := press(t, sized(t, 120, 40), "?")

	if !toggled.help.ShowAll {
		t.Error("? did not turn on the full help")
	}
}

func TestOpFailureShowsTheErrorTextVerbatim(t *testing.T) {
	model := sized(t, 120, 40)

	_, cmd := model.Update(opFailedMsg{Err: &op.OpError{
		Args:   []string{"vault", "list"},
		Stderr: "you are not currently signed in",
	}})
	if cmd == nil {
		t.Fatal("opFailedMsg produced no status command")
	}
}

func TestModelUsesTheSharedKeymap(t *testing.T) {
	if got, want := New(op.Client{}).keys.Quit.Keys(), keymap.Default().Quit.Keys(); len(got) != len(want) {
		t.Errorf("Quit keys = %v, want %v", got, want)
	}
}
