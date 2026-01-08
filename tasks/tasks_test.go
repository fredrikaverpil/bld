package tasks_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tasks"
	"github.com/fredrikaverpil/pocket/tasks/golang"
	"github.com/fredrikaverpil/pocket/tasks/markdown"
)

func TestNew_CustomTasks(t *testing.T) {
	customTask := &pocket.Task{
		Name:  "my-custom-task",
		Usage: "a custom task for testing",
	}

	cfg := pocket.Config{
		Tasks: []*pocket.Task{customTask},
	}

	result := tasks.New(cfg)

	// Verify custom task is registered.
	if len(result.Tasks) != 1 {
		t.Fatalf("expected 1 custom task, got %d", len(result.Tasks))
	}
	if result.Tasks[0].Name != "my-custom-task" {
		t.Errorf("expected custom task name 'my-custom-task', got %q", result.Tasks[0].Name)
	}
}

func TestNew_MultipleCustomTasks(t *testing.T) {
	cfg := pocket.Config{
		Tasks: []*pocket.Task{
			{Name: "deploy", Usage: "deploy the app"},
			{Name: "release", Usage: "create a release"},
		},
	}

	result := tasks.New(cfg)

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 custom tasks, got %d", len(result.Tasks))
	}
}

func TestNew_GoTaskGroup(t *testing.T) {
	cfg := pocket.Config{
		TaskGroups: []*pocket.TaskGroup{
			golang.NewTaskGroup(),
		},
	}
	result := tasks.New(cfg)

	// Check that Go tasks are present.
	taskNames := make(map[string]bool)
	for _, task := range result.TaskGroupTasks {
		taskNames[task.Name] = true
	}

	expected := []string{"go-format", "go-lint", "go-test", "go-vulncheck", "go-all"}
	for _, name := range expected {
		if !taskNames[name] {
			t.Errorf("expected %q in tasks, but not found", name)
		}
	}
}

func TestNew_MarkdownTaskGroup(t *testing.T) {
	cfg := pocket.Config{
		TaskGroups: []*pocket.TaskGroup{
			markdown.NewTaskGroup(),
		},
	}
	result := tasks.New(cfg)

	// Check that markdown tasks are present.
	var foundMdFormat bool
	for _, task := range result.TaskGroupTasks {
		if task.Name == "md-format" {
			foundMdFormat = true
			break
		}
	}

	if !foundMdFormat {
		t.Error("expected md-format in tasks")
	}
}

func TestNew_GenerateAlwaysPresent(t *testing.T) {
	// Even with empty config, generate should be present.
	result := tasks.New(pocket.Config{})

	if result.Generate == nil {
		t.Error("'generate' task should always be present")
	}
	if result.Generate.Name != "generate" {
		t.Errorf("expected generate task name, got %q", result.Generate.Name)
	}
}

func TestNew_NoTaskGroupsRegistered(t *testing.T) {
	result := tasks.New(pocket.Config{})

	// Should have Generate, All, Update, GitDiff defined.
	if result.Generate == nil {
		t.Error("Generate task should be defined")
	}
	if result.All == nil {
		t.Error("All task should be defined")
	}
	if result.Update == nil {
		t.Error("Update task should be defined")
	}
	if result.GitDiff == nil {
		t.Error("GitDiff task should be defined")
	}

	// No task group tasks should be registered.
	if len(result.TaskGroupTasks) != 0 {
		t.Errorf("expected 0 task group tasks, got %d", len(result.TaskGroupTasks))
	}
}

func TestAllTasks_ReturnsAllTasks(t *testing.T) {
	cfg := pocket.Config{
		TaskGroups: []*pocket.TaskGroup{
			golang.NewTaskGroup(),
		},
		Tasks: []*pocket.Task{
			{Name: "custom", Usage: "custom task"},
		},
	}

	result := tasks.New(cfg)
	allTasks := result.AllTasks()

	// Should include All, Generate, Update, GitDiff, task group tasks, and custom tasks.
	taskNames := make(map[string]bool)
	for _, task := range allTasks {
		taskNames[task.Name] = true
	}

	expected := []string{
		"all",
		"generate",
		"update",
		"git-diff",
		"custom",
		"go-format",
		"go-lint",
		"go-test",
		"go-vulncheck",
		"go-all",
	}
	for _, name := range expected {
		if !taskNames[name] {
			t.Errorf("expected %q in AllTasks(), but not found", name)
		}
	}
}

func TestDeps_ParallelExecution(t *testing.T) {
	var count atomic.Int32
	task1 := &pocket.Task{
		Name: "task1",
		Action: func(_ context.Context, _ map[string]string) error {
			count.Add(1)
			return nil
		},
	}
	task2 := &pocket.Task{
		Name: "task2",
		Action: func(_ context.Context, _ map[string]string) error {
			count.Add(1)
			return nil
		},
	}

	err := pocket.Deps(context.Background(), task1, task2)
	if err != nil {
		t.Fatalf("Deps failed: %v", err)
	}

	if count.Load() != 2 {
		t.Errorf("expected both tasks to run, got count=%d", count.Load())
	}
}

func TestSerialDeps_SequentialExecution(t *testing.T) {
	var order []string
	task1 := &pocket.Task{
		Name: "task1",
		Action: func(_ context.Context, _ map[string]string) error {
			order = append(order, "task1")
			return nil
		},
	}
	task2 := &pocket.Task{
		Name: "task2",
		Action: func(_ context.Context, _ map[string]string) error {
			order = append(order, "task2")
			return nil
		},
	}

	err := pocket.SerialDeps(context.Background(), task1, task2)
	if err != nil {
		t.Fatalf("SerialDeps failed: %v", err)
	}

	if len(order) != 2 || order[0] != "task1" || order[1] != "task2" {
		t.Errorf("expected [task1, task2], got %v", order)
	}
}

func TestTask_RunsOnlyOnce(t *testing.T) {
	runCount := 0
	task := &pocket.Task{
		Name: "once",
		Action: func(_ context.Context, _ map[string]string) error {
			runCount++
			return nil
		},
	}

	ctx := context.Background()
	_ = pocket.Run(ctx, task)
	_ = pocket.Run(ctx, task)
	_ = pocket.Run(ctx, task)

	if runCount != 1 {
		t.Errorf("expected task to run once, but ran %d times", runCount)
	}
}
