package app

import "github.com/ktb-soft/tui-tool.1password/internal/op"

// vaultsLoadedMsg carries the result of ListVaults.
type vaultsLoadedMsg []op.Vault

// itemsLoadedMsg carries the result of ListItems for the selected vault.
type itemsLoadedMsg []op.Item

// itemLoadedMsg carries the result of GetItem for the selected item.
type itemLoadedMsg op.Item

// writeSucceededMsg reports a completed create, edit, or delete, and names the
// pane whose contents must be reloaded.
type writeSucceededMsg struct {
	Reload  Focus
	Status  string
	Removed bool
}

// opFailedMsg carries any error from internal/op. Its text is shown verbatim.
type opFailedMsg struct{ Err error }

// selectionChangedMsg fires after the debounce interval when the cursor has
// settled on a new row in the given pane.
type selectionChangedMsg struct {
	Pane Focus
	ID   string
}
