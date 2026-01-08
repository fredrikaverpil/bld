// Package tasks provides the unified task entry point for pocket.
// It automatically creates tasks based on the provided Config.
package tasks

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tasks/generate"
	"github.com/fredrikaverpil/pocket/tasks/gitdiff"
	"github.com/fredrikaverpil/pocket/tasks/update"
)

// Tasks holds all registered tasks based on the Config.
type Tasks struct {
	// All runs all configured tasks.
	All *pocket.Task

	// Generate regenerates all generated files.
	Generate *pocket.Task

	// Update updates pocket and regenerates files.
	Update *pocket.Task

	// GitDiff fails if there are uncommitted changes.
	GitDiff *pocket.Task

	// Tasks holds standalone tasks registered in config.
	Tasks []*pocket.Task

	// TaskGroupTasks holds all tasks from registered task groups.
	TaskGroupTasks []*pocket.Task
}

// New creates tasks based on the provided Config.
func New(cfg pocket.Config) *Tasks {
	cfg = cfg.WithDefaults()
	t := &Tasks{}

	// Generate runs first - other tasks may need generated files.
	t.Generate = generate.Task(cfg)

	// Update is standalone (not part of "all").
	t.Update = update.Task(cfg)

	// GitDiff is available as a standalone task.
	t.GitDiff = gitdiff.Task()

	// Collect orchestrator tasks from task groups (hidden tasks that control order).
	var orchestratorTasks []*pocket.Task

	// Create tasks from task groups.
	for _, tg := range cfg.TaskGroups {
		for _, task := range tg.Tasks() {
			t.TaskGroupTasks = append(t.TaskGroupTasks, task)
			if task.Hidden {
				// Hidden tasks are orchestrators that control execution order.
				orchestratorTasks = append(orchestratorTasks, task)
			}
		}
	}

	// Add standalone tasks from config.
	t.Tasks = cfg.Tasks

	// Create the "all" task that runs everything, then checks for uncommitted changes.
	t.All = &pocket.Task{
		Name:  "all",
		Usage: "run all tasks",
		Action: func(ctx context.Context, _ map[string]string) error {
			// Generate first.
			if err := pocket.SerialDeps(ctx, t.Generate); err != nil {
				return err
			}

			// Run all task group orchestrators (each handles its own ordering).
			if err := pocket.SerialDeps(ctx, orchestratorTasks...); err != nil {
				return err
			}

			// Run custom user tasks in parallel.
			if err := pocket.Deps(ctx, t.Tasks...); err != nil {
				return err
			}

			// Git diff at the end (if not skipped).
			if !cfg.SkipGitDiff {
				return pocket.SerialDeps(ctx, t.GitDiff)
			}
			return nil
		},
	}

	return t
}

// AllTasks returns all tasks including the "all" task.
// This is used by the CLI to register all available tasks.
func (t *Tasks) AllTasks() []*pocket.Task {
	tasks := []*pocket.Task{t.All, t.Generate, t.Update, t.GitDiff}
	tasks = append(tasks, t.TaskGroupTasks...)
	tasks = append(tasks, t.Tasks...)
	return tasks
}
