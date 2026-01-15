// Package github provides GitHub-related tasks.
package github

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fredrikaverpil/pocket"
)

//go:embed workflows/*.yml
var workflowFiles embed.FS

// WorkflowsOptions configures which workflows to bootstrap.
type WorkflowsOptions struct {
	PR      bool `arg:"pr"      usage:"include PR validation workflow"`
	Release bool `arg:"release" usage:"include release-please workflow"`
	Stale   bool `arg:"stale"   usage:"include stale issues workflow"`
	All     bool `arg:"all"     usage:"include all workflows (default if none specified)"`
}

// Workflows bootstraps GitHub workflow files into .github/workflows/.
// By default, all workflows are copied. Use flags to select specific ones.
var Workflows = pocket.Func("github-workflows", "bootstrap GitHub workflow files", workflows).
	With(WorkflowsOptions{})

func workflows(ctx context.Context) error {
	opts := pocket.Options[WorkflowsOptions](ctx)

	// If no specific workflows selected, include all
	includeAll := opts.All || (!opts.PR && !opts.Release && !opts.Stale)

	// Ensure .github/workflows directory exists
	workflowDir := pocket.FromGitRoot(".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		return fmt.Errorf("create workflows dir: %w", err)
	}

	// Map of workflow files to copy
	workflowsToCopy := map[string]bool{
		"pr.yml":      includeAll || opts.PR,
		"release.yml": includeAll || opts.Release,
		"stale.yml":   includeAll || opts.Stale,
	}

	copied := 0
	for filename, include := range workflowsToCopy {
		if !include {
			continue
		}

		content, err := workflowFiles.ReadFile(filepath.Join("workflows", filename))
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", filename, err)
		}

		destPath := filepath.Join(workflowDir, filename)

		// Check if file already exists
		if _, err := os.Stat(destPath); err == nil {
			if pocket.Verbose(ctx) {
				pocket.Printf(ctx, "  %s (already exists, skipping)\n", filename)
			}
			continue
		}

		if err := os.WriteFile(destPath, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}

		pocket.Printf(ctx, "  Created %s\n", destPath)
		copied++
	}

	if copied == 0 {
		pocket.Println(ctx, "  All workflows already exist")
	} else {
		pocket.Printf(ctx, "  Bootstrapped %d workflow(s)\n", copied)
	}

	return nil
}
