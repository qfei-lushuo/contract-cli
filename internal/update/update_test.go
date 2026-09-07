package update

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCheckUsesLatestMetadataEndpoint(t *testing.T) {
	t.Parallel()
	result, err := Check(context.Background(), Options{
		HTTPClient:     jsonHTTPClient(`{"version":"1.8.4"}`),
		RegistryURL:    "https://registry.test/@qfeius%2fcontract-cli/latest",
		CurrentVersion: "1.8.3",
		Now:            fixedNow,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.UpdateAvailable || result.LatestVersion != "1.8.4" {
		t.Fatalf("result = %+v", result)
	}
	if result.InstallCommand != UpdateCommand {
		t.Fatalf("latest-only result = %+v", result)
	}
}

func TestCheckAcceptsPackumentLatestFallback(t *testing.T) {
	t.Parallel()
	result, err := Check(context.Background(), Options{
		HTTPClient:     jsonHTTPClient(`{"dist-tags":{"latest":"1.8.4","beta":"1.9.0-beta.1"}}`),
		RegistryURL:    "https://registry.test/@qfeius%2fcontract-cli",
		CurrentVersion: "1.8.3",
		Now:            fixedNow,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.LatestVersion != "1.8.4" {
		t.Fatalf("LatestVersion = %q", result.LatestVersion)
	}
}

func TestCheckReportsUpToDate(t *testing.T) {
	t.Parallel()
	result, err := Check(context.Background(), Options{
		HTTPClient:     jsonHTTPClient(`{"version":"1.8.3"}`),
		CurrentVersion: "1.8.3",
		Now:            fixedNow,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.UpdateAvailable {
		t.Fatal("UpdateAvailable = true, want false")
	}
}

func TestCheckSkipsNonSemanticCurrentVersion(t *testing.T) {
	t.Parallel()
	result, err := Check(context.Background(), Options{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("non-semver version should not request registry")
			return nil, nil
		})},
		CurrentVersion: "b4a2091-dirty",
		Now:            fixedNow,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Skipped || result.Reason == "" {
		t.Fatalf("result = %+v", result)
	}
}

func TestCheckSkipsGitDescribeBuild(t *testing.T) {
	t.Parallel()
	result, err := Check(context.Background(), Options{
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("git-describe build should not request registry")
			return nil, nil
		})},
		CurrentVersion: "v1.8.3-2-gabcdef1-dirty",
		Now:            fixedNow,
	})
	if err != nil || !result.Skipped {
		t.Fatalf("Check() result=%+v err=%v", result, err)
	}
}

func TestCompareSemanticVersionsWithPrerelease(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "beta increments", a: "0.1.0-beta.1", b: "0.1.0-beta.2", want: -1},
		{name: "release wins prerelease", a: "0.1.0-beta.2", b: "0.1.0", want: -1},
		{name: "equal", a: "v0.1.0", b: "0.1.0", want: 0},
		{name: "major wins", a: "1.0.0", b: "0.9.9", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CompareSemver(tt.a, tt.b)
			if err != nil || got != tt.want {
				t.Fatalf("CompareSemver(%q, %q) = %d, %v", tt.a, tt.b, got, err)
			}
		})
	}
}

func TestCacheRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "update-check.json")
	cache := Cache{
		CheckedAt:       fixedNow(),
		Channel:         LatestChannel,
		CurrentVersion:  "1.8.3",
		LatestVersion:   "1.8.4",
		UpdateAvailable: true,
		InstallCommand:  UpdateCommand,
	}
	if err := SaveCache(path, cache); err != nil {
		t.Fatalf("SaveCache() error = %v", err)
	}
	loaded, ok, err := LoadCache(path)
	if err != nil || !ok || loaded.LatestVersion != cache.LatestVersion {
		t.Fatalf("LoadCache() = %+v, %v, %v", loaded, ok, err)
	}
	if err := os.WriteFile(path, []byte(`{bad json`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadCache(path); err == nil {
		t.Fatal("LoadCache(bad JSON) error = nil")
	}
}

func TestConcurrentCacheWritesRemainValidJSON(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "update-check.json")
	var group sync.WaitGroup
	for index := 0; index < 20; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := SaveCache(path, Cache{
				CheckedAt:     fixedNow(),
				Channel:       LatestChannel,
				LatestVersion: "1.8.4",
			}); err != nil {
				t.Errorf("SaveCache() error = %v", err)
			}
		}()
	}
	group.Wait()
	if _, ok, err := LoadCache(path); err != nil || !ok {
		t.Fatalf("LoadCache() after concurrent writes ok=%v err=%v", ok, err)
	}
}

func TestCacheFreshnessUsesLatestAndTTL(t *testing.T) {
	t.Parallel()
	cache := Cache{CheckedAt: fixedNow().Add(-23 * time.Hour), Channel: LatestChannel}
	if !CacheFresh(cache, fixedNow(), CacheTTL) {
		t.Fatal("fresh latest cache rejected")
	}
	if CacheFresh(Cache{CheckedAt: fixedNow(), Channel: "beta"}, fixedNow(), CacheTTL) {
		t.Fatal("legacy beta cache accepted")
	}
	if CacheFresh(Cache{CheckedAt: fixedNow().Add(-25 * time.Hour), Channel: LatestChannel}, fixedNow(), CacheTTL) {
		t.Fatal("stale cache accepted")
	}
}

func TestNoticeFromCacheUsesUpdateCommand(t *testing.T) {
	t.Parallel()
	cache := Cache{Channel: LatestChannel, LatestVersion: "1.8.4", UpdateAvailable: true}
	notice := NoticeFromCache(cache, "1.8.3")
	if notice == nil {
		t.Fatal("NoticeFromCache() = nil")
	}
	if notice.Command != UpdateCommand || !strings.Contains(notice.Message, "contract-cli 1.8.4 available") {
		t.Fatalf("notice = %+v", notice)
	}
	if NoticeFromCache(cache, "1.8.4") != nil {
		t.Fatal("up-to-date cache produced notice")
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 4, 20, 16, 0, 0, 0, time.FixedZone("CST", 8*60*60))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func jsonHTTPClient(body string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
}
