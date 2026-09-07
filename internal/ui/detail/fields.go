package detail

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// fieldGroup is the fields sharing one section label. The sectionless group
// carries an empty label and renders without a heading.
type fieldGroup struct {
	label  string
	fields []op.Field
}

// groupFields buckets an item's fields by section label, sectionless first and
// the rest in the order their sections first appear.
func groupFields(fields []op.Field) []fieldGroup {
	groups := []fieldGroup{{}}
	index := map[string]int{"": 0}

	for _, field := range fields {
		label := ""
		if field.Section != nil {
			label = field.Section.Label
		}
		position, seen := index[label]
		if !seen {
			position = len(groups)
			index[label] = position
			groups = append(groups, fieldGroup{label: label})
		}
		groups[position].fields = append(groups[position].fields, field)
	}

	return withFields(groups)
}

func withFields(groups []fieldGroup) []fieldGroup {
	populated := make([]fieldGroup, 0, len(groups))
	for _, group := range groups {
		if len(group.fields) > 0 {
			populated = append(populated, group)
		}
	}
	return populated
}

// renderGroups renders each group as a heading over a table, joined vertically.
func (d Detail) renderGroups(groups []fieldGroup, width int) string {
	blocks := make([]string, 0, len(groups))
	for _, group := range groups {
		blocks = append(blocks, d.renderGroup(group, width))
	}
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func (d Detail) renderGroup(group fieldGroup, width int) string {
	rows := make([][]string, 0, len(group.fields))
	for _, field := range group.fields {
		rows = append(rows, []string{field.Label, d.displayValue(field)})
	}

	fields := table.New().
		Border(lipgloss.HiddenBorder()).
		BorderTop(false).
		BorderBottom(false).
		StyleFunc(theme.FieldCell).
		Width(width).
		Rows(rows...).
		String()

	if group.label == "" {
		return fields
	}
	return lipgloss.JoinVertical(lipgloss.Left, theme.Section.Render(group.label), fields)
}

// displayValue masks a concealed field unless it has been revealed.
func (d Detail) displayValue(field op.Field) string {
	if field.IsSecret() && !d.revealed[field.ID] {
		return theme.Mask
	}
	return field.Value
}
