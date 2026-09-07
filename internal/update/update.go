package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultPackageName = "@qfeius/contract-cli"
	DefaultRegistryURL = "https://registry.npmjs.org/@qfeius%2fcontract-cli/latest"
	CacheTTL           = 24 * time.Hour
	FetchTimeout       = 15 * time.Second
	LatestChannel      = "latest"
	UpdateCommand      = "contract-cli update"
)

type Options struct {
	HTTPClient     *http.Client
	Logger         *slog.Logger
	RegistryURL    string
	PackageName    string
	CurrentVersion string
	Now            func() time.Time
}

type Result struct {
	CheckedAt       time.Time
	PackageName     string
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	InstallCommand  string
	Skipped         bool
	Reason          string
}

type Notice struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Message string `json:"message"`
	Command string `json:"command"`
}

type Cache struct {
	CheckedAt       time.Time `json:"checked_at"`
	Channel         string    `json:"channel"`
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	InstallCommand  string    `json:"install_command,omitempty"`
}

func Check(ctx context.Context, options Options) (Result, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, FetchTimeout)
		defer cancel()
	}
	now := optionNow(options.Now)
	packageName := defaultString(options.PackageName, DefaultPackageName)
	currentVersion := strings.TrimSpace(options.CurrentVersion)
	result := Result{
		CheckedAt:      now,
		PackageName:    packageName,
		CurrentVersion: currentVersion,
	}

	logger := options.Logger
	if logger != nil {
		logger.Info("update check started", "package", packageName, "current_version", emptyFallback(currentVersion, "<empty>"), "channel", LatestChannel)
	}

	if shouldSkipVersion(currentVersion) {
		result.Skipped = true
		result.Reason = "current version is a local/dev build; skip remote update check"
		if logger != nil {
			logger.Info("update check skipped", "reason", result.Reason)
		}
		return result, nil
	}
	if _, err := parseSemver(currentVersion); err != nil {
		result.Skipped = true
		result.Reason = "current version is not semantic version; skip remote update check"
		if logger != nil {
			logger.Info("update check skipped", "reason", result.Reason, "current_version", currentVersion)
		}
		return result, nil
	}

	registryURL := defaultString(options.RegistryURL, DefaultRegistryURL)
	client := options.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, registryURL, nil)
	if err != nil {
		if logger != nil {
			logger.Error("build update check request failed", "registry_url", registryURL, "error", err.Error())
		}
		return Result{}, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if logger != nil {
			logger.Error("send update check request failed", "registry_url", registryURL, "error", err.Error())
		}
		return Result{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		if logger != nil {
			logger.Error("read update check response failed", "status_code", resp.StatusCode, "error", err.Error())
		}
		return Result{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("npm registry returned status %d: %s", resp.StatusCode, responseSummary(body))
		if logger != nil {
			logger.Error("update check registry status failed", "status_code", resp.StatusCode, "error", err.Error())
		}
		return Result{}, err
	}

	var metadata struct {
		Version  string            `json:"version"`
		DistTags map[string]string `json:"dist-tags"`
	}
	if err := json.Unmarshal(body, &metadata); err != nil {
		if logger != nil {
			logger.Error("decode update check response failed", "error", err.Error())
		}
		return Result{}, err
	}
	latestVersion := strings.TrimSpace(metadata.Version)
	if latestVersion == "" {
		// Accept a full npm packument as a compatibility fallback for private
		// registries and older test fixtures. Production uses the /latest
		// endpoint, matching lark-cli's fixed latest-only update model.
		latestVersion = strings.TrimSpace(metadata.DistTags[LatestChannel])
	}
	if latestVersion == "" {
		err := fmt.Errorf("npm latest version not found for %s", packageName)
		if logger != nil {
			logger.Error("update check dist-tag missing", "package", packageName, "channel", LatestChannel, "error", err.Error())
		}
		return Result{}, err
	}

	compare, err := CompareSemver(currentVersion, latestVersion)
	if err != nil {
		if logger != nil {
			logger.Error("compare update versions failed", "current_version", currentVersion, "latest_version", latestVersion, "error", err.Error())
		}
		return Result{}, err
	}

	result.LatestVersion = latestVersion
	result.UpdateAvailable = compare < 0
	result.InstallCommand = UpdateCommand
	if logger != nil {
		logger.Info("update check completed", "package", packageName, "current_version", currentVersion, "latest_version", latestVersion, "channel", LatestChannel, "update_available", result.UpdateAvailable)
	}
	return result, nil
}

func CompareSemver(a, b string) (int, error) {
	left, err := parseSemver(a)
	if err != nil {
		return 0, err
	}
	right, err := parseSemver(b)
	if err != nil {
		return 0, err
	}

	for i := range left.core {
		if left.core[i] < right.core[i] {
			return -1, nil
		}
		if left.core[i] > right.core[i] {
			return 1, nil
		}
	}
	return comparePrerelease(left.prerelease, right.prerelease), nil
}

func LoadCache(path string) (Cache, bool, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Cache{}, false, nil
	}
	if err != nil {
		return Cache{}, false, err
	}

	var cache Cache
	if err := json.Unmarshal(content, &cache); err != nil {
		return Cache{}, false, err
	}
	return cache, true, nil
}

func SaveCache(path string, cache Cache) error {
	content, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteFile(path, append(content, '\n'), 0o600)
}

func CacheFromResult(result Result) Cache {
	return Cache{
		CheckedAt:       result.CheckedAt,
		Channel:         LatestChannel,
		CurrentVersion:  result.CurrentVersion,
		LatestVersion:   result.LatestVersion,
		UpdateAvailable: result.UpdateAvailable,
		InstallCommand:  result.InstallCommand,
	}
}

func CacheFresh(cache Cache, now time.Time, ttl time.Duration) bool {
	if strings.TrimSpace(cache.Channel) != LatestChannel {
		return false
	}
	if cache.CheckedAt.IsZero() {
		return false
	}
	if ttl <= 0 {
		return false
	}
	return now.Before(cache.CheckedAt.Add(ttl))
}

func NoticeFromResult(result Result) *Notice {
	if result.Skipped || !result.UpdateAvailable {
		return nil
	}
	return newNotice(result.CurrentVersion, result.LatestVersion, result.InstallCommand)
}

func NoticeFromCache(cache Cache, currentVersion string) *Notice {
	currentVersion = strings.TrimSpace(currentVersion)
	latestVersion := strings.TrimSpace(cache.LatestVersion)
	if shouldSkipVersion(currentVersion) || latestVersion == "" {
		return nil
	}
	if _, err := parseSemver(currentVersion); err != nil {
		return nil
	}
	compare, err := CompareSemver(currentVersion, latestVersion)
	if err != nil || compare >= 0 {
		return nil
	}
	command := strings.TrimSpace(cache.InstallCommand)
	if command == "" {
		command = UpdateCommand
	}
	return newNotice(currentVersion, latestVersion, command)
}

func (n *Notice) Map() map[string]any {
	if n == nil {
		return nil
	}
	return map[string]any{
		"current": n.Current,
		"latest":  n.Latest,
		"message": n.Message,
		"command": n.Command,
	}
}

func newNotice(current string, latest string, command string) *Notice {
	return &Notice{
		Current: current,
		Latest:  latest,
		Message: fmt.Sprintf("contract-cli %s available, current %s, run: %s", latest, current, command),
		Command: command,
	}
}

type semver struct {
	core       [3]int
	prerelease []string
}

func parseSemver(version string) (semver, error) {
	normalized := strings.TrimSpace(strings.TrimPrefix(version, "v"))
	if normalized == "" {
		return semver{}, errors.New("empty semantic version")
	}
	if before, _, ok := strings.Cut(normalized, "+"); ok {
		normalized = before
	}

	corePart := normalized
	var prerelease []string
	if before, after, ok := strings.Cut(normalized, "-"); ok {
		corePart = before
		prerelease = strings.Split(after, ".")
	}

	parts := strings.Split(corePart, ".")
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("invalid semantic version %q", version)
	}

	var parsed semver
	parsed.prerelease = prerelease
	for i, part := range parts {
		if part == "" {
			return semver{}, fmt.Errorf("invalid semantic version %q", version)
		}
		number, err := strconv.Atoi(part)
		if err != nil {
			return semver{}, fmt.Errorf("invalid semantic version %q: %w", version, err)
		}
		parsed.core[i] = number
	}
	return parsed, nil
}

func comparePrerelease(a, b []string) int {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	if len(a) == 0 {
		return 1
	}
	if len(b) == 0 {
		return -1
	}

	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	for i := 0; i < limit; i++ {
		result := comparePrereleaseIdentifier(a[i], b[i])
		if result != 0 {
			return result
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func comparePrereleaseIdentifier(a, b string) int {
	aNumber, aErr := strconv.Atoi(a)
	bNumber, bErr := strconv.Atoi(b)
	aIsNumber := aErr == nil
	bIsNumber := bErr == nil

	switch {
	case aIsNumber && bIsNumber:
		if aNumber < bNumber {
			return -1
		}
		if aNumber > bNumber {
			return 1
		}
		return 0
	case aIsNumber:
		return -1
	case bIsNumber:
		return 1
	default:
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
		return 0
	}
}

func shouldSkipVersion(version string) bool {
	trimmed := strings.TrimSpace(version)
	switch strings.ToLower(trimmed) {
	case "", "dev", "unknown":
		return true
	default:
		return gitDescribeVersionPattern.MatchString(strings.TrimPrefix(trimmed, "v"))
	}
}

var gitDescribeVersionPattern = regexp.MustCompile(`-\d+-g[0-9a-fA-F]{7,}`)

func optionNow(now func() time.Time) time.Time {
	if now != nil {
		return now()
	}
	return time.Now()
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func responseSummary(body []byte) string {
	const limit = 512
	summary := strings.TrimSpace(string(body))
	if len(summary) > limit {
		return summary[:limit] + "..."
	}
	return summary
}
