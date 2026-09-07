// Package op adapts the 1Password CLI. It is the only package in the project
// that runs a subprocess; every other package receives Go values.
package op

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// DefaultPath is the `op` binary used when Client.Path is empty.
const DefaultPath = "op"

const (
	formatFlag  = "--format=json"
	accountFlag = "--account"
)

// errNotImplemented marks a signature that slice 0 froze but did not fill in.
var errNotImplemented = errors.New("not implemented")

// Client runs the `op` binary. The zero value uses DefaultPath and whatever
// account `op` is already signed in to.
type Client struct {
	Path    string
	Account string
}

// OpError is a non-zero exit from `op`, carrying the command that failed and
// the stderr text `op` wrote. The controller shows Stderr verbatim.
type OpError struct {
	Args   []string
	Stderr string
	Err    error
}

// Error renders `op vault delete: <stderr>`.
func (e *OpError) Error() string {
	detail := strings.TrimSpace(e.Stderr)
	if detail == "" && e.Err != nil {
		detail = e.Err.Error()
	}
	return fmt.Sprintf("op %s: %s", strings.Join(e.Args, " "), detail)
}

// Unwrap exposes the underlying exec error.
func (e *OpError) Unwrap() error { return e.Err }

// run is the single exec choke point: it appends --format=json and --account,
// pipes stdin, captures stdout and stderr separately, and turns a non-zero
// exit into an *OpError.
func (c Client) run(args []string, stdin []byte) ([]byte, error) {
	full := c.commandArgs(args)

	path := c.Path
	if path == "" {
		path = DefaultPath
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(path, full...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = bytes.NewReader(stdin)

	if err := cmd.Run(); err != nil {
		return nil, &OpError{Args: full, Stderr: stderr.String(), Err: err}
	}
	return stdout.Bytes(), nil
}

// commandArgs appends the flags every invocation carries.
func (c Client) commandArgs(args []string) []string {
	full := make([]string, 0, len(args)+3)
	full = append(full, args...)
	full = append(full, formatFlag)
	if c.Account != "" {
		full = append(full, accountFlag, c.Account)
	}
	return full
}
