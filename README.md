# bld

An opinonated, cross-platform, build system for git projects, powered by
[goyek](https://github.com/goyek/goyek).

> [!WARNING]
>
> Under heavy development. Breaking changes will occur.

## Features

- **Cross-platform**: No Makefiles - works on Windows, macOS, and Linux
- **Task management**: Defines tasks like `go-test`, `python-lint`...
- **Tool management**: Downloads and caches tools in `.bld/`, which are used by
  tasks
- **CI workflow generation**: Generates GitHub Actions workflows from templates
- **Simple invocation**: Just `./bld <task>` or `./bld -h` to list all tasks

## Bootstrap a new project

Run the init command in your project root (must have a `go.mod`):

```bash
go run github.com/fredrikaverpil/bld/cmd/bld@latest init
```

This creates:

- `.bld/` - build module with config and tasks
- `./bld` - wrapper script

### Run tasks

```bash
./bld            # run all tasks (generate, lint, format, test)
./bld update     # update bld to latest version
./bld generate   # regenerate shim and CI workflows
```

Run `./bld -h` for a list of all possible tasks to run.

### Shell alias (optional)

For even shorter commands, add an alias to your shell profile:

```bash
# ~/.bashrc or ~/.zshrc
alias bld='./bld'
```

Then run tasks with just `bld <task>`.

### Configuration

```go
bld.Config{
    // Go configuration (nil = no Go tasks)
    Go: &bld.GoConfig{
        Modules: map[string]bld.GoModuleOptions{
            ".":          {},                         // all tasks enabled
            "subdir/lib": {SkipFormat: true},         // skip format for this module
            "generated":  {SkipLint: true},           // skip lint for generated code
        },
    },

    // GitHub Actions configuration (nil = no GitHub workflows)
    GitHub: &bld.GitHubConfig{
        // Go versions are automatically extracted from each module's go.mod.
        // Add extra versions to test against (e.g., "stable", "oldstable", pre-releases).
        ExtraGoVersions: []string{"stable"},
        OSVersions:      []string{"ubuntu-latest"},  // default

        // Skip generic workflows
        SkipPR:      false,  // semantic PR validation
        SkipStale:   false,  // stale issue/PR management
        SkipRelease: false,  // release-please
        SkipSync:    false,  // auto-sync bld updates
    },
}
```

Task skips in `GoModuleOptions` affect both local execution and CI:

- If all modules skip format → no format job in CI workflow
- `go-fmt` task only runs on modules where `SkipFormat: false`

### Project Structure

```
your-project/
├── .bld/
│   ├── main.go      # generated (do not edit)
│   ├── config.go    # project config (edit this)
│   └── go.mod
├── .github/workflows/
│   └── bld-*.yml    # generated
└── ...
```

### Custom Tasks

Add your own tasks in `.bld/config.go`:

```go
import (
    "github.com/fredrikaverpil/bld"
    "github.com/goyek/goyek/v3"
)

var Config = bld.Config{
    Go: &bld.GoConfig{...},

    // Custom tasks per folder
    Custom: map[string][]goyek.Task{
        ".": {  // available from root ./bld
            {
                Name:  "deploy",
                Usage: "deploy to production",
                Action: func(a *goyek.A) {
                    a.Log("Deploying...")
                    // your logic here
                },
            },
        },
    },
}
```

Custom tasks appear in `./bld -h` and run as part of `./bld all`.

For multi-module projects, you can define context-specific tasks that only
appear when running the shim from that folder:

```go
Custom: map[string][]goyek.Task{
    ".":            {rootTask},
    "services/api": {apiTask},  // only in ./services/api/bld
}
```

See [goyek documentation](https://github.com/goyek/goyek) for more task options
like dependencies, parallel execution, and error handling.

### Releases

The GitHub release workflow requires the following repository settings:

- Actions → General → Workflow Permissions:
  - [x] **Read and write permissions**
  - [x] **Allow GitHub Actions to create and approve pull requests**

### Windows Support

By default, bld generates a Posix-compatible `./bld` script that works with
bash. For native Windows support, you can enable Windows shims in your
`.bld/config.go`:

```go
var Config = bld.Config{
    Shim: &bld.ShimConfig{
        Name:       "bld",  // base name (default: "bld")
        Posix:      true,   // ./bld (bash) - default
        Windows:    true,   // bld.cmd (requires Go in PATH)
        PowerShell: true,   // bld.ps1 (can auto-download Go)
    },
    // ... rest of config
}
```

After updating the config, run `./bld generate` to create the Windows shims.

**Shim types:**

| Shim          | File      | Go Auto-Download | Notes                        |
| ------------- | --------- | ---------------- | ---------------------------- |
| Posix         | `./bld`   | Yes              | Works with bash, Git Bash    |
| Windows (CMD) | `bld.cmd` | No               | Requires Go in PATH          |
| PowerShell    | `bld.ps1` | Yes              | Full-featured Windows option |

**Using the shims on Windows:**

```batch
rem CMD
bld.cmd go-test

rem PowerShell
.\bld.ps1 go-test
```

**Alternative approaches:**

- Git Bash (included with Git for Windows) - use `./bld` directly
- WSL (Windows Subsystem for Linux) - use `./bld` directly
- GitHub Actions - uses `shell: bash` automatically in generated workflows

## Adding a New Ecosystem

To add support for a new language/ecosystem (e.g., Python, Lua):

1. **Create tools** in `tools/<toolname>/tool.go`
   - Export `Prepare(ctx) error`, `Command(ctx, args) (*Cmd, error)`,
     `Run(ctx, args) error`
   - Both `Command()` and `Run()` auto-prepare the tool
   - Add Renovate comment for version updates
1. **Create tasks** in `tasks/<ecosystem>/tasks.go`
   - Define goyek tasks that use the tools
   - Use `tool.Command()` when you need to customize the command (e.g., set
     `cmd.Dir`), or `tool.Run()` for simple execution
1. **Add tool tests** in `tools/tools_test.go`
   - Add one line to the `tools` table
1. **Wire up in config** - add config options in `config.go` if needed

## Terminology

### Tools

- Binaries downloaded to `.bld/tools/` and symlinked to `.bld/bin/`
- Examples: golangci-lint, govulncheck, mdformat, uv
- Have versions, download URLs, Renovate comments
- Expose `Prepare()`, `Command()`, `Run()` functions

### Tasks (goyek tasks)

- What users execute: `go-fmt`, `go-lint`, `go-test`
- Use tools via their Go API
- Defined in `tasks/`
