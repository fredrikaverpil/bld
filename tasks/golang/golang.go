// Package golang provides Go development tasks.
package golang

import (
	"context"

	"github.com/fredrikaverpil/pocket"
)

// Tool versions - managed by renovate.
const (
	// renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
	golangciLintVersion = "v2.0.2"

	// renovate: datasource=go depName=golang.org/x/vuln
	govulncheckVersion = "latest"
)

// golangci-lint configuration.
var golangciLintConfig = pocket.ToolConfig{
	UserFiles:   []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"},
	DefaultFile: "", // No default config - use golangci-lint defaults
}

// Task definitions.
var (
	// GoFormat formats Go code using go fmt.
	GoFormat = pocket.Func("go-format", "format Go code", goFormat)

	// GoLint runs golangci-lint.
	GoLint = pocket.Func("go-lint", "run golangci-lint", goLint)

	// GoTest runs tests with race detection.
	GoTest = pocket.Func("go-test", "run tests with race detection", goTest)

	// GoVulncheck runs govulncheck for vulnerability scanning.
	GoVulncheck = pocket.Func("go-vulncheck", "run govulncheck", goVulncheck)

	// InstallGolangciLint installs golangci-lint.
	InstallGolangciLint = pocket.Func("install:golangci-lint", "install golangci-lint", installGolangciLint).Hidden()

	// InstallGovulncheck installs govulncheck.
	InstallGovulncheck = pocket.Func("install:govulncheck", "install govulncheck", installGovulncheck).Hidden()
)

// Tasks returns the default Go tasks as a serial runnable.
// Use this with pocket.Paths().DetectBy(golang.Detect()) for auto-detection.
//
// Example:
//
//	pocket.Paths(golang.Tasks()).DetectBy(golang.Detect())
func Tasks() pocket.Runnable {
	return pocket.Serial(GoFormat, GoLint, GoTest, GoVulncheck)
}

// Detect returns a detection function for Go modules.
// It finds directories containing go.mod files.
func Detect() func() []string {
	return func() []string {
		return pocket.DetectByFile("go.mod")
	}
}

// Task implementations.

func goFormat(ctx context.Context) error {
	return pocket.Exec(ctx, "go", "fmt", "./...")
}

func goLint(ctx context.Context) error {
	// Install golangci-lint if needed (runs once per execution due to dedup)
	pocket.Serial(ctx, InstallGolangciLint)

	args := []string{"run"}

	// Check for user config file.
	configPath, err := pocket.ConfigPath("golangci-lint", golangciLintConfig)
	if err != nil {
		return err
	}
	if configPath != "" {
		args = append(args, "-c", configPath)
	}

	args = append(args, "./...")
	return pocket.Exec(ctx, "golangci-lint", args...)
}

func goTest(ctx context.Context) error {
	return pocket.Exec(ctx, "go", "test", "-race", "./...")
}

func goVulncheck(ctx context.Context) error {
	// Install govulncheck if needed (runs once per execution due to dedup)
	pocket.Serial(ctx, InstallGovulncheck)
	return pocket.Exec(ctx, "govulncheck", "./...")
}

// Install functions.

func installGolangciLint(ctx context.Context) error {
	pocket.Printf(ctx, "Installing golangci-lint %s...\n", golangciLintVersion)
	return pocket.InstallGo(ctx, "github.com/golangci/golangci-lint/v2/cmd/golangci-lint", golangciLintVersion)
}

func installGovulncheck(ctx context.Context) error {
	pocket.Printf(ctx, "Installing govulncheck %s...\n", govulncheckVersion)
	return pocket.InstallGo(ctx, "golang.org/x/vuln/cmd/govulncheck", govulncheckVersion)
}
