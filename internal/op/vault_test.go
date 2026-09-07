package op

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fakeOpServing writes a fake op that prints the named fixture and exits 0.
func fakeOpServing(t *testing.T, fixture string) (path, argvFile string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(fixtureDir, fixture))
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}

	path, argvFile, _ = fakeOp(t, string(data), 0)
	return path, argvFile
}

func TestListVaultsDecodesTheVaultList(t *testing.T) {
	path, argvFile := fakeOpServing(t, "vault-list.json")

	vaults, err := Client{Path: path}.ListVaults()
	if err != nil {
		t.Fatalf("ListVaults: %v", err)
	}

	if len(vaults) != 3 {
		t.Fatalf("len = %d, want 3", len(vaults))
	}
	if got, want := vaults[0].Name, "Personal"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := vaults[2].Description(), "0 items"; got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}

	want := []string{"vault", "list", formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
}

func TestListVaultsPassesTheAccountThrough(t *testing.T) {
	path, argvFile := fakeOpServing(t, "vault-list.json")

	if _, err := (Client{Path: path, Account: "example"}).ListVaults(); err != nil {
		t.Fatalf("ListVaults: %v", err)
	}

	want := []string{"vault", "list", formatFlag, accountFlag, "example"}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
}

func TestListVaultsReturnsEmptyForAnAccountWithNoVaults(t *testing.T) {
	path, _ := fakeOpServing(t, "empty-list.json")

	vaults, err := Client{Path: path}.ListVaults()
	if err != nil {
		t.Fatalf("ListVaults: %v", err)
	}
	if len(vaults) != 0 {
		t.Errorf("len = %d, want 0", len(vaults))
	}
}

func TestListVaultsReportsNotSignedIn(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	_, err := Client{Path: path}.ListVaults()
	if err == nil {
		t.Fatal("ListVaults: want error, got nil")
	}

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %T, want it to wrap *OpError", err)
	}
	if !strings.Contains(err.Error(), "not signed in") {
		t.Errorf("err = %q, want the stderr text", err)
	}
	if !strings.HasPrefix(err.Error(), "list vaults:") {
		t.Errorf("err = %q, want it to name the operation", err)
	}
}

func TestListVaultsReportsAMissingBinary(t *testing.T) {
	_, err := Client{Path: filepath.Join(t.TempDir(), "absent")}.ListVaults()

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %v, want it to wrap *OpError", err)
	}
}

func TestListVaultsReportsMalformedJSON(t *testing.T) {
	path, _, _ := fakeOp(t, `[{"id":`, 0)

	_, err := Client{Path: path}.ListVaults()
	if err == nil {
		t.Fatal("ListVaults: want a decode error, got nil")
	}

	var opErr *OpError
	if errors.As(err, &opErr) {
		t.Errorf("err = %v, want a decode error, not an *OpError", err)
	}
	if !strings.Contains(err.Error(), "list vaults: decode:") {
		t.Errorf("err = %q, want it to name the operation and the decode", err)
	}
}

func TestVaultWritesRemainUnimplemented(t *testing.T) {
	client := Client{Path: filepath.Join(t.TempDir(), "absent")}

	if _, err := client.CreateVault("Personal"); !errors.Is(err, errNotImplemented) {
		t.Errorf("CreateVault err = %v, want errNotImplemented", err)
	}
	if _, err := client.EditVault("id", "Personal"); !errors.Is(err, errNotImplemented) {
		t.Errorf("EditVault err = %v, want errNotImplemented", err)
	}
	if err := client.DeleteVault("id"); !errors.Is(err, errNotImplemented) {
		t.Errorf("DeleteVault err = %v, want errNotImplemented", err)
	}
}
