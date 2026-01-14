package golang

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/govulncheck"
)

// Vulncheck runs govulncheck for vulnerability scanning.
var Vulncheck = pocket.Func("go-vulncheck", "run govulncheck", vulncheck)

func vulncheck(ctx context.Context) error {
	return govulncheck.Exec(ctx, "./...")
}
