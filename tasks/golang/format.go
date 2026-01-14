package golang

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// Format formats Go code using go fmt.
var Format = pocket.Func("go-format", "format Go code", format)

func format(ctx context.Context) error {
	return pocket.Exec(ctx, "go", "fmt", "./...")
}
