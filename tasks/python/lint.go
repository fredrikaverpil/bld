package python

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

// LintOptions configures the py-lint task.
type LintOptions struct {
	PythonVersion string `arg:"python"   usage:"Python version (for target-version inference)"`
	SkipFix       bool   `arg:"skip-fix" usage:"don't auto-fix issues"`
}

// Lint lints Python files using ruff check with auto-fix enabled by default.
// Requires ruff as a project dependency in pyproject.toml.
var Lint = pocket.Task("py-lint", "lint Python files",
	pocket.Serial(uv.Install, lintSyncCmd(), lintCmd()),
	pocket.Opts(LintOptions{}),
)

func lintSyncCmd() pocket.Runnable {
	return pocket.Do(func(ctx context.Context) error {
		opts := pocket.Options[LintOptions](ctx)
		return uv.Sync(ctx, opts.PythonVersion, true)
	})
}

func lintCmd() pocket.Runnable {
	return pocket.Do(func(ctx context.Context) error {
		opts := pocket.Options[LintOptions](ctx)

		args := []string{
			"check",
			"--exclude", ".pocket", // Exclude pocket-managed directories
		}
		if pocket.Verbose(ctx) {
			args = append(args, "--verbose")
		}
		if !opts.SkipFix {
			args = append(args, "--fix")
		}
		if opts.PythonVersion != "" {
			args = append(args, "--target-version", pythonVersionToRuff(opts.PythonVersion))
		}
		args = append(args, pocket.Path(ctx))

		return uv.Run(ctx, opts.PythonVersion, "ruff", args...)
	})
}
