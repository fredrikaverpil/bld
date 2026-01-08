// Package ruff provides ruff (Python linter and formatter) tool integration.
// ruff is installed via uv into a virtual environment.
package ruff

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tool"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

const name = "ruff"

// renovate: datasource=pypi depName=ruff
const version = "0.14.0"

// pythonVersion specifies the Python version for the virtual environment.
const pythonVersion = "3.12"

//go:embed ruff.toml
var defaultConfig []byte

// Command prepares the tool and returns an exec.Cmd for running ruff.
func Command(ctx context.Context, args ...string) (*exec.Cmd, error) {
	if err := Prepare(ctx); err != nil {
		return nil, err
	}
	return pocket.Command(ctx, pocket.FromBinDir(pocket.BinaryName(name)), args...), nil
}

// Run installs (if needed) and executes ruff.
func Run(ctx context.Context, args ...string) error {
	cmd, err := Command(ctx, args...)
	if err != nil {
		return err
	}
	return cmd.Run()
}

// ConfigPath returns the path to the ruff config file.
// It checks for ruff.toml or pyproject.toml in the repo root first,
// then falls back to the bundled default config.
func ConfigPath() (string, error) {
	// Check for user config in repo root.
	for _, configName := range []string{"ruff.toml", ".ruff.toml", "pyproject.toml"} {
		repoConfig := pocket.FromGitRoot(configName)
		if _, err := os.Stat(repoConfig); err == nil {
			return repoConfig, nil
		}
	}

	// Write bundled config to .pocket/tools/ruff/ruff.toml.
	configDir := pocket.FromToolsDir(name)
	configPath := filepath.Join(configDir, "ruff.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return "", fmt.Errorf("create config dir: %w", err)
		}
		if err := os.WriteFile(configPath, defaultConfig, 0o644); err != nil {
			return "", fmt.Errorf("write default config: %w", err)
		}
	}

	return configPath, nil
}

// Prepare ensures ruff is installed.
func Prepare(ctx context.Context) error {
	// Use version-based path: .pocket/tools/ruff/<version>/
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

	// Install ruff.
	if err := uv.PipInstall(ctx, venvDir, name+"=="+version); err != nil {
		return err
	}

	// Create symlink (or copy on Windows) to .pocket/bin/.
	_, err := tool.CreateSymlink(binary)
	return err
}
