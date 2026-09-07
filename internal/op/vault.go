package op

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Vault is a 1Password vault, shaped as `op vault list --format=json` returns it.
type Vault struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Note  string `json:"description,omitempty"`
	Items int    `json:"items,omitempty"`
}

// VaultRef is the abbreviated vault an item carries.
type VaultRef struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// Title satisfies list.DefaultItem.
func (v Vault) Title() string { return v.Name }

// Description satisfies list.DefaultItem, reporting the vault's item count.
func (v Vault) Description() string { return strconv.Itoa(v.Items) + " items" }

// FilterValue satisfies list.Item.
func (v Vault) FilterValue() string { return v.Name }

// ListVaults returns every vault the signed-in account can see.
func (c Client) ListVaults() ([]Vault, error) {
	out, err := c.run([]string{"vault", "list"}, nil)
	if err != nil {
		return nil, fmt.Errorf("list vaults: %w", err)
	}

	var vaults []Vault
	if err := json.Unmarshal(out, &vaults); err != nil {
		return nil, fmt.Errorf("list vaults: decode: %w", err)
	}
	return vaults, nil
}

// CreateVault creates a vault with the given name and returns it.
func (c Client) CreateVault(name string) (Vault, error) { return Vault{}, errNotImplemented }

// EditVault renames the vault with the given ID.
func (c Client) EditVault(id, name string) (Vault, error) { return Vault{}, errNotImplemented }

// DeleteVault deletes the vault with the given ID and every item in it.
func (c Client) DeleteVault(id string) error { return errNotImplemented }
