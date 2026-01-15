package golang

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// Fix runs go fix to update code for newer Go versions.
var Fix = pocket.Func("go-fix", "update code for newer Go versions", fix)

func fix(ctx context.Context) error {
	return pocket.Exec(ctx, "go", "fix", "./...")
}
