// Package govulncheck provides govulncheck tool integration.
package govulncheck

import (
	"context"

	"github.com/fredrikaverpil/pocket/tool"
)

const name = "govulncheck"

// renovate: datasource=go depName=golang.org/x/vuln
const version = "v1.1.4"

// T is the tool instance for use with TaskContext.Tool().
// Example: tc.Tool(govulncheck.T).Run(ctx, "./...").
var T = &tool.Tool{Name: name, Prepare: Prepare}

// Prepare ensures govulncheck is installed.
func Prepare(ctx context.Context) error {
	_, err := tool.GoInstall(ctx, "golang.org/x/vuln/cmd/govulncheck", version)
	return err
}
