package op

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// fakeOp writes a script that records its argv and stdin, prints stdout, and
// exits with the given code.
func fakeOp(t *testing.T, stdout string, exitCode int) (path, argvFile, stdinFile string) {
	t.Helper()

	dir := t.TempDir()
	path = filepath.Join(dir, "op")
	argvFile = filepath.Join(dir, "argv")
	stdinFile = filepath.Join(dir, "stdin")

	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" > " + argvFile + "\n" +
		"cat > " + stdinFile + "\n" +
		"printf '%s' '" + stdout + "'\n" +
		"printf 'not signed in\\n' >&2\n" +
		"exit " + strconv.Itoa(exitCode) + "\n"

	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake op: %v", err)
	}
	return path, argvFile, stdinFile
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func TestRunAppendsFormatFlag(t *testing.T) {
	path, argvFile, _ := fakeOp(t, "[]", 0)

	out, err := Client{Path: path}.run([]string{"vault", "list"}, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if string(out) != "[]" {
		t.Errorf("stdout = %q, want %q", out, "[]")
	}

	want := []string{"vault", "list", formatFlag}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
}

func TestRunAppendsAccountWhenSet(t *testing.T) {
	path, argvFile, _ := fakeOp(t, "[]", 0)

	if _, err := (Client{Path: path, Account: "example"}).run([]string{"vault", "list"}, nil); err != nil {
		t.Fatalf("run: %v", err)
	}

	want := []string{"vault", "list", formatFlag, accountFlag, "example"}
	if got := readLines(t, argvFile); !slices.Equal(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
}

func TestRunPipesStdin(t *testing.T) {
	path, _, stdinFile := fakeOp(t, "{}", 0)

	template := []byte(`{"title":"Example Login"}`)
	if _, err := (Client{Path: path}).run([]string{"item", "create", "-"}, template); err != nil {
		t.Fatalf("run: %v", err)
	}

	got, err := os.ReadFile(stdinFile)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if string(got) != string(template) {
		t.Errorf("stdin = %q, want %q", got, template)
	}
}

func TestRunNeverPutsSecretsInArgv(t *testing.T) {
	path, argvFile, _ := fakeOp(t, "{}", 0)

	if _, err := (Client{Path: path}).run([]string{"item", "edit", "abc", "-"}, []byte("hunter2")); err != nil {
		t.Fatalf("run: %v", err)
	}

	for _, arg := range readLines(t, argvFile) {
		if strings.Contains(arg, "hunter2") {
			t.Fatalf("secret leaked into argv: %q", arg)
		}
	}
}

func TestRunWrapsNonZeroExitInOpError(t *testing.T) {
	path, _, _ := fakeOp(t, "", 1)

	_, err := Client{Path: path}.run([]string{"vault", "delete", "abc"}, nil)
	if err == nil {
		t.Fatal("run: want error, got nil")
	}

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %T, want *OpError", err)
	}
	if !strings.Contains(opErr.Error(), "not signed in") {
		t.Errorf("Error() = %q, want it to carry the stderr text", opErr.Error())
	}
	if !strings.HasPrefix(opErr.Error(), "op vault delete abc") {
		t.Errorf("Error() = %q, want it to name the command", opErr.Error())
	}
}

func TestRunReportsMissingBinary(t *testing.T) {
	_, err := Client{Path: filepath.Join(t.TempDir(), "absent")}.run([]string{"vault", "list"}, nil)

	var opErr *OpError
	if !errors.As(err, &opErr) {
		t.Fatalf("err = %v, want *OpError", err)
	}
	if opErr.Unwrap() == nil {
		t.Error("Unwrap() = nil, want the exec error")
	}
}

func TestOpErrorFallsBackToWrappedError(t *testing.T) {
	err := &OpError{Args: []string{"vault", "list"}, Err: errors.New("exec format error")}

	if got, want := err.Error(), "op vault list: exec format error"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
