package theme

// DetailIndent is the left inset huh gives a form field — a one-cell border
// plus a one-cell pad — so the detail pane's empty state lines up with the
// fields that replace it.
const DetailIndent = 2

// EmptyDetail styles the detail pane's empty state at the same indent as the
// form it stands in for.
var EmptyDetail = Empty.PaddingLeft(DetailIndent)

// SectionSeparator joins a field's section label to its own label, so the
// grouping an item carries survives in a flat form.
const SectionSeparator = " · "
