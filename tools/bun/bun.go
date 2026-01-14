// Package bun provides bun runtime integration.
// Bun is a JavaScript runtime used by other tools (e.g., prettier).
// This is a "runtime tool" - it only provides Install, not direct actions.
package bun

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/fredrikaverpil/pocket"
)

// Install ensures bun is available in PATH.
// This is a hidden dependency - other tools call this, users don't.
var Install = pocket.Func("install:bun", "ensure bun is available", install).Hidden()

func install(ctx context.Context) error {
	if _, err := exec.LookPath("bun"); err != nil {
		return fmt.Errorf("bun not found in PATH - install from https://bun.sh")
	}
	return nil
}
