package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/config"
)

type fakeAuthorizationCallback struct {
	wait func(context.Context, string) (string, error)
}

type authRoundTripFunc func(*http.Request) (*http.Response, error)

func (f authRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func (f fakeAuthorizationCallback) Wait(ctx context.Context, state string) (string, error) {
	return f.wait(ctx, state)
}

func TestUserAuthLoginNoOpenBrowserPrintsAuthorizationURLBeforeWaiting(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	startedCallback := false
	provider := userAuthProvider{
		httpClient: &http.Client{},
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		openBrowser: func(string) error {
			t.Fatal("open browser should not be called when --no-open-browser is set")
			return nil
		},
		authorizationURLWriter: stdout,
		startCallbackServer: func(redirectURL string) (authorizationCallback, error) {
			startedCallback = true
			if redirectURL != "http://127.0.0.1:8000/callback" {
				t.Fatalf("redirect URL = %q", redirectURL)
			}
			return fakeAuthorizationCallback{
				wait: func(context.Context, string) (string, error) {
					if !strings.Contains(stdout.String(), "Open this URL and finish authorization:") {
						t.Fatalf("authorization URL should be printed before waiting, got: %s", stdout.String())
					}
					return "", errors.New("forced wait error")
				},
			}, nil
		},
	}
	profile := &config.Profile{
		Name:         "contract",
		ClientName:   "contract-cli",
		BusinessType: "contract",
		Scopes:       []string{"mcp:tools"},
		Identities: config.Identities{
			User: config.UserIdentity{
				ClientID:              "client-123",
				AuthorizationEndpoint: "https://example.test/oauth/authorize/contract",
				TokenEndpoint:         "https://example.test/oauth/token/contract",
				RegistrationEndpoint:  "https://example.test/oauth/register/contract",
				RedirectURL:           "http://127.0.0.1:8000/callback",
			},
		},
	}

	message, err := provider.Login(context.Background(), profile, authCommandOptions{
		Timeout:       time.Second,
		NoOpenBrowser: true,
	})
	if err == nil || !strings.Contains(err.Error(), "forced wait error") {
		t.Fatalf("Login() error = %v", err)
	}
	if message != "" {
		t.Fatalf("Login() message = %q, want empty on failure", message)
	}
	if !startedCallback {
		t.Fatal("callback server was not started")
	}

	output := stdout.String()
	for _, want := range []string{
		"Open this URL and finish authorization:",
		"https://example.test/oauth/authorize/contract",
		"client_id=client-123",
		"redirect_uri=http%3A%2F%2F127.0.0.1%3A8000%2Fcallback",
		"scope=mcp%3Atools",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("authorization URL output missing %q: %s", want, output)
		}
	}
}

func TestUserAuthLoginAutoOpenDoesNotReturnAuthorizationURL(t *testing.T) {
	t.Parallel()

	var openedURL string
	provider := userAuthProvider{
		httpClient: &http.Client{
			Transport: authRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "https://example.test/oauth/token/contract" {
					t.Fatalf("unexpected token request URL: %s", req.URL.String())
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"access_token":"user-token","token_type":"Bearer","expires_in":3600}`)),
				}, nil
			}),
		},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		openBrowser: func(url string) error {
			openedURL = url
			return nil
		},
		startCallbackServer: func(string) (authorizationCallback, error) {
			return fakeAuthorizationCallback{
				wait: func(context.Context, string) (string, error) {
					return "authorization-code", nil
				},
			}, nil
		},
	}
	profile := &config.Profile{
		Name:         "contract",
		ClientName:   "contract-cli",
		BusinessType: "contract",
		Scopes:       []string{"mcp:tools"},
		Identities: config.Identities{
			User: config.UserIdentity{
				ClientID:              "client-123",
				AuthorizationEndpoint: "https://example.test/oauth/authorize/contract",
				TokenEndpoint:         "https://example.test/oauth/token/contract",
				RegistrationEndpoint:  "https://example.test/oauth/register/contract",
				RedirectURL:           "http://127.0.0.1:8000/callback",
			},
		},
	}

	message, err := provider.Login(context.Background(), profile, authCommandOptions{
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !strings.Contains(openedURL, "https://example.test/oauth/authorize/contract") {
		t.Fatalf("browser was not opened with authorization URL: %s", openedURL)
	}
	if !strings.Contains(message, `Authorization succeeded for profile "contract".`) {
		t.Fatalf("unexpected login message: %s", message)
	}
	if strings.Contains(message, "Open this URL") || strings.Contains(message, "https://example.test/oauth/authorize/contract") {
		t.Fatalf("login message should not expose authorization URL when browser was opened: %s", message)
	}
}
