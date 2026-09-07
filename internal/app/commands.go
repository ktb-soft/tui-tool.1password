package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// debounceSelection emits a selectionChangedMsg once the cursor has rested for
// SelectionDebounce, so holding a movement key issues one load, not many.
//
//nolint:unused // frozen contract seam, wired up in wave 1 (ADR 07)
func debounceSelection(target Focus, id string) tea.Cmd {
	return tea.Tick(SelectionDebounce, func(_ time.Time) tea.Msg {
		return selectionChangedMsg{Pane: target, ID: id}
	})
}
