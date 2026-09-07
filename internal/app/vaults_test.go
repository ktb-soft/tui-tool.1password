package app

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

func unreachableClient(t *testing.T) op.Client {
	t.Helper()
	return op.Client{Path: filepath.Join(t.TempDir(), "absent")}
}

func loadedWithVaults(t *testing.T, vaults ...op.Vault) Model {
	t.Helper()

	next, cmd := sized(t, 120, 40).Update(vaultsLoadedMsg(vaults))
	loaded, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	if cmd != nil {
		cmd()
	}
	return loaded
}

func TestVaultsLoadedFillsTheVaultPane(t *testing.T) {
	model := loadedWithVaults(t,
		op.Vault{ID: "a", Name: "Personal", Items: 4},
		op.Vault{ID: "b", Name: "Shared", Items: 2})

	if got := len(model.vaults.List().Items()); got != 2 {
		t.Fatalf("vault pane holds %d rows, want 2", got)
	}

	selected, ok := model.vaults.SelectedItem().(op.Vault)
	if !ok {
		t.Fatalf("selected item is %T, want op.Vault", model.vaults.SelectedItem())
	}
	if got, want := selected.Name, "Personal"; got != want {
		t.Errorf("selected vault = %q, want %q", got, want)
	}

	if content := model.View().Content; !strings.Contains(content, "Personal") {
		t.Errorf("view does not show the loaded vault: %q", content)
	}
}

func TestVaultsLoadedWithNoVaultsLeavesThePaneEmpty(t *testing.T) {
	model := loadedWithVaults(t)

	if got := len(model.vaults.List().Items()); got != 0 {
		t.Errorf("vault pane holds %d rows, want 0", got)
	}
	if model.vaults.SelectedItem() != nil {
		t.Errorf("SelectedItem() = %v, want nil", model.vaults.SelectedItem())
	}
}

func TestVaultsLoadedReplacesThePreviousContents(t *testing.T) {
	model := loadedWithVaults(t, op.Vault{ID: "a", Name: "Personal"})

	next, cmd := model.Update(vaultsLoadedMsg([]op.Vault{{ID: "b", Name: "Shared"}}))
	if cmd != nil {
		cmd()
	}
	reloaded, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}

	if got := len(reloaded.vaults.List().Items()); got != 1 {
		t.Fatalf("vault pane holds %d rows, want 1", got)
	}
	if content := reloaded.View().Content; strings.Contains(content, "Personal") {
		t.Error("view still shows the vault the reload replaced")
	}
}

func TestLoadVaultsReportsFailureAsAnOpFailure(t *testing.T) {
	msg := loadVaults(unreachableClient(t))()

	failure, ok := msg.(opFailedMsg)
	if !ok {
		t.Fatalf("loadVaults returned %T, want opFailedMsg", msg)
	}
	if failure.Err == nil {
		t.Fatal("opFailedMsg carries no error")
	}
}

func TestVaultLoadFailureShowsAStatusAndKeepsThePane(t *testing.T) {
	model := loadedWithVaults(t, op.Vault{ID: "a", Name: "Personal"})

	next, cmd := model.Update(opFailedMsg{Err: &op.OpError{
		Args:   []string{"vault", "list"},
		Stderr: "you are not currently signed in",
	}})
	if cmd == nil {
		t.Fatal("a vault load failure produced no status command")
	}
	failed, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	if got := len(failed.vaults.List().Items()); got != 1 {
		t.Errorf("vault pane holds %d rows, want the failure to leave it alone", got)
	}
}

func TestAVaultWriteReloadsTheVaultPane(t *testing.T) {
	model := loadedWithVaults(t)
	model.client = unreachableClient(t)

	_, cmd := model.updateVaults(writeSucceededMsg{Reload: VaultPane, Status: theme.CreatedStatus})
	if cmd == nil {
		t.Fatal("a completed vault write issued no reload")
	}
	if _, ok := cmd().(opFailedMsg); !ok {
		t.Errorf("reload produced %T, want it to run ListVaults", cmd())
	}
}

func TestVaultPaneIgnoresMessagesThatAreNotItsOwn(t *testing.T) {
	model := loadedWithVaults(t, op.Vault{ID: "a", Name: "Personal"})

	next, cmd := model.updateVaults(tea.WindowSizeMsg{Width: 120, Height: 40})
	if cmd != nil {
		t.Errorf("updateVaults returned a command for an unrelated message")
	}
	unchanged, ok := next.(Model)
	if !ok {
		t.Fatalf("updateVaults returned %T, want Model", next)
	}
	if got := len(unchanged.vaults.List().Items()); got != 1 {
		t.Errorf("vault pane holds %d rows, want 1", got)
	}
}
