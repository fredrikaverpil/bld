package pocket

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Task represents a runnable task.
type Task struct {
	Name   string
	Usage  string
	Action func(ctx context.Context) error
	Hidden bool

	// once ensures the task runs exactly once per execution.
	once sync.Once
	// err stores the result of the task execution.
	err error
}

// contextKey is the type for context keys used by this package.
type contextKey int

const (
	// verboseKey is the context key for verbose mode.
	verboseKey contextKey = iota
)

// WithVerbose returns a context with verbose mode set.
func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, verboseKey, verbose)
}

// IsVerbose returns true if verbose mode is enabled in the context.
func IsVerbose(ctx context.Context) bool {
	v, _ := ctx.Value(verboseKey).(bool)
	return v
}

// run executes the task's action exactly once.
func (t *Task) run(ctx context.Context) error {
	t.once.Do(func() {
		if t.Action == nil {
			return
		}
		// Always show task name for progress feedback.
		fmt.Printf("=== %s\n", t.Name)
		t.err = t.Action(ctx)
	})
	return t.err
}

// Deps runs the given tasks in parallel and waits for all to complete.
// Each task runs at most once regardless of how many times Deps is called.
// If any task fails, Deps returns the first error encountered.
func Deps(ctx context.Context, tasks ...*Task) error {
	if len(tasks) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, task := range tasks {
		if task == nil {
			continue
		}
		g.Go(func() error {
			return task.run(ctx)
		})
	}
	return g.Wait()
}

// SerialDeps runs the given tasks sequentially in order.
// Each task runs at most once regardless of how many times SerialDeps is called.
// If any task fails, SerialDeps returns immediately with the error.
func SerialDeps(ctx context.Context, tasks ...*Task) error {
	for _, task := range tasks {
		if task == nil {
			continue
		}
		if err := task.run(ctx); err != nil {
			return err
		}
	}
	return nil
}
