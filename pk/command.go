package pk

import (
	"context"
	"fmt"
	"os/exec"
)

// RunCommand executes a command in the directory specified by the context path.
// It gets the path from PathFromContext(ctx), resolves it relative to git root,
// and executes the command with that working directory.
// Returns combined stdout+stderr and error.
func RunCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	// Get path from context
	path := PathFromContext(ctx)

	// Construct absolute path to target directory
	targetDir := FromGitRoot(path)

	// Create command
	cmd := exec.Command(name, args...)
	cmd.Dir = targetDir

	// Run and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("command failed in %s: %w", path, err)
	}

	return output, nil
}

// RunCommandString executes a command and returns output as a string.
// Useful for commands like pwd, git, etc. that return simple text.
func RunCommandString(ctx context.Context, name string, args ...string) (string, error) {
	output, err := RunCommand(ctx, name, args...)
	return string(output), err
}
