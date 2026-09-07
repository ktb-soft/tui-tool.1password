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

const createdVaultJSON = `{"id":"vault-new","name":"Archive","items":0}`

func TestCreateVaultSendsTheNameAndDecodesTheResult(t *testing.T) {
	path, argvFile, stdinFile := fakeOp(t, createdVaultJSON, 0)

	vault, err := Client{Path: path}.CreateVault("Archive")
	if err != nil {
		t.Fatalf("CreateVault: %v", err)
	}

	if got, want := vault.ID, "vault-new"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	if got, want := vault.Name, "Archive"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}

	want := []string{"vault", "create", "Archive", formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if data, err := os.ReadFile(stdinFile); err != nil || len(data) != 0 {
		t.Errorf("stdin = %q (err %v), want empty", data, err)
	}
}

func TestCreateVaultReportsARejectedName(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	_, err := Client{Path: path}.CreateVault("Personal")
	if err == nil {
		t.Fatal("CreateVault: want an error for a duplicate name, got nil")
	}

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %T, want it to wrap *OpError", err)
	}
	if !strings.HasPrefix(err.Error(), "create vault:") {
		t.Errorf("err = %q, want it to name the operation", err)
	}
	if !strings.Contains(err.Error(), "not signed in") {
		t.Errorf("err = %q, want the stderr text verbatim", err)
	}
}

func TestCreateVaultReportsMalformedJSON(t *testing.T) {
	path, _, _ := fakeOp(t, `{"id":`, 0)

	_, err := Client{Path: path}.CreateVault("Archive")
	if err == nil {
		t.Fatal("CreateVault: want a decode error, got nil")
	}
	if !strings.Contains(err.Error(), "create vault: decode:") {
		t.Errorf("err = %q, want it to name the operation and the decode", err)
	}
}

func TestEditVaultRenamesByID(t *testing.T) {
	path, argvFile, stdinFile := fakeOp(t, "", 0)

	vault, err := Client{Path: path}.EditVault("vault-a", "Renamed")
	if err != nil {
		t.Fatalf("EditVault: %v", err)
	}

	if got, want := vault.ID, "vault-a"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	if got, want := vault.Name, "Renamed"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}

	want := []string{"vault", "edit", "vault-a", nameFlag, "Renamed", formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if data, err := os.ReadFile(stdinFile); err != nil || len(data) != 0 {
		t.Errorf("stdin = %q (err %v), want empty", data, err)
	}
}

func TestEditVaultReportsARejectedName(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	_, err := Client{Path: path}.EditVault("vault-a", "Personal")
	if err == nil {
		t.Fatal("EditVault: want an error, got nil")
	}

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %T, want it to wrap *OpError", err)
	}
	if !strings.HasPrefix(err.Error(), "edit vault:") {
		t.Errorf("err = %q, want it to name the operation", err)
	}
}

func TestEditVaultPassesTheAccountThrough(t *testing.T) {
	path, argvFile, _ := fakeOp(t, "", 0)

	if _, err := (Client{Path: path, Account: "example"}).EditVault("vault-a", "Renamed"); err != nil {
		t.Fatalf("EditVault: %v", err)
	}

	want := []string{"vault", "edit", "vault-a", nameFlag, "Renamed", formatFlag, accountFlag, "example"}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
}

func TestDeleteVaultDeletesByID(t *testing.T) {
	path, argvFile, _ := fakeOp(t, "", 0)

	if err := (Client{Path: path}).DeleteVault("vault-a"); err != nil {
		t.Fatalf("DeleteVault: %v", err)
	}

	want := []string{"vault", "delete", "vault-a", formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
}

func TestDeleteVaultReportsFailure(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	err := Client{Path: path}.DeleteVault("vault-a")
	if err == nil {
		t.Fatal("DeleteVault: want an error, got nil")
	}

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %T, want it to wrap *OpError", err)
	}
	if !strings.HasPrefix(err.Error(), "delete vault:") {
		t.Errorf("err = %q, want it to name the operation", err)
	}
}

func TestDeleteVaultReportsAMissingBinary(t *testing.T) {
	err := Client{Path: filepath.Join(t.TempDir(), "absent")}.DeleteVault("vault-a")

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %v, want it to wrap *OpError", err)
	}
}
