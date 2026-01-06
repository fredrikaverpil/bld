// Package generate provides the generate task for regenerating all generated files.
package generate

import (
	"github.com/fredrikaverpil/bld"
	"github.com/fredrikaverpil/bld/internal/shim"
	"github.com/fredrikaverpil/bld/internal/workflows"
	"github.com/goyek/goyek/v3"
)

// Task returns a goyek task that regenerates all generated files (shim, workflows).
func Task(cfg bld.Config) *goyek.DefinedTask {
	return goyek.Define(goyek.Task{
		Name:  "generate",
		Usage: "regenerate all generated files (shim, workflows)",
		Action: func(a *goyek.A) {
			if err := shim.Generate(); err != nil {
				a.Fatal(err)
			}
			a.Log("Generated ./bld shim")

			if err := workflows.Generate(cfg); err != nil {
				a.Fatal(err)
			}
			a.Log("Generated workflows in .github/workflows/")
		},
	})
}
