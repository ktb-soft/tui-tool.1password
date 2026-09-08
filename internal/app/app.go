// Package app wires internal/op to internal/ui. It holds the only tea.Model
// the program runs.
package app

import (
	"time"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/op"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/detail"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/keymap"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/pane"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// Focus names the pane that receives pane-scoped keys.
type Focus int

// The three panes, left to right.
const (
	VaultPane Focus = iota
	ItemPane
	DetailPane
)

// SelectionDebounce is how long the cursor must rest before a load is issued,
// so holding j does not launch one op process per keystroke.
const SelectionDebounce = 150 * time.Millisecond

// Model is the root model.
type Model struct {
	client  op.Client
	cache   *cache
	keys    keymap.KeyMap
	help    help.Model
	vaults  pane.Pane
	items   pane.Pane
	detail  detail.Detail
	overlay *huh.Form
	// overlaySubmit runs when the open form completes. The vertical that
	// opened the overlay sets it, so update.go never learns which form is up.
	overlaySubmit func(Model) (tea.Model, tea.Cmd)
	focus         Focus
	width         int
	height        int
}

// New builds the root model around the given op client.
func New(client op.Client) Model {
	model := Model{
		client: client,
		cache:  newCache(),
		keys:   keymap.Default(),
		help:   help.New(),
		vaults: pane.New(theme.VaultPaneTitle, theme.VaultNoun, theme.VaultNounPlural),
		items:  pane.New(theme.ItemPaneTitle, theme.ItemNoun, theme.ItemNounPlural),
		detail: detail.New(),
	}
	model.applyFocus()
	return model
}

// Init loads the vault list.
func (m Model) Init() tea.Cmd { return loadVaults(m.client, m.cache) }

// View joins the three panes and the footer, then places any overlay on top.
func (m Model) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = theme.AppName
	return view
}

func (m Model) render() string {
	if m.width < theme.MinWidth || m.height < theme.MinHeight {
		return theme.Empty.Render(theme.TerminalTooSmall)
	}

	columns := lipgloss.JoinHorizontal(lipgloss.Top,
		m.vaults.View(), m.items.View(), m.detail.View())
	screen := lipgloss.JoinVertical(lipgloss.Left, columns, theme.Footer(m.width).Render(m.help.View(m.keys)))

	if m.overlay == nil {
		return screen
	}
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, m.overlay.View())
}

// resizePanes divides the terminal width by the theme ratios, giving the last
// column the rounding remainder so the three always sum to the full width.
func (m *Model) resizePanes(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)

	paneHeight := height - theme.FooterHeight
	vaultWidth := int(float64(width) * theme.VaultPaneRatio)
	itemWidth := int(float64(width) * theme.ItemPaneRatio)
	detailWidth := width - vaultWidth - itemWidth

	m.vaults.SetSize(vaultWidth, paneHeight)
	m.items.SetSize(itemWidth, paneHeight)
	m.detail.SetSize(detailWidth, paneHeight)
}

// focusLeft moves focus one pane left, clamped.
func (m *Model) focusLeft() {
	if m.focus > VaultPane {
		m.focus--
		m.applyFocus()
	}
}

// focusRight moves focus one pane right, clamped, and only onto a pane that
// has content to act on.
func (m *Model) focusRight() {
	if m.focus < DetailPane && m.canFocus(m.focus+1) {
		m.focus++
		m.applyFocus()
	}
}

func (m Model) canFocus(target Focus) bool {
	switch target {
	case ItemPane:
		return len(m.items.List().Items()) > 0
	case DetailPane:
		return m.detail.Item().ID != ""
	default:
		return true
	}
}

func (m *Model) applyFocus() {
	m.vaults.Blur()
	m.items.Blur()
	m.detail.Blur()

	switch m.focus {
	case VaultPane:
		m.vaults.Focus()
	case ItemPane:
		m.items.Focus()
	case DetailPane:
		m.detail.Focus()
	}
}

// focusedPane returns the focused list pane, or nil when the detail pane has
// focus, since it has no list.
func (m *Model) focusedPane() *pane.Pane {
	switch m.focus {
	case VaultPane:
		return &m.vaults
	case ItemPane:
		return &m.items
	default:
		return nil
	}
}
