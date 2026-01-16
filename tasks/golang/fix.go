package golang

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// Fix runs go fix to update code for newer Go versions.
var Fix = pocket.Func("go-fix", "update code for newer Go versions", fixCmd())

func fixCmd() pocket.Runnable {
	return pocket.RunWith("go", func(ctx context.Context) []string {
		args := []string{"fix"}
		if pocket.Verbose(ctx) {
			args = append(args, "-v")
		}
		args = append(args, "./...")
		return args
	})
}
