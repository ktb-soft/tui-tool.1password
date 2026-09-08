package form

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// FieldDraft is one editable row of the item form. Clearing Label removes the
// field from the item the draft produces.
type FieldDraft struct {
	Label string
	Type  string
	Value string

	id      string
	purpose string
	section *op.Section
}

// ItemDraft is the mutable state an item form binds to. Read it once the form
// reports huh.StateCompleted.
type ItemDraft struct {
	Title    string
	Category string
	Fields   []FieldDraft

	base op.Item
}

// NewItem builds the overlay form for creating an item, returning the draft its
// inputs are bound to. An item with no ID is a create, so the category is
// selectable; an existing item shows its category read-only, because `op` will
// not change it.
//
// The draft always carries one blank field row, which is how a field is added.
func NewItem(item op.Item) (*huh.Form, *ItemDraft) {
	draft := newItemDraft(item)
	draft.Fields = append(draft.Fields, FieldDraft{Type: op.FieldTypeString})

	groups := []*huh.Group{draft.headerGroup()}
	for index := range draft.Fields {
		groups = append(groups, draft.fieldGroups(index)...)
	}
	return huh.NewForm(groups...), draft
}

// NewInlineItem builds the detail pane's form: the item's title over one input
// per field, each titled with the field's label and bound to its value. Every
// field is in a single group, so the whole item is on screen at once rather
// than paged.
//
// A concealed field echoes as a password unless revealed, so its value never
// reaches the screen.
func NewInlineItem(item op.Item, revealed bool) (*huh.Form, *ItemDraft) {
	draft := newItemDraft(item)

	fields := []huh.Field{huh.NewInput().
		Title(theme.ItemTitlePrompt).Value(&draft.Title).Validate(requireValue)}
	for index := range draft.Fields {
		fields = append(fields, draft.valueInput(index, revealed))
	}
	if len(draft.Fields) == 0 {
		fields = append(fields, huh.NewNote().Description(theme.NoFields))
	}
	return huh.NewForm(huh.NewGroup(fields...)), draft
}

// valueInput is one field of the inline form, masked when it holds a secret
// that has not been revealed. huh fixes an input's echo mode at construction,
// so revealing rebuilds the form rather than mutating it.
func (d *ItemDraft) valueInput(index int, revealed bool) huh.Field {
	field := &d.Fields[index]
	input := huh.NewInput().Title(field.title()).Value(&field.Value)
	if field.Type == op.FieldTypeConcealed && !revealed {
		return input.EchoMode(huh.EchoModePassword)
	}
	return input
}

// NewDeleteItem builds the delete confirmation, defaulted to No.
func NewDeleteItem(name string) (*huh.Form, *bool) {
	confirmed := new(bool)
	confirmation := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title(fmt.Sprintf(theme.DeleteItemPrompt, name)).Value(confirmed),
	))
	return confirmation, confirmed
}

// Item folds the draft back onto the item it was seeded from. Rows whose label
// is blank are dropped, so clearing a label deletes that field and an untouched
// blank row adds nothing.
func (d ItemDraft) Item() op.Item {
	item := d.base
	item.Name = strings.TrimSpace(d.Title)
	item.Category = d.Category
	item.Fields = nil

	for _, draft := range d.Fields {
		label := strings.TrimSpace(draft.Label)
		if label == "" {
			continue
		}
		item.Fields = append(item.Fields, op.Field{
			ID:      draft.id,
			Label:   label,
			Type:    draft.Type,
			Purpose: draft.purpose,
			Value:   draft.Value,
			Section: draft.section,
		})
	}
	return item
}

func newItemDraft(item op.Item) *ItemDraft {
	draft := &ItemDraft{Title: item.Name, Category: item.Category, base: item}
	if draft.Category == "" {
		draft.Category = op.Categories[0]
	}

	for _, field := range item.Fields {
		draft.Fields = append(draft.Fields, FieldDraft{
			Label:   field.Label,
			Type:    field.Type,
			Value:   field.Value,
			id:      field.ID,
			purpose: field.Purpose,
			section: field.Section,
		})
	}
	return draft
}

// title names the field, qualified by its section when it has one, so the
// grouping the detail pane used to draw as headings is still visible in a flat
// form.
func (f FieldDraft) title() string {
	if f.section == nil || f.section.Label == "" {
		return f.Label
	}
	return f.section.Label + theme.SectionSeparator + f.Label
}

func requireValue(value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(theme.RequiredValidation)
	}
	return nil
}

func (d *ItemDraft) headerGroup() *huh.Group {
	title := huh.NewInput().Title(theme.ItemTitlePrompt).Value(&d.Title).
		Validate(requireValue)

	if d.base.ID != "" {
		return huh.NewGroup(title,
			huh.NewNote().Title(theme.ItemCategoryPrompt).Description(d.Category))
	}
	return huh.NewGroup(title, huh.NewSelect[string]().
		Title(theme.ItemCategoryPrompt).
		Options(huh.NewOptions(op.Categories...)...).
		Value(&d.Category))
}

// fieldGroups pages one field: its label and type, then its value. The value
// appears twice, once masked and once not, with only the one matching the
// chosen type shown, because huh fixes an input's echo mode at construction.
func (d *ItemDraft) fieldGroups(index int) []*huh.Group {
	field := &d.Fields[index]

	editor := huh.NewGroup(
		huh.NewInput().Title(theme.FieldLabelPrompt).
			Description(theme.FieldRemoveHint).Value(&field.Label),
		huh.NewSelect[string]().Title(theme.FieldTypePrompt).
			Options(huh.NewOptions(op.FieldTypes...)...).Value(&field.Type),
	)

	shows := func(concealed bool) bool {
		return strings.TrimSpace(field.Label) != "" &&
			(field.Type == op.FieldTypeConcealed) == concealed
	}
	plain := huh.NewGroup(
		huh.NewInput().Title(theme.FieldValuePrompt).Value(&field.Value),
	).WithHideFunc(func() bool { return !shows(false) })
	secret := huh.NewGroup(
		huh.NewInput().Title(theme.FieldValuePrompt).Value(&field.Value).
			EchoMode(huh.EchoModePassword),
	).WithHideFunc(func() bool { return !shows(true) })

	return []*huh.Group{editor, plain, secret}
}
