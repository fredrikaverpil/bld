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

// toolTest defines a tool to test.
type toolTest struct {
	name        string
	install     *pocket.FuncDef
	exec        func(ctx context.Context, args ...string) error
	versionArgs []string
}

var tools = []toolTest{
	{"golangci-lint", golangcilint.Install, golangcilint.Exec, []string{"version"}},
	{"govulncheck", govulncheck.Install, govulncheck.Exec, []string{"-version"}},
	{"uv", uv.Install, uv.Exec, []string{"--version"}},
	{"mdformat", mdformat.Install, mdformat.Exec, []string{"--version"}},
	{"ruff", ruff.Install, ruff.Exec, []string{"--version"}},
	{"mypy", mypy.Install, mypy.Exec, []string{"--version"}},
	{"basedpyright", basedpyright.Install, basedpyright.Exec, []string{"--version"}},
	{"stylua", stylua.Install, stylua.Exec, []string{"--version"}},
	{"bun", bun.Install, bun.Exec, []string{"--version"}},
	{"prettier", prettier.Install, prettier.Exec, []string{"--version"}},
}

func TestTools(t *testing.T) {
	// Create execution context for testing.
	out := pocket.StdOutput()
	out.Stdout = os.Stdout
	out.Stderr = os.Stderr

	for _, tool := range tools {
		t.Run(tool.name, func(t *testing.T) {
			ctx := pocket.TestContext(out)

			// Install the tool.
			pocket.Serial(ctx, tool.install)

			// Run the tool to verify it works.
			if err := tool.exec(ctx, tool.versionArgs...); err != nil {
				t.Fatalf("Exec %v: %v", tool.versionArgs, err)
			}
		})
	}
}
