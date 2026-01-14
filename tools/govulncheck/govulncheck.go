// Package govulncheck provides govulncheck integration.
// This is an "action tool" - it provides Install, Check, and Exec.
package govulncheck

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// renovate: datasource=go depName=golang.org/x/vuln
const Version = "v1.1.4" // NOTE: May need updating for Go 1.25+ compatibility

// Install ensures govulncheck is available.
// This is a hidden dependency used by Check and Exec.
var Install = pocket.Func("install:govulncheck", "install govulncheck", install).Hidden()

func install(ctx context.Context) error {
	pocket.Printf(ctx, "Installing govulncheck %s...\n", Version)
	return pocket.InstallGo(ctx, "golang.org/x/vuln/cmd/govulncheck", Version)
}

// Check runs govulncheck vulnerability scanner.
// This is visible in CLI and can be used directly in config.
var Check = pocket.Func("govulncheck", "run govulncheck", check)

func check(ctx context.Context) error {
	pocket.Serial(ctx, Install)
	return pocket.Exec(ctx, "govulncheck", "./...")
}

// Exec runs govulncheck with the given arguments.
// This is for programmatic use when you need full control.
//
// Example:
//
//	govulncheck.Exec(ctx, "-json", "./...")
func Exec(ctx context.Context, args ...string) error {
	pocket.Serial(ctx, Install)
	return pocket.Exec(ctx, "govulncheck", args...)
}
