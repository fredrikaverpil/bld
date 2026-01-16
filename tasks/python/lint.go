package python

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/ruff"
)

// LintOptions configures the py-lint task.
type LintOptions struct {
	RuffConfig string `arg:"ruff-config" usage:"path to ruff config file"`
	SkipFix    bool   `arg:"skip-fix"    usage:"don't auto-fix issues"`
}

// Lint lints Python files using ruff check with auto-fix enabled by default.
var Lint = pocket.Func("py-lint", "lint Python files", pocket.Serial(
	ruff.Install,
	lintCmd(),
)).With(LintOptions{})

func lintCmd() pocket.Runnable {
	return pocket.RunWith(ruff.Name, func(ctx context.Context) []string {
		opts := pocket.Options[LintOptions](ctx)
		configPath := opts.RuffConfig
		if configPath == "" {
			configPath, _ = pocket.ConfigPath(ctx, "ruff", ruff.Config)
		}

		args := []string{"check"}
		if pocket.Verbose(ctx) {
			args = append(args, "--verbose")
		}
		if !opts.SkipFix {
			args = append(args, "--fix")
		}
		if configPath != "" {
			args = append(args, "--config", configPath)
		}
		args = append(args, pocket.Path(ctx))

		return args
	})
}
