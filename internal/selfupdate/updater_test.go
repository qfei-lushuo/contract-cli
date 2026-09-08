package selfupdate

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGlobalBinaryFor(t *testing.T) {
	root := filepath.Join(t.TempDir(), "global", "node_modules")
	name := "contract-cli"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	global := filepath.Join(root, "@qfeius", "contract-cli", "bin", name)
	makeBinary := func(path string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("binary"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	makeBinary(global)
	if path, ok := globalBinaryFor(global, root); !ok || path != global {
		t.Fatalf("global match: %q %t", path, ok)
	}
	for _, dir := range []string{"project", "_npx/cache", "other-global", ".pnpm/local"} {
		local := filepath.Join(t.TempDir(), dir, "node_modules", "@qfeius", "contract-cli", "bin", name)
		makeBinary(local)
		if _, ok := globalBinaryFor(local, root); ok {
			t.Fatalf("accepted non-global binary %s", local)
		}
	}
	for _, invalid := range []string{"", "relative/node_modules", root + "\nwarning", filepath.Join(root, "missing")} {
		if _, ok := globalBinaryFor(global, invalid); ok {
			t.Fatalf("accepted root %q", invalid)
		}
	}
	t.Run("pnpm global symlink", func(t *testing.T) {
		store := filepath.Join(t.TempDir(), "pnpm", "store", "pkg")
		binary := filepath.Join(store, "bin", name)
		makeBinary(binary)
		linkedRoot := filepath.Join(t.TempDir(), "global", "node_modules")
		if err := os.MkdirAll(filepath.Join(linkedRoot, "@qfeius"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(store, filepath.Join(linkedRoot, "@qfeius", "contract-cli")); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		stable, ok := globalBinaryFor(binary, linkedRoot)
		if !ok || stable != filepath.Join(linkedRoot, "@qfeius", "contract-cli", "bin", name) {
			t.Fatalf("symlink match: %s %t", stable, ok)
		}
	})
}

func TestDetectOverrideRunsForEachExplicitDetection(t *testing.T) {
	calls := 0
	updater := &Updater{DetectOverride: func() DetectResult {
		calls++
		return DetectResult{Method: InstallNpm, NpmAvailable: true}
	}}
	// Overrides deliberately remain uncached because tests may alter them.
	updater.DetectInstallMethod()
	updater.DetectInstallMethod()
	if calls != 2 {
		t.Fatalf("override calls = %d, want 2", calls)
	}
}

func TestVersionFromOutput(t *testing.T) {
	t.Parallel()
	actual, err := versionFromOutput([]byte("contract-cli version 1.8.4 (commit abc123, built 2026-09-02)\n"))
	if err != nil || actual != "1.8.4" {
		t.Fatalf("versionFromOutput() = %q, %v", actual, err)
	}
	if _, err := versionFromOutput([]byte("contract-cli unknown")); err == nil {
		t.Fatal("versionFromOutput(invalid) error = nil")
	}
}

func TestInstallOverridesReceiveExactVersion(t *testing.T) {
	t.Parallel()
	npmVersion := ""
	pnpmVersion := ""
	updater := &Updater{
		NpmInstallOverride: func(version string) *CommandResult {
			npmVersion = version
			return &CommandResult{}
		},
		PnpmInstallOverride: func(version string) *CommandResult {
			pnpmVersion = version
			return &CommandResult{Err: errors.New("expected")}
		},
	}
	if result := updater.RunNpmInstall("1.8.4"); result.Err != nil {
		t.Fatalf("RunNpmInstall() error = %v", result.Err)
	}
	if result := updater.RunPnpmInstall("1.8.4"); result.Err == nil {
		t.Fatal("RunPnpmInstall() error = nil")
	}
	if npmVersion != "1.8.4" || pnpmVersion != "1.8.4" {
		t.Fatalf("npm=%q pnpm=%q", npmVersion, pnpmVersion)
	}
}

func TestVerifyOverride(t *testing.T) {
	t.Parallel()
	want := errors.New("verify failed")
	updater := &Updater{VerifyOverride: func(version string) error {
		if version != "1.8.4" {
			t.Fatalf("version = %q", version)
		}
		return want
	}}
	if err := updater.VerifyBinary("1.8.4"); !errors.Is(err, want) {
		t.Fatalf("VerifyBinary() error = %v", err)
	}
}

func TestTruncateKeepsTail(t *testing.T) {
	t.Parallel()
	if got := Truncate("abcdef", 3); got != "def" {
		t.Fatalf("Truncate() = %q", got)
	}
}
