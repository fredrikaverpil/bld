package pocket

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
)

// ArgDef defines an argument that a task accepts.
type ArgDef struct {
	Name    string // argument name (used as key in key=value)
	Usage   string // description for help output
	Default string // default value if not provided
}

// RunContext provides runtime context to Actions.
type RunContext struct {
	Args    map[string]string // CLI arguments (key=value pairs)
	Paths   []string          // resolved paths for this task (from Paths wrapper)
	Cwd     string            // current working directory (relative to git root)
	Verbose bool              // verbose mode enabled
}

// ForEachPath executes fn for each path in the context.
// This is a convenience helper for the common pattern of iterating over paths.
func (rc *RunContext) ForEachPath(fn func(dir string) error) error {
	for _, dir := range rc.Paths {
		if err := fn(dir); err != nil {
			return err
		}
	}
	return nil
}

// Task represents a runnable task.
type Task struct {
	Name    string
	Usage   string
	Args    []ArgDef // declared arguments this task accepts
	Action  func(ctx context.Context, rc *RunContext) error
	Hidden  bool
	Builtin bool // true for core tasks like generate, update, git-diff

	// once ensures the task runs exactly once per execution.
	once sync.Once
	// err stores the result of the task execution.
	err error
	// args stores the parsed arguments for this execution.
	args map[string]string
	// paths stores the resolved paths for this execution.
	paths []string
}

// contextKey is the type for context keys used by this package.
type contextKey int

const runConfigKey contextKey = 0

// skipRule defines a rule for skipping a task.
type skipRule struct {
	taskName string
	paths    []string // empty = skip everywhere, non-empty = skip only in these paths
}

// runConfig holds runtime configuration passed through context.
// This consolidates all run-time state into a single context value.
type runConfig struct {
	verbose   bool
	cwd       string     // relative to git root, defaults to "."
	skipRules []skipRule // task skip rules from PathFilter.Skip()
}

// runConfigFromContext returns the run configuration from context.
func runConfigFromContext(ctx context.Context) *runConfig {
	if cfg, ok := ctx.Value(runConfigKey).(*runConfig); ok {
		return cfg
	}
	return &runConfig{cwd: "."}
}

// withRunConfig returns a context with the run configuration set.
func withRunConfig(ctx context.Context, cfg *runConfig) context.Context {
	return context.WithValue(ctx, runConfigKey, cfg)
}

// withSkipRules returns a new context with skip rules added to the run config.
func withSkipRules(ctx context.Context, rules []skipRule) context.Context {
	cfg := runConfigFromContext(ctx)
	newCfg := &runConfig{
		verbose:   cfg.verbose,
		cwd:       cfg.cwd,
		skipRules: rules,
	}
	return withRunConfig(ctx, newCfg)
}

// isSkipped returns true if the task should be skipped for the given path.
func isSkipped(ctx context.Context, name, path string) bool {
	cfg := runConfigFromContext(ctx)
	for _, rule := range cfg.skipRules {
		if rule.taskName != name {
			continue
		}
		if len(rule.paths) == 0 {
			return true // global skip
		}
		if slices.Contains(rule.paths, path) {
			return true // path-specific skip
		}
	}
	return false
}

// SetArgs sets the arguments for this task execution.
// Arguments are merged with defaults defined in Args.
func (t *Task) SetArgs(args map[string]string) {
	t.args = make(map[string]string)
	// Apply defaults first.
	for _, arg := range t.Args {
		if arg.Default != "" {
			t.args[arg.Name] = arg.Default
		}
	}
	// Override with provided args.
	maps.Copy(t.args, args)
}

// SetPaths sets the resolved paths for this task execution.
func (t *Task) SetPaths(paths []string) {
	t.paths = paths
}

// Run executes the task's action exactly once.
// Implements the Runnable interface.
// Skip rules from context are checked:
// - Global skip (no paths): task doesn't run at all
// - Path-specific skip: those paths are filtered from execution.
func (t *Task) Run(ctx context.Context) error {
	// Check for global skip (rule with no paths).
	if isSkipped(ctx, t.Name, "") {
		return nil
	}
	t.once.Do(func() {
		if t.Action == nil {
			return
		}
		cfg := runConfigFromContext(ctx)
		// Determine paths, defaulting to cwd if not set.
		paths := t.paths
		if len(paths) == 0 {
			paths = []string{cfg.cwd}
		}
		// Filter out paths that should be skipped.
		var filteredPaths []string
		var skippedPaths []string
		for _, p := range paths {
			if !isSkipped(ctx, t.Name, p) {
				filteredPaths = append(filteredPaths, p)
			} else {
				skippedPaths = append(skippedPaths, p)
			}
		}
		// If all paths are skipped, don't run.
		if len(filteredPaths) == 0 {
			fmt.Fprintf(Stdout(ctx), "=== %s (skipped)\n", t.Name)
			return
		}
		// Show task name with any skipped paths.
		if len(skippedPaths) > 0 {
			fmt.Fprintf(Stdout(ctx), "=== %s (skipped in: %s)\n", t.Name, strings.Join(skippedPaths, ", "))
		} else {
			fmt.Fprintf(Stdout(ctx), "=== %s\n", t.Name)
		}
		// Ensure args map exists (may be nil if SetArgs wasn't called).
		if t.args == nil {
			t.SetArgs(nil)
		}
		// Build RunContext from run config.
		rc := &RunContext{
			Args:    t.args,
			Paths:   filteredPaths,
			Cwd:     cfg.cwd,
			Verbose: cfg.verbose,
		}
		t.err = t.Action(ctx, rc)
	})
	return t.err
}

// Tasks returns this task as a slice (implements Runnable interface).
func (t *Task) Tasks() []*Task {
	return []*Task{t}
}
