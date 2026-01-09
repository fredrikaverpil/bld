// Package govulncheck provides govulncheck tool integration.
package govulncheck

import (
	"context"

	"github.com/fredrikaverpil/pocket/tool"
)

const name = "govulncheck"

// renovate: datasource=go depName=golang.org/x/vuln
const version = "v1.1.4"

var t = &tool.Tool{Name: name, Prepare: Prepare}

// Command prepares the tool and returns an exec.Cmd for running govulncheck.
var Command = t.Command

// Run installs (if needed) and executes govulncheck.
var Run = t.Run

// Prepare ensures govulncheck is installed.
func Prepare(ctx context.Context) error {
	_, err := tool.GoInstall(ctx, "golang.org/x/vuln/cmd/govulncheck", version)
	return err
}
