package pocket

import "context"

// TaskGroup holds a collection of tasks with execution and detection semantics.
// Use NewTaskGroup to create a group, then chain methods to configure it.
//
// Example (simple, parallel execution):
//
//	func Tasks() pocket.Runnable {
//	    return pocket.NewTaskGroup(FormatTask(), LintTask()).
//	        DetectBy(pocket.DetectByFile("go.mod"))
//	}
//
// Example (custom execution order):
//
//	func Tasks() pocket.Runnable {
//	    format, lint := FormatTask(), LintTask()
//	    test, vulncheck := TestTask(), VulncheckTask()
//
//	    return pocket.NewTaskGroup(format, lint, test, vulncheck).
//	        RunWith(pocket.Serial(format, lint, pocket.Parallel(test, vulncheck))).
//	        DetectBy(pocket.DetectByFile("go.mod"))
//	}
type TaskGroup struct {
	tasks    []*Task
	runner   Runnable
	detectFn func() []string
}

// NewTaskGroup creates a new task group with the given tasks.
// By default, tasks run in parallel. Use RunWith to customize execution order.
func NewTaskGroup(tasks ...*Task) *TaskGroup {
	return &TaskGroup{
		tasks: tasks,
	}
}

// RunWith sets a custom execution order for the task group.
// If not called, tasks run in parallel by default.
//
// Use Serial() and Parallel() to compose the execution order:
//
//	group.RunWith(pocket.Serial(format, lint, pocket.Parallel(test, vulncheck)))
func (g *TaskGroup) RunWith(r Runnable) *TaskGroup {
	g.runner = r
	return g
}

// DetectBy sets the detection function for auto-detection.
// This makes the TaskGroup implement Detectable.
//
// Example:
//
//	DetectBy(pocket.DetectByFile("go.mod"))
//	DetectBy(pocket.DetectByExtension(".py"))
//	DetectBy(func() []string { return []string{"."} })
func (g *TaskGroup) DetectBy(fn func() []string) *TaskGroup {
	g.detectFn = fn
	return g
}

// Run executes the task group.
// If RunWith was called, uses the custom Runnable.
// Otherwise, runs all tasks in parallel.
func (g *TaskGroup) Run(ctx context.Context, exec *Execution) error {
	if g.runner != nil {
		return g.runner.Run(ctx, exec)
	}
	// Default: run all tasks in parallel.
	runnables := make([]Runnable, len(g.tasks))
	for i, t := range g.tasks {
		runnables[i] = t
	}
	return Parallel(runnables...).Run(ctx, exec)
}

// Tasks returns all tasks in the group.
func (g *TaskGroup) Tasks() []*Task {
	return g.tasks
}

// DefaultDetect returns the detection function.
// Implements the Detectable interface.
func (g *TaskGroup) DefaultDetect() func() []string {
	return g.detectFn
}

// DetectByFile is a convenience method that detects directories containing
// any of the specified files (e.g., "go.mod", "pyproject.toml").
func (g *TaskGroup) DetectByFile(filenames ...string) *TaskGroup {
	return g.DetectBy(func() []string { return DetectByFile(filenames...) })
}

// DetectByExtension is a convenience method that detects directories containing
// files with any of the specified extensions (e.g., ".py", ".md").
func (g *TaskGroup) DetectByExtension(extensions ...string) *TaskGroup {
	return g.DetectBy(func() []string { return DetectByExtension(extensions...) })
}
