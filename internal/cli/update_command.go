package cli

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cn.qfei/contract-cli/internal/build"
	"cn.qfei/contract-cli/internal/output"
	"cn.qfei/contract-cli/internal/selfupdate"
	updatecheck "cn.qfei/contract-cli/internal/update"
)

const (
	contractCLIRepository = "https://github.com/qfeius/contract-cli"
	maxUpdateOutput       = 2000
	automaticUpdateBudget = 1500 * time.Millisecond
)

func (a *App) runUpdate(ctx context.Context, args []string) error {
	// Keep the old `update check` spelling as a hidden compatibility alias.
	legacyCheck := len(args) > 0 && args[0] == "check"
	if legacyCheck {
		args = args[1:]
	}
	parsed, err := parseArgs(args, map[string]struct{}{
		"--channel": {}, // hidden migration compatibility; latest only
	}, map[string]struct{}{
		"--check": {},
		"--force": {},
		"--json":  {},
	})
	if err != nil {
		return err
	}
	if len(parsed.positionals) != 0 {
		return fmt.Errorf("update does not accept positional arguments: %s", strings.Join(parsed.positionals, " "))
	}
	if channel := strings.TrimSpace(parsed.String("--channel")); channel != "" && channel != updatecheck.LatestChannel {
		return fmt.Errorf("--channel %q is no longer supported; contract-cli update follows npm latest", channel)
	}

	checkOnly := legacyCheck || parsed.Bool("--check")
	jsonOutput := parsed.Bool("--json")
	force := parsed.Bool("--force")
	result, err := a.checkUpdate(ctx)
	if err != nil {
		return fmt.Errorf("check latest contract-cli version: %w", err)
	}
	if !result.Skipped {
		a.saveUpdateCache(result)
	}

	updater := a.newSelfUpdater()
	if !checkOnly {
		updater.CleanupStaleFiles()
	}
	detection := selfupdate.DetectResult{Method: selfupdate.InstallManual}
	if result.UpdateAvailable || force {
		detection = updater.DetectInstallMethod()
	}

	if checkOnly || result.Skipped {
		return a.reportUpdateCheck(result, detection, jsonOutput)
	}
	if !result.UpdateAvailable && !force {
		return a.reportAlreadyUpToDate(result, jsonOutput)
	}
	if !detection.CanAutoUpdate() {
		return a.reportManualUpdate(result, detection, jsonOutput)
	}
	return a.performAutomaticUpdate(result, detection, updater, jsonOutput)
}

func (a *App) reportUpdateCheck(result updatecheck.Result, detection selfupdate.DetectResult, jsonOutput bool) error {
	if jsonOutput {
		data := updateResultJSON(result)
		if result.UpdateAvailable {
			data["auto_update"] = detection.CanAutoUpdate()
			data["url"] = updateReleaseURL(result.LatestVersion)
			data["changelog"] = updateChangelogURL()
		}
		return output.NewRenderer(a.stdout).Render(output.FormatJSON, data)
	}
	if result.Skipped {
		_, err := fmt.Fprintf(a.stderr, "Update check skipped: %s\n", result.Reason)
		return err
	}
	if !result.UpdateAvailable {
		_, err := fmt.Fprintf(a.stderr, "contract-cli %s is already up to date\n", result.CurrentVersion)
		return err
	}
	if _, err := fmt.Fprintf(a.stderr, "Update available: %s -> %s\n", result.CurrentVersion, result.LatestVersion); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(a.stderr, "  Release:   %s\n", updateReleaseURL(result.LatestVersion))
	_, _ = fmt.Fprintf(a.stderr, "  Changelog: %s\n", updateChangelogURL())
	if detection.CanAutoUpdate() {
		_, _ = fmt.Fprintln(a.stderr, "\nRun `contract-cli update` to install.")
	} else {
		_, _ = fmt.Fprintln(a.stderr, "\nDownload the release above to update manually.")
	}
	return nil
}

func (a *App) reportAlreadyUpToDate(result updatecheck.Result, jsonOutput bool) error {
	if jsonOutput {
		return output.NewRenderer(a.stdout).Render(output.FormatJSON, updateResultJSON(result))
	}
	_, err := fmt.Fprintf(a.stderr, "contract-cli %s is already up to date\n", result.CurrentVersion)
	return err
}

func (a *App) reportManualUpdate(result updatecheck.Result, detection selfupdate.DetectResult, jsonOutput bool) error {
	message := fmt.Sprintf("automatic update unavailable: %s", detection.ManualReason())
	if jsonOutput {
		data := updateResultJSON(result)
		data["action"] = "manual_required"
		data["auto_update"] = false
		data["message"] = message
		data["url"] = updateReleaseURL(result.LatestVersion)
		data["changelog"] = updateChangelogURL()
		data["install_method"] = detection.Method
		if detection.ResolvedPath != "" {
			data["resolved_path"] = detection.ResolvedPath
		}
		return output.NewRenderer(a.stdout).Render(output.FormatJSON, data)
	}
	_, _ = fmt.Fprintf(a.stderr, "Automatic update unavailable: %s", detection.ManualReason())
	if detection.ResolvedPath != "" {
		_, _ = fmt.Fprintf(a.stderr, " (path: %s)", detection.ResolvedPath)
	}
	_, _ = fmt.Fprintln(a.stderr, ".")
	_, _ = fmt.Fprintf(a.stderr, "Release: %s\n", updateReleaseURL(result.LatestVersion))
	if detection.Method == selfupdate.InstallManual {
		_, _ = fmt.Fprintln(a.stderr, "Update this copy through its original installer or project dependency.")
	} else if detection.Method == selfupdate.InstallPnpm {
		_, _ = fmt.Fprintf(a.stderr, "Or run: pnpm add -g %s@%s\n", selfupdate.NpmPackage, result.LatestVersion)
	} else {
		_, _ = fmt.Fprintf(a.stderr, "Or run: npm install -g %s@%s\n", selfupdate.NpmPackage, result.LatestVersion)
	}
	return nil
}

func (a *App) performAutomaticUpdate(result updatecheck.Result, detection selfupdate.DetectResult, updater *selfupdate.Updater, jsonOutput bool) error {
	manager := "npm"
	install := updater.RunNpmInstall
	if detection.Method == selfupdate.InstallPnpm {
		manager = "pnpm"
		install = updater.RunPnpmInstall
	}
	restore, err := updater.PrepareSelfReplace()
	if err != nil {
		return fmt.Errorf("prepare contract-cli update: %w", err)
	}
	if !jsonOutput {
		_, _ = fmt.Fprintf(a.stderr, "Updating contract-cli %s -> %s via %s ...\n", result.CurrentVersion, result.LatestVersion, manager)
	}
	installResult := install(result.LatestVersion)
	if installResult.Err != nil {
		restore()
		detail := strings.TrimSpace(selfupdate.Truncate(installResult.CombinedOutput(), maxUpdateOutput))
		if detail != "" {
			return fmt.Errorf("%s install failed: %w: %s", manager, installResult.Err, detail)
		}
		return fmt.Errorf("%s install failed: %w", manager, installResult.Err)
	}
	if err := updater.VerifyBinary(result.LatestVersion); err != nil {
		restore()
		if updater.CanRestorePreviousVersion() {
			return fmt.Errorf("new binary verification failed and previous version was restored: %w", err)
		}
		return fmt.Errorf("new binary verification failed: %w", err)
	}
	updater.CleanupStaleFiles()

	updated := result
	updated.CurrentVersion = result.LatestVersion
	updated.UpdateAvailable = false
	updated.InstallCommand = updatecheck.UpdateCommand
	a.saveUpdateCache(updated)
	if jsonOutput {
		data := updateResultJSON(updated)
		data["previous_version"] = result.CurrentVersion
		data["action"] = "updated"
		data["message"] = fmt.Sprintf("contract-cli updated from %s to %s", result.CurrentVersion, result.LatestVersion)
		data["url"] = updateReleaseURL(result.LatestVersion)
		data["changelog"] = updateChangelogURL()
		return output.NewRenderer(a.stdout).Render(output.FormatJSON, data)
	}
	_, _ = fmt.Fprintf(a.stderr, "Successfully updated contract-cli from %s to %s\n", result.CurrentVersion, result.LatestVersion)
	_, _ = fmt.Fprintf(a.stderr, "Changelog: %s\n", updateChangelogURL())
	return nil
}

func updateResultJSON(result updatecheck.Result) map[string]any {
	action := "already_up_to_date"
	message := fmt.Sprintf("contract-cli %s is already up to date", result.CurrentVersion)
	if result.UpdateAvailable {
		action = "update_available"
		message = fmt.Sprintf("contract-cli %s -> %s available", result.CurrentVersion, result.LatestVersion)
	}
	if result.Skipped {
		action = "skipped"
		message = result.Reason
	}
	data := map[string]any{
		"ok":               true,
		"package":          result.PackageName,
		"previous_version": result.CurrentVersion,
		"current_version":  result.CurrentVersion,
		"latest_version":   result.LatestVersion,
		"action":           action,
		"message":          message,
	}
	if result.Skipped {
		data["reason"] = result.Reason
	}
	return data
}

func updateReleaseURL(version string) string {
	return contractCLIRepository + "/releases/tag/v" + strings.TrimPrefix(strings.TrimSpace(version), "v")
}

func updateChangelogURL() string { return contractCLIRepository + "/blob/master/CHANGELOG.md" }

func (a *App) maybePrepareUpdateNotice(ctx context.Context, args []string) func() {
	noop := func() {}
	if !a.shouldAutoCheckUpdate(args) {
		return noop
	}
	runID := a.currentUpdateRunID()
	currentVersion := a.currentUpdateVersion()
	now := a.now()
	cache, cacheOK, err := updatecheck.LoadCache(a.updateCachePath())
	if err != nil {
		a.logger.Debug("load update check cache failed", "path", a.updateCachePath(), "error", err.Error())
	}
	if cacheOK && strings.TrimSpace(cache.Channel) == updatecheck.LatestChannel {
		if notice := updatecheck.NoticeFromCache(cache, currentVersion); notice != nil {
			a.setUpdateNotice(runID, map[string]any{"update": notice.Map()})
		}
		if updatecheck.CacheFresh(cache, now, updatecheck.CacheTTL) {
			return noop
		}
	}

	// Business execution proceeds immediately. Join the optional check before
	// process exit, within one budget measured from the start of the check.
	checkCtx, cancel := context.WithTimeout(ctx, automaticUpdateBudget)
	type checked struct {
		result updatecheck.Result
		err    error
	}
	done := make(chan checked, 1)
	go func() {
		result, err := a.checkUpdateVersionWithLogger(checkCtx, currentVersion, nil)
		done <- checked{result: result, err: err}
	}()
	return func() {
		defer cancel()
		var check checked
		// A long business command can finish after the budget. Prefer a result
		// already obtained within that budget over the now-expired context.
		select {
		case check = <-done:
		default:
			select {
			case check = <-done:
			case <-checkCtx.Done():
				return
			}
		}
		if check.err != nil {
			a.logger.Debug("automatic update cache refresh failed", "error", check.err.Error())
			return
		}
		if !check.result.Skipped {
			a.saveUpdateCache(check.result)
		}
	}
}

func (a *App) shouldAutoCheckUpdate(args []string) bool {
	if len(args) == 0 || isHelpRequest(args) {
		return false
	}
	switch args[0] {
	case "version", "--version", "-version", "-v", "update":
		return false
	}
	for _, key := range []string{"CONTRACT_CLI_NO_UPDATE_NOTIFIER", "CONTRACT_CLI_NO_UPDATE_CHECK"} {
		if value, ok := a.lookupEnv(key); ok && truthy(value) {
			return false
		}
	}
	return !isCIUpdateEnv(a.lookupEnv)
}

func (a *App) checkUpdate(ctx context.Context) (updatecheck.Result, error) {
	return a.checkUpdateVersionWithLogger(ctx, a.currentUpdateVersion(), a.logger)
}

func (a *App) checkUpdateVersionWithLogger(ctx context.Context, currentVersion string, logger *slog.Logger) (updatecheck.Result, error) {
	return updatecheck.Check(ctx, updatecheck.Options{
		HTTPClient:     a.httpClient,
		Logger:         logger,
		RegistryURL:    a.updateURL,
		PackageName:    updatecheck.DefaultPackageName,
		CurrentVersion: currentVersion,
		Now:            a.now,
	})
}

func (a *App) saveUpdateCache(result updatecheck.Result) {
	if err := updatecheck.SaveCache(a.updateCachePath(), updatecheck.CacheFromResult(result)); err != nil {
		a.logger.Warn("save update check cache failed", "path", a.updateCachePath(), "error", err.Error())
	}
}

func (a *App) currentUpdateVersion() string {
	if strings.TrimSpace(a.updateVersion) != "" {
		return strings.TrimSpace(a.updateVersion)
	}
	return build.Current().Version
}

func truthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func isCIUpdateEnv(lookup func(string) (string, bool)) bool {
	for _, key := range []string{"CI", "BUILD_NUMBER", "RUN_ID"} {
		if value, ok := lookup(key); ok && strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}
