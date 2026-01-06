// Package update provides the update task for updating bld dependencies.
package update

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fredrikaverpil/bld"
	"github.com/fredrikaverpil/bld/internal/scaffold"
	"github.com/goyek/goyek/v3"
)

// Task returns a goyek task that updates bld and regenerates files.
func Task(cfg bld.Config) *goyek.DefinedTask {
	return goyek.Define(goyek.Task{
		Name:  "update",
		Usage: "update bld dependency and regenerate files",
		Action: func(a *goyek.A) {
			bldDir := filepath.Join(bld.FromGitRoot(), bld.DirName)

			// Update bld dependency
			a.Log("Updating github.com/fredrikaverpil/bld@latest")
			cmd := exec.CommandContext(a.Context(), "go", "get", "-u", "github.com/fredrikaverpil/bld@latest")
			cmd.Dir = bldDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				a.Fatalf("go get -u: %v", err)
			}

			// Run go mod tidy
			a.Log("Running go mod tidy")
			cmd = exec.CommandContext(a.Context(), "go", "mod", "tidy")
			cmd.Dir = bldDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				a.Fatalf("go mod tidy: %v", err)
			}

			// Regenerate all files
			a.Log("Regenerating files")
			if err := scaffold.GenerateAll(&cfg); err != nil {
				a.Fatal(err)
			}

			a.Log("Done!")
		},
	})
}
