package python

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

// SyncOptions configures the py-sync task.
type SyncOptions struct {
	PythonVersion string `arg:"python" usage:"Python version to use (e.g., 3.12)"`
}

// Sync installs Python dependencies using uv sync.
var Sync = pocket.Func("py-sync", "install Python dependencies", pocket.Serial(
	uv.Install,
	uvSync,
)).With(SyncOptions{})

func uvSync(ctx context.Context) error {
	opts := pocket.Options[SyncOptions](ctx)

	args := []string{"sync"}
	if pocket.Verbose(ctx) {
		args = append(args, "--verbose")
	}
	if opts.PythonVersion != "" {
		args = append(args, "--python", opts.PythonVersion)
	}

	return pocket.ExecIn(ctx, pocket.FromGitRoot(pocket.Path(ctx)), uv.Name, args...)
}
