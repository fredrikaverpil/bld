package tool

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fredrikaverpil/pocket"
)

// ConfigSpec describes how to find or create a tool's configuration file.
type ConfigSpec struct {
	// ToolName is used to create the fallback config directory.
	ToolName string
	// UserConfigNames are filenames to search for in the repo root.
	// Checked in order; first match wins.
	UserConfigNames []string
	// DefaultConfigName is the filename for the bundled default config.
	DefaultConfigName string
	// DefaultConfig is the bundled default configuration content.
	DefaultConfig []byte
}

// Path returns the path to the tool's config file.
// It checks for user config files in the repo root first,
// then falls back to writing the bundled default config.
func (c ConfigSpec) Path() (string, error) {
	// Check for user config in repo root.
	for _, configName := range c.UserConfigNames {
		repoConfig := pocket.FromGitRoot(configName)
		if _, err := os.Stat(repoConfig); err == nil {
			return repoConfig, nil
		}
	}

	// Write bundled config to .pocket/tools/<name>/<default-name>.
	configDir := pocket.FromToolsDir(c.ToolName)
	configPath := filepath.Join(configDir, c.DefaultConfigName)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return "", fmt.Errorf("create config dir: %w", err)
		}
		if err := os.WriteFile(configPath, c.DefaultConfig, 0o644); err != nil {
			return "", fmt.Errorf("write default config: %w", err)
		}
	}

	return configPath, nil
}
