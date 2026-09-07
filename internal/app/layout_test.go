package app

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestPaneWidthsSumToTerminalWidth(t *testing.T) {
	for _, width := range []int{80, 100, 120, 133, 201} {
		model := sized(t, width, 40)
		topRow := strings.SplitN(model.View().Content, "\n", 2)[0]

		if got := lipgloss.Width(topRow); got != width {
			t.Errorf("terminal width %d: top row is %d cells", width, got)
		}
	}
}

func TestPaneHeightLeavesRoomForTheFooter(t *testing.T) {
	const height = 40

	rows := strings.Split(sized(t, 120, height).View().Content, "\n")
	if len(rows) != height {
		t.Errorf("view is %d rows, want %d", len(rows), height)
	}
}
