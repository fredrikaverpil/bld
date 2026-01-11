package pocket

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
)

// TaskAction is the function signature for task actions.
// ctx carries cancellation signals and deadlines.
// rc provides task-specific data: paths, options, output writers.
type TaskAction func(ctx context.Context, rc *RunContext) error

// RunContext provides all runtime state for task execution.
// It is passed explicitly through the Runnable chain.
type RunContext struct {
	// Public fields for task actions
	Paths   []string // resolved paths for this task (from Paths wrapper)
	Verbose bool     // verbose mode enabled
	Out     *Output  // output writers for stdout/stderr

	// Internal composition
	state         *executionState // shared across execution tree
	setup         *taskSetup      // accumulated during tree traversal
	parsedOptions any             // per-task: typed options, access via GetOptions[T](rc)
}

// NewRunContext creates a RunContext for a new execution.
func NewRunContext(out *Output, verbose bool, cwd string) *RunContext {
	return &RunContext{
		Out:     out,
		Verbose: verbose,
		state:   newExecutionState(cwd, verbose),
		setup:   newTaskSetup(),
	}
}

// CWD returns the current working directory relative to git root.
func (rc *RunContext) CWD() string {
	return rc.state.cwd
}

// withOutput returns a copy with different output (for parallel buffering).
// Shares the same execution tracking.
func (rc *RunContext) withOutput(out *Output) *RunContext {
	cp := *rc
	cp.Out = out
	// Share execution (tracks what's done across all children)
	// Share maps (they're set up front, not modified during execution)
	return &cp
}

// withSkipRules returns a copy with additional skip rules.
func (rc *RunContext) withSkipRules(rules []skipRule) *RunContext {
	cp := *rc
	cp.setup = rc.setup.withSkipRules(rules)
	return &cp
}

// setTaskPaths sets resolved paths for a task.
func (rc *RunContext) setTaskPaths(taskName string, paths []string) {
	rc.setup.paths[taskName] = paths
}

// SetTaskArgs sets CLI arguments for a task. This is used by the CLI
// to pass parsed command-line arguments to the task.
func (rc *RunContext) SetTaskArgs(taskName string, args map[string]string) {
	rc.setup.args[taskName] = args
}

// ForEachPath executes fn for each path in the context.
// This is a convenience helper for the common pattern of iterating over paths.
// Iteration stops early if the context is cancelled (e.g., another parallel task failed).
func (rc *RunContext) ForEachPath(ctx context.Context, fn func(dir string) error) error {
	for _, dir := range rc.Paths {
		// Check for cancellation before each iteration.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := fn(dir); err != nil {
			return err
		}
	}
	return nil
}

// isSkipped checks if a task should be skipped for a given path.
func (rc *RunContext) isSkipped(taskName, path string) bool {
	for _, rule := range rc.setup.skipRules {
		if rule.taskName != taskName {
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

// shouldSkipGlobally checks if a task should be skipped entirely (global skip rule).
func (rc *RunContext) shouldSkipGlobally(taskName string) bool {
	return rc.isSkipped(taskName, "")
}

// resolveAndFilterPaths returns the paths for a task after applying skip filters.
// Returns the paths to run and the paths that were skipped.
func (rc *RunContext) resolveAndFilterPaths(taskName string) (paths, skipped []string) {
	// Determine paths, defaulting to cwd if not set.
	all := rc.setup.paths[taskName]
	if len(all) == 0 {
		all = []string{rc.state.cwd}
	}

	// Filter out paths that should be skipped.
	for _, p := range all {
		if !rc.isSkipped(taskName, p) {
			paths = append(paths, p)
		} else {
			skipped = append(skipped, p)
		}
	}
	return paths, skipped
}

// printTaskHeader writes the task execution header to output.
func (rc *RunContext) printTaskHeader(taskName string, skippedPaths []string) {
	if len(skippedPaths) > 0 {
		fmt.Fprintf(rc.Out.Stdout, "=== %s (skipped in: %s)\n", taskName, strings.Join(skippedPaths, ", "))
	} else {
		fmt.Fprintf(rc.Out.Stdout, "=== %s\n", taskName)
	}
}

// buildTaskContext creates a task-specific RunContext with the given paths and options.
func (rc *RunContext) buildTaskContext(paths []string, opts any) *RunContext {
	return &RunContext{
		Paths:         paths,
		Verbose:       rc.Verbose,
		Out:           rc.Out,
		state:         rc.state, // shared
		setup:         rc.setup, // shared
		parsedOptions: opts,
	}
}

// execution tracks which tasks have run in a single execution.
// This is shared across the entire Runnable tree.
type execution struct {
	mu     sync.Mutex
	done   map[string]bool
	errors map[string]error
}

func newExecution() *execution {
	return &execution{
		done:   make(map[string]bool),
		errors: make(map[string]error),
	}
}

func (e *execution) markDone(name string, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.done[name] = true
	if err != nil {
		e.errors[name] = err
	}
}

func (e *execution) isDone(name string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.done[name] {
		return true, e.errors[name]
	}
	return false, nil
}

// skipRule defines a rule for skipping a task.
type skipRule struct {
	taskName string
	paths    []string // empty = skip everywhere, non-empty = skip only in these paths
}

// Task represents an immutable task definition.
//
// Create tasks using NewTask:
//
//	pocket.NewTask("my-task", "description", func(ctx context.Context, rc *pocket.RunContext) error {
//	    return nil
//	}).WithOptions(MyOptions{})
type Task struct {
	// Public fields (read-only after creation)
	Name    string
	Usage   string
	Options TaskOptions // typed options struct for CLI parsing (see args.go)
	Action  TaskAction  // function to execute when task runs
	Hidden  bool        // hide from CLI shim
	Builtin bool        // true for core tasks like generate, update, git-diff
}

// TaskName returns the task's CLI command name.
func (t *Task) TaskName() string {
	return t.Name
}

// NewTask creates an immutable task definition.
// Name is the CLI command name (e.g., "go-format").
// Usage is the help text shown in CLI.
// Action is the function executed when the task runs.
//
// Example:
//
//	pocket.NewTask("deploy", "deploy to environment", func(ctx context.Context, rc *pocket.RunContext) error {
//	    opts := pocket.GetOptions[DeployOptions](rc)
//	    return deploy(ctx, opts.Env)
//	}).WithOptions(DeployOptions{Env: "staging"})
func NewTask(name, usage string, action TaskAction) *Task {
	if name == "" {
		panic("pocket.NewTask: name is required")
	}
	if usage == "" {
		panic("pocket.NewTask: usage is required")
	}
	if action == nil {
		panic("pocket.NewTask: action is required")
	}
	return &Task{
		Name:   name,
		Usage:  usage,
		Action: action,
	}
}

// WithOptions returns a new Task with typed options for CLI flag parsing.
// Options must be a struct with exported fields of type bool, string, or int.
// Use struct tags to customize: `usage:"help text"` and `arg:"flag-name"`.
//
// Example:
//
//	type DeployOptions struct {
//	    Env    string `usage:"target environment"`
//	    DryRun bool   `usage:"print without executing"`
//	}
//
//	NewTask("deploy", "deploy app", deployAction).
//	    WithOptions(DeployOptions{Env: "staging"})
func (t *Task) WithOptions(opts any) *Task {
	if opts != nil {
		if _, err := inspectArgs(opts); err != nil {
			panic(fmt.Sprintf("pocket.Task.WithOptions: %v", err))
		}
	}
	cp := *t
	cp.Options = opts
	return &cp
}

// AsHidden returns a new Task marked as hidden from CLI help output.
// Hidden tasks can still be run directly by name.
func (t *Task) AsHidden() *Task {
	cp := *t
	cp.Hidden = true
	return &cp
}

// AsBuiltin returns a new Task marked as a built-in task.
// This is used internally for core tasks like generate, update, git-diff.
func (t *Task) AsBuiltin() *Task {
	cp := *t
	cp.Builtin = true
	return &cp
}

// Run executes the task's action exactly once per execution.
// Implements the Runnable interface.
// Skip rules from RunContext are checked:
// - Global skip (no paths): task doesn't run at all
// - Path-specific skip: those paths are filtered from execution.
func (t *Task) Run(ctx context.Context, rc *RunContext) error {
	dedup := rc.state.dedup

	// Check if already done in this execution.
	if done, err := dedup.isDone(t.Name); done {
		return err
	}

	// Check for global skip.
	if rc.shouldSkipGlobally(t.Name) {
		dedup.markDone(t.Name, nil)
		return nil
	}

	// Resolve paths and filter skipped ones.
	paths, skipped := rc.resolveAndFilterPaths(t.Name)
	if len(paths) == 0 {
		fmt.Fprintf(rc.Out.Stdout, "=== %s (skipped)\n", t.Name)
		dedup.markDone(t.Name, nil)
		return nil
	}

	// Print task header.
	rc.printTaskHeader(t.Name, skipped)

	// Parse typed options.
	opts, err := parseOptionsFromCLI(t.Options, rc.setup.args[t.Name])
	if err != nil {
		dedup.markDone(t.Name, fmt.Errorf("parse options: %w", err))
		return err
	}

	// Build task-specific RunContext and run the action.
	taskRC := rc.buildTaskContext(paths, opts)
	err = t.Action(ctx, taskRC)
	dedup.markDone(t.Name, err)
	return err
}

// Tasks returns this task as a slice (implements Runnable interface).
func (t *Task) Tasks() []*Task {
	return []*Task{t}
}
