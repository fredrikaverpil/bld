// Package prettier provides prettier (code formatter) integration.
// This is an "action tool" - it provides Install, Format, and Exec.
package prettier

import (
	"context"
	_ "embed"
	"os"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/bun"
)

// renovate: datasource=npm depName=prettier
const Version = "3.7.4"

//go:embed prettierrc.json
var defaultConfig []byte

//go:embed prettierignore
var defaultIgnore []byte

// Install ensures prettier (via bun) is available.
// This is a hidden dependency used by Format and Exec.
var Install = pocket.Func("install:prettier", "ensure prettier is available", install).Hidden()

func install(ctx context.Context) error {
	pocket.Serial(ctx, bun.Install)
	return nil
}

// Format formats files using prettier.
// This is visible in CLI and can be used directly in config.
//
// Example in config:
//
//	pocket.Paths(prettier.Format).In("docs")
var Format = pocket.Func("prettier", "format files with prettier", format)

// FormatOptions configures prettier formatting.
type FormatOptions struct {
	Patterns []string // file patterns to format (default: all supported files)
	Check    bool     // check only, don't write (--check flag)
}

func format(ctx context.Context) error {
	pocket.Serial(ctx, Install)

	opts := pocket.Options[FormatOptions](ctx)

	args := []string{}
	if opts.Check {
		args = append(args, "--check")
	} else {
		args = append(args, "--write")
	}

	// Add config if available
	if configPath, err := pocket.ConfigPath("prettier", Config); err == nil && configPath != "" {
		args = append(args, "--config", configPath)
	}

	// Add ignore file if available
	if ignorePath, err := EnsureIgnoreFile(); err == nil {
		args = append(args, "--ignore-path", ignorePath)
	}

	// Add patterns or default
	if len(opts.Patterns) > 0 {
		args = append(args, opts.Patterns...)
	} else {
		args = append(args, ".")
	}

	return Exec(ctx, args...)
}

// Exec runs prettier with the given arguments.
// This is for programmatic use when you need full control over arguments.
//
// Example:
//
//	prettier.Exec(ctx, "--write", "**/*.md")
func Exec(ctx context.Context, args ...string) error {
	pocket.Serial(ctx, Install)

	allArgs := append([]string{"prettier@" + Version}, args...)
	return pocket.Exec(ctx, "bunx", allArgs...)
}

// Config for prettier configuration file lookup.
var Config = pocket.ToolConfig{
	UserFiles: []string{
		".prettierrc",
		".prettierrc.json",
		".prettierrc.yaml",
		".prettierrc.yml",
		"prettier.config.js",
		"prettier.config.mjs",
	},
	DefaultFile: ".prettierrc",
	DefaultData: defaultConfig,
}

// DefaultIgnore returns the default .prettierignore content.
func DefaultIgnore() []byte {
	return defaultIgnore
}

// EnsureIgnoreFile ensures a .prettierignore file exists at git root.
func EnsureIgnoreFile() (string, error) {
	ignoreFile := pocket.FromGitRoot(".prettierignore")

	if _, err := os.Stat(ignoreFile); err == nil {
		return ignoreFile, nil
	}

	if err := os.WriteFile(ignoreFile, defaultIgnore, 0o644); err != nil {
		return "", err
	}
	return ignoreFile, nil
}
