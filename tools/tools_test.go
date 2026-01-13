package tools_test

import (
	"context"
	"os"
	"testing"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/basedpyright"
	"github.com/fredrikaverpil/pocket/tools/golangcilint"
	"github.com/fredrikaverpil/pocket/tools/govulncheck"
	"github.com/fredrikaverpil/pocket/tools/mdformat"
	"github.com/fredrikaverpil/pocket/tools/mypy"
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
}

func TestTools(t *testing.T) {
	// Create a minimal TaskContext for testing.
	tc := &pocket.TaskContext{
		Out: &pocket.Output{Stdout: os.Stdout, Stderr: os.Stderr},
	}

	for _, tool := range tools {
		t.Run(tool.name, func(t *testing.T) {
			ctx := context.Background()
			if err := tool.tool.Install(ctx, tc); err != nil {
				t.Fatalf("Install: %v", err)
			}
			if err := tool.tool.Run(ctx, tc, tool.versionArgs...); err != nil {
				t.Fatalf("Run %v: %v", tool.versionArgs, err)
			}
		})
	}
}
