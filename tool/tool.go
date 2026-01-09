package tool

import (
	"context"
	"os/exec"

	"github.com/fredrikaverpil/pocket"
)

// Tool represents a tool that can be prepared (installed) and executed.
// It provides a standard Command and Run pattern that all tools share.
type Tool struct {
	// Name is the binary name (without .exe extension).
	Name string
	// Prepare ensures the tool is installed. It is called before Command.
	Prepare func(ctx context.Context) error
}

// Command prepares the tool and returns an exec.Cmd for running it.
func (t *Tool) Command(ctx context.Context, args ...string) (*exec.Cmd, error) {
	if err := t.Prepare(ctx); err != nil {
		return nil, err
	}
	return pocket.Command(ctx, pocket.FromBinDir(pocket.BinaryName(t.Name)), args...), nil
}

// Run prepares and executes the tool.
func (t *Tool) Run(ctx context.Context, args ...string) error {
	cmd, err := t.Command(ctx, args...)
	if err != nil {
		return err
	}
	return cmd.Run()
}
