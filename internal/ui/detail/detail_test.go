package detail_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/detail"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

const (
	secretValue = "hunter2"
	paneWidth   = 100
	paneHeight  = 30
)

func loadItem(t *testing.T, name string) op.Item {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var item op.Item
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return item
}

func sizedPane(t *testing.T, item op.Item) detail.Detail {
	t.Helper()

	pane := detail.New()
	pane.SetSize(paneWidth, paneHeight)
	pane.SetItem(item)
	return pane
}

func TestViewMasksConcealedFieldUntilRevealed(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-login.json"))

	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("concealed value rendered before reveal")
	}
	if !strings.Contains(pane.View(), theme.Mask) {
		t.Fatal("mask not rendered for concealed field")
	}

	pane.ToggleReveal()
	if !strings.Contains(pane.View(), secretValue) {
		t.Fatal("concealed value not rendered after reveal")
	}

	pane.ToggleReveal()
	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("concealed value still rendered after masking again")
	}
}

func TestRevealResetsWhenItemChanges(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-login.json"))
	pane.ToggleReveal()

	pane.SetItem(loadItem(t, "item-sections.json"))

	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("reveal survived an item change")
	}
	if pane.Revealed("password") {
		t.Fatal("reveal state survived an item change")
	}
}

func TestRevealResetsOnClear(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-login.json"))
	pane.ToggleReveal()

	pane.Clear()

	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("reveal survived a clear")
	}
	if !strings.Contains(pane.View(), theme.NoSelection) {
		t.Fatal("cleared pane does not show the empty state")
	}
}

func TestViewShowsEmptyStateWithNoItem(t *testing.T) {
	pane := detail.New()
	pane.SetSize(paneWidth, paneHeight)

	if !strings.Contains(pane.View(), theme.NoSelection) {
		t.Fatalf("empty pane does not show %q", theme.NoSelection)
	}
}

func TestViewShowsNoFieldsForFieldlessItem(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-empty.json"))

	if !strings.Contains(pane.View(), theme.NoFields) {
		t.Fatalf("fieldless item does not show %q", theme.NoFields)
	}
}

func TestViewGroupsSectionlessFieldsFirst(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-sections.json"))
	view := pane.View()

	sectionless := strings.Index(view, "base station name")
	admin := strings.Index(view, "Admin")
	guest := strings.Index(view, "Guest network")

	switch {
	case sectionless < 0 || admin < 0 || guest < 0:
		t.Fatalf("missing field or section heading in view:\n%s", view)
	case sectionless > admin:
		t.Fatal("sectionless field rendered after a section")
	case admin > guest:
		t.Fatal("sections rendered out of first-appearance order")
	}
}

func TestViewRendersFieldWithEmptyValue(t *testing.T) {
	item := op.Item{
		ID:     "id",
		Fields: []op.Field{{ID: "blank", Label: "blank label", Type: op.FieldTypeString}},
	}
	pane := sizedPane(t, item)

	if !strings.Contains(pane.View(), "blank label") {
		t.Fatal("field with an empty value is not rendered")
	}
}

func TestViewKeepsLongValueInsidePaneWidth(t *testing.T) {
	item := op.Item{
		ID:     "id",
		Fields: []op.Field{{ID: "long", Label: "long", Type: op.FieldTypeString, Value: strings.Repeat("x", 500)}},
	}
	pane := sizedPane(t, item)

	if width := lipgloss.Width(pane.View()); width > paneWidth {
		t.Fatalf("long value widened the pane to %d, want at most %d", width, paneWidth)
	}
}

func TestViewSurvivesTerminalTooNarrowForPane(t *testing.T) {
	pane := detail.New()
	pane.SetSize(1, 1)
	pane.SetItem(loadItem(t, "item-sections.json"))

	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("concealed value rendered in a narrow pane")
	}
}
