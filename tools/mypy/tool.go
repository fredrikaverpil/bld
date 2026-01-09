// Package mypy provides mypy (Python static type checker) tool integration.
// mypy is installed via uv into a virtual environment.
package mypy

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tool"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

const name = "mypy"

// renovate: datasource=pypi depName=mypy
const version = "1.19.1"

// pythonVersion specifies the Python version for the virtual environment.
const pythonVersion = "3.12"

var t = &tool.Tool{Name: name, Prepare: Prepare}

// Command prepares the tool and returns an exec.Cmd for running mypy.
var Command = t.Command

// Run installs (if needed) and executes mypy.
var Run = t.Run

// Prepare ensures mypy is installed.
func Prepare(ctx context.Context) error {
	// Use version-based path: .pocket/tools/mypy/<version>/
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

	// Install mypy.
	if err := uv.PipInstall(ctx, venvDir, name+"=="+version); err != nil {
		return err
	}

	// Create symlink (or copy on Windows) to .pocket/bin/.
	_, err := tool.CreateSymlink(binary)
	return err
}
