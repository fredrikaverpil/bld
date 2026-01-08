// Package shim provides generation of the ./pok wrapper scripts.
package shim

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	pocket "github.com/fredrikaverpil/pocket"
)

//go:embed pok.sh.tmpl
var posixTemplate string

//go:embed pok.cmd.tmpl
var windowsTemplate string

//go:embed pok.ps1.tmpl
var powershellTemplate string

// shimData holds the template data for generating a shim.
type shimData struct {
	GoVersion   string
	PocketDir   string
	Context     string
	GoChecksums GoChecksums // SHA256 checksums keyed by "os-arch"
}

// shimType represents a type of shim to generate.
type shimType struct {
	name      string // Template name for errors.
	template  string // Template content.
	extension string // File extension (empty for posix).
	pathSep   string // Path separator to use in template output.
}

// Generate creates or updates wrapper scripts for all contexts.
// It generates shims at the root and one in each unique module directory.
func Generate(cfg pocket.Config) error {
	return GenerateWithRoot(cfg, pocket.GitRoot())
}

// GenerateWithRoot creates or updates wrapper scripts for all contexts
// using the specified root directory. This is useful for testing.
func GenerateWithRoot(cfg pocket.Config, rootDir string) error {
	cfg = cfg.WithDefaults()

	goVersion, err := extractGoVersionFromDir(filepath.Join(rootDir, pocket.DirName))
	if err != nil {
		return fmt.Errorf("reading Go version: %w", err)
	}

	// Fetch checksums for Go downloads.
	checksums, err := FetchGoChecksums(context.Background(), goVersion)
	if err != nil {
		return fmt.Errorf("fetching Go checksums: %w", err)
	}

	// Determine which shim types to generate.
	var types []shimType
	if cfg.Shim.Posix {
		types = append(types, shimType{
			name:      "posix",
			template:  posixTemplate,
			extension: "",
			pathSep:   "/",
		})
	}
	if cfg.Shim.Windows {
		types = append(types, shimType{
			name:      "windows",
			template:  windowsTemplate,
			extension: ".cmd",
			pathSep:   "\\",
		})
	}
	if cfg.Shim.PowerShell {
		types = append(types, shimType{
			name:      "powershell",
			template:  powershellTemplate,
			extension: ".ps1",
			pathSep:   "\\",
		})
	}

	// Generate each shim type at root.
	for _, st := range types {
		tmpl, err := template.New(st.name).Parse(st.template)
		if err != nil {
			return fmt.Errorf("parsing %s template: %w", st.name, err)
		}

		err = generateShim(tmpl, cfg.Shim.Name, st.extension, st.pathSep, goVersion, checksums, rootDir)
		if err != nil {
			return fmt.Errorf("generating %s shim: %w", st.name, err)
		}
	}

	return nil
}

// extractGoVersionFromDir reads a go.mod file from the given directory
// and returns the Go version specified in the "go" directive.
func extractGoVersionFromDir(dir string) (string, error) {
	gomodPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(gomodPath)
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	// Parse the go directive from the file.
	// Look for a line starting with "go " followed by the version.
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "go "); ok {
			version := after
			return strings.TrimSpace(version), nil
		}
	}

	return "", fmt.Errorf("no go directive in %s", gomodPath)
}

// generateShim creates a single shim at root.
func generateShim(
	tmpl *template.Template,
	shimName, extension, _, goVersion string,
	checksums GoChecksums,
	rootDir string,
) error {
	data := shimData{
		GoVersion:   goVersion,
		PocketDir:   ".pocket",
		Context:     ".",
		GoChecksums: checksums,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing shim template: %w", err)
	}

	shimPath := filepath.Join(rootDir, shimName+extension)
	if err := os.WriteFile(shimPath, buf.Bytes(), 0o755); err != nil {
		return fmt.Errorf("writing shim: %w", err)
	}

	return nil
}
