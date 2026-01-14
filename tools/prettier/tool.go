// Package prettier provides prettier (code formatter) tool integration.
// prettier is installed via bun into a local node_modules directory.
package prettier

import (
	"context"
	_ "embed"
	"os"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/bun"
)

const name = "prettier"

// renovate: datasource=npm depName=prettier
const version = "3.7.4"

//go:embed prettierrc.json
var defaultConfig []byte

//go:embed prettierignore
var defaultIgnore []byte

// Tool is the prettier tool.
//
// Example usage in a task action:
//
//	prettier.Tool.Exec(ctx, tc, "--write", "**/*.md")
var Tool = pocket.NewTool(name, version, install).
	WithConfig(pocket.ToolConfig{
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
	})

// DefaultIgnore returns the default .prettierignore content.
// Tasks can use this to write an ignore file if needed.
func DefaultIgnore() []byte {
	return defaultIgnore
}

func install(ctx context.Context, tc *pocket.TaskContext) error {
	installDir := pocket.FromToolsDir(name, version)
	binary := bun.BinaryPath(installDir, name)

	// Skip if already installed.
	if _, err := os.Stat(binary); err == nil {
		_, err := pocket.CreateSymlink(binary)
		return err
	}

	// Create install directory.
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return err
	}

	// Install prettier using bun.
	if err := bun.Install(ctx, tc, installDir, name+"@"+version); err != nil {
		return err
	}

	// Create symlink to .pocket/bin/.
	_, err := pocket.CreateSymlink(binary)
	return err
}
