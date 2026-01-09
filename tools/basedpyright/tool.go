// Package basedpyright provides basedpyright (Python static type checker) tool integration.
// basedpyright is installed via uv into a virtual environment.
package basedpyright

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tool"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

const name = "basedpyright"

// renovate: datasource=pypi depName=basedpyright
const version = "1.37.0"

// pythonVersion specifies the Python version for the virtual environment.
const pythonVersion = "3.12"

var t = &tool.Tool{Name: name, Prepare: Prepare}

// Command prepares the tool and returns an exec.Cmd for running basedpyright.
var Command = t.Command

// Run installs (if needed) and executes basedpyright.
var Run = t.Run

// Prepare ensures basedpyright is installed.
func Prepare(ctx context.Context) error {
	// Use version-based path: .pocket/tools/basedpyright/<version>/
	venvDir := pocket.FromToolsDir(name, version)

	// On Windows, venv uses Scripts/ instead of bin/, and .exe extension.
	var binary string
	if runtime.GOOS == "windows" {
		binary = filepath.Join(venvDir, "Scripts", name+".exe")
	} else {
		binary = filepath.Join(venvDir, "bin", name)
	}

	// Skip if already installed.
	if _, err := os.Stat(binary); err == nil {
		// Ensure symlink/copy exists.
		_, err := tool.CreateSymlink(binary)
		return err
	}

	// Create virtual environment.
	if err := uv.CreateVenv(ctx, venvDir, pythonVersion); err != nil {
		return err
	}

	// Install basedpyright.
	if err := uv.PipInstall(ctx, venvDir, name+"=="+version); err != nil {
		return err
	}

	// Create symlink (or copy on Windows) to .pocket/bin/.
	_, err := tool.CreateSymlink(binary)
	return err
}
