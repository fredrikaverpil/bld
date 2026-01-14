// Package golangcilint provides golangci-lint integration.
// This is an "action tool" - it provides Install, Lint, Fmt, and Exec.
package golangcilint

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
const Version = "v2.0.2"

// Install ensures golangci-lint is available.
// This is a hidden dependency used by Lint, Fmt, and Exec.
var Install = pocket.Func("install:golangci-lint", "install golangci-lint", install).Hidden()

func install(ctx context.Context) error {
	pocket.Printf(ctx, "Installing golangci-lint %s...\n", Version)
	return pocket.InstallGo(ctx, "github.com/golangci/golangci-lint/v2/cmd/golangci-lint", Version)
}

// Lint runs golangci-lint linter.
// This is visible in CLI and can be used directly in config.
var Lint = pocket.Func("golangci-lint", "run golangci-lint", lint)

func lint(ctx context.Context) error {
	pocket.Serial(ctx, Install)

	args := []string{"run"}
	if configPath, err := pocket.ConfigPath("golangci-lint", Config); err == nil && configPath != "" {
		args = append(args, "-c", configPath)
	}
	args = append(args, "./...")

	return pocket.Exec(ctx, "golangci-lint", args...)
}

// Fmt runs golangci-lint formatter.
var Fmt = pocket.Func("golangci-lint-fmt", "format with golangci-lint", fmtFunc)

func fmtFunc(ctx context.Context) error {
	pocket.Serial(ctx, Install)

	args := []string{"fmt"}
	if configPath, err := pocket.ConfigPath("golangci-lint", Config); err == nil && configPath != "" {
		args = append(args, "-c", configPath)
	}
	args = append(args, "./...")

	return pocket.Exec(ctx, "golangci-lint", args...)
}

// Exec runs golangci-lint with the given arguments.
// This is for programmatic use when you need full control.
//
// Example:
//
//	golangcilint.Exec(ctx, "run", "--fix", "./...")
func Exec(ctx context.Context, args ...string) error {
	pocket.Serial(ctx, Install)
	return pocket.Exec(ctx, "golangci-lint", args...)
}

// Config for golangci-lint configuration file lookup.
var Config = pocket.ToolConfig{
	UserFiles: []string{
		".golangci.yml",
		".golangci.yaml",
		".golangci.toml",
		".golangci.json",
	},
	DefaultFile: "", // No default - use golangci-lint defaults
}
