package golang

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/govulncheck"
)

// Vulncheck runs govulncheck for vulnerability scanning.
var Vulncheck = pocket.Func("go-vulncheck", "run govulncheck", pocket.Serial(
	govulncheck.Install,
	vulncheck,
))

func vulncheck(ctx context.Context) error {
	args := []string{}
	if pocket.Verbose(ctx) {
		args = append(args, "-v")
	}
	args = append(args, "./...")
	return pocket.Exec(ctx, govulncheck.Name, args...)
}
