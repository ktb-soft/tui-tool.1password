package app

import (
	"sync"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

// cache holds what `op` has already returned, for the life of the process and
// in memory only. Nothing here is written to disk, logged, or put in an error;
// field values are secrets. See docs/adr/09-00-00-in-memory-cache.md.
//
// Entries never expire. A write invalidates what it changed and the refresh
// binding drops everything, so no cached value survives a change the user made
// and none is refetched on a timer.
type cache struct {
	mutex     sync.Mutex
	vaults    []op.Vault
	hasVaults bool
	itemLists map[string][]op.Item
	items     map[string]map[string]op.Item
}

// newCache returns an empty cache.
func newCache() *cache {
	return &cache{
		itemLists: map[string][]op.Item{},
		items:     map[string]map[string]op.Item{},
	}
}

// getVaults returns the cached vault list, calling fetch only on a miss.
func (c *cache) getVaults(fetch func() ([]op.Vault, error)) ([]op.Vault, error) {
	c.mutex.Lock()
	vaults, ok := c.vaults, c.hasVaults
	c.mutex.Unlock()
	if ok {
		return vaults, nil
	}

	vaults, err := fetch()
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.vaults, c.hasVaults = vaults, true
	return vaults, nil
}

// getItems returns the cached item list for one vault, calling fetch only on a
// miss.
func (c *cache) getItems(vaultID string, fetch func() ([]op.Item, error)) ([]op.Item, error) {
	c.mutex.Lock()
	items, ok := c.itemLists[vaultID]
	c.mutex.Unlock()
	if ok {
		return items, nil
	}

	items, err := fetch()
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.itemLists[vaultID] = items
	return items, nil
}

// getItem returns the cached item, field values included, calling fetch only
// on a miss.
func (c *cache) getItem(vaultID, itemID string, fetch func() (op.Item, error)) (op.Item, error) {
	c.mutex.Lock()
	item, ok := c.items[vaultID][itemID]
	c.mutex.Unlock()
	if ok {
		return item, nil
	}

	item, err := fetch()
	if err != nil {
		return op.Item{}, err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.items[vaultID] == nil {
		c.items[vaultID] = map[string]op.Item{}
	}
	c.items[vaultID][itemID] = item
	return item, nil
}

// putItems stores a vault's items with their field values, as one bulk read
// returned them, so selecting any row in that vault reaches no subprocess.
func (c *cache) putItems(vaultID string, items []op.Item) {
	byID := make(map[string]op.Item, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[vaultID] = byID
}

// removeVaults drops the vault list, so the next read refetches it.
func (c *cache) removeVaults() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.vaults, c.hasVaults = nil, false
}

// removeVaultItems drops one vault's item list and every item cached from it,
// so an item create, edit, or delete cannot leave a stale row or a stale value
// on screen.
func (c *cache) removeVaultItems(vaultID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.itemLists, vaultID)
	delete(c.items, vaultID)
}

// reset drops everything, which is what the refresh binding does.
func (c *cache) reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.vaults, c.hasVaults = nil, false
	c.itemLists = map[string][]op.Item{}
	c.items = map[string]map[string]op.Item{}
}

// invalidate drops what a completed write changed. The pane the write named is
// the whole of what the program knows about the change, and it is enough: a
// vault write changes the vault list, and an item write changes the selected
// vault's items and any of their values already fetched.
func (m Model) invalidate(written Focus) {
	if written == VaultPane {
		m.cache.removeVaults()
		return
	}
	m.cache.removeVaultItems(m.selectedVaultID())
}

// refresh drops every cached result and reloads from the vault list down, so
// the whole selection cascade refetches.
func (m Model) refresh() tea.Cmd {
	m.cache.reset()
	return loadVaults(m.client, m.cache)
}
