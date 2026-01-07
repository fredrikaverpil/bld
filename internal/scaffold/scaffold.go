// Package scaffold provides generation of .bld/ scaffold files.
package scaffold

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fredrikaverpil/bld"
	"github.com/fredrikaverpil/bld/internal/shim"
	"github.com/fredrikaverpil/bld/internal/workflows"
)

//go:embed main.go.tmpl
var MainTemplate []byte

//go:embed config.go.tmpl
var ConfigTemplate []byte

//go:embed gitignore.tmpl
var GitignoreTemplate []byte

// GenerateAll regenerates all generated files.
// Creates one-time files (config.go, .gitignore) if they don't exist.
// Always regenerates main.go and shim.
// If cfg is not nil, also generates workflows.
func GenerateAll(cfg *bld.Config) error {
	bldDir := filepath.Join(bld.FromGitRoot(), bld.DirName)

	// Ensure .bld/ exists
	if err := os.MkdirAll(bldDir, 0o755); err != nil {
		return fmt.Errorf("creating .bld/: %w", err)
	}

	// Create config.go if not exists (user-editable, never overwritten)
	configPath := filepath.Join(bldDir, "config.go")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, ConfigTemplate, 0o644); err != nil {
			return fmt.Errorf("writing config.go: %w", err)
		}
	}

	// Create .gitignore if not exists
	gitignorePath := filepath.Join(bldDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, GitignoreTemplate, 0o644); err != nil {
			return fmt.Errorf("writing .gitignore: %w", err)
		}
	}

	// Always regenerate main.go
	if err := GenerateMain(); err != nil {
		return err
	}

	// Always regenerate shim(s).
	// Use provided config or a minimal default for initial scaffold.
	shimCfg := bld.Config{}
	if cfg != nil {
		shimCfg = *cfg
	}
	if err := shim.Generate(shimCfg); err != nil {
		return err
	}

	// Generate workflows if config provided.
	if cfg != nil {
		if err := workflows.Generate(*cfg); err != nil {
			return err
		}
	}

	return nil
}

// GenerateMain creates or updates .bld/main.go from the template.
func GenerateMain() error {
	mainPath := filepath.Join(bld.FromGitRoot(), bld.DirName, "main.go")
	if err := os.WriteFile(mainPath, MainTemplate, 0o644); err != nil {
		return fmt.Errorf("writing .bld/main.go: %w", err)
	}
	return nil
}
