package cli

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/credential"
	"cn.qfei/contract-cli/internal/invocation"
	"cn.qfei/contract-cli/internal/openplatform"
)

func developmentMetadata() string {
	return `{"issuer":"https://dev-myaccount.qtech.cn/contract","authorization_endpoint":"https://dev-myaccount.qtech.cn/api/public/oauth/authorize/contract","device_authorization_endpoint":"https://dev-myaccount.qtech.cn/api/public/oauth/device-authorization/contract","token_endpoint":"https://dev-myaccount.qtech.cn/api/public/oauth/token/contract","revocation_endpoint":"https://dev-myaccount.qtech.cn/api/public/oauth/revoke/contract","registration_endpoint":"https://dev-myaccount.qtech.cn/api/public/oauth/register/contract"}`
}

func validDevelopmentProfile() config.Profile {
	preset := developmentPreset()
	return config.Profile{
		Name: developmentProfileName, Environment: developmentEnvironment,
		OpenPlatformBaseURL: preset.OpenPlatformBaseURL, Resource: preset.Resource,
		AppTokenEndpoint: preset.AppTokenEndpoint, AuthorizationServerMetadataURL: preset.AuthorizationServerMetadataURL,
		BusinessType: preset.BusinessType, ClientName: preset.ClientName, Scopes: preset.Scopes,
		DefaultIdentity: config.IdentityUser,
		Identities: config.Identities{User: config.UserIdentity{
			AuthMode: config.UserAuthModeDevice, DeviceClientID: preset.DeviceClientID, DeviceScope: preset.DeviceScope,
			AuthorizationEndpoint:       developmentAccountOrigin + "/api/public/oauth/authorize/contract",
			DeviceAuthorizationEndpoint: developmentAccountOrigin + "/api/public/oauth/device-authorization/contract",
			TokenEndpoint:               developmentAccountOrigin + "/api/public/oauth/token/contract",
			RevocationEndpoint:          developmentAccountOrigin + "/api/public/oauth/revoke/contract",
			RegistrationEndpoint:        developmentAccountOrigin + "/api/public/oauth/register/contract",
			RedirectURL:                 preset.RedirectURL,
		}},
	}
}

func newDevelopmentTestApp(t *testing.T, handler productionRoundTripFunc) (*App, *deviceMemoryCredentialStore) {
	t.Helper()
	dir, workspace := t.TempDir(), t.TempDir()
	credentials := &deviceMemoryCredentialStore{values: map[string]credential.DeviceCredential{}}
	app := New(Options{
		Stdout: io.Discard, Stderr: io.Discard, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Store: config.NewStore(dir), Secrets: config.NewSecretsStore(dir), CredentialStore: credentials,
		HTTPClient: &http.Client{Transport: handler},
		LookupEnv: func(name string) (string, bool) {
			if name == "SKILL_SESSION_WORKSPACE" {
				return workspace, true
			}
			return "", false
		},
	})
	return app, credentials
}

func TestDevelopmentConfigPreservesProductionAndDefault(t *testing.T) {
	for _, withProduction := range []bool{false, true} {
		t.Run(map[bool]string{false: "fresh", true: "existing prod"}[withProduction], func(t *testing.T) {
			calls := 0
			app, credentials := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
				calls++
				if req.URL.String() != developmentPreset().AuthorizationServerMetadataURL {
					t.Fatal(req.URL)
				}
				return productionResponse(200, developmentMetadata()), nil
			})
			prod := validProductionProfile()
			prod.Identities.User.Token = &config.Token{AccessToken: "prod-fixture"}
			if withProduction {
				if err := app.store.UpsertProfile(prod, true); err != nil {
					t.Fatal(err)
				}
				credentials.values["contract"] = credential.DeviceCredential{DeviceProfile: snapshotDeviceProfile(prod), Token: prod.Identities.User.Token}
			}
			for i := 0; i < 2; i++ {
				if err := app.runConfigAdd(context.Background(), []string{"--env", "dev", "--name", "contract-dev"}); err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := app.store.Load()
			if err != nil {
				t.Fatal(err)
			}
			wantCurrent := ""
			if withProduction {
				wantCurrent = "contract"
				if !reflect.DeepEqual(cfg.Profiles["contract"], prod) {
					t.Fatal("production profile changed")
				}
				if credentials.values["contract"].Token.AccessToken != "prod-fixture" {
					t.Fatal("production credential changed")
				}
			}
			if cfg.CurrentProfile != wantCurrent {
				t.Fatalf("default = %q", cfg.CurrentProfile)
			}
			dev := cfg.Profiles[developmentProfileName]
			if err := validateDevelopmentProfile(dev); err != nil {
				t.Fatal(err)
			}
			if dev.Identities.User.Token != nil || dev.Identities.App.Token != nil || calls != 2 {
				t.Fatal("unexpected auth state or discovery count")
			}
		})
	}
}

func TestDevelopmentProfileRequiresExplicitNameAndSelection(t *testing.T) {
	app, _ := newDevelopmentTestApp(t, func(*http.Request) (*http.Response, error) { t.Fatal("unexpected network call"); return nil, nil })
	for _, args := range [][]string{
		{"--env", "dev"}, {"--env", "dev", "--name", "contract"}, {"--env", "prod", "--name", "contract-dev"}, {"--env", "test"},
	} {
		if err := app.runConfigAdd(context.Background(), args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if err := app.store.UpsertProfile(validDevelopmentProfile(), true); err != nil {
		t.Fatal(err)
	}
	if _, err := app.loadDeviceAwareProfile(""); err == nil {
		t.Fatal("implicitly selected dev")
	}
	if _, err := app.loadDeviceAwareProfile(developmentProfileName); err != nil {
		t.Fatal(err)
	}
}

func TestDevelopmentConfigRejectsProductionDiscoveryWithoutSaving(t *testing.T) {
	app, _ := newDevelopmentTestApp(t, func(*http.Request) (*http.Response, error) {
		return productionResponse(200, strings.ReplaceAll(developmentMetadata(), developmentAccountOrigin, productionAccountOrigin)), nil
	})
	if err := app.runConfigAdd(context.Background(), []string{"--env", "dev", "--name", "contract-dev"}); err == nil {
		t.Fatal("accepted production authorization endpoints")
	}
	if _, found, err := app.store.LookupProfile(developmentProfileName); err != nil || found {
		t.Fatalf("saved invalid profile: found=%v err=%v", found, err)
	}
}

func TestDevelopmentRejectsMixedProfileURLs(t *testing.T) {
	for _, field := range []string{"base", "resource", "app", "metadata", "authorization", "device", "token", "revoke", "registration", "secret"} {
		t.Run(field, func(t *testing.T) {
			profile := validDevelopmentProfile()
			switch field {
			case "base":
				profile.OpenPlatformBaseURL = productionOpenPlatformOrigin
			case "resource":
				profile.Resource = productionOpenPlatformOrigin
			case "app":
				profile.AppTokenEndpoint = productionOpenPlatformOrigin + "/token"
			case "metadata":
				profile.AuthorizationServerMetadataURL = productionAccountOrigin + "/metadata"
			case "authorization":
				profile.Identities.User.AuthorizationEndpoint = productionAccountOrigin + "/authorize"
			case "device":
				profile.Identities.User.DeviceAuthorizationEndpoint = productionAccountOrigin + "/device"
			case "token":
				profile.Identities.User.TokenEndpoint = productionAccountOrigin + "/token"
			case "revoke":
				profile.Identities.User.RevocationEndpoint = productionAccountOrigin + "/revoke"
			case "registration":
				profile.Identities.User.RegistrationEndpoint = productionAccountOrigin + "/register"
			case "secret":
				profile.Identities.App.SecretRef = config.AppSecretKey("contract")
			}
			if validateProductionProfile(profile) == nil {
				t.Fatal("accepted mixed environment")
			}
		})
	}
}

func TestDevelopmentRejectsForeignDeviceState(t *testing.T) {
	dev := validDevelopmentProfile()
	for _, stored := range []credential.DeviceCredential{
		{Token: &config.Token{AccessToken: "unknown-origin"}},
		{DeviceProfile: snapshotDeviceProfile(validProductionProfile())},
		{DeviceProfile: snapshotDeviceProfile(dev), Pending: &credential.PendingTransaction{TokenEndpoint: productionAccountOrigin + "/token"}},
		{DeviceProfile: snapshotDeviceProfile(dev), Pending: &credential.PendingTransaction{TokenEndpoint: dev.Identities.User.TokenEndpoint, VerificationURIComplete: productionAccountOrigin + "/device"}},
	} {
		if validateProductionDeviceCredential(developmentProfileName, stored) == nil {
			t.Fatal("accepted foreign device state")
		}
	}
}

func TestDevelopmentEncryptedProfileRecovery(t *testing.T) {
	app, credentials := newDevelopmentTestApp(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("recovery must not make network requests")
		return nil, nil
	})
	if err := app.store.UpsertProfile(validProductionProfile(), true); err != nil {
		t.Fatal(err)
	}
	credentials.values[developmentProfileName] = credential.DeviceCredential{
		DeviceProfile: snapshotDeviceProfile(validDevelopmentProfile()),
		Token:         &config.Token{AccessToken: "dev-fixture", Expiry: time.Now().Add(time.Hour)},
	}
	profile, err := app.loadDeviceAwareProfile(developmentProfileName)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Environment != developmentEnvironment || profile.Identities.User.Token != nil {
		t.Fatal("invalid recovered profile")
	}
	cfg, err := app.store.Load()
	if err != nil || cfg.CurrentProfile != "contract" {
		t.Fatal("changed production selection during recovery")
	}
}

func TestDevelopmentTransportBlocksForeignDestinationsAndRedirects(t *testing.T) {
	for _, destination := range []string{productionOpenPlatformOrigin, productionAccountOrigin, "https://example.com", "http://dev-open.qtech.cn", "https://dev-open.qtech.cn.evil.example", "https://dev-open.qtech.cn:444"} {
		t.Run(destination, func(t *testing.T) {
			calls := 0
			app, _ := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
				calls++
				response := productionResponse(307, "")
				response.Header.Set("Location", destination+"/redirect")
				return response, nil
			})
			client := clientForEnvironment(app.httpClient, developmentEnvironment)
			if _, err := client.Get(developmentOpenPlatformOrigin + "/start"); err == nil {
				t.Fatal("allowed foreign redirect")
			}
			if calls != 1 {
				t.Fatalf("requests = %d", calls)
			}
			if _, err := app.httpClient.Get(developmentOpenPlatformOrigin); err == nil {
				t.Fatal("shared production guard was relaxed")
			}
			if calls != 1 {
				t.Fatal("production request reached dev")
			}
		})
	}
}

// All environment packages use the same existing credential inputs and priority.
func TestEnvironmentAppCredentialsPreserveExistingInputs(t *testing.T) {
	for _, environment := range []string{"prod", "dev", "test", "blue"} {
		t.Run(environment, func(t *testing.T) {
			profile := validProductionProfile()
			if environment != "prod" {
				profile = environmentProfile(nonProductionEnvironments[environment])
			}
			secrets := config.NewSecretsStore(t.TempDir())
			profile.Identities.App.AppID = "saved-id"
			profile.Identities.App.SecretRef = config.AppSecretKey(profile.Name)
			if err := secrets.Set(profile.Identities.App.SecretRef, "saved-secret"); err != nil {
				t.Fatal(err)
			}
			for _, test := range []struct {
				name               string
				values             map[string]string
				options            authCommandOptions
				wantID, wantSecret string
			}{
				{"saved", nil, authCommandOptions{}, "saved-id", "saved-secret"},
				{"standard env", map[string]string{envAppID: "environment-id", envAppSecret: "environment-secret"}, authCommandOptions{}, "environment-id", "environment-secret"},
				{"legacy env", map[string]string{legacyEnvAppID: "legacy-id", legacyEnvAppSecret: "legacy-secret"}, authCommandOptions{}, "legacy-id", "legacy-secret"},
				{"legacy bot env", map[string]string{legacyEnvBotAppID: "bot-id", legacyEnvBotAppSecret: "bot-secret"}, authCommandOptions{}, "bot-id", "bot-secret"},
				{"standard before legacy", map[string]string{envAppID: "environment-id", envAppSecret: "environment-secret", legacyEnvAppID: "legacy-id", legacyEnvAppSecret: "legacy-secret"}, authCommandOptions{}, "environment-id", "environment-secret"},
				{"flags before env", map[string]string{envAppID: "environment-id", envAppSecret: "environment-secret"}, authCommandOptions{AppID: "flag-id", AppSecret: "flag-secret"}, "flag-id", "flag-secret"},
			} {
				t.Run(test.name, func(t *testing.T) {
					provider := appAuthProvider{secrets: secrets, lookupEnv: func(key string) (string, bool) { value, ok := test.values[key]; return value, ok }}
					got, err := provider.resolveCredentials(profile, test.options, true)
					if err != nil {
						t.Fatal(err)
					}
					if got.appID != test.wantID || got.appSecret != test.wantSecret {
						t.Fatal("credential input priority changed")
					}
				})
			}
		})
	}
}

func TestDevelopmentDeviceLifecycleAndPerRequestHeaders(t *testing.T) {
	initCalls, exchangeCalls, refreshCalls, revokeCalls, businessCalls, detectCalls := 0, 0, 0, 0, 0, 0
	traces := map[string]bool{}
	app, credentials := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/api/public/oauth/device-authorization/contract":
			initCalls++
			if err := req.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if req.Form.Get("resource") != developmentOpenPlatformOrigin || req.Form.Get("client_id") != developmentPreset().DeviceClientID {
				t.Fatal("wrong dev auth parameters")
			}
			return productionResponse(200, `{"device_code":"test-device","user_code":"test-user","verification_uri_complete":"https://dev-myaccount.qtech.cn/device?user_code=test-user","expires_in":600}`), nil
		case "/api/public/oauth/token/contract":
			if err := req.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if req.Form.Get("grant_type") == "refresh_token" {
				refreshCalls++
				return productionResponse(200, `{"access_token":"dev-refreshed","refresh_token":"dev-refresh-2","token_type":"Bearer","expires_in":3600}`), nil
			}
			exchangeCalls++
			return productionResponse(200, `{"access_token":"dev-initial","refresh_token":"dev-refresh-1","token_type":"Bearer","expires_in":1}`), nil
		case "/api/public/oauth/revoke/contract":
			revokeCalls++
			return productionResponse(200, `{}`), nil
		case "/open-apis/contract/v1/mcp/contracts/search":
			businessCalls++
			if req.URL.Host != "dev-open.qtech.cn" || req.Header.Get("Authorization") != "Bearer dev-refreshed" {
				t.Fatal("wrong dev business destination or credential")
			}
			wantSource := []string{"doubao", "workbuddy"}[businessCalls-1]
			for key, want := range map[string]string{
				invocation.HeaderChannelType: "cli", invocation.HeaderAgentSourceType: wantSource,
				invocation.HeaderProductCode: "contract", invocation.HeaderEvidenceType: "test",
				invocation.HeaderConfidence: "high", invocation.HeaderDetectorVersion: invocation.DetectorVersion,
				invocation.HeaderRuleID: "test.rule",
			} {
				if req.Header.Get(key) != want {
					t.Fatalf("%s = %s", key, req.Header.Get(key))
				}
			}
			trace := req.Header.Get("Traceparent")
			if len(trace) != 55 || req.Header.Get("X-Log-Id") == "" || traces[trace] {
				t.Fatal("missing or reused per-request trace")
			}
			traces[trace] = true
			return productionResponse(200, `{"code":0,"data":[]}`), nil
		default:
			t.Fatalf("unexpected dev path %s", req.URL.Path)
			return nil, nil
		}
	})
	dev := validDevelopmentProfile()
	if err := app.store.UpsertProfile(validProductionProfile(), true); err != nil {
		t.Fatal(err)
	}
	if err := app.saveEnvironmentProfile(dev); err != nil {
		t.Fatal(err)
	}
	app.inspectEnvironment = func(context.Context, int) invocation.Result {
		source := []string{"doubao", "workbuddy"}[detectCalls]
		detectCalls++
		return invocation.Result{ChannelType: "cli", AgentSourceType: source, ProductCode: "contract", EvidenceType: "test", Confidence: "high", DetectorVersion: invocation.DetectorVersion, RuleID: "test.rule"}
	}
	ctx := context.Background()
	if err := app.runAuthDeviceInit(ctx, []string{"--profile", "contract-dev"}); err != nil {
		t.Fatal(err)
	}
	if err := app.runAuthDeviceComplete(ctx, []string{"--profile", "contract-dev"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		client, requestContext, err := app.openPlatformClientAndContext(developmentProfileName, "user", "/open-apis/contract/v1/mcp/contracts/search", openplatform.IdentityPolicyUserOnly)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Do(ctx, requestContext, openplatform.Request{Method: http.MethodGet, Path: "/open-apis/contract/v1/mcp/contracts/search"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := app.deviceAuthLogout(ctx, dev); err != nil {
		t.Fatal(err)
	}
	if initCalls != 1 || exchangeCalls != 1 || refreshCalls != 1 || revokeCalls != 1 || businessCalls != 2 || detectCalls != 2 {
		t.Fatalf("calls init=%d exchange=%d refresh=%d revoke=%d business=%d detect=%d", initCalls, exchangeCalls, refreshCalls, revokeCalls, businessCalls, detectCalls)
	}
	if _, ok := credentials.values[developmentProfileName]; ok {
		t.Fatal("dev credential not cleared")
	}
	cfg, _ := app.store.Load()
	if cfg.CurrentProfile != "contract" {
		t.Fatal("changed default profile")
	}
}

func TestDevelopmentBusinessWithLegacyUserAndAppTokens(t *testing.T) {
	for _, identity := range []config.IdentityKind{config.IdentityUser, config.IdentityApp} {
		t.Run(string(identity), func(t *testing.T) {
			calls := 0
			app, _ := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
				calls++
				if !strings.HasPrefix(req.URL.String(), developmentOpenPlatformOrigin+"/") || req.Header.Get("Authorization") != "Bearer dev-fixture" {
					t.Fatal("wrong destination or token")
				}
				return productionResponse(200, `{"code":0}`), nil
			})
			profile := validDevelopmentProfile()
			profile.Identities.User.AuthMode = config.UserAuthModeAuthorizationCode
			token := &config.Token{AccessToken: "dev-fixture", Expiry: time.Now().Add(time.Hour)}
			if identity == config.IdentityUser {
				profile.Identities.User.Token = token
			} else {
				profile.Identities.App.Token = token
			}
			if err := app.saveEnvironmentProfile(profile); err != nil {
				t.Fatal(err)
			}
			app.inspectEnvironment = func(context.Context, int) invocation.Result { return invocation.Result{AgentSourceType: "unknown"} }
			client, requestContext, err := app.openPlatformClientAndContext(developmentProfileName, string(identity), "/open-apis/contract/v1/contracts", openplatform.IdentityPolicyAny)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.Do(context.Background(), requestContext, openplatform.Request{Method: http.MethodGet, Path: "/open-apis/contract/v1/contracts"}); err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}

func developmentPreset() environmentPreset { return nonProductionEnvironments["dev"].preset() }
func validateDevelopmentProfile(profile config.Profile) error {
	return validateNonProductionProfile(profile)
}
