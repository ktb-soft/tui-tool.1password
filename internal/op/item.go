package op

import (
	"encoding/json"
	"fmt"
	"time"
)

// vaultFlag scopes an item command to one vault.
const vaultFlag = "--vault"

// stdinArg tells `op` to read the item template from standard input, so no
// field value is ever passed in argv. See
// docs/adr/04-00-00-secrets-never-in-argv.md.
const stdinArg = "-"

// Field types as `op` spells them.
const (
	FieldTypeString    = "STRING"
	FieldTypeConcealed = "CONCEALED"
	FieldTypeURL       = "URL"
	FieldTypeOTP       = "OTP"
)

// FieldTypes are the types the field editor offers.
var FieldTypes = []string{FieldTypeString, FieldTypeConcealed, FieldTypeURL, FieldTypeOTP}

// Categories are the item categories the create form offers.
var Categories = []string{
	"LOGIN",
	"PASSWORD",
	"SECURE_NOTE",
	"API_CREDENTIAL",
	"CREDIT_CARD",
	"DATABASE",
	"IDENTITY",
	"SERVER",
	"SSH_KEY",
	"SOFTWARE_LICENSE",
	"WIRELESS_ROUTER",
}

// Item is a 1Password item, shaped as `op item get --format=json` returns it.
type Item struct {
	ID        string    `json:"id,omitempty"`
	Name      string    `json:"title"`
	Category  string    `json:"category"`
	Vault     VaultRef  `json:"vault"`
	Tags      []string  `json:"tags,omitempty"`
	Fields    []Field   `json:"fields,omitempty"`
	URLs      []URL     `json:"urls,omitempty"`
	Sections  []Section `json:"sections,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
}

// Field is one value on an item.
type Field struct {
	ID      string   `json:"id,omitempty"`
	Label   string   `json:"label"`
	Type    string   `json:"type"`
	Purpose string   `json:"purpose,omitempty"`
	Value   string   `json:"value,omitempty"`
	Section *Section `json:"section,omitempty"`
}

// Section groups fields within an item.
type Section struct {
	ID    string `json:"id"`
	Label string `json:"label,omitempty"`
}

// URL is a website attached to an item.
type URL struct {
	Label   string `json:"label,omitempty"`
	Primary bool   `json:"primary,omitempty"`
	HRef    string `json:"href"`
}

// IsSecret reports whether the field's value must be masked until revealed.
func (f Field) IsSecret() bool { return f.Type == FieldTypeConcealed }

// Title satisfies list.DefaultItem.
func (i Item) Title() string { return i.Name }

// Description satisfies list.DefaultItem, reporting the item's category.
func (i Item) Description() string { return i.Category }

// FilterValue satisfies list.Item.
func (i Item) FilterValue() string { return i.Name }

// ListItems returns the items in the given vault, without their field values.
func (c Client) ListItems(vaultID string) ([]Item, error) {
	out, err := c.run([]string{"item", "list", vaultFlag, vaultID}, nil)
	if err != nil {
		return nil, err
	}

	var items []Item
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("list items in vault %s: %w", vaultID, err)
	}
	return items, nil
}

// GetItem returns one item, including its field values. Concealed values come
// back in plaintext and are masked by the detail pane, never persisted.
func (c Client) GetItem(vaultID, id string) (Item, error) {
	out, err := c.run([]string{"item", "get", id, vaultFlag, vaultID}, nil)
	if err != nil {
		return Item{}, err
	}

	var item Item
	if err := json.Unmarshal(out, &item); err != nil {
		return Item{}, fmt.Errorf("get item %s: %w", id, err)
	}
	return item, nil
}

// CreateItem creates an item in item.Vault.ID from a JSON template piped on
// stdin, and returns the item `op` wrote.
func (c Client) CreateItem(item Item) (Item, error) {
	return c.writeItem([]string{"item", "create", vaultFlag, item.Vault.ID, stdinArg}, item)
}

// EditItem replaces an item from a JSON template piped on stdin, and returns
// the item `op` wrote.
func (c Client) EditItem(item Item) (Item, error) {
	return c.writeItem([]string{"item", "edit", item.ID, stdinArg}, item)
}

// DeleteItem deletes one item.
func (c Client) DeleteItem(vaultID, id string) error {
	_, err := c.run([]string{"item", "delete", id, vaultFlag, vaultID}, nil)
	return err
}

// writeItem pipes the item to `op` as a JSON template. The template carries
// every field value, so no value reaches argv.
func (c Client) writeItem(args []string, item Item) (Item, error) {
	template, err := json.Marshal(item)
	if err != nil {
		return Item{}, fmt.Errorf("write item %s: encode: %w", item.Name, err)
	}

	out, err := c.run(args, template)
	if err != nil {
		return Item{}, err
	}

	var written Item
	if err := json.Unmarshal(out, &written); err != nil {
		return Item{}, fmt.Errorf("write item %s: decode: %w", item.Name, err)
	}
	return written, nil
}
