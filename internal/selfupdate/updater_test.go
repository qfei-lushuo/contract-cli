package selfupdate

import (
	"errors"
	"testing"
)

func TestDetectFromResolvedPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		npm     bool
		pnpm    bool
		method  InstallMethod
		canAuto bool
	}{
		{name: "npm macOS", path: "/usr/local/lib/node_modules/@qfeius/contract-cli/bin/contract-cli", npm: true, method: InstallNpm, canAuto: true},
		{name: "pnpm Linux", path: "/home/user/.local/share/pnpm/store/v10/links/pkg/node_modules/@qfeius/contract-cli/bin/contract-cli", pnpm: true, method: InstallPnpm, canAuto: true},
		{name: "pnpm Windows", path: `C:\Users\Lucas\AppData\Local\pnpm\store\v10\links\pkg\node_modules\@qfeius\contract-cli\bin\contract-cli.exe`, pnpm: true, method: InstallPnpm, canAuto: true},
		{name: "npm Windows", path: `C:\Users\Lucas\AppData\Roaming\npm\node_modules\@qfeius\contract-cli\bin\contract-cli.exe`, npm: true, method: InstallNpm, canAuto: true},
		{name: "manual", path: "/usr/local/bin/contract-cli", npm: true, pnpm: true, method: InstallManual, canAuto: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := DetectFromResolvedPath(test.path, test.npm, test.pnpm)
			if result.Method != test.method || result.CanAutoUpdate() != test.canAuto {
				t.Fatalf("DetectFromResolvedPath() = %+v", result)
			}
		})
	}
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
