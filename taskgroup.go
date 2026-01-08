package pocket

import "context"

// TaskGroup is a named collection of related tasks.
type TaskGroup struct {
	name  string
	tasks []*Task
}

// Name returns the task group name (e.g., "go", "lua", "markdown").
func (g *TaskGroup) Name() string { return g.name }

// Tasks returns all tasks in this group, including the orchestrator task.
func (g *TaskGroup) Tasks() []*Task { return g.tasks }

// NewTaskGroup creates a task group with the given name and tasks.
// An orchestrator task (name-all) is automatically added that runs all tasks serially.
func NewTaskGroup(name string, tasks ...*Task) *TaskGroup {
	// Create orchestrator task that runs all tasks serially.
	allTask := &Task{
		Name:   name + "-all",
		Usage:  "run all " + name + " tasks",
		Hidden: true,
		Action: func(ctx context.Context, _ map[string]string) error {
			return SerialDeps(ctx, tasks...)
		},
	}

	return &TaskGroup{
		name:  name,
		tasks: append(tasks, allTask),
	}
}
