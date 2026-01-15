package golang

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// Fix runs go fix to update code for newer Go versions.
var Fix = pocket.Func("go-fix", "update code for newer Go versions", fix)

func fix(ctx context.Context) error {
	args := []string{"fix"}
	if pocket.Verbose(ctx) {
		args = append(args, "-v")
	}
	args = append(args, "./...")
	return pocket.Exec(ctx, "go", args...)
}
