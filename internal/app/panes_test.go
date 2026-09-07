package app

import (
	"strings"
	"testing"
)

func TestAllThreePanesRenderTheSameHeight(t *testing.T) {
	model := sized(t, 100, 24)

	rows := func(view string) int { return strings.Count(view, "\n") + 1 }
	vaults, items, detail := rows(model.vaults.View()), rows(model.items.View()), rows(model.detail.View())

	if vaults != items || items != detail {
		t.Errorf("pane heights = %d/%d/%d, want all equal", vaults, items, detail)
	}
}
