package pocket

import "encoding/json"

// TaskInfo represents a task for export.
// This is the public type used by the export API for CI/CD integration.
type TaskInfo struct {
	Name   string   `json:"name"`             // CLI command name
	Usage  string   `json:"usage"`            // Description/help text
	Paths  []string `json:"paths,omitempty"`  // Directories this task runs in
	Hidden bool     `json:"hidden,omitempty"` // Whether task is hidden from help
}

// CollectTasks extracts task information from a Runnable tree.
// This walks the tree and returns all TaskDefs with their path mappings.
// Tasks without RunIn() wrappers get ["."] (root only).
func CollectTasks(r Runnable) []TaskInfo {
	if r == nil {
		return nil
	}

	pathMappings := collectPathMappings(r)
	funcs := r.funcs()

	result := make([]TaskInfo, 0, len(funcs))
	for _, f := range funcs {
		info := TaskInfo{
			Name:   f.name,
			Usage:  f.usage,
			Hidden: f.hidden,
		}

		// Get paths from mapping, default to ["."] for root-only tasks
		if pf, ok := pathMappings[f.name]; ok {
			info.Paths = pf.Resolve()
		} else {
			info.Paths = []string{"."}
		}

		result = append(result, info)
	}

	return result
}

// ExportJSON exports task information as JSON bytes.
// This is the underlying implementation for the export command.
func ExportJSON(r Runnable) ([]byte, error) {
	tasks := CollectTasks(r)
	return json.MarshalIndent(tasks, "", "  ")
}
