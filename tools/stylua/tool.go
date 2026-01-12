// Package stylua provides stylua tool integration.
package stylua

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tool"
)

const name = "stylua"

// renovate: datasource=github-releases depName=JohnnyMorganz/StyLua
const version = "2.3.1"

//go:embed stylua.toml
var defaultConfig []byte

// T is the tool instance for use with TaskContext.Tool().
// Example: tc.Tool(stylua.T).Run(ctx, ".").
var T = &tool.Tool{Name: name, Prepare: Prepare}

var configSpec = tool.ConfigSpec{
	ToolName:          name,
	UserConfigNames:   []string{"stylua.toml", ".stylua.toml"},
	DefaultConfigName: "stylua.toml",
	DefaultConfig:     defaultConfig,
}

// ConfigPath returns the path to the stylua config file.
// It checks for stylua.toml in the repo root first, then falls back
// to the bundled default config.
var ConfigPath = configSpec.Path

// Prepare ensures stylua is installed.
func Prepare(ctx context.Context) error {
	binDir := pocket.FromToolsDir(name, version, "bin")
	binaryName := pocket.BinaryName(name)
	binary := filepath.Join(binDir, binaryName)

	binURL := fmt.Sprintf(
		"https://github.com/JohnnyMorganz/StyLua/releases/download/v%s/stylua-%s-%s.zip",
		version,
		osName(),
		archName(),
	)

	return tool.FromRemote(
		ctx,
		binURL,
		tool.WithDestinationDir(binDir),
		tool.WithUnzip(),
		tool.WithSkipIfFileExists(binary),
		tool.WithSymlink(binary),
	)
}

func osName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macos"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}

func archName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return runtime.GOARCH
	}
}
