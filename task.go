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

	skipRules []skipRule // internal: task skip rules (not visible to Actions in other packages)
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

const runContextKey contextKey = 0

// skipRule defines a rule for skipping a task.
type skipRule struct {
	taskName string
	paths    []string // empty = skip everywhere, non-empty = skip only in these paths
}

// getRunContext returns the RunContext from context.
func getRunContext(ctx context.Context) *RunContext {
	if rc, ok := ctx.Value(runContextKey).(*RunContext); ok {
		return rc
	}
	return &RunContext{Cwd: "."}
}

// withRunContext returns a context with the RunContext set.
func withRunContext(ctx context.Context, rc *RunContext) context.Context {
	return context.WithValue(ctx, runContextKey, rc)
}

// withSkipRules returns a new context with skip rules added.
func withSkipRules(ctx context.Context, rules []skipRule) context.Context {
	rc := getRunContext(ctx)
	return withRunContext(ctx, &RunContext{
		Verbose:   rc.Verbose,
		Cwd:       rc.Cwd,
		skipRules: rules,
	})
}

// isSkipped returns true if the task should be skipped for the given path.
func isSkipped(ctx context.Context, name, path string) bool {
	rc := getRunContext(ctx)
	for _, rule := range rc.skipRules {
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
		base := getRunContext(ctx)
		// Determine paths, defaulting to cwd if not set.
		paths := t.paths
		if len(paths) == 0 {
			paths = []string{base.Cwd}
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
		// Build RunContext for this task.
		rc := &RunContext{
			Args:    t.args,
			Paths:   filteredPaths,
			Cwd:     base.Cwd,
			Verbose: base.Verbose,
		}
		t.err = t.Action(ctx, rc)
	})
	return t.err
}

// Tasks returns this task as a slice (implements Runnable interface).
func (t *Task) Tasks() []*Task {
	return []*Task{t}
}
