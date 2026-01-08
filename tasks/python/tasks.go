// Package python provides Python-related build tasks using ruff and mypy.
package python

import (
	"slices"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/mypy"
	"github.com/fredrikaverpil/pocket/tools/ruff"
	"github.com/goyek/goyek/v3"
)

const name = "python"

// Config defines the configuration for the Python task group.
type Config struct {
	// Modules maps context paths to module options.
	// The key is the path relative to the git root (use "." for root).
	Modules map[string]Options
}

// Options defines options for a Python module.
type Options struct {
	// Skip lists task names to skip (e.g., "format", "lint", "typecheck").
	Skip []string
	// Only lists task names to run (empty = run all).
	// If non-empty, only these tasks run (Skip is ignored).
	Only []string
}

// ShouldRun returns true if the given task should run based on Skip/Only options.
func (o Options) ShouldRun(task string) bool {
	if len(o.Only) > 0 {
		return slices.Contains(o.Only, task)
	}
	return !slices.Contains(o.Skip, task)
}

// New creates a Python task group with the given configuration.
func New(cfg Config) pocket.TaskGroup {
	return &taskGroup{config: cfg}
}

type taskGroup struct {
	config Config
}

func (tg *taskGroup) Name() string { return name }

func (tg *taskGroup) Modules() map[string]pocket.ModuleConfig {
	modules := make(map[string]pocket.ModuleConfig, len(tg.config.Modules))
	for path, opts := range tg.config.Modules {
		modules[path] = opts
	}
	return modules
}

func (tg *taskGroup) ForContext(context string) pocket.TaskGroup {
	if context == "." {
		return tg
	}
	if opts, ok := tg.config.Modules[context]; ok {
		return &taskGroup{config: Config{
			Modules: map[string]Options{context: opts},
		}}
	}
	return nil
}

func (tg *taskGroup) Tasks(cfg pocket.Config) []*goyek.DefinedTask {
	_ = cfg.WithDefaults()
	var tasks []*goyek.DefinedTask

	if modules := pocket.ModulesFor(tg, "format"); len(modules) > 0 {
		tasks = append(tasks, goyek.Define(FormatTask(modules)))
	}

	if modules := pocket.ModulesFor(tg, "lint"); len(modules) > 0 {
		tasks = append(tasks, goyek.Define(LintTask(modules)))
	}

	if modules := pocket.ModulesFor(tg, "typecheck"); len(modules) > 0 {
		tasks = append(tasks, goyek.Define(TypecheckTask(modules)))
	}

	return tasks
}

// FormatTask returns a task that formats Python files using ruff format.
func FormatTask(modules []string) goyek.Task {
	return goyek.Task{
		Name:  "py-format",
		Usage: "format Python files",
		Action: func(a *goyek.A) {
			configPath, err := ruff.ConfigPath()
			if err != nil {
				a.Errorf("get ruff config: %v", err)
				return
			}
			for _, mod := range modules {
				if err := ruff.Run(a.Context(), "format", "--config", configPath, mod); err != nil {
					a.Errorf("ruff format failed in %s: %v", mod, err)
				}
			}
		},
	}
}

// LintTask returns a task that lints Python files using ruff check.
func LintTask(modules []string) goyek.Task {
	return goyek.Task{
		Name:  "py-lint",
		Usage: "lint Python files",
		Action: func(a *goyek.A) {
			configPath, err := ruff.ConfigPath()
			if err != nil {
				a.Errorf("get ruff config: %v", err)
				return
			}
			for _, mod := range modules {
				if err := ruff.Run(a.Context(), "check", "--config", configPath, mod); err != nil {
					a.Errorf("ruff check failed in %s: %v", mod, err)
				}
			}
		},
	}
}

// TypecheckTask returns a task that type-checks Python files using mypy.
func TypecheckTask(modules []string) goyek.Task {
	return goyek.Task{
		Name:  "py-typecheck",
		Usage: "type-check Python files",
		Action: func(a *goyek.A) {
			for _, mod := range modules {
				if err := mypy.Run(a.Context(), mod); err != nil {
					a.Errorf("mypy failed in %s: %v", mod, err)
				}
			}
		},
	}
}
