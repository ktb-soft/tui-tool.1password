package keymap

import (
	"slices"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

func TestKeyMapImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = Default()
}

func TestMovementBindsArrowAndVimKeysTogether(t *testing.T) {
	keys := Default()

	cases := []struct {
		name    string
		binding key.Binding
		want    []string
	}{
		{"Up", keys.Up, []string{"up", "k"}},
		{"Down", keys.Down, []string{"down", "j"}},
		{"PaneLeft", keys.PaneLeft, []string{"left", "h"}},
		{"PaneRight", keys.PaneRight, []string{"right", "l"}},
	}

	for _, tc := range cases {
		if got := tc.binding.Keys(); !slices.Equal(got, tc.want) {
			t.Errorf("%s keys = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestEveryBindingHasHelpText(t *testing.T) {
	keys := Default()

	for _, column := range keys.FullHelp() {
		for _, binding := range column {
			if binding.Help().Key == "" || binding.Help().Desc == "" {
				t.Errorf("binding %v has incomplete help: %+v", binding.Keys(), binding.Help())
			}
		}
	}
}

func TestFullHelpCoversEveryBinding(t *testing.T) {
	keys := Default()

	documented := map[string]bool{}
	for _, column := range keys.FullHelp() {
		for _, binding := range column {
			for _, k := range binding.Keys() {
				documented[k] = true
			}
		}
	}

	all := []key.Binding{
		keys.Up, keys.Down, keys.PaneLeft, keys.PaneRight,
		keys.Create, keys.Edit, keys.Delete, keys.Reveal, keys.Copy,
		keys.Filter, keys.Help, keys.Confirm, keys.Cancel, keys.Quit,
	}
	for _, binding := range all {
		for _, k := range binding.Keys() {
			if !documented[k] {
				t.Errorf("key %q appears in no FullHelp column", k)
			}
		}
	}
}

func TestShortHelpIsASubsetOfTheBindings(t *testing.T) {
	keys := Default()

	if len(keys.ShortHelp()) == 0 {
		t.Fatal("ShortHelp is empty")
	}
	for _, binding := range keys.ShortHelp() {
		if len(binding.Keys()) == 0 {
			t.Errorf("ShortHelp holds a binding with no keys: %+v", binding.Help())
		}
	}
}

func TestQuitBindsBothQAndCtrlC(t *testing.T) {
	if got, want := Default().Quit.Keys(), []string{"q", "ctrl+c"}; !slices.Equal(got, want) {
		t.Errorf("Quit keys = %v, want %v", got, want)
	}
}
