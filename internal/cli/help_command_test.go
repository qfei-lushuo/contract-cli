package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cn.qfei/contract-cli/internal/cli"
	"cn.qfei/contract-cli/internal/config"
)

func TestHelpRequestsRenderExpectedTopics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		args        []string
		contains    []string
		notContains []string
	}{
		{
			name: "top level help flag",
			args: []string{"--help"},
			contains: []string{
				"Name:",
				"contract-cli",
				"Usage:",
				"contract-cli <command> [flags]",
				"Commands:",
				"contract-cli contract <subcommand> [flags]",
				"contract-cli mdm vendor <subcommand> [flags]",
			},
		},
		{
			name: "help command path",
			args: []string{"help", "contract", "upload-file"},
			contains: []string{
				"Name:",
				"contract upload-file",
				"Usage:",
				"contract-cli contract upload-file --file <path> --file-type <type> [flags]",
				"--file <path>",
				"--file-type <type>",
				"user/bot",
				"200MB",
				"不接受 --input-file / --data",
			},
		},
		{
			name: "subcommand help flag",
			args: []string{"contract", "upload-file", "--help"},
			contains: []string{
				"contract upload-file",
				"--file-name <name>",
				"multipart/form-data",
			},
		},
		{
			name: "contract download file help",
			args: []string{"contract", "download-file", "--help"},
			contains: []string{
				"contract download-file",
				"--output-file <path>",
				"--force",
				"默认拉起保存文件弹窗",
				"bot-only",
			},
		},
		{
			name: "contract patch help",
			args: []string{"contract", "patch", "--help"},
			contains: []string{
				"contract patch",
				"contract-cli contract patch <contract-id> --input-file <path>|--data <json> [flags]",
				"PATCH /open-apis/contract/v1/contracts/{contract_id}",
				"bot-only",
			},
		},
		{
			name: "leaf help ignores positional example",
			args: []string{"contract", "get", "contract-1", "--help"},
			contains: []string{
				"Name:",
				"contract get",
				"Usage:",
				"contract-cli contract get <contract-id> [flags]",
				"--user-id-type <type>",
			},
		},
		{
			name: "contract search explains identity routing",
			args: []string{"contract", "search", "--help"},
			contains: []string{
				"contract search",
				"--contract-number <number>",
				"user: /open-apis/contract/v1/mcp/contracts/search",
				"bot: /open-apis/contract/v1/contracts/search",
			},
		},
		{
			name: "contract enum list user only",
			args: []string{"contract", "enum", "list", "--help"},
			contains: []string{
				"contract enum list",
				"--type <enum-type>",
				"仅支持 --as user",
			},
		},
		{
			name: "mdm vendor list flags",
			args: []string{"mdm", "vendor", "list", "--help"},
			contains: []string{
				"mdm vendor list",
				"--name <name>",
				"--page-size <n>",
				"--page-token <token>",
			},
		},
		{
			name: "auth login flags",
			args: []string{"auth", "login", "--help"},
			contains: []string{
				"auth login",
				"--timeout <duration>",
				"--app-id <id>",
				"--app-secret <secret>",
			},
		},
		{
			name: "config add production defaults",
			args: []string{"config", "add", "--help"},
			contains: []string{
				"config add",
				"--env <prod>",
				"默认 prod",
				"--name <profile>",
				"默认 contract",
				"contract-cli config add --env prod --name contract",
			},
			notContains: []string{
				"dev",
			},
		},
		{
			name: "update check production example",
			args: []string{"update", "check", "--help"},
			contains: []string{
				"update check",
				"--channel <latest|beta>",
				"--json",
				"contract-cli update check --channel latest --json",
			},
			notContains: []string{
				"contract-cli update check --channel beta",
			},
		},
		{
			name: "skills install flags",
			args: []string{"skills", "install", "--help"},
			contains: []string{
				"skills install",
				"--target <dir>",
				"--force",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			app := cli.New(cli.Options{
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
				Store:  config.NewStore(t.TempDir()),
			})

			if err := app.Run(context.Background(), tc.args); err != nil {
				t.Fatalf("Run(%v) error = %v", tc.args, err)
			}
			for _, want := range tc.contains {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("help output missing %q:\n%s", want, stdout.String())
				}
			}
			for _, forbidden := range tc.notContains {
				if strings.Contains(stdout.String(), forbidden) {
					t.Fatalf("help output should not contain %q:\n%s", forbidden, stdout.String())
				}
			}
		})
	}
}

func TestHelpDoesNotTriggerProfilesHTTPUpdateOrLogs(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	requests := 0
	store := config.NewStore(t.TempDir())
	app := cli.New(cli.Options{
		Stdout:               stdout,
		Stderr:               stderr,
		Store:                store,
		UpdateRegistryURL:    "https://registry.test/@qfeius%2fcontract-cli",
		UpdateCurrentVersion: "0.1.0-beta.1",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				requests++
				return nil, fmt.Errorf("help must not send HTTP request to %s", req.URL.String())
			}),
		},
	})

	if err := app.Run(context.Background(), []string{"contract", "upload-file", "--help"}); err != nil {
		t.Fatalf("Run(help) error = %v", err)
	}
	if requests != 0 {
		t.Fatalf("help sent %d HTTP requests, want 0", requests)
	}
	if stderr.String() != "" {
		t.Fatalf("help should not log to stderr, got: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "contract upload-file") {
		t.Fatalf("missing help output: %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(store.Path()), "update-check.json")); !os.IsNotExist(err) {
		t.Fatalf("help should not write update cache, stat error = %v", err)
	}
}

func TestAllCurrentHelpTopicsRender(t *testing.T) {
	t.Parallel()

	topics := []string{
		"version",
		"config",
		"config add",
		"auth",
		"auth login",
		"auth status",
		"auth logout",
		"auth use",
		"skills",
		"skills list",
		"skills install",
		"update",
		"update check",
		"contract",
		"contract search",
		"contract get",
		"contract sync-user-groups",
		"contract text",
		"contract create",
		"contract upload-file",
		"contract submit",
		"contract resubmit",
		"contract patch",
		"contract download-file",
		"contract delete",
		"contract print-file",
		"contract share",
		"contract share get",
		"contract cooperation",
		"contract cooperation link get",
		"contract cooperation record get",
		"contract approval",
		"contract approval start",
		"contract approval get",
		"contract category",
		"contract category list",
		"contract template",
		"contract template list",
		"contract template get",
		"contract template instantiate",
		"contract enum",
		"contract enum list",
		"payment",
		"payment create",
		"payment update",
		"payment get",
		"payment list",
		"payment plan",
		"payment plan notify",
		"payment plan search",
		"payment record",
		"payment record create",
		"payment record update",
		"payment record get",
		"payment record list",
		"mdm",
		"mdm vendor",
		"mdm vendor list",
		"mdm vendor get",
		"mdm legal",
		"mdm legal list",
		"mdm legal get",
		"mdm fields",
		"mdm fields list",
	}

	for _, topic := range topics {
		topic := topic
		t.Run(topic, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			app := cli.New(cli.Options{
				Stdout: stdout,
				Stderr: &bytes.Buffer{},
				Store:  config.NewStore(t.TempDir()),
			})
			args := append([]string{"help"}, strings.Fields(topic)...)
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Run(%v) error = %v", args, err)
			}
			if !strings.Contains(stdout.String(), "Name:\n  "+topic) {
				t.Fatalf("help output for %q has unexpected name:\n%s", topic, stdout.String())
			}
		})
	}
}

func TestUnknownHelpTopicReturnsClearError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		args    []string
		wantErr string
	}{
		{
			args:    []string{"help", "vendor"},
			wantErr: `unknown help topic "vendor"`,
		},
		{
			args:    []string{"contract", "unknown", "--help"},
			wantErr: `unknown help topic "contract unknown"`,
		},
		{
			args:    []string{"help", "api", "call"},
			wantErr: `unknown help topic "api call"`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			t.Parallel()

			app := cli.New(cli.Options{
				Stdout: &bytes.Buffer{},
				Stderr: &bytes.Buffer{},
				Store:  config.NewStore(t.TempDir()),
			})

			err := app.Run(context.Background(), tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) || !strings.Contains(err.Error(), "contract-cli help") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
