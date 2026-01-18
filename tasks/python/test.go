package python

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

// TestOptions configures the py-test task.
type TestOptions struct {
	PythonVersion string `arg:"python"        usage:"Python version to use (e.g., 3.9)"`
	SkipCoverage  bool   `arg:"skip-coverage" usage:"disable coverage generation"`
}

// Test runs Python tests using pytest with coverage by default.
// Requires pytest and coverage as project dependencies in pyproject.toml.
var Test = pocket.Task("py-test", "run Python tests",
	pocket.Serial(uv.Install, testSyncCmd(), testCmd()),
	pocket.Opts(TestOptions{}),
)

func testSyncCmd() pocket.Runnable {
	return pocket.Do(func(ctx context.Context) error {
		opts := pocket.Options[TestOptions](ctx)
		return uv.Sync(ctx, opts.PythonVersion, true)
	})
}

func testCmd() pocket.Runnable {
	return pocket.Do(func(ctx context.Context) error {
		opts := pocket.Options[TestOptions](ctx)

		if opts.SkipCoverage {
			// Run pytest directly without coverage
			args := []string{}
			if pocket.Verbose(ctx) {
				args = append(args, "-vv")
			}
			return uv.Run(ctx, opts.PythonVersion, "pytest", args...)
		}

		// Run with coverage: coverage run --parallel-mode -m pytest
		// --parallel-mode creates .coverage.<hostname>.<pid> files to avoid conflicts
		// when running multiple test processes in parallel (e.g., TestMatrix)
		args := []string{"run", "--parallel-mode", "-m", "pytest"}
		if pocket.Verbose(ctx) {
			args = append(args, "-vv")
		}
		if err := uv.Run(ctx, opts.PythonVersion, "coverage", args...); err != nil {
			return err
		}

		// Combine parallel coverage files before reporting
		if err := uv.Run(ctx, opts.PythonVersion, "coverage", "combine"); err != nil {
			// Ignore error if no parallel files to combine (single run)
			pocket.Printf(ctx, "Note: coverage combine skipped (may be single run)\n")
		}

		// Show coverage report
		if err := uv.Run(ctx, opts.PythonVersion, "coverage", "report"); err != nil {
			return err
		}

		// Generate HTML report
		return uv.Run(ctx, opts.PythonVersion, "coverage", "html")
	})
}
