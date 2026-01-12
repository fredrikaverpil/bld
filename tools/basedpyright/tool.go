// Package basedpyright provides basedpyright (Python static type checker) tool integration.
// basedpyright is installed via uv into a virtual environment.
package basedpyright

import (
	"context"
	"os"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tool"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

const name = "basedpyright"

// renovate: datasource=pypi depName=basedpyright
const version = "1.37.0"

// pythonVersion specifies the Python version for the virtual environment.
const pythonVersion = "3.12"

// T is the tool instance for use with TaskContext.Tool().
// Example: tc.Tool(basedpyright.T).Run(ctx, ".").
var T = &tool.Tool{Name: name, Prepare: Prepare}

// Prepare ensures basedpyright is installed.
func Prepare(ctx context.Context) error {
	venvDir := pocket.FromToolsDir(name, version)
	binary := tool.VenvBinaryPath(venvDir, name)

	// Skip if already installed.
	if _, err := os.Stat(binary); err == nil {
		_, err := tool.CreateSymlink(binary)
		return err
	}

	// Create virtual environment.
	if err := uv.CreateVenv(ctx, venvDir, pythonVersion); err != nil {
		return err
	}

	// Install the package.
	if err := uv.PipInstall(ctx, venvDir, name+"=="+version); err != nil {
		return err
	}

	// Create symlink to .pocket/bin/.
	_, err := tool.CreateSymlink(binary)
	return err
}
