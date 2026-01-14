package pocket

import "context"

// RunConfig is the main entry point for running a pocket configuration.
// It parses CLI flags, discovers functions, and runs the appropriate ones.
//
// Example usage in .pocket/main.go:
//
//	func main() {
//	    pocket.RunConfig(Config)
//	}
func RunConfig(cfg Config) {
	cfg = cfg.WithDefaults()

	// Collect all functions and path mappings from AutoRun.
	var allFuncs []*FuncDef
	pathMappings := make(map[string]*PathFilter)
	autoRunNames := make(map[string]bool)

	if cfg.AutoRun != nil {
		allFuncs = cfg.AutoRun.funcs()
		pathMappings = CollectPathMappings(cfg.AutoRun)
		for _, f := range allFuncs {
			autoRunNames[f.name] = true
		}
	}

	// Create an "all" function that runs the entire AutoRun tree.
	var allFunc *FuncDef
	if cfg.AutoRun != nil {
		allFunc = Func("all", "run all tasks", func(ctx context.Context) error {
			return cfg.AutoRun.run(ctx)
		})
	}

	// Add manual run functions (if any - ManualRun is []Runnable in old Config).
	for _, r := range cfg.ManualRun {
		allFuncs = append(allFuncs, r.funcs()...)
	}

	// Call the CLI main function.
	Main(allFuncs, allFunc, nil, pathMappings, autoRunNames)
}

// RunConfig2 is the main entry point for running a pocket v2 configuration.
// It supports the new Cmd-based manual run.
//
// Example usage in .pocket/main.go:
//
//	func main() {
//	    pocket.RunConfig2(Config)
//	}
func RunConfig2(cfg Config2) {
	cfg = cfg.WithDefaults2()

	// Collect all functions and path mappings from AutoRun.
	var allFuncs []*FuncDef
	pathMappings := make(map[string]*PathFilter)
	autoRunNames := make(map[string]bool)

	if cfg.AutoRun != nil {
		allFuncs = cfg.AutoRun.funcs()
		pathMappings = CollectPathMappings(cfg.AutoRun)
		for _, f := range allFuncs {
			autoRunNames[f.name] = true
		}
	}

	// Create an "all" function that runs the entire AutoRun tree.
	var allFunc *FuncDef
	if cfg.AutoRun != nil {
		allFunc = Func("all", "run all tasks", func(ctx context.Context) error {
			return cfg.AutoRun.run(ctx)
		})
	}

	// Call the CLI main function with commands.
	Main(allFuncs, allFunc, cfg.ManualRun, pathMappings, autoRunNames)
}
