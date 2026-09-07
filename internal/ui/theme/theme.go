// Package theme holds every color, border, dimension, ratio, and user-facing
// string in the program. No other package calls lipgloss.Color or writes a
// literal width.
package theme

import "charm.land/lipgloss/v2"

// Colors.
var (
	Accent    = lipgloss.Color("#4C9AFF")
	Muted     = lipgloss.Color("#6C7086")
	Text      = lipgloss.Color("#CDD6F4")
	Subtle    = lipgloss.Color("#9399B2")
	Danger    = lipgloss.Color("#F38BA8")
	Success   = lipgloss.Color("#A6E3A1")
	Highlight = lipgloss.Color("#F9E2AF")
)

// Layout ratios and fixed dimensions.
const (
	VaultPaneRatio  = 0.20
	ItemPaneRatio   = 0.35
	DetailPaneRatio = 0.45

	FooterHeight = 1
	BorderWidth  = 2

	MinWidth  = 60
	MinHeight = 12

	OverlayWidthRatio  = 0.50
	OverlayHeightRatio = 0.50
)

// Borders. Both use the same character set so nothing shifts when focus moves.
var (
	FocusedBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Accent)

	BlurredBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted)
)

// Border returns the pane border for the given focus state.
func Border(focused bool) lipgloss.Style {
	if focused {
		return FocusedBorder
	}
	return BlurredBorder
}

// Text styles.
var (
	PaneTitle = lipgloss.NewStyle().Foreground(Accent).Bold(true)
	FieldName = lipgloss.NewStyle().Foreground(Subtle)
	FieldVal  = lipgloss.NewStyle().Foreground(Text)
	Section   = lipgloss.NewStyle().Foreground(Accent).Bold(true)
	Empty     = lipgloss.NewStyle().Foreground(Muted).Italic(true)
	ErrorText = lipgloss.NewStyle().Foreground(Danger)
	OKText    = lipgloss.NewStyle().Foreground(Success)
)

// Footer caps the help row at the terminal width, since help.Model renders its
// full short help regardless of the width it was given.
func Footer(width int) lipgloss.Style { return lipgloss.NewStyle().MaxWidth(width) }

// FieldCell styles one cell of the detail pane's field table. Column 0 is the
// field name, column 1 the value.
func FieldCell(_, col int) lipgloss.Style {
	if col == 0 {
		return FieldName.PaddingRight(2)
	}
	return FieldVal
}

// Pane titles.
const (
	VaultPaneTitle  = "Vaults"
	ItemPaneTitle   = "Items"
	DetailPaneTitle = "Details"
)

// Singular and plural nouns for list.SetStatusBarItemName.
const (
	VaultNoun       = "vault"
	VaultNounPlural = "vaults"
	ItemNoun        = "item"
	ItemNounPlural  = "items"
)

// Empty and placeholder strings.
const (
	NoSelection      = "select an item"
	NoFields         = "no fields"
	TerminalTooSmall = "terminal too small"
	Mask             = "••••••••"
)

// Status messages.
const (
	CopiedStatus  = "copied"
	DeletedStatus = "deleted"
	SavedStatus   = "saved"
	CreatedStatus = "created"
)

// Form prompts.
const (
	VaultNamePrompt     = "Vault name"
	ItemTitlePrompt     = "Title"
	ItemCategoryPrompt  = "Category"
	FieldLabelPrompt    = "Label"
	FieldTypePrompt     = "Type"
	FieldValuePrompt    = "Value"
	DeleteVaultPrompt   = "Delete vault %q?"
	DeleteItemPrompt    = "Delete item %q?"
	DeleteFieldPrompt   = "Delete field %q?"
	ConfirmVaultPrompt  = "Type the vault name to confirm"
	RequiredValidation  = "required"
	NameMismatchMessage = "name does not match"
)

// AppName is the window title prefix.
const AppName = "optui"
