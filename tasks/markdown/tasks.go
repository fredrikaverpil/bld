// Package markdown provides Markdown-related build tasks.
package markdown

import (
	"context"
	"fmt"
	"os"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/mdformat"
	"github.com/fredrikaverpil/pocket/tools/prettier"
)

// Tasks returns a Runnable that executes all Markdown tasks.
// Runs from repository root since markdown files are typically scattered.
// Use pocket.Paths(markdown.Tasks()).DetectBy(markdown.Detect()) to enable path filtering.
func Tasks() pocket.Runnable {
	return FormatTask()
}

// Detect returns a detection function that finds Markdown projects.
// It returns the repository root since markdown files are typically scattered.
//
// Usage:
//
//	pocket.Paths(markdown.Tasks()).DetectBy(markdown.Detect())
func Detect() func() []string {
	return func() []string {
		return []string{"."}
	}
}

// FormatTask returns a task that formats Markdown files using prettier.
func FormatTask() *pocket.Task {
	return pocket.NewTask("md-format", "format Markdown files", formatAction)
}

// formatAction is the action for the md-format task.
func formatAction(ctx context.Context, tc *pocket.TaskContext) error {
	args := []string{"--write"}
	if configPath, _ := prettier.Tool.ConfigPath(); configPath != "" {
		args = append(args, "--config", configPath)
	}
	if ignorePath, err := ensureIgnoreFile(); err == nil {
		args = append(args, "--ignore-path", ignorePath)
	}
	args = append(args, "**/*.md")

	if err := prettier.Tool.Exec(ctx, tc, args...); err != nil {
		return fmt.Errorf("prettier failed: %w", err)
	}
	return nil
}

// ensureIgnoreFile ensures a .prettierignore file exists at git root.
// Returns the path to the ignore file.
func ensureIgnoreFile() (string, error) {
	ignoreFile := pocket.FromGitRoot(".prettierignore")

	// Check if file already exists (user-defined or previously written).
	if _, err := os.Stat(ignoreFile); err == nil {
		return ignoreFile, nil
	}

	// Write default ignore file to git root.
	if err := os.WriteFile(ignoreFile, prettier.DefaultIgnore(), 0o644); err != nil {
		return "", err
	}
	return ignoreFile, nil
}

// MdformatTask returns a task that formats Markdown files using mdformat.
// This task is not included in Tasks() but can be used directly if preferred.
func MdformatTask() *pocket.Task {
	return pocket.NewTask("mdformat", "format Markdown files with mdformat", mdformatAction)
}

// mdformatAction is the action for the mdformat task.
func mdformatAction(ctx context.Context, tc *pocket.TaskContext) error {
	absDir := pocket.FromGitRoot(tc.Path)
	if err := mdformat.Tool.Exec(ctx, tc, "--number", "--wrap", "80", absDir); err != nil {
		return fmt.Errorf("mdformat failed in %s: %w", tc.Path, err)
	}
	return nil
}
