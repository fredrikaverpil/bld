package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/fredrikaverpil/bld/internal/scaffold"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`bld - bootstrap bld in your project

Usage:
  bld init      Initialize .bld/ in current directory

Examples:
  go run github.com/fredrikaverpil/bld/cmd/bld@latest init`)
}

func runInit() error {
	// Check we're in a Go module
	if _, err := os.Stat("go.mod"); err != nil {
		return fmt.Errorf("not in a Go module (no go.mod found)")
	}

	// Check .bld doesn't already exist
	if _, err := os.Stat(".bld"); err == nil {
		return fmt.Errorf(".bld/ already exists")
	}

	fmt.Println("Initializing bld...")

	// Create .bld directory
	if err := os.MkdirAll(".bld", 0o755); err != nil {
		return fmt.Errorf("creating .bld/: %w", err)
	}

	// Create go.mod
	fmt.Println("  Creating .bld/go.mod")
	if err := runCommand(".bld", "go", "mod", "init", "bld"); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}

	// Get dependencies
	deps := []string{
		"github.com/fredrikaverpil/bld@latest",
		"github.com/goyek/goyek/v3@latest",
		"github.com/goyek/x@latest",
	}
	for _, dep := range deps {
		fmt.Printf("  Adding %s\n", dep)
		if err := runCommand(".bld", "go", "get", dep); err != nil {
			return fmt.Errorf("go get %s: %w", dep, err)
		}
	}

	// Generate all scaffold files (config.go, .gitignore, main.go, shim)
	// No config = no workflows
	fmt.Println("  Generating scaffold files")
	if err := scaffold.GenerateAll(nil); err != nil {
		return fmt.Errorf("generating scaffold: %w", err)
	}

	// Run go mod tidy (after main.go is created)
	fmt.Println("  Running go mod tidy")
	if err := runCommand(".bld", "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	fmt.Println()
	fmt.Println("Done! You can now run:")
	fmt.Println("  ./bld -h          # list available tasks")
	fmt.Println("  ./bld             # run all tasks")
	fmt.Println("  ./bld update      # update bld to latest version")

	return nil
}

func runCommand(dir, name string, args ...string) error {
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
