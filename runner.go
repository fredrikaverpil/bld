package pocket

import (
	"context"
	"fmt"
)

// Run executes the given task.
// If the task has already been run, it returns the cached result.
func Run(ctx context.Context, task *Task) error {
	if task == nil {
		return fmt.Errorf("cannot run nil task")
	}
	return task.run(ctx)
}
