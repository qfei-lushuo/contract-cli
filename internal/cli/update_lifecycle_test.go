package cli_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
	updatecheck "cn.qfei/contract-cli/internal/update"
)

// Exit immediately after Run, as the production entry point does. Waiting in
// the parent test must not accidentally keep the child's goroutines alive.
func TestUpdateCacheSurvivesShortCommandExit(t *testing.T) {
	directory := t.TempDir()
	for attempt := 0; attempt < 2; attempt++ {
		command := exec.Command(os.Args[0], "-test.run=^TestUpdateCacheProcessHelper$")
		command.Env = append(os.Environ(), "CONTRACT_REVIEW_CACHE_HELPER="+directory)
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("helper: %v %s", err, output)
		}
		cache, ok, err := updatecheck.LoadCache(filepath.Join(directory, "update-check.json"))
		if err != nil || !ok || cache.LatestVersion != "1.8.4" {
			t.Fatalf("cache missing after process exit: %+v %t %v", cache, ok, err)
		}
	}
}

func TestUpdateCacheProcessHelper(t *testing.T) {
	directory := os.Getenv("CONTRACT_REVIEW_CACHE_HELPER")
	if directory == "" {
		return
	}
	app := cli.New(cli.Options{
		Stdout: io.Discard, Stderr: io.Discard, Store: config.NewStore(directory),
		UpdateCurrentVersion: "1.8.3", UpdateRegistryURL: "https://registry.test/latest",
		LookupEnv: func(string) (string, bool) { return "", false },
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			select {
			case <-time.After(300 * time.Millisecond):
				return jsonResponse(`{"version":"1.8.4"}`), nil
			case <-request.Context().Done():
				return nil, request.Context().Err()
			}
		})},
	})
	if err := app.Run(context.Background(), []string{"skills", "list"}); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestAutomaticUpdateWaitIsBoundedAndRespectsCancellation(t *testing.T) {
	for _, cancelParent := range []bool{false, true} {
		t.Run(map[bool]string{false: "registry timeout", true: "parent cancelled"}[cancelParent], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			stopped := make(chan struct{})
			app := cli.New(cli.Options{
				Stdout: io.Discard, Stderr: io.Discard, Store: config.NewStore(t.TempDir()),
				UpdateCurrentVersion: "1.8.3", UpdateRegistryURL: "https://registry.test/latest",
				LookupEnv: func(string) (string, bool) { return "", false },
				HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
					if cancelParent {
						cancel()
					}
					<-request.Context().Done()
					close(stopped)
					return nil, request.Context().Err()
				})},
			})
			start := time.Now()
			if err := app.Run(ctx, []string{"skills", "list"}); err != nil {
				t.Fatal(err)
			}
			limit := 3 * time.Second
			if cancelParent {
				limit = time.Second
			}
			if elapsed := time.Since(start); elapsed > limit {
				t.Fatalf("optional check held command for %v", elapsed)
			}
			select {
			case <-stopped:
			case <-time.After(time.Second):
				t.Fatal("registry request was not cancelled")
			}
		})
	}
}
