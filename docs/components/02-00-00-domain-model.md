# Domain model

Three types, in `internal/op`, mirroring the shape `op --format=json` returns.
No mapping layer sits between the JSON and the UI — a second set of types
would be an abstraction earned by nothing.

```go
type Vault struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Items       int    `json:"items,omitempty"`
}

type Item struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Category string    `json:"category"`
	Vault    VaultRef  `json:"vault"`
	Tags     []string  `json:"tags,omitempty"`
	Fields   []Field   `json:"fields,omitempty"`
	URLs     []URL     `json:"urls,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Field struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Type     string  `json:"type"`
	Purpose  string  `json:"purpose,omitempty"`
	Value    string  `json:"value,omitempty"`
	Section  *Section `json:"section,omitempty"`
}
```

## Concealment

`Field.Type == "CONCEALED"` is the one piece of domain logic the UI needs:

```go
func (f Field) IsSecret() bool { return f.Type == "CONCEALED" }
```

The detail pane calls it to decide whether to mask. Nothing else branches on
field type.

## List identity

`bubbles/list` requires items to satisfy `list.Item`. `Vault` and `Item` get
`Title()`, `Description()`, and `FilterValue()` methods here rather than in a
wrapper type, so the panes hold domain values directly and
`list.SelectedItem()` type-asserts straight back to `op.Vault` / `op.Item`
(`charm-bubbles.md:2035`).

`Description()` returns the item count for a vault and the category for an
item.

## Sections

An item's fields carry an optional section. The detail pane groups by
`Field.Section.Label`, with sectionless fields first. That grouping is a
rendering concern and lives in `internal/ui/detail`, not here.
