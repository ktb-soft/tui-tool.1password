package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"charm.land/huh/v2"
	"github.com/ktb-soft/tui-tool.1password/internal/ui/theme"
)

// Update dispatches on message type and delegates. Every branch calls a
// handler that lives in the file owning that vertical.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resizePanes(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case vaultsLoadedMsg:
		return m.updateVaults(msg)

	case itemsLoadedMsg:
		return m.updateItems(msg)

	case itemLoadedMsg:
		return m.updateDetail(msg)

	case selectionChangedMsg:
		return m.routeSelection(msg)

	case writeSucceededMsg:
		return m.handleWriteSucceeded(msg)

	case opFailedMsg:
		return m.handleOpFailed(msg)
	}

	// huh advances a form on its own messages, not on key presses alone, so an
	// open overlay takes everything the cases above did not claim. Without
	// this, a form can never reach StateCompleted and nothing is ever
	// submitted.
	if m.overlay != nil {
		return m.updateOverlay(msg)
	}
	return m.forwardToFocused(msg)
}

func (m Model) routeSelection(msg selectionChangedMsg) (tea.Model, tea.Cmd) {
	return m.updatePane(msg.Pane, msg)
}

// updatePane hands a message to the vertical owning the named pane.
func (m Model) updatePane(target Focus, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch target {
	case VaultPane:
		return m.updateVaults(msg)
	case ItemPane:
		return m.updateItems(msg)
	default:
		return m.updateDetail(msg)
	}
}

// handleWriteSucceeded refreshes the pane the completed write named. A delete
// also empties the panes fed by it, so the record just removed — and any
// concealed field revealed on it — cannot stay on screen.
func (m Model) handleWriteSucceeded(msg writeSucceededMsg) (tea.Model, tea.Cmd) {
	status := m.statusCmd(msg.Status)
	if msg.Removed {
		m.clearBelow(msg.Reload)
	}
	model, reload := m.updatePane(msg.Reload, msg)
	return model, tea.Batch(status, reload)
}

// handleOpFailed shows the op error verbatim in the focused pane's status bar,
// leaving the pane's contents untouched.
func (m Model) handleOpFailed(msg opFailedMsg) (tea.Model, tea.Cmd) {
	return m, m.statusCmd(msg.Err.Error())
}

func (m *Model) statusCmd(text string) tea.Cmd {
	focused := m.focusedPane()
	if text == "" || focused == nil {
		return nil
	}
	return focused.List().NewStatusMessage(text)
}

// handleKey follows the routing order in
// docs/components/04-00-00-navigation-keymap.md, stopping at the first match.
func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.overlay != nil {
		if key.Matches(msg, m.keys.Cancel) {
			m.closeOverlay()
			return m, nil
		}
		return m.updateOverlay(msg)
	}
	if m.isFiltering() {
		return m.forwardToFocused(msg)
	}
	if model, cmd, handled := m.handleGlobalKey(msg); handled {
		return model, cmd
	}
	return m.handlePaneKey(msg)
}

func (m Model) isFiltering() bool {
	focused := m.focusedPane()
	return focused != nil && focused.List().SettingFilter()
}

func (m Model) handleGlobalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit, true
	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll
		return m, nil, true
	default:
		return m, nil, false
	}
}

// handlePaneKey gives movement to the app and everything else to the focused
// vertical.
func (m Model) handlePaneKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.PaneLeft):
		m.focusLeft()
		return m, nil
	case key.Matches(msg, m.keys.PaneRight):
		m.focusRight()
		return m, nil
	case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
		return m.moveCursor(msg)
	}

	var model tea.Model
	var cmd tea.Cmd
	var handled bool
	switch m.focus {
	case VaultPane:
		model, cmd, handled = m.handleVaultKey(msg)
	case ItemPane:
		model, cmd, handled = m.handleItemKey(msg)
	default:
		model, cmd, handled = m.handleDetailKey(msg)
	}
	if handled {
		return model, cmd
	}
	return m.forwardToFocused(msg)
}

// updateOverlay routes every message to the open form and, when the form
// settles, clears the overlay and runs the submit the vertical supplied.
func (m Model) updateOverlay(msg tea.Msg) (tea.Model, tea.Cmd) {
	_, cmd := m.overlay.Update(msg)

	switch m.overlay.State {
	case huh.StateCompleted:
		submit := m.overlaySubmit
		m.closeOverlay()
		if submit == nil {
			return m, cmd
		}
		submitted, submitCmd := submit(m)
		return submitted, tea.Batch(cmd, submitCmd)
	case huh.StateAborted:
		m.closeOverlay()
		return m, cmd
	default:
		return m, cmd
	}
}

// openOverlay shows a form over the panes, sized from the theme ratios, and
// records the submit to run when it completes. Every vertical opens its forms
// through here so they are all sized and dismissed the same way.
func (m *Model) openOverlay(form *huh.Form, submit func(Model) (tea.Model, tea.Cmd)) tea.Cmd {
	m.overlay = form.
		WithWidth(int(float64(m.width) * theme.OverlayWidthRatio)).
		WithHeight(int(float64(m.height) * theme.OverlayHeightRatio))
	m.overlaySubmit = submit
	return m.overlay.Init()
}

// closeOverlay dismisses the open form. huh only aborts on ctrl+c, so esc is
// handled here to match the keymap.
func (m *Model) closeOverlay() {
	m.overlay, m.overlaySubmit = nil, nil
}

// forwardToFocused hands a message to the focused pane's component.
func (m Model) forwardToFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case VaultPane:
		m.vaults, cmd = m.vaults.Update(msg)
	case ItemPane:
		m.items, cmd = m.items.Update(msg)
	default:
		m.detail, cmd = m.detail.Update(msg)
	}
	return m, cmd
}
