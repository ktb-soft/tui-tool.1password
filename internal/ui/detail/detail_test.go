package detail_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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

	pane.ToggleReveal()
	if !strings.Contains(pane.View(), secretValue) {
		t.Fatal("concealed value not rendered after reveal")
	}

	pane.ToggleReveal()
	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("concealed value still rendered after masking again")
	}
}

func TestFocusingThePaneKeepsAConcealedFieldMasked(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-login.json"))
	pane.Focus()

	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("focusing the pane rendered the concealed value")
	}
}

func TestRevealResetsWhenItemChanges(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-login.json"))
	pane.ToggleReveal()

	pane.SetItem(loadItem(t, "item-sections.json"))

	if strings.Contains(pane.View(), secretValue) {
		t.Fatal("reveal survived an item change")
	}
	if pane.Revealed() {
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

// TestEmptyStateIsIndentedLikeTheFields keeps the empty state from sitting
// flush against the border while the fields that replace it are inset.
func TestEmptyStateIsIndentedLikeTheFields(t *testing.T) {
	pane := detail.New()
	pane.SetSize(paneWidth, paneHeight)

	line := lineContaining(t, pane.View(), theme.NoSelection)
	if got := indentOf(line, theme.NoSelection); got != theme.DetailIndent {
		t.Fatalf("empty state indented %d cells past the border, want %d", got, theme.DetailIndent)
	}
}

const borderLeft = "│"

var ansiCodes = regexp.MustCompile("\x1b\\[[0-9;]*m")

// indentOf counts the cells between the pane's left border and the given text.
func indentOf(line, text string) int {
	plain := ansiCodes.ReplaceAllString(line, "")
	return strings.Index(plain, text) - strings.Index(plain, borderLeft) - len(borderLeft)
}

func lineContaining(t *testing.T, view, text string) string {
	t.Helper()

	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, text) {
			return line
		}
	}
	t.Fatalf("no line of the view contains %q:\n%s", text, view)
	return ""
}

func TestViewShowsNoFieldsForFieldlessItem(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-empty.json"))

	if !strings.Contains(pane.View(), theme.NoFields) {
		t.Fatalf("fieldless item does not show %q", theme.NoFields)
	}
}

func TestViewNamesFieldsBySection(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-sections.json"))
	view := pane.View()

	if !strings.Contains(view, "base station name") {
		t.Fatalf("sectionless field is not rendered:\n%s", view)
	}
	if !strings.Contains(view, "Admin"+theme.SectionSeparator) {
		t.Fatalf("a sectioned field is not qualified by its section:\n%s", view)
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

// TestBlurDiscardsEdits is the esc path: whatever was typed into the form is
// gone the moment the pane loses focus.
func TestBlurDiscardsEdits(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-login.json"))
	original := pane.Item().Name

	pane.Focus()
	typed := typeInto(t, pane, "ZZZ")
	if typed.EditedItem().Name == original {
		t.Fatal("typing did not reach the form, so the discard is untested")
	}

	typed.Blur()
	if got := typed.EditedItem().Name; got != original {
		t.Fatalf("edited title survived a blur as %q, want %q", got, original)
	}
	if !strings.Contains(typed.View(), original) {
		t.Fatal("the form still shows the discarded edit")
	}
}

func TestABlurredPaneIgnoresKeys(t *testing.T) {
	pane := sizedPane(t, loadItem(t, "item-login.json"))
	original := pane.Item().Name

	typed := typeInto(t, pane, "ZZZ")
	if got := typed.EditedItem().Name; got != original {
		t.Fatalf("a blurred pane took keys: title became %q", got)
	}
}

func typeInto(t *testing.T, pane detail.Detail, text string) detail.Detail {
	t.Helper()

	for _, char := range text {
		pane, _ = pane.Update(keyPress(char))
	}
	return pane
}

func keyPress(char rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: char, Text: string(char)}
}
