// Package shim provides generation of the ./bld wrapper script.
package shim

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/fredrikaverpil/bld"
)

//go:embed bld.sh.tmpl
var shimTemplate string

// Generate creates or updates the ./bld wrapper script.
func Generate() error {
	goVersion, err := bld.ExtractGoVersion(bld.DirName)
	if err != nil {
		return fmt.Errorf("reading Go version: %w", err)
	}

	tmpl, err := template.New("shim").Parse(shimTemplate)
	if err != nil {
		return fmt.Errorf("parsing shim template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"GoVersion": goVersion}); err != nil {
		return fmt.Errorf("executing shim template: %w", err)
	}

	shimPath := bld.FromGitRoot("bld")
	if err := os.WriteFile(shimPath, buf.Bytes(), 0o755); err != nil {
		return fmt.Errorf("writing bld shim: %w", err)
	}

	return nil
}
