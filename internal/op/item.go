package op

import "time"

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
	ID        string    `json:"id"`
	Name      string    `json:"title"`
	Category  string    `json:"category"`
	Vault     VaultRef  `json:"vault"`
	Tags      []string  `json:"tags,omitempty"`
	Fields    []Field   `json:"fields,omitempty"`
	URLs      []URL     `json:"urls,omitempty"`
	Sections  []Section `json:"sections,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Field is one value on an item.
type Field struct {
	ID      string   `json:"id"`
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
func (c Client) ListItems(vaultID string) ([]Item, error) { return nil, errNotImplemented }

// GetItem returns one item, including its field values.
func (c Client) GetItem(vaultID, id string) (Item, error) { return Item{}, errNotImplemented }

// CreateItem creates an item from a JSON template piped on stdin.
func (c Client) CreateItem(item Item) (Item, error) { return Item{}, errNotImplemented }

// EditItem replaces an item from a JSON template piped on stdin.
func (c Client) EditItem(item Item) (Item, error) { return Item{}, errNotImplemented }

// DeleteItem deletes one item.
func (c Client) DeleteItem(vaultID, id string) error { return errNotImplemented }
