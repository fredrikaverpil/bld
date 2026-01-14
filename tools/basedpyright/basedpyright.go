// Package basedpyright provides basedpyright (Python static type checker) tool integration.
// basedpyright is installed via uv into a virtual environment.
package basedpyright

import (
	"context"
	"os"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

// renovate: datasource=pypi depName=basedpyright
const Version = "1.37.0"

// pythonVersion specifies the Python version for the virtual environment.
const pythonVersion = "3.12"

// Install ensures basedpyright is available.
// This is a hidden dependency used by Exec.
var Install = pocket.Func("install:basedpyright", "install basedpyright", install).Hidden()

func install(ctx context.Context) error {
	venvDir := pocket.FromToolsDir("basedpyright", Version)
	binary := pocket.VenvBinaryPath(venvDir, "basedpyright")

	// Skip if already installed.
	if _, err := os.Stat(binary); err == nil {
		_, err := pocket.CreateSymlink(binary)
		return err
	}

	pocket.Printf(ctx, "Installing basedpyright %s...\n", Version)

	// Create virtual environment (uv auto-installs if needed).
	if err := uv.CreateVenv(ctx, venvDir, pythonVersion); err != nil {
		return err
	}

	// Install the package.
	if err := uv.PipInstall(ctx, venvDir, "basedpyright=="+Version); err != nil {
		return err
	}

	// Create symlink to .pocket/bin/.
	_, err := pocket.CreateSymlink(binary)
	return err
}

// Exec runs basedpyright with the given arguments.
func Exec(ctx context.Context, args ...string) error {
	pocket.Serial(ctx, Install)
	return pocket.Exec(ctx, "basedpyright", args...)
}
