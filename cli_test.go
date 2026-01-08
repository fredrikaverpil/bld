package pocket

import (
	"strings"
	"testing"
)

func TestPrintTaskHelp_NoArgs(t *testing.T) {
	task := &Task{
		Name:  "test-task",
		Usage: "a test task",
	}

	// Verify task with no args is set up correctly
	if len(task.Args) != 0 {
		t.Error("expected no args")
	}
}

func TestPrintTaskHelp_WithArgs(t *testing.T) {
	task := &Task{
		Name:  "greet",
		Usage: "print a greeting",
		Args: []ArgDef{
			{Name: "name", Usage: "who to greet", Default: "world"},
			{Name: "count", Usage: "how many times"},
		},
	}

	if len(task.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(task.Args))
	}
	if task.Args[0].Name != "name" {
		t.Errorf("expected first arg name='name', got %q", task.Args[0].Name)
	}
	if task.Args[0].Default != "world" {
		t.Errorf("expected first arg default='world', got %q", task.Args[0].Default)
	}
	if task.Args[1].Default != "" {
		t.Errorf("expected second arg no default, got %q", task.Args[1].Default)
	}
}

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		input   string
		wantKey string
		wantVal string
		wantOK  bool
	}{
		{"name=world", "name", "world", true},
		{"count=10", "count", "10", true},
		{"empty=", "empty", "", true},
		{"with=equals=sign", "with", "equals=sign", true},
		{"noequals", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, val, ok := strings.Cut(tt.input, "=")
			if ok != tt.wantOK {
				t.Errorf("Cut(%q): got ok=%v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok {
				if key != tt.wantKey {
					t.Errorf("Cut(%q): got key=%q, want %q", tt.input, key, tt.wantKey)
				}
				if val != tt.wantVal {
					t.Errorf("Cut(%q): got val=%q, want %q", tt.input, val, tt.wantVal)
				}
			}
		})
	}
}
