package bld

import (
	"fmt"
	"os"
)

// GenerateShim creates or updates the ./bld wrapper script.
func GenerateShim() error {
	goVersion, err := ExtractGoVersion(DirName)
	if err != nil {
		return fmt.Errorf("reading Go version: %w", err)
	}

	shimPath := FromGitRoot("bld")
	if err := os.WriteFile(shimPath, []byte(shimScript(goVersion)), 0o755); err != nil {
		return fmt.Errorf("writing bld shim: %w", err)
	}

	return nil
}

func shimScript(goVersion string) string {
	return fmt.Sprintf(`#!/bin/bash
set -e

BLD_DIR=".bld"
GO_VERSION="%s"
GO_INSTALL_DIR="$BLD_DIR/tools/go/$GO_VERSION"
GO_BIN="$GO_INSTALL_DIR/go/bin/go"

# Find Go binary
if command -v go &> /dev/null; then
    GO_CMD="go"
elif [[ -x "$GO_BIN" ]]; then
    GO_CMD="$GO_BIN"
else
    # Download Go
    echo "Go not found, downloading go$GO_VERSION..."
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [[ "$ARCH" == "x86_64" ]] && ARCH="amd64"
    [[ "$ARCH" == "aarch64" || "$ARCH" == "arm64" ]] && ARCH="arm64"

    mkdir -p "$GO_INSTALL_DIR"
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.${OS}-${ARCH}.tar.gz" | tar -xz -C "$GO_INSTALL_DIR"
    GO_CMD="$GO_BIN"
    echo "Go $GO_VERSION installed to $GO_INSTALL_DIR"
fi

"$GO_CMD" run -C "$BLD_DIR" . -v "$@"
`, goVersion)
}
