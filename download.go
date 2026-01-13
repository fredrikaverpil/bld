package pocket

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DownloadOpts configures binary download and extraction.
type DownloadOpts struct {
	// DestDir is the directory to extract files to.
	DestDir string
	// Format specifies the archive format: "tar.gz", "tar", "zip", or "" for raw copy.
	Format string
	// ExtractFiles limits extraction to files matching these base names (flattened to DestDir).
	ExtractFiles []string
	// SkipIfExists skips download if this file already exists.
	SkipIfExists string
	// Symlink creates a symlink in .pocket/bin/ if true.
	Symlink bool
	// HTTPHeaders adds headers to the download request.
	HTTPHeaders map[string]string
}

// DownloadBinary downloads and extracts a binary from a URL.
// Progress and status messages are written to tc.Out.
//
// Example:
//
//	func install(ctx context.Context, tc *pocket.TaskContext) error {
//	    return pocket.DownloadBinary(ctx, tc, url, pocket.DownloadOpts{
//	        DestDir:      pocket.FromToolsDir("mytool", version, "bin"),
//	        Format:       "tar.gz",
//	        ExtractFiles: []string{pocket.BinaryName("mytool")},
//	        Symlink:      true,
//	    })
//	}
func DownloadBinary(ctx context.Context, tc *TaskContext, url string, opts DownloadOpts) error {
	binaryName := ""
	if len(opts.ExtractFiles) > 0 {
		binaryName = opts.ExtractFiles[0]
	}
	binaryPath := filepath.Join(opts.DestDir, binaryName)

	// Check if we can skip.
	skipPath := opts.SkipIfExists
	if skipPath == "" && binaryName != "" {
		skipPath = binaryPath
	}
	if skipPath != "" {
		if _, err := os.Stat(skipPath); err == nil {
			// Already installed, just ensure symlink exists.
			if opts.Symlink && binaryPath != "" {
				if _, err := CreateSymlink(binaryPath); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Create destination directory.
	if opts.DestDir != "" {
		if err := os.MkdirAll(opts.DestDir, 0o755); err != nil {
			return fmt.Errorf("create destination dir: %w", err)
		}
	}

	tc.Out.Printf("  Downloading %s\n", url)

	// Download to temp file.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	for k, v := range opts.HTTPHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "pocket-download-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("download: %w", err)
	}
	tmpFile.Close()

	// Extract or copy.
	if err := extractFile(tmpPath, opts); err != nil {
		return err
	}

	// Create symlink if requested.
	if opts.Symlink && binaryPath != "" {
		if _, err := CreateSymlink(binaryPath); err != nil {
			return err
		}
	}

	return nil
}

// GoInstall installs a Go binary using 'go install'.
// The binary is installed to .pocket/tools/go/<pkg>/<version>/
// and symlinked to .pocket/bin/.
//
// Example:
//
//	func install(ctx context.Context, tc *pocket.TaskContext) error {
//	    tc.Out.Printf("Installing govulncheck %s...\n", version)
//	    _, err := pocket.GoInstall(ctx, tc, "golang.org/x/vuln/cmd/govulncheck", version)
//	    return err
//	}
func GoInstall(ctx context.Context, tc *TaskContext, pkg, version string) (string, error) {
	// Determine binary name from package path.
	binaryName := goBinaryName(pkg)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Destination directory: .pocket/tools/go/<pkg>/<version>/
	toolDir := FromToolsDir("go", pkg, version)
	binaryPath := filepath.Join(toolDir, binaryName)

	// Check if already installed.
	if _, err := os.Stat(binaryPath); err == nil {
		// Already installed, ensure symlink exists.
		if _, err := CreateSymlink(binaryPath); err != nil {
			return "", err
		}
		return binaryPath, nil
	}

	// Create tool directory.
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		return "", fmt.Errorf("create tool dir: %w", err)
	}

	// Run go install with GOBIN set.
	pkgWithVersion := pkg + "@" + version
	tc.Out.Printf("  go install %s\n", pkgWithVersion)

	cmd := tc.Command(ctx, "go", "install", pkgWithVersion)
	cmd.Env = append(cmd.Environ(), "GOBIN="+toolDir)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go install %s: %w", pkgWithVersion, err)
	}

	// Create symlink.
	if _, err := CreateSymlink(binaryPath); err != nil {
		return "", err
	}

	return binaryPath, nil
}

// CreateSymlink creates a symlink in .pocket/bin/ pointing to the given binary.
// On Windows, it copies the file instead since symlinks require admin privileges.
// Returns the path to the symlink (or copy on Windows).
func CreateSymlink(binaryPath string) (string, error) {
	binDir := FromBinDir()
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", fmt.Errorf("create bin dir: %w", err)
	}

	// Ensure tools/go.mod exists to prevent go mod tidy issues.
	if err := ensureToolsGoMod(); err != nil {
		return "", err
	}

	name := filepath.Base(binaryPath)
	linkPath := filepath.Join(binDir, name)

	// Remove existing file/symlink if it exists.
	if _, err := os.Lstat(linkPath); err == nil {
		if err := os.Remove(linkPath); err != nil {
			return "", fmt.Errorf("remove existing file: %w", err)
		}
	}

	// On Windows, copy the file instead of creating a symlink.
	if runtime.GOOS == "windows" {
		if err := copyFile(binaryPath, linkPath); err != nil {
			return "", fmt.Errorf("copy binary: %w", err)
		}
		return linkPath, nil
	}

	// Create relative symlink on Unix.
	relPath, err := filepath.Rel(binDir, binaryPath)
	if err != nil {
		return "", fmt.Errorf("compute relative path: %w", err)
	}

	if err := os.Symlink(relPath, linkPath); err != nil {
		return "", fmt.Errorf("create symlink: %w", err)
	}

	return linkPath, nil
}

// VenvBinaryPath returns the cross-platform path to a binary in a Python venv.
func VenvBinaryPath(venvDir, name string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvDir, "Scripts", name+".exe")
	}
	return filepath.Join(venvDir, "bin", name)
}

// goBinaryName extracts the binary name from a Go package path.
func goBinaryName(pkg string) string {
	parts := strings.Split(pkg, "/")
	// If path ends with /cmd/<name>, use <name>
	if len(parts) >= 2 && parts[len(parts)-2] == "cmd" {
		return parts[len(parts)-1]
	}
	// Otherwise use last non-version component
	for i := len(parts) - 1; i >= 0; i-- {
		if !strings.HasPrefix(parts[i], "v") || !isGoVersion(parts[i]) {
			return parts[i]
		}
	}
	return parts[len(parts)-1]
}

func isGoVersion(s string) bool {
	if len(s) < 2 || s[0] != 'v' {
		return false
	}
	for _, c := range s[1:] {
		if c != '.' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

func extractFile(path string, opts DownloadOpts) error {
	switch opts.Format {
	case "tar.gz":
		return extractTarGz(path, opts.DestDir, opts.ExtractFiles)
	case "tar":
		return extractTar(path, opts.DestDir, opts.ExtractFiles)
	case "zip":
		return extractZip(path, opts.DestDir, opts.ExtractFiles)
	default:
		// Just copy the file.
		if opts.DestDir != "" {
			dst := filepath.Join(opts.DestDir, filepath.Base(path))
			return copyFile(path, dst)
		}
		return nil
	}
}

func extractTarGz(src, destDir string, extractOnly []string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	return extractTarReader(tar.NewReader(gzr), destDir, extractOnly)
}

func extractTar(src, destDir string, extractOnly []string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return extractTarReader(tar.NewReader(f), destDir, extractOnly)
}

func extractTarReader(tr *tar.Reader, destDir string, extractOnly []string) error {
	extractSet := make(map[string]bool)
	for _, name := range extractOnly {
		extractSet[name] = true
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := header.Name
		baseName := filepath.Base(name)

		// If extractOnly is set, only extract matching files (flattened).
		if len(extractOnly) > 0 {
			if !extractSet[baseName] {
				continue
			}
			name = baseName
		}

		target := filepath.Join(destDir, name)

		// Security check: ensure we don't escape destDir.
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path: %s", name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if len(extractOnly) > 0 {
				continue
			}
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func extractZip(src, destDir string, extractOnly []string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	extractSet := make(map[string]bool)
	for _, name := range extractOnly {
		extractSet[name] = true
	}

	for _, f := range r.File {
		name := f.Name
		baseName := filepath.Base(name)

		// If extractOnly is set, only extract matching files (flattened).
		if len(extractOnly) > 0 {
			if !extractSet[baseName] {
				continue
			}
			name = baseName
		}

		target := filepath.Join(destDir, name)

		// Security check.
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path: %s", name)
		}

		if f.FileInfo().IsDir() {
			if len(extractOnly) > 0 {
				continue
			}
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Chmod(0o755)
}

// ensureToolsGoMod creates .pocket/tools/go.mod if it doesn't exist.
// This prevents go mod tidy from scanning downloaded tools which may
// contain test files with relative imports that break module mode.
func ensureToolsGoMod() error {
	toolsDir := FromToolsDir()
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		return fmt.Errorf("create tools dir: %w", err)
	}

	goModPath := filepath.Join(toolsDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		return nil // Already exists
	}

	// Read Go version from .pocket/go.mod
	goVersion, err := GoVersionFromDir(FromPocketDir())
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`// This file prevents go mod tidy from scanning downloaded tools.
// Downloaded tools (like Go SDK) contain test files with relative imports
// that break module mode.

module pocket-tools

go %s
`, goVersion)

	if err := os.WriteFile(goModPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write tools/go.mod: %w", err)
	}
	return nil
}
