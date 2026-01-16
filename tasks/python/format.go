package python

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/ruff"
)

// FormatOptions configures the py-format task.
type FormatOptions struct {
	RuffConfig string `arg:"ruff-config" usage:"path to ruff config file"`
}

// Format formats Python files using ruff format.
var Format = pocket.Func("py-format", "format Python files", pocket.Serial(
	ruff.Install,
	formatCmd(),
)).With(FormatOptions{})

func formatCmd() pocket.Runnable {
	return pocket.RunWith(ruff.Name, func(ctx context.Context) []string {
		opts := pocket.Options[FormatOptions](ctx)
		configPath := opts.RuffConfig
		if configPath == "" {
			configPath, _ = pocket.ConfigPath(ctx, "ruff", ruff.Config)
		}

		args := []string{"format"}
		if pocket.Verbose(ctx) {
			args = append(args, "--verbose")
		}
		if configPath != "" {
			args = append(args, "--config", configPath)
		}
		args = append(args, pocket.Path(ctx))

		return args
	})
}
