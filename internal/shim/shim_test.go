package shim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pocket "github.com/fredrikaverpil/pocket"
)

func TestGenerate_PosixShim(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create .pocket directory with go.mod.
	pocketDir := filepath.Join(tmpDir, ".pocket")
	if err := os.MkdirAll(pocketDir, 0o755); err != nil {
		t.Fatalf("creating .pocket dir: %v", err)
	}
	gomod := "module pocket\n\ngo 1.24.4\n"
	if err := os.WriteFile(filepath.Join(pocketDir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	cfg := pocket.Config{
		Shim: &pocket.ShimConfig{
			Name:  "pok",
			Posix: true,
		},
	}

	if err := GenerateWithRoot(cfg, tmpDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Check shim was created.
	shimPath := filepath.Join(tmpDir, "pok")
	if _, err := os.Stat(shimPath); os.IsNotExist(err) {
		t.Fatal("shim was not created")
	}

	// Check shim content.
	content, err := os.ReadFile(shimPath)
	if err != nil {
		t.Fatalf("reading shim: %v", err)
	}

	if !strings.Contains(string(content), "#!/") {
		t.Error("shim missing shebang")
	}
	if !strings.Contains(string(content), ".pocket") {
		t.Error("shim missing .pocket reference")
	}
}

func TestGenerate_WindowsShim(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create .pocket directory with go.mod.
	pocketDir := filepath.Join(tmpDir, ".pocket")
	if err := os.MkdirAll(pocketDir, 0o755); err != nil {
		t.Fatalf("creating .pocket dir: %v", err)
	}
	gomod := "module pocket\n\ngo 1.24.4\n"
	if err := os.WriteFile(filepath.Join(pocketDir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	cfg := pocket.Config{
		Shim: &pocket.ShimConfig{
			Name:    "pok",
			Windows: true,
		},
	}

	if err := GenerateWithRoot(cfg, tmpDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Check shim was created.
	shimPath := filepath.Join(tmpDir, "pok.cmd")
	if _, err := os.Stat(shimPath); os.IsNotExist(err) {
		t.Fatal("shim was not created")
	}
}

func TestGenerate_PowerShellShim(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create .pocket directory with go.mod.
	pocketDir := filepath.Join(tmpDir, ".pocket")
	if err := os.MkdirAll(pocketDir, 0o755); err != nil {
		t.Fatalf("creating .pocket dir: %v", err)
	}
	gomod := "module pocket\n\ngo 1.24.4\n"
	if err := os.WriteFile(filepath.Join(pocketDir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	cfg := pocket.Config{
		Shim: &pocket.ShimConfig{
			Name:       "pok",
			PowerShell: true,
		},
	}

	if err := GenerateWithRoot(cfg, tmpDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Check shim was created.
	shimPath := filepath.Join(tmpDir, "pok.ps1")
	if _, err := os.Stat(shimPath); os.IsNotExist(err) {
		t.Fatal("shim was not created")
	}
}

func TestGenerate_AllShimTypes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create .pocket directory with go.mod.
	pocketDir := filepath.Join(tmpDir, ".pocket")
	if err := os.MkdirAll(pocketDir, 0o755); err != nil {
		t.Fatalf("creating .pocket dir: %v", err)
	}
	gomod := "module pocket\n\ngo 1.24.4\n"
	if err := os.WriteFile(filepath.Join(pocketDir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	cfg := pocket.Config{
		Shim: &pocket.ShimConfig{
			Name:       "build",
			Posix:      true,
			Windows:    true,
			PowerShell: true,
		},
	}

	if err := GenerateWithRoot(cfg, tmpDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Check all shims were created.
	for _, shimFile := range []string{"build", "build.cmd", "build.ps1"} {
		shimPath := filepath.Join(tmpDir, shimFile)
		if _, err := os.Stat(shimPath); os.IsNotExist(err) {
			t.Errorf("shim %q was not created", shimFile)
		}
	}
}
