// Package form builds the huh forms shown over the three panes. It constructs
// and validates them; it never submits one and never calls op.
package form

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// NewVault builds the one-input create/edit form, bound to name. Create passes
// a pointer to an empty string; edit passes one seeded with the current name.
func NewVault(name *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(theme.VaultNamePrompt).
			Value(name).
			Validate(ValidateVaultName),
	))
}

// NewVaultDelete builds the two-gate deletion form for the vault called
// vaultName: the name has to be typed back into typed before the confirmation
// bound to confirmed is reachable, because deleting a vault takes every item
// in it with it. The caller must check both gates before deleting.
func NewVaultDelete(vaultName string, typed *string, confirmed *bool) *huh.Form {
	*confirmed = false

	return huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(theme.ConfirmVaultPrompt).
			Value(typed).
			Validate(matchVaultName(vaultName)),
		huh.NewConfirm().
			Title(fmt.Sprintf(theme.DeleteVaultPrompt, vaultName)).
			Value(confirmed),
	))
}

// ValidateVaultName rejects a name that is empty or only whitespace.
func ValidateVaultName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New(theme.RequiredValidation)
	}
	return nil
}

// IsVaultDeleteConfirmed reports whether the deletion form's two gates both
// passed. The controller calls it before running the delete, so a mistyped
// name cannot destroy a vault even if the form completes.
func IsVaultDeleteConfirmed(vaultName, typed string, confirmed bool) bool {
	return confirmed && typed == vaultName
}

// matchVaultName builds the validator that demands the exact vault name back.
func matchVaultName(vaultName string) func(string) error {
	return func(typed string) error {
		if typed != vaultName {
			return errors.New(theme.NameMismatchMessage)
		}
		return nil
	}
}
