package lua

import (
	"context"
	"fmt"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/stylua"
)

// LintOptions configures the lua-lint task.
type LintOptions struct {
	StyluaConfig string `arg:"stylua-config" usage:"path to stylua config file"`
}

// Lint checks Lua files using stylua --check.
var Lint = pocket.Func("lua-lint", "check Lua formatting", pocket.Serial(
	stylua.Install,
	lint,
)).With(LintOptions{})

func lint(ctx context.Context) error {
	opts := pocket.Options[LintOptions](ctx)
	configPath := opts.StyluaConfig
	if configPath == "" {
		var err error
		configPath, err = pocket.ConfigPath("stylua", stylua.Config)
		if err != nil {
			return fmt.Errorf("get stylua config: %w", err)
		}
	}

	absDir := pocket.FromGitRoot(pocket.Path(ctx))

	args := []string{"--check"}
	if configPath != "" {
		args = append(args, "-f", configPath)
	}
	args = append(args, absDir)

	return pocket.Exec(ctx, stylua.Name, args...)
}
