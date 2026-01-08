// Package python provides Python-related build tasks using ruff and mypy.
package python

import (
	"context"
	"fmt"
	"slices"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/mypy"
	"github.com/fredrikaverpil/pocket/tools/ruff"
)

const name = "python"

// Options defines options for a Python module within a task group.
type Options struct {
	// Skip lists task names to skip (e.g., "format", "lint", "typecheck").
	Skip []string
	// Only lists task names to run (empty = run all).
	// If non-empty, only these tasks run (Skip is ignored).
	Only []string

	// Task-specific options
	Format    FormatOptions
	Lint      LintOptions
	Typecheck TypecheckOptions
}

// ShouldRun returns true if the given task should run based on Skip/Only options.
func (o Options) ShouldRun(task string) bool {
	if len(o.Only) > 0 {
		return slices.Contains(o.Only, task)
	}
	return !slices.Contains(o.Skip, task)
}

// FormatOptions defines options for the format task.
type FormatOptions struct {
	// ConfigFile overrides the default ruff config file.
	ConfigFile string
}

// LintOptions defines options for the lint task.
type LintOptions struct {
	// ConfigFile overrides the default ruff config file.
	ConfigFile string
}

// TypecheckOptions defines options for the typecheck task.
type TypecheckOptions struct {
	// placeholder for future options
}

// New creates a Python task group with the given module configuration.
func New(modules map[string]Options) pocket.TaskGroup {
	return &taskGroup{modules: modules}
}

type taskGroup struct {
	modules map[string]Options
}

func (tg *taskGroup) Name() string { return name }

func (tg *taskGroup) Modules() map[string]pocket.ModuleConfig {
	modules := make(map[string]pocket.ModuleConfig, len(tg.modules))
	for path, opts := range tg.modules {
		modules[path] = opts
	}
	return modules
}

func (tg *taskGroup) ForContext(context string) pocket.TaskGroup {
	if context == "." {
		return tg
	}
	if opts, ok := tg.modules[context]; ok {
		return &taskGroup{modules: map[string]Options{context: opts}}
	}
	return nil
}

func (tg *taskGroup) Tasks(cfg pocket.Config) []*pocket.Task {
	_ = cfg.WithDefaults()
	var tasks []*pocket.Task

	var formatTask, lintTask, typecheckTask *pocket.Task

	if mods := tg.modulesFor("format"); len(mods) > 0 {
		formatTask = FormatTask(mods)
		tasks = append(tasks, formatTask)
	}

	if mods := tg.modulesFor("lint"); len(mods) > 0 {
		lintTask = LintTask(mods)
		tasks = append(tasks, lintTask)
	}

	if mods := tg.modulesFor("typecheck"); len(mods) > 0 {
		typecheckTask = TypecheckTask(mods)
		tasks = append(tasks, typecheckTask)
	}

	// Create orchestrator task that controls execution order.
	allTask := &pocket.Task{
		Name:   "py-all",
		Usage:  "run all Python tasks",
		Hidden: true,
		Action: func(ctx context.Context) error {
			// Format and lint run serially (they modify files).
			if err := pocket.SerialDeps(ctx, formatTask, lintTask); err != nil {
				return err
			}
			// Typecheck runs after format/lint.
			return pocket.SerialDeps(ctx, typecheckTask)
		},
	}
	tasks = append(tasks, allTask)

	return tasks
}

// modulesFor returns modules with their task-specific options for a given task.
func (tg *taskGroup) modulesFor(task string) map[string]Options {
	result := make(map[string]Options)
	for path, opts := range tg.modules {
		if opts.ShouldRun(task) {
			result[path] = opts
		}
	}
	return result
}

// FormatTask returns a task that formats Python files using ruff format.
// The modules map specifies which directories to format and their options.
func FormatTask(modules map[string]Options) *pocket.Task {
	return &pocket.Task{
		Name:  "py-format",
		Usage: "format Python files",
		Action: func(ctx context.Context) error {
			for mod, opts := range modules {
				configPath := opts.Format.ConfigFile
				if configPath == "" {
					var err error
					configPath, err = ruff.ConfigPath()
					if err != nil {
						return fmt.Errorf("get ruff config: %w", err)
					}
				}
				if err := ruff.Run(ctx, "format", "--config", configPath, mod); err != nil {
					return fmt.Errorf("ruff format failed in %s: %w", mod, err)
				}
			}
			return nil
		},
	}
}

// LintTask returns a task that lints Python files using ruff check.
// The modules map specifies which directories to lint and their options.
func LintTask(modules map[string]Options) *pocket.Task {
	return &pocket.Task{
		Name:  "py-lint",
		Usage: "lint Python files",
		Action: func(ctx context.Context) error {
			for mod, opts := range modules {
				configPath := opts.Lint.ConfigFile
				if configPath == "" {
					var err error
					configPath, err = ruff.ConfigPath()
					if err != nil {
						return fmt.Errorf("get ruff config: %w", err)
					}
				}
				if err := ruff.Run(ctx, "check", "--config", configPath, mod); err != nil {
					return fmt.Errorf("ruff check failed in %s: %w", mod, err)
				}
			}
			return nil
		},
	}
}

// TypecheckTask returns a task that type-checks Python files using mypy.
// The modules map specifies which directories to check and their options.
func TypecheckTask(modules map[string]Options) *pocket.Task {
	return &pocket.Task{
		Name:  "py-typecheck",
		Usage: "type-check Python files",
		Action: func(ctx context.Context) error {
			for mod := range modules {
				if err := mypy.Run(ctx, mod); err != nil {
					return fmt.Errorf("mypy failed in %s: %w", mod, err)
				}
			}
			return nil
		},
	}
}
