package app

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// debounceSelection emits a selectionChangedMsg once the cursor has rested for
// SelectionDebounce, so holding a movement key issues one load, not many.
func debounceSelection(target Focus, id string) tea.Cmd {
	return tea.Tick(SelectionDebounce, func(_ time.Time) tea.Msg {
		return selectionChangedMsg{Pane: target, ID: id}
	})
}
