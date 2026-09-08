package selfupdate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const (
	NpmPackage        = "@qfeius/contract-cli"
	installTimeout    = 10 * time.Minute
	verificationLimit = 10 * time.Second
	detectionTimeout  = 5 * time.Second
)

type InstallMethod string

const (
	InstallNpm    InstallMethod = "npm"
	InstallPnpm   InstallMethod = "pnpm"
	InstallManual InstallMethod = "manual"
)

type DetectResult struct {
	Method           InstallMethod `json:"method"`
	ResolvedPath     string        `json:"resolved_path,omitempty"`
	GlobalBinaryPath string        `json:"global_binary_path,omitempty"`
	NpmAvailable     bool          `json:"npm_available,omitempty"`
	PnpmAvailable    bool          `json:"pnpm_available,omitempty"`
}

func (d DetectResult) CanAutoUpdate() bool {
	switch d.Method {
	case InstallNpm:
		return d.NpmAvailable
	case InstallPnpm:
		return d.PnpmAvailable
	default:
		return false
	}
}

func (d DetectResult) ManualReason() string {
	switch {
	case d.Method == InstallNpm && !d.NpmAvailable:
		return "installed via npm, but npm is not available in PATH"
	case d.Method == InstallPnpm && !d.PnpmAvailable:
		return "installed via pnpm, but pnpm is not available in PATH"
	default:
		return "current binary is not a verified global npm or pnpm installation; update it through its original installer"
	}
}

type CommandResult struct {
	Stdout bytes.Buffer
	Stderr bytes.Buffer
	Err    error
}

func (r *CommandResult) CombinedOutput() string {
	if r == nil {
		return ""
	}
	return r.Stdout.String() + r.Stderr.String()
}

// Updater contains the mutation part of self-update. The override fields are
// intentionally small seams used by unit tests; production callers use New.
type Updater struct {
	DetectOverride      func() DetectResult
	NpmInstallOverride  func(string) *CommandResult
	PnpmInstallOverride func(string) *CommandResult
	VerifyOverride      func(string) error

	detectCache   *DetectResult
	backupCreated bool
}

func New() *Updater { return &Updater{} }

func (u *Updater) DetectInstallMethod() DetectResult {
	if u.DetectOverride != nil {
		return u.DetectOverride()
	}
	if u.detectCache != nil {
		return *u.detectCache
	}
	result := detectInstallMethod()
	u.detectCache = &result
	return result
}

func detectInstallMethod() DetectResult {
	executable, err := os.Executable()
	if err != nil {
		return DetectResult{Method: InstallManual}
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return DetectResult{Method: InstallManual, ResolvedPath: executable}
	}
	// Inspect the manager's current global root, not a guessed node_modules path.
	// Resolve package symlinks so pnpm's store layout is supported too.
	managers := []InstallMethod{InstallNpm, InstallPnpm}
	if containsPnpmMarker(resolved) {
		managers[0], managers[1] = managers[1], managers[0]
	}
	ctx, cancel := context.WithTimeout(context.Background(), detectionTimeout)
	defer cancel()
	for _, manager := range managers {
		result := runPackageManagerContext(ctx, string(manager), []string{"root", "-g"})
		if result.Err != nil {
			continue
		}
		root := strings.TrimSpace(result.Stdout.String())
		if candidate, ok := globalBinaryFor(resolved, root); ok {
			return DetectResult{Method: manager, ResolvedPath: resolved, GlobalBinaryPath: candidate,
				NpmAvailable: manager == InstallNpm, PnpmAvailable: manager == InstallPnpm}
		}
	}
	return DetectResult{Method: InstallManual, ResolvedPath: resolved}
}

func globalBinaryFor(executable, root string) (string, bool) {
	if !filepath.IsAbs(root) {
		return "", false
	}
	binary := "contract-cli"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	candidate := filepath.Join(root, "@qfeius", "contract-cli", "bin", binary)
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", false
	}
	current, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", false
	}
	equal := filepath.Clean(current) == filepath.Clean(resolved)
	if runtime.GOOS == "windows" {
		equal = strings.EqualFold(filepath.Clean(current), filepath.Clean(resolved))
	}
	return candidate, equal
}

func containsPnpmMarker(path string) bool {
	parts := strings.Split(strings.ReplaceAll(path, `\`, "/"), "/")
	for index, part := range parts {
		if part == ".pnpm" {
			return true
		}
		if part == "pnpm" && index+1 < len(parts) && parts[index+1] == "store" {
			return true
		}
	}
	return false
}

func (u *Updater) RunNpmInstall(version string) *CommandResult {
	if u.NpmInstallOverride != nil {
		return u.NpmInstallOverride(version)
	}
	return runPackageManager("npm", []string{"install", "-g", NpmPackage + "@" + version})
}

func (u *Updater) RunPnpmInstall(version string) *CommandResult {
	if u.PnpmInstallOverride != nil {
		return u.PnpmInstallOverride(version)
	}
	return runPackageManager("pnpm", []string{"add", "-g", NpmPackage + "@" + version})
}

func runPackageManager(name string, args []string) *CommandResult {
	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()
	return runPackageManagerContext(ctx, name, args)
}

func runPackageManagerContext(ctx context.Context, name string, args []string) *CommandResult {
	result := &CommandResult{}
	path, err := exec.LookPath(name)
	if err != nil {
		result.Err = fmt.Errorf("%s not found in PATH: %w", name, err)
		return result
	}
	command, err := packageManagerCommand(ctx, path, args)
	if err != nil {
		result.Err = err
		return result
	}
	command.Stdout = &result.Stdout
	command.Stderr = &result.Stderr
	command.WaitDelay = time.Second
	result.Err = command.Run()
	if ctx.Err() != nil {
		result.Err = fmt.Errorf("%s command cancelled: %w", name, ctx.Err())
	}
	return result
}

var versionOutputPattern = regexp.MustCompile(`(?i)\bversion\s+v?([^\s()]+)`)

func (u *Updater) VerifyBinary(expectedVersion string) error {
	if u.VerifyOverride != nil {
		return u.VerifyOverride(expectedVersion)
	}
	executable := ""
	if u.detectCache != nil {
		executable = u.detectCache.GlobalBinaryPath
	}
	// Prefer the native binary path detected before the package manager mutates
	// the installation. This also avoids a Windows .cmd wrapper deleting the
	// rollback file before the exact version comparison has succeeded.
	if executable == "" && u.detectCache != nil && strings.TrimSpace(u.detectCache.ResolvedPath) != "" {
		if _, err := os.Stat(u.detectCache.ResolvedPath); err == nil {
			executable = u.detectCache.ResolvedPath
		}
	}
	if executable == "" {
		executable, _ = exec.LookPath("contract-cli")
	}
	if executable == "" {
		var err error
		executable, err = os.Executable()
		if err != nil {
			return fmt.Errorf("cannot locate contract-cli: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), verificationLimit)
	defer cancel()
	output, err := exec.CommandContext(ctx, executable, "--version").Output()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("binary verification timed out after %s", verificationLimit)
	}
	if err != nil {
		return fmt.Errorf("binary not executable: %w", err)
	}
	actual, err := versionFromOutput(output)
	if err != nil {
		return err
	}
	expected := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(expectedVersion), "v"), "V")
	if actual != expected {
		return fmt.Errorf("expected version %s, got %q", expectedVersion, actual)
	}
	return nil
}

func versionFromOutput(output []byte) (string, error) {
	trimmed := strings.TrimSpace(string(output))
	match := versionOutputPattern.FindStringSubmatch(trimmed)
	if len(match) != 2 {
		return "", fmt.Errorf("unrecognized version output %q", trimmed)
	}
	return strings.TrimPrefix(strings.TrimPrefix(match[1], "v"), "V"), nil
}

func (u *Updater) resolveExecutable() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(executable)
}

func Truncate(value string, maxRunes int) string {
	runes := []rune(value)
	if maxRunes <= 0 {
		return ""
	}
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[len(runes)-maxRunes:])
}
