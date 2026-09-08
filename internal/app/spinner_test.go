package app

import (
	"errors"
	"testing"

	"charm.land/bubbles/v2/spinner"
)

func TestALoadingVaultSpinsTheItemPane(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))
	model.client = fakeOpClient(t, "[]", 0)

	_, cmd := model.Update(selectionChangedMsg{Pane: VaultPane, ID: personalVaultID})

	spun := false
	for _, msg := range batched(cmd) {
		if _, ok := msg.(spinner.TickMsg); ok {
			spun = true
		}
	}
	if !spun {
		t.Error("the item pane did not start its spinner while the vault's items load")
	}
}

func TestTheArrivingItemListStopsTheSpinner(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))
	model.client = fakeOpClient(t, "[]", 0)

	spinning, _ := dispatch(t, model, selectionChangedMsg{Pane: VaultPane, ID: personalVaultID})
	if _, cmd := spinning.Update(spinner.TickMsg{}); cmd == nil {
		t.Fatal("the spinner stopped before the item list arrived")
	}

	loaded, _ := dispatch(t, spinning, itemsLoadedMsg{{ID: "item1", Name: "GitHub"}})
	if _, cmd := loaded.Update(spinner.TickMsg{}); cmd != nil {
		t.Error("the spinner kept running after the item list arrived")
	}
}

func TestAFailedLoadStopsTheSpinner(t *testing.T) {
	model := withVaults(t, sized(t, 120, 40))
	model.client = fakeOpClient(t, "[]", 0)

	spinning, _ := dispatch(t, model, selectionChangedMsg{Pane: VaultPane, ID: personalVaultID})
	failed, _ := dispatch(t, spinning, opFailedMsg{Err: errors.New("not signed in")})

	if _, cmd := failed.Update(spinner.TickMsg{}); cmd != nil {
		t.Error("the spinner kept running after the load failed")
	}
}
