// Command optui is a three-pane terminal UI over the 1Password CLI.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/ktb-soft/tui-tool.1password/internal/app"
	"github.com/ktb-soft/tui-tool.1password/internal/op"
)

func main() {
	client := op.Client{Path: op.DefaultPath, Account: os.Getenv("OP_ACCOUNT")}

	if _, err := tea.NewProgram(app.New(client)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
