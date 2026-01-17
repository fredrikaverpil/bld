package github

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/fredrikaverpil/pocket"
)

// MatrixConfig configures GitHub Actions matrix generation.
type MatrixConfig struct {
	// DefaultPlatforms for all tasks. Default: ["ubuntu-latest"]
	DefaultPlatforms []string

	// TaskOverrides provides per-task platform configuration.
	TaskOverrides map[string]TaskOverride

	// ExcludeTasks removes tasks from the matrix entirely.
	ExcludeTasks []string

	// WindowsShell determines which shim to use on Windows.
	// Options: "cmd" (pok.cmd), "powershell" (pok.ps1)
	// Default: "cmd"
	WindowsShell string
}

// TaskOverride configures a single task in the matrix.
type TaskOverride struct {
	// Platforms overrides DefaultPlatforms for this task.
	// Empty means use DefaultPlatforms.
	Platforms []string
}

// DefaultMatrixConfig returns sensible defaults.
func DefaultMatrixConfig() MatrixConfig {
	return MatrixConfig{
		DefaultPlatforms: []string{"ubuntu-latest"},
		WindowsShell:     "cmd",
	}
}

// matrixEntry is a single entry in the GHA matrix.
type matrixEntry struct {
	Task string `json:"task"`
	OS   string `json:"os"`
	Shim string `json:"shim"`
}

// matrixOutput is the JSON structure for fromJson().
type matrixOutput struct {
	Include []matrixEntry `json:"include"`
}

// GenerateMatrix creates the GitHub Actions matrix JSON from tasks.
func GenerateMatrix(tasks []pocket.TaskInfo, cfg MatrixConfig) ([]byte, error) {
	if cfg.DefaultPlatforms == nil {
		cfg.DefaultPlatforms = []string{"ubuntu-latest"}
	}
	if cfg.WindowsShell == "" {
		cfg.WindowsShell = "cmd"
	}

	excludeSet := make(map[string]bool)
	for _, name := range cfg.ExcludeTasks {
		excludeSet[name] = true
	}

	entries := make([]matrixEntry, 0)
	for _, task := range tasks {
		// Skip hidden and excluded tasks
		if task.Hidden || excludeSet[task.Name] {
			continue
		}

		// Determine platforms for this task
		platforms := cfg.DefaultPlatforms
		if override, ok := cfg.TaskOverrides[task.Name]; ok && len(override.Platforms) > 0 {
			platforms = override.Platforms
		}

		// Create entry for each platform
		for _, platform := range platforms {
			entries = append(entries, matrixEntry{
				Task: task.Name,
				OS:   platform,
				Shim: shimForPlatform(platform, cfg.WindowsShell),
			})
		}
	}

	return json.Marshal(matrixOutput{Include: entries})
}

// shimForPlatform returns the appropriate shim command for the platform.
func shimForPlatform(platform, windowsShell string) string {
	if strings.Contains(platform, "windows") {
		switch windowsShell {
		case "powershell":
			return ".\\pok.ps1"
		default:
			return ".\\pok.cmd"
		}
	}
	return "./pok"
}

// MatrixTask creates the gha-matrix task.
// Users pass their AutoRun and MatrixConfig to generate the matrix.
//
// Example usage in .pocket/config.go:
//
//	var autoRun = pocket.Parallel(
//	    pocket.RunIn(golang.Tasks(), pocket.Detect(golang.Detect())),
//	)
//
//	var Config = pocket.Config{
//	    AutoRun: autoRun,
//	    ManualRun: []pocket.Runnable{
//	        github.MatrixTask(autoRun, github.MatrixConfig{
//	            DefaultPlatforms: []string{"ubuntu-latest", "macos-latest"},
//	            TaskOverrides: map[string]github.TaskOverride{
//	                "go-lint": {Platforms: []string{"ubuntu-latest"}},
//	            },
//	        }),
//	    },
//	}
func MatrixTask(autoRun pocket.Runnable, cfg MatrixConfig) *pocket.TaskDef {
	return pocket.Task("gha-matrix", "output GitHub Actions matrix JSON",
		matrixCmd(autoRun, cfg),
		pocket.AsSilent(),
	)
}

func matrixCmd(autoRun pocket.Runnable, cfg MatrixConfig) pocket.Runnable {
	return pocket.Do(func(ctx context.Context) error {
		tasks := pocket.CollectTasks(autoRun)
		data, err := GenerateMatrix(tasks, cfg)
		if err != nil {
			return err
		}
		pocket.Printf(ctx, "%s\n", data)
		return nil
	})
}
