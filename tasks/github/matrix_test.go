package github

import (
	"encoding/json"
	"testing"

	"github.com/fredrikaverpil/pocket"
)

func TestGenerateMatrix_Default(t *testing.T) {
	tasks := []pocket.TaskInfo{
		{Name: "lint", Usage: "lint code"},
		{Name: "test", Usage: "run tests"},
	}

	cfg := DefaultMatrixConfig()
	data, err := GenerateMatrix(tasks, cfg)
	if err != nil {
		t.Fatalf("GenerateMatrix() failed: %v", err)
	}

	var output matrixOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Default is ubuntu-latest only, so 2 tasks = 2 entries
	if len(output.Include) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(output.Include))
	}

	for _, entry := range output.Include {
		if entry.OS != "ubuntu-latest" {
			t.Errorf("expected os 'ubuntu-latest', got %q", entry.OS)
		}
		if entry.Shim != "./pok" {
			t.Errorf("expected shim './pok', got %q", entry.Shim)
		}
	}
}

func TestGenerateMatrix_MultiplePlatforms(t *testing.T) {
	tasks := []pocket.TaskInfo{
		{Name: "test", Usage: "run tests"},
	}

	cfg := MatrixConfig{
		DefaultPlatforms: []string{"ubuntu-latest", "macos-latest", "windows-latest"},
	}
	data, err := GenerateMatrix(tasks, cfg)
	if err != nil {
		t.Fatalf("GenerateMatrix() failed: %v", err)
	}

	var output matrixOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// 1 task * 3 platforms = 3 entries
	if len(output.Include) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(output.Include))
	}

	// Check that we have all platforms
	platforms := make(map[string]bool)
	for _, entry := range output.Include {
		platforms[entry.OS] = true
	}

	expected := []string{"ubuntu-latest", "macos-latest", "windows-latest"}
	for _, p := range expected {
		if !platforms[p] {
			t.Errorf("expected platform %q in output", p)
		}
	}
}

func TestGenerateMatrix_TaskOverrides(t *testing.T) {
	tasks := []pocket.TaskInfo{
		{Name: "lint", Usage: "lint code"},
		{Name: "test", Usage: "run tests"},
	}

	cfg := MatrixConfig{
		DefaultPlatforms: []string{"ubuntu-latest", "macos-latest", "windows-latest"},
		TaskOverrides: map[string]TaskOverride{
			"lint": {Platforms: []string{"ubuntu-latest"}}, // lint only on ubuntu
		},
	}
	data, err := GenerateMatrix(tasks, cfg)
	if err != nil {
		t.Fatalf("GenerateMatrix() failed: %v", err)
	}

	var output matrixOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// lint: 1 platform, test: 3 platforms = 4 entries
	if len(output.Include) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(output.Include))
	}

	// Count lint entries
	lintCount := 0
	for _, entry := range output.Include {
		if entry.Task == "lint" {
			lintCount++
			if entry.OS != "ubuntu-latest" {
				t.Errorf("lint should only run on ubuntu-latest, got %q", entry.OS)
			}
		}
	}
	if lintCount != 1 {
		t.Errorf("expected 1 lint entry, got %d", lintCount)
	}
}

func TestGenerateMatrix_ExcludeTasks(t *testing.T) {
	tasks := []pocket.TaskInfo{
		{Name: "format", Usage: "format code"},
		{Name: "lint", Usage: "lint code"},
		{Name: "test", Usage: "run tests"},
	}

	cfg := MatrixConfig{
		DefaultPlatforms: []string{"ubuntu-latest"},
		ExcludeTasks:     []string{"format"},
	}
	data, err := GenerateMatrix(tasks, cfg)
	if err != nil {
		t.Fatalf("GenerateMatrix() failed: %v", err)
	}

	var output matrixOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// 3 tasks - 1 excluded = 2 entries
	if len(output.Include) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(output.Include))
	}

	for _, entry := range output.Include {
		if entry.Task == "format" {
			t.Error("format task should be excluded")
		}
	}
}

func TestGenerateMatrix_HiddenTasksExcluded(t *testing.T) {
	tasks := []pocket.TaskInfo{
		{Name: "lint", Usage: "lint code", Hidden: false},
		{Name: "install:tool", Usage: "install tool", Hidden: true},
	}

	cfg := DefaultMatrixConfig()
	data, err := GenerateMatrix(tasks, cfg)
	if err != nil {
		t.Fatalf("GenerateMatrix() failed: %v", err)
	}

	var output matrixOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Only non-hidden task
	if len(output.Include) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(output.Include))
	}
	if output.Include[0].Task != "lint" {
		t.Errorf("expected task 'lint', got %q", output.Include[0].Task)
	}
}

func TestGenerateMatrix_WindowsShim(t *testing.T) {
	tasks := []pocket.TaskInfo{
		{Name: "test", Usage: "run tests"},
	}

	tests := []struct {
		name         string
		windowsShell string
		wantShim     string
	}{
		{"default", "", ".\\pok.cmd"},
		{"cmd", "cmd", ".\\pok.cmd"},
		{"powershell", "powershell", ".\\pok.ps1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := MatrixConfig{
				DefaultPlatforms: []string{"windows-latest"},
				WindowsShell:     tt.windowsShell,
			}
			data, err := GenerateMatrix(tasks, cfg)
			if err != nil {
				t.Fatalf("GenerateMatrix() failed: %v", err)
			}

			var output matrixOutput
			if err := json.Unmarshal(data, &output); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if len(output.Include) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(output.Include))
			}
			if output.Include[0].Shim != tt.wantShim {
				t.Errorf("expected shim %q, got %q", tt.wantShim, output.Include[0].Shim)
			}
		})
	}
}

func TestGenerateMatrix_ShimForPlatform(t *testing.T) {
	tests := []struct {
		platform     string
		windowsShell string
		want         string
	}{
		{"ubuntu-latest", "cmd", "./pok"},
		{"macos-latest", "cmd", "./pok"},
		{"windows-latest", "cmd", ".\\pok.cmd"},
		{"windows-2022", "cmd", ".\\pok.cmd"},
		{"windows-latest", "powershell", ".\\pok.ps1"},
		{"ubuntu-22.04", "cmd", "./pok"},
	}

	for _, tt := range tests {
		t.Run(tt.platform+"_"+tt.windowsShell, func(t *testing.T) {
			got := shimForPlatform(tt.platform, tt.windowsShell)
			if got != tt.want {
				t.Errorf("shimForPlatform(%q, %q) = %q, want %q",
					tt.platform, tt.windowsShell, got, tt.want)
			}
		})
	}
}

func TestGenerateMatrix_EmptyTasks(t *testing.T) {
	cfg := DefaultMatrixConfig()
	data, err := GenerateMatrix(nil, cfg)
	if err != nil {
		t.Fatalf("GenerateMatrix() failed: %v", err)
	}

	var output matrixOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(output.Include) != 0 {
		t.Errorf("expected 0 entries, got %d", len(output.Include))
	}

	// Verify the JSON structure is correct for GHA
	expected := `{"include":[]}`
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestDefaultMatrixConfig(t *testing.T) {
	cfg := DefaultMatrixConfig()

	if len(cfg.DefaultPlatforms) != 1 {
		t.Errorf("expected 1 default platform, got %d", len(cfg.DefaultPlatforms))
	}
	if cfg.DefaultPlatforms[0] != "ubuntu-latest" {
		t.Errorf("expected default platform 'ubuntu-latest', got %q", cfg.DefaultPlatforms[0])
	}
	if cfg.WindowsShell != "cmd" {
		t.Errorf("expected default WindowsShell 'cmd', got %q", cfg.WindowsShell)
	}
}
