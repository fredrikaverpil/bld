package tasks_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tasks"
	"github.com/fredrikaverpil/pocket/tasks/golang"
	"github.com/fredrikaverpil/pocket/tasks/lua"
	"github.com/fredrikaverpil/pocket/tasks/markdown"
)

func TestNew_CustomTasks(t *testing.T) {
	customTask := &pocket.Task{
		Name:  "my-custom-task",
		Usage: "a custom task for testing",
	}

	cfg := pocket.Config{
		Tasks: map[string][]*pocket.Task{
			".": {customTask},
		},
	}

	result := tasks.New(cfg, ".")

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
		Tasks: map[string][]*pocket.Task{
			".": {
				{Name: "deploy", Usage: "deploy the app"},
				{Name: "release", Usage: "create a release"},
			},
		},
	}

	result := tasks.New(cfg, ".")

	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 custom tasks, got %d", len(result.Tasks))
	}
}

func TestNew_GoTaskGroupConfigDriven(t *testing.T) {
	tests := []struct {
		name         string
		taskGroup    pocket.TaskGroup
		wantTasks    []string
		wantNotTasks []string
	}{
		{
			name: "all Go tasks enabled",
			taskGroup: golang.New(map[string]golang.Options{
				".": {},
			}),
			wantTasks:    []string{"go-format", "go-lint", "go-test", "go-vulncheck"},
			wantNotTasks: nil,
		},
		{
			name: "skip format excludes go-format",
			taskGroup: golang.New(map[string]golang.Options{
				".": {Skip: []string{"go-format"}},
			}),
			wantTasks:    []string{"go-lint", "go-test", "go-vulncheck"},
			wantNotTasks: []string{"go-format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := pocket.Config{
				TaskGroups: []pocket.TaskGroup{tt.taskGroup},
			}
			result := tasks.New(cfg, ".")

			// Build map of task names.
			taskNames := make(map[string]bool)
			for _, task := range result.TaskGroupTasks {
				taskNames[task.Name] = true
			}

			for _, want := range tt.wantTasks {
				if !taskNames[want] {
					t.Errorf("expected %q in tasks, but not found", want)
				}
			}

			for _, notWant := range tt.wantNotTasks {
				if taskNames[notWant] {
					t.Errorf("expected %q NOT in tasks, but found", notWant)
				}
			}
		})
	}
}

func TestNew_LuaTaskGroupConfigDriven(t *testing.T) {
	tests := []struct {
		name          string
		taskGroup     pocket.TaskGroup
		wantLuaFormat bool
	}{
		{
			name: "lua format enabled",
			taskGroup: lua.New(map[string]lua.Options{
				".": {},
			}),
			wantLuaFormat: true,
		},
		{
			name: "lua format skipped",
			taskGroup: lua.New(map[string]lua.Options{
				".": {Skip: []string{"lua-format"}},
			}),
			wantLuaFormat: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := pocket.Config{
				TaskGroups: []pocket.TaskGroup{tt.taskGroup},
			}
			result := tasks.New(cfg, ".")

			found := false
			for _, task := range result.TaskGroupTasks {
				if task.Name == "lua-format" {
					found = true
					break
				}
			}

			if found != tt.wantLuaFormat {
				t.Errorf("lua-format in tasks = %v, want %v", found, tt.wantLuaFormat)
			}
		})
	}
}

func TestNew_MarkdownTaskGroupConfigDriven(t *testing.T) {
	tests := []struct {
		name         string
		taskGroup    pocket.TaskGroup
		wantMdFormat bool
	}{
		{
			name: "markdown format enabled",
			taskGroup: markdown.New(map[string]markdown.Options{
				".": {},
			}),
			wantMdFormat: true,
		},
		{
			name: "markdown format skipped",
			taskGroup: markdown.New(map[string]markdown.Options{
				".": {Skip: []string{"md-format"}},
			}),
			wantMdFormat: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := pocket.Config{
				TaskGroups: []pocket.TaskGroup{tt.taskGroup},
			}
			result := tasks.New(cfg, ".")

			found := false
			for _, task := range result.TaskGroupTasks {
				if task.Name == "md-format" {
					found = true
					break
				}
			}

			if found != tt.wantMdFormat {
				t.Errorf("md-format in tasks = %v, want %v", found, tt.wantMdFormat)
			}
		})
	}
}

func TestNew_GenerateAlwaysPresent(t *testing.T) {
	// Even with empty config, generate should be present.
	result := tasks.New(pocket.Config{}, ".")

	if result.Generate == nil {
		t.Error("'generate' task should always be present")
	}
	if result.Generate.Name != "generate" {
		t.Errorf("expected generate task name, got %q", result.Generate.Name)
	}
}

func TestNew_NoTaskGroupsRegistered(t *testing.T) {
	result := tasks.New(pocket.Config{}, ".")

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

func TestNew_ContextFiltering(t *testing.T) {
	// Create a task group with modules in different contexts.
	goTaskGroup := golang.New(map[string]golang.Options{
		".":     {},
		"tests": {},
	})

	cfg := pocket.Config{
		TaskGroups: []pocket.TaskGroup{goTaskGroup},
		Tasks: map[string][]*pocket.Task{
			".":      {{Name: "root-task", Usage: "root only"}},
			"tests":  {{Name: "tests-task", Usage: "tests only"}},
			"deploy": {{Name: "deploy-task", Usage: "deploy only"}},
		},
	}

	t.Run("root context includes all", func(t *testing.T) {
		result := tasks.New(cfg, ".")

		// Should include root custom task.
		found := false
		for _, task := range result.Tasks {
			if task.Name == "root-task" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected root-task for root context")
		}

		// Task group tasks should be present.
		foundFormat := false
		for _, task := range result.TaskGroupTasks {
			if task.Name == "go-format" {
				foundFormat = true
				break
			}
		}
		if !foundFormat {
			t.Error("expected go-format for root context")
		}
	})

	t.Run("tests context filters to tests only", func(t *testing.T) {
		result := tasks.New(cfg, "tests")

		// Should include tests custom task.
		found := false
		for _, task := range result.Tasks {
			if task.Name == "tests-task" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected tests-task for tests context")
		}

		// Should NOT include root custom task.
		foundRoot := false
		for _, task := range result.Tasks {
			if task.Name == "root-task" {
				foundRoot = true
				break
			}
		}
		if foundRoot {
			t.Error("did not expect root-task for tests context")
		}
	})

	t.Run("deploy context has no task group tasks", func(t *testing.T) {
		result := tasks.New(cfg, "deploy")

		// Should include deploy custom task.
		found := false
		for _, task := range result.Tasks {
			if task.Name == "deploy-task" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected deploy-task for deploy context")
		}

		// Task group doesn't have a deploy module, so no go-format.
		foundFormat := false
		for _, task := range result.TaskGroupTasks {
			if task.Name == "go-format" {
				foundFormat = true
				break
			}
		}
		if foundFormat {
			t.Error("did not expect go-format for deploy context (task group has no deploy module)")
		}
	})
}

func TestAllTasks_ReturnsAllTasks(t *testing.T) {
	cfg := pocket.Config{
		TaskGroups: []pocket.TaskGroup{
			golang.New(map[string]golang.Options{".": {}}),
		},
		Tasks: map[string][]*pocket.Task{
			".": {{Name: "custom", Usage: "custom task"}},
		},
	}

	result := tasks.New(cfg, ".")
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
