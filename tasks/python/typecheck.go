package python

import (
	"context"

	"github.com/fredrikaverpil/pocket"
	"github.com/fredrikaverpil/pocket/tools/mypy"
)

// Typecheck type-checks Python files using mypy.
var Typecheck = pocket.Func("py-typecheck", "type-check Python files", pocket.Serial(
	mypy.Install,
	typecheckCmd(),
))

func typecheckCmd() pocket.Runnable {
	return pocket.RunWith(mypy.Name, func(ctx context.Context) []string {
		args := []string{}
		if pocket.Verbose(ctx) {
			args = append(args, "-v")
		}
		args = append(args, pocket.Path(ctx))
		return args
	})
}
