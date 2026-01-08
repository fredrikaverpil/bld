package pocket

import (
	"context"
	"testing"
)

func TestConfig_WithDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		config       Config
		wantShimName string
		wantPosix    bool
	}{
		{
			name:         "empty config gets default shim name and posix",
			config:       Config{},
			wantShimName: "pok",
			wantPosix:    true,
		},
		{
			name: "custom shim name is preserved",
			config: Config{
				Shim: &ShimConfig{Name: "build", Posix: true},
			},
			wantShimName: "build",
			wantPosix:    true,
		},
		{
			name: "empty shim name gets default",
			config: Config{
				Shim: &ShimConfig{Posix: true},
			},
			wantShimName: "pok",
			wantPosix:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.config.WithDefaults()

			if got.Shim == nil {
				t.Fatal("WithDefaults().Shim is nil")
			}
			if got.Shim.Name != tt.wantShimName {
				t.Errorf("WithDefaults().Shim.Name = %q, want %q", got.Shim.Name, tt.wantShimName)
			}
		})
	}
}

func TestNewTaskGroup(t *testing.T) {
	t.Parallel()

	task1 := &Task{
		Name:  "test-format",
		Usage: "format test files",
		Action: func(_ context.Context, _ map[string]string) error {
			return nil
		},
	}
	task2 := &Task{
		Name:  "test-lint",
		Usage: "lint test files",
		Action: func(_ context.Context, _ map[string]string) error {
			return nil
		},
	}

	tg := NewTaskGroup("test", task1, task2)

	// Check name.
	if tg.Name() != "test" {
		t.Errorf("Name() = %q, want %q", tg.Name(), "test")
	}

	// Check tasks include originals plus orchestrator.
	tasks := tg.Tasks()
	if len(tasks) != 3 {
		t.Errorf("Tasks() length = %d, want 3", len(tasks))
	}

	// Check orchestrator task exists.
	var foundOrchestrator bool
	for _, task := range tasks {
		if task.Name == "test-all" {
			foundOrchestrator = true
			if !task.Hidden {
				t.Error("orchestrator task should be hidden")
			}
		}
	}
	if !foundOrchestrator {
		t.Error("orchestrator task not found")
	}
}

func TestConfig_TaskGroups(t *testing.T) {
	t.Parallel()

	tg1 := NewTaskGroup("go",
		&Task{Name: "go-format", Usage: "format Go"},
		&Task{Name: "go-lint", Usage: "lint Go"},
	)
	tg2 := NewTaskGroup("python",
		&Task{Name: "py-format", Usage: "format Python"},
	)

	cfg := Config{
		TaskGroups: []*TaskGroup{tg1, tg2},
	}

	if len(cfg.TaskGroups) != 2 {
		t.Errorf("TaskGroups length = %d, want 2", len(cfg.TaskGroups))
	}
	if cfg.TaskGroups[0].Name() != "go" {
		t.Errorf("TaskGroups[0].Name() = %q, want %q", cfg.TaskGroups[0].Name(), "go")
	}
	if cfg.TaskGroups[1].Name() != "python" {
		t.Errorf("TaskGroups[1].Name() = %q, want %q", cfg.TaskGroups[1].Name(), "python")
	}
}

func TestConfig_Tasks(t *testing.T) {
	t.Parallel()

	task1 := &Task{Name: "deploy", Usage: "deploy app"}
	task2 := &Task{Name: "release", Usage: "release app"}

	cfg := Config{
		Tasks: []*Task{task1, task2},
	}

	if len(cfg.Tasks) != 2 {
		t.Errorf("Tasks length = %d, want 2", len(cfg.Tasks))
	}
}
