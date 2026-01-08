package pocket

// Config defines the configuration for a project using pocket.
type Config struct {
	// TaskGroups contains task groups (e.g., Go, Lua, Markdown) to enable.
	//
	// Example:
	//
	//	TaskGroups: []*pocket.TaskGroup{
	//	    golang.NewTaskGroup(),
	//	    markdown.NewTaskGroup(),
	//	},
	TaskGroups []*TaskGroup

	// Tasks contains custom standalone tasks.
	//
	// Example:
	//
	//	Tasks: []*pocket.Task{
	//	    {Name: "deploy", Usage: "deploy the app", Action: deployAction},
	//	},
	Tasks []*Task

	// Shim controls shim script generation.
	// By default, only Posix (./pok) is generated with name "pok".
	Shim *ShimConfig

	// SkipGitDiff disables the git diff check at the end of the "all" task.
	// By default, "all" fails if there are uncommitted changes after running all tasks.
	// Set to true to disable this check.
	SkipGitDiff bool
}

// ShimConfig controls shim script generation.
type ShimConfig struct {
	// Name is the base name of the generated shim scripts (without extension).
	// Default: "pok"
	Name string

	// Posix generates a bash script (./pok).
	// This is enabled by default if ShimConfig is nil.
	Posix bool

	// Windows generates a batch file (pok.cmd).
	// The batch file requires Go to be installed and in PATH.
	Windows bool

	// PowerShell generates a PowerShell script (pok.ps1).
	// The PowerShell script can auto-download Go if not found.
	PowerShell bool
}

// WithDefaults returns a copy of the config with default values applied.
func (c Config) WithDefaults() Config {
	// Default to Posix shim only if no Shim config is provided.
	if c.Shim == nil {
		c.Shim = &ShimConfig{Posix: true}
	}
	// Apply shim defaults.
	shim := *c.Shim
	if shim.Name == "" {
		shim.Name = "pok"
	}
	c.Shim = &shim

	return c
}
