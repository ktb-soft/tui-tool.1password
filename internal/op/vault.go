package op

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// nameFlag renames a vault. A vault name is not a secret, so it stays in argv;
// see docs/adr/04-00-00-secrets-never-in-argv.md.
const nameFlag = "--name"

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
func (c Client) CreateVault(name string) (Vault, error) {
	out, err := c.run([]string{"vault", "create", name}, nil)
	if err != nil {
		return Vault{}, fmt.Errorf("create vault: %w", err)
	}

	var vault Vault
	if err := json.Unmarshal(out, &vault); err != nil {
		return Vault{}, fmt.Errorf("create vault: decode: %w", err)
	}
	return vault, nil
}

// EditVault renames the vault with the given ID. `op vault edit` writes
// nothing to stdout, so the renamed vault is composed from the arguments.
func (c Client) EditVault(id, name string) (Vault, error) {
	if _, err := c.run([]string{"vault", "edit", id, nameFlag, name}, nil); err != nil {
		return Vault{}, fmt.Errorf("edit vault: %w", err)
	}
	return Vault{ID: id, Name: name}, nil
}

// DeleteVault deletes the vault with the given ID and every item in it.
func (c Client) DeleteVault(id string) error {
	if _, err := c.run([]string{"vault", "delete", id}, nil); err != nil {
		return fmt.Errorf("delete vault: %w", err)
	}
	return nil
}
