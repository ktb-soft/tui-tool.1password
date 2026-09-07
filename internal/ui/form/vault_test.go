package form

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// started runs the form's Init the way the bubbletea runtime does, which is
// what focuses its first field.
func started(form *huh.Form) *huh.Form {
	drain(form, form.Init())
	return form
}

// drain feeds a command's message back to the form, flattening batches.
func drain(form *huh.Form, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case nil:
	case tea.BatchMsg:
		for _, inner := range msg {
			drain(form, inner)
		}
	default:
		form.Update(msg)
	}
}

// typeInto pushes each rune of text at the form as a key press, which is all
// an input needs to update the value it is bound to.
func typeInto(form *huh.Form, text string) {
	for _, char := range text {
		form.Update(tea.KeyPressMsg{Code: char, Text: string(char)})
	}
}

func TestNewVaultBindsWhatIsTypedToTheName(t *testing.T) {
	name := new(string)
	vaultForm := started(NewVault(name))

	typeInto(vaultForm, "Archive")

	if got, want := *name, "Archive"; got != want {
		t.Errorf("name = %q, want %q", got, want)
	}
	if !strings.Contains(vaultForm.View(), theme.VaultNamePrompt) {
		t.Errorf("form does not show the prompt %q", theme.VaultNamePrompt)
	}
}

func TestNewVaultStartsFromTheSeededName(t *testing.T) {
	name := new(string)
	*name = "Personal"

	if !strings.Contains(started(NewVault(name)).View(), "Personal") {
		t.Error("edit form does not show the current vault name")
	}
}

func TestValidateVaultNameRejectsBlankNames(t *testing.T) {
	cases := []struct {
		name    string
		wantErr bool
	}{
		{"Archive", false},
		{"a", false},
		{" Archive ", false},
		{"", true},
		{"   ", true},
		{"\t\n", true},
	}

	for _, tc := range cases {
		err := ValidateVaultName(tc.name)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateVaultName(%q) = %v, want error: %v", tc.name, err, tc.wantErr)
		}
		if tc.wantErr && err.Error() != theme.RequiredValidation {
			t.Errorf("ValidateVaultName(%q) = %q, want %q", tc.name, err, theme.RequiredValidation)
		}
	}
}

func TestNewVaultDoesNotCompleteOnAnEmptyName(t *testing.T) {
	vaultForm := started(NewVault(new(string)))

	vaultForm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if vaultForm.State == huh.StateCompleted {
		t.Error("the create form accepted an empty vault name")
	}
}

func TestNewVaultDeleteShowsBothGates(t *testing.T) {
	view := started(NewVaultDelete("Shared", new(string), new(bool))).View()

	if !strings.Contains(view, theme.ConfirmVaultPrompt) {
		t.Errorf("delete form does not ask for the name back: %q", view)
	}
	if prompt := fmt.Sprintf(theme.DeleteVaultPrompt, "Shared"); !strings.Contains(view, prompt) {
		t.Errorf("delete form is missing the confirmation %q", prompt)
	}
}

func TestNewVaultDeleteStartsUnconfirmed(t *testing.T) {
	confirmed := new(bool)
	*confirmed = true

	NewVaultDelete("Shared", new(string), confirmed)

	if *confirmed {
		t.Error("the delete form did not default the confirmation to No")
	}
}

func TestNewVaultDeleteRefusesToAdvanceOnAMistypedName(t *testing.T) {
	typed, confirmed := new(string), new(bool)
	deleteForm := started(NewVaultDelete("Shared", typed, confirmed))

	typeInto(deleteForm, "Share")
	deleteForm.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if len(deleteForm.Errors()) == 0 {
		t.Error("a mistyped vault name raised no validation error")
	}
	if *confirmed {
		t.Error("a mistyped vault name reached the confirmation")
	}
}

func TestIsVaultDeleteConfirmedNeedsBothGates(t *testing.T) {
	cases := []struct {
		typed     string
		confirmed bool
		want      bool
	}{
		{"Shared", true, true},
		{"Shared", false, false},
		{"shared", true, false},
		{"Share", true, false},
		{"Shared ", true, false},
		{"", true, false},
		{"", false, false},
	}

	for _, tc := range cases {
		got := IsVaultDeleteConfirmed("Shared", tc.typed, tc.confirmed)
		if got != tc.want {
			t.Errorf("IsVaultDeleteConfirmed(%q, %q, %v) = %v, want %v",
				"Shared", tc.typed, tc.confirmed, got, tc.want)
		}
	}
}
