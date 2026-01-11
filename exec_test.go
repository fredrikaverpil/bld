package pocket

import (
	"os"
	"testing"
)

func TestComputeColorEnv(t *testing.T) {
	tests := []struct {
		name       string
		isTTY      bool
		noColorSet bool
		wantColors bool
	}{
		{
			name:       "TTY without NO_COLOR returns color vars",
			isTTY:      true,
			noColorSet: false,
			wantColors: true,
		},
		{
			name:       "TTY with NO_COLOR returns nil",
			isTTY:      true,
			noColorSet: true,
			wantColors: false,
		},
		{
			name:       "non-TTY without NO_COLOR returns nil",
			isTTY:      false,
			noColorSet: false,
			wantColors: false,
		},
		{
			name:       "non-TTY with NO_COLOR returns nil",
			isTTY:      false,
			noColorSet: true,
			wantColors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeColorEnv(tt.isTTY, tt.noColorSet)
			if tt.wantColors {
				if len(got) == 0 {
					t.Error("expected color env vars, got none")
				}
				// Verify expected vars are present.
				hasForceColor := false
				for _, v := range got {
					if v == "FORCE_COLOR=1" {
						hasForceColor = true
					}
				}
				if !hasForceColor {
					t.Error("expected FORCE_COLOR=1 in color env vars")
				}
			} else {
				if len(got) != 0 {
					t.Errorf("expected no color env vars, got %v", got)
				}
			}
		})
	}
}

func TestPrependPath(t *testing.T) {
	sep := string(os.PathListSeparator)
	tests := []struct {
		name     string
		env      []string
		dir      string
		wantPath string
	}{
		{
			name:     "prepends to existing PATH",
			env:      []string{"HOME=/home/user", "PATH=/usr/bin"},
			dir:      "/custom/bin",
			wantPath: "PATH=/custom/bin" + sep + "/usr/bin",
		},
		{
			name:     "creates PATH if not exists",
			env:      []string{"HOME=/home/user"},
			dir:      "/custom/bin",
			wantPath: "PATH=/custom/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrependPath(tt.env, tt.dir)
			found := false
			for _, v := range got {
				if v == tt.wantPath {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("PrependPath() = %v, want PATH containing %q", got, tt.wantPath)
			}
		})
	}
}
