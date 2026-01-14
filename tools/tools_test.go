package tools_test

import (
	"context"
	"os"
	"testing"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/basedpyright"
	"github.com/fredrikaverpil/pocket/tools/bun"
	"github.com/fredrikaverpil/pocket/tools/golangcilint"
	"github.com/fredrikaverpil/pocket/tools/govulncheck"
	"github.com/fredrikaverpil/pocket/tools/mdformat"
	"github.com/fredrikaverpil/pocket/tools/mypy"
	"github.com/fredrikaverpil/pocket/tools/prettier"
	"github.com/fredrikaverpil/pocket/tools/ruff"
	"github.com/fredrikaverpil/pocket/tools/stylua"
	"github.com/fredrikaverpil/pocket/tools/uv"
)

var tools = []struct {
	name        string
	tool        *pocket.Tool
	versionArgs []string
}{
	{"golangci-lint", golangcilint.Tool, []string{"version"}},
	{"govulncheck", govulncheck.Tool, []string{"-version"}},
	{"uv", uv.Tool, []string{"--version"}},
	{"mdformat", mdformat.Tool, []string{"--version"}},
	{"ruff", ruff.Tool, []string{"--version"}},
	{"mypy", mypy.Tool, []string{"--version"}},
	{"basedpyright", basedpyright.Tool, []string{"--version"}},
	{"stylua", stylua.Tool, []string{"--version"}},
	{"bun", bun.Tool, []string{"--version"}},
	{"prettier", prettier.Tool, []string{"--version"}},
}

func TestTools(t *testing.T) {
	// Create an Execution for testing tool installation.
	out := &pocket.Output{Stdout: os.Stdout, Stderr: os.Stderr}
	exec := pocket.NewExecution(out, false, ".")
	tc := exec.TaskContext(".")

	for _, tool := range tools {
		t.Run(tool.name, func(t *testing.T) {
			ctx := context.Background()
			// Install the tool (Tool implements Runnable).
			if err := tool.tool.Run(ctx, exec); err != nil {
				t.Fatalf("Install: %v", err)
			}
			// Run the tool binary to verify it works.
			if err := tool.tool.Exec(ctx, tc, tool.versionArgs...); err != nil {
				t.Fatalf("Exec %v: %v", tool.versionArgs, err)
			}
		})
	}
}
