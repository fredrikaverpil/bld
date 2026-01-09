// Package golangcilint provides golangci-lint tool integration.
package golangcilint

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tool"
)

const name = "golangci-lint"

// renovate: datasource=github-releases depName=golangci/golangci-lint
const version = "2.7.1"

//go:embed golangci.yml
var defaultConfig []byte

var t = &tool.Tool{Name: name, Prepare: Prepare}

// Command prepares the tool and returns an exec.Cmd for running golangci-lint.
var Command = t.Command

// Run installs (if needed) and executes golangci-lint.
var Run = t.Run

var configSpec = tool.ConfigSpec{
	ToolName:          name,
	UserConfigNames:   []string{".golangci.yml", ".golangci.yaml"},
	DefaultConfigName: "golangci.yml",
	DefaultConfig:     defaultConfig,
}

// ConfigPath returns the path to the golangci-lint config file.
// It checks for .golangci.yml or .golangci.yaml in the repo root first,
// then falls back to the bundled default config.
var ConfigPath = configSpec.Path

// Prepare ensures golangci-lint is installed.
func Prepare(ctx context.Context) error {
	binDir := pocket.FromToolsDir(name, version, "bin")
	binaryName := pocket.BinaryName(name)
	binary := filepath.Join(binDir, binaryName)

	// Windows uses .zip, others use .tar.gz.
	var binURL string
	var opts []tool.Opt
	if runtime.GOOS == "windows" {
		binURL = fmt.Sprintf(
			"https://github.com/golangci/golangci-lint/releases/download/v%s/golangci-lint-%s-%s-%s.zip",
			version,
			version,
			runtime.GOOS,
			archName(),
		)
		opts = []tool.Opt{
			tool.WithDestinationDir(binDir),
			tool.WithUnzip(),
			tool.WithExtractFiles(binaryName),
			tool.WithSkipIfFileExists(binary),
			tool.WithSymlink(binary),
		}
	} else {
		binURL = fmt.Sprintf(
			"https://github.com/golangci/golangci-lint/releases/download/v%s/golangci-lint-%s-%s-%s.tar.gz",
			version,
			version,
			runtime.GOOS,
			archName(),
		)
		opts = []tool.Opt{
			tool.WithDestinationDir(binDir),
			tool.WithUntarGz(),
			tool.WithExtractFiles(binaryName),
			tool.WithSkipIfFileExists(binary),
			tool.WithSymlink(binary),
		}
	}

	return tool.FromRemote(ctx, binURL, opts...)
}

func archName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}
