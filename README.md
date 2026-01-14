# pocket

A cross-platform build system inspired by [Mage](https://magefile.org/) and
[Sage](https://github.com/einride/sage). Define functions, compose them with
`Serial`/`Parallel`, and let pocket handle tool installation.

> [!NOTE]
>
> You don't need Go installed to use Pocket. The `./pok` shim automatically
> downloads Go to `.pocket/` if needed.

> [!WARNING]
>
> Under heavy development. Breaking changes will occur until the initial
> release.

## Features

- **Cross-platform**: Works on Windows, macOS, and Linux (no Makefiles)
- **Function-based**: Define functions with `pocket.Func()`, compose with
  `Serial()`/`Parallel()`
- **Dependency management**: Functions can depend on other functions with
  automatic deduplication
- **Tool management**: Downloads and caches tools in `.pocket/`
- **Path filtering**: Run different functions in different directories

## Quickstart

### Bootstrap

Run in your project root (requires Go for this step only):

```bash
go run github.com/fredrikaverpil/pocket/cmd/pocket@latest init
```

This creates `.pocket/` and `./pok` (the wrapper script).

### Your first function

Edit `.pocket/config.go`:

```go
package main

import (
    "context"
    "github.com/fredrikaverpil/pocket"
)

var Config = pocket.Config{
    ManualRun: []pocket.Runnable{Hello},
}

var Hello = pocket.Func("hello", "say hello", hello)

func hello(ctx context.Context) error {
    pocket.Printf(ctx, "Hello from pocket!\n")
    return nil
}
```

```bash
./pok -h      # list functions
./pok hello   # run function
```

### Composition

Use `Serial()` and `Parallel()` to control execution order:

```go
var Config = pocket.Config{
    AutoRun: pocket.Serial(
        Format,              // first
        pocket.Parallel(     // then these in parallel
            Lint,
            Test,
        ),
        Build,               // last
    ),
}
```

### Dependencies

Functions can depend on other functions. Dependencies are deduplicated
automatically - each function runs at most once per execution.

```go
var Install = pocket.Func("install:tool", "install tool", install).Hidden()
var Lint = pocket.Func("lint", "run linter", lint)

func lint(ctx context.Context) error {
    // Ensure tool is installed (runs once, even if called multiple times)
    pocket.Serial(ctx, Install)
    return pocket.Exec(ctx, "tool", "lint", "./...")
}
```

## Concepts

### Functions

Everything in Pocket is a function created with `pocket.Func()`:

```go
var MyFunc = pocket.Func("name", "description", implementation)

func implementation(ctx context.Context) error {
    // do work
    return nil
}
```

Functions can be:

- **Visible**: Shown in `./pok -h` and callable from CLI
- **Hidden**: Not shown in help, used as dependencies (`.Hidden()`)

### Serial and Parallel

These have two modes based on the first argument:

**Composition mode** (no context) - returns a Runnable for config:

```go
pocket.Serial(fn1, fn2, fn3)    // run in sequence
pocket.Parallel(fn1, fn2, fn3)  // run concurrently
```

**Execution mode** (with context) - runs immediately:

```go
pocket.Serial(ctx, fn1, fn2)    // run dependencies in sequence
pocket.Parallel(ctx, fn1, fn2)  // run dependencies concurrently
```

### Tools vs Tasks

Pocket distinguishes between **tools** (provide capabilities) and **tasks** (do
work):

| Type             | Purpose            | Exports                               | Example          |
| ---------------- | ------------------ | ------------------------------------- | ---------------- |
| **Runtime Tool** | Provides a runtime | `Install` (hidden)                    | bun, uv          |
| **Action Tool**  | Does something     | `Install` + action func + `Exec()`    | prettier, ruff   |
| **Task**         | Orchestrates tools | Action funcs + `Tasks()` + `Detect()` | markdown, golang |

**Runtime Tool** (e.g., `tools/bun/bun.go`):

```go
// Only provides Install - used as a dependency by other tools
var Install = pocket.Func("install:bun", "ensure bun available", install).Hidden()
```

**Action Tool** (e.g., `tools/prettier/prettier.go`):

```go
// Hidden install function
var Install = pocket.Func("install:prettier", "ensure prettier available", install).Hidden()

// Visible action function - can be used directly in config
var Format = pocket.Func("prettier", "format with prettier", format)

// Programmatic helper for other packages
func Exec(ctx context.Context, args ...string) error {
    pocket.Serial(ctx, Install)
    return pocket.Exec(ctx, "bunx", "prettier@"+Version, args...)
}
```

**Task** (e.g., `tasks/markdown/markdown.go`):

```go
// Visible function
var Format = pocket.Func("md-format", "format Markdown files", format)

// Returns all tasks as a Runnable for config
func Tasks() pocket.Runnable { return Format }

// Detection function for auto-discovery
func Detect() func() []string {
    return func() []string { return []string{"."} }
}

func format(ctx context.Context) error {
    return prettier.Exec(ctx, "--write", "**/*.md")
}
```

### Config Usage

```go
var Config = pocket.Config{
    AutoRun: pocket.Serial(
        // Use task collections with auto-detection
        pocket.Paths(golang.Tasks()).DetectBy(golang.Detect()),
        pocket.Paths(markdown.Tasks()).DetectBy(markdown.Detect()),

        // Or use tools directly
        pocket.Paths(prettier.Format).In("docs"),
    ),
    ManualRun: []pocket.Runnable{
        Deploy,
    },
}
```

## Path Filtering

For monorepos, use `Paths()` to control where functions run:

```go
// Run in specific directories
pocket.Paths(myFunc).In("services/api", "services/web")

// Auto-detect directories containing go.mod
pocket.Paths(golang.Tasks()).DetectBy(golang.Detect())

// Exclude directories
pocket.Paths(golang.Tasks()).DetectBy(golang.Detect()).Except("vendor")

// Skip specific functions in specific paths
pocket.Paths(golang.Tasks()).DetectBy(golang.Detect()).Skip(golang.GoTest, "docs")
```

## Options

Functions can accept options:

```go
type DeployOptions struct {
    Env    string `arg:"env" usage:"target environment"`
    DryRun bool   `arg:"dry-run" usage:"print without executing"`
}

var Deploy = pocket.Func("deploy", "deploy to environment", deploy).
    With(DeployOptions{Env: "staging"})

func deploy(ctx context.Context) error {
    opts := pocket.Options[DeployOptions](ctx)
    if opts.DryRun {
        pocket.Printf(ctx, "Would deploy to %s\n", opts.Env)
        return nil
    }
    // deploy...
    return nil
}
```

```bash
./pok deploy                     # uses default (staging)
./pok deploy -env=prod -dry-run  # override at runtime
```

## Bundled Packages

### Tasks

```go
import (
    "github.com/fredrikaverpil/pocket/tasks/golang"
    "github.com/fredrikaverpil/pocket/tasks/markdown"
)

var Config = pocket.Config{
    AutoRun: pocket.Serial(
        pocket.Paths(golang.Tasks()).DetectBy(golang.Detect()),
        pocket.Paths(markdown.Tasks()).DetectBy(markdown.Detect()),
    ),
}
```

### Tools

```go
import "github.com/fredrikaverpil/pocket/tools/prettier"

func myFormat(ctx context.Context) error {
    return prettier.Exec(ctx, "--write", ".")
}
```

## Reference

### Helper Functions

```go
// Execution
pocket.Exec(ctx, "command", "arg1", "arg2")  // run command
pocket.Printf(ctx, "format %s", arg)          // print output

// Paths
pocket.GitRoot()              // git repository root
pocket.FromGitRoot("subdir")  // path relative to git root
pocket.FromPocketDir("file")  // path relative to .pocket/
pocket.FromBinDir("tool")     // path relative to .pocket/bin/

// Context
pocket.Options[T](ctx)        // get typed options
pocket.Path(ctx)              // current path (for path-filtered functions)

// Detection
pocket.DetectByFile("go.mod")       // find dirs with file
pocket.DetectByExtension(".lua")    // find dirs with extension

// Installation
pocket.InstallGo(ctx, "pkg/path", "version")  // go install
pocket.ConfigPath("tool", config)              // find/create config file
```

### Config Structure

```go
var Config = pocket.Config{
    // AutoRun: runs on ./pok (no arguments)
    AutoRun: pocket.Serial(...),

    // ManualRun: requires ./pok <name>
    ManualRun: []pocket.Runnable{...},

    // Shim: configure wrapper scripts
    Shim: &pocket.ShimConfig{
        Name:       "pok",   // base name
        Posix:      true,    // ./pok
        Windows:    true,    // pok.cmd
        PowerShell: true,    // pok.ps1
    },
}
```

## Acknowledgements

- [einride/sage](https://github.com/einride/sage) - Inspiration for the
  function-based architecture and dependency pattern
- [magefile/mage](https://github.com/magefile/mage) - Inspiration for the
  Go-based build system approach
