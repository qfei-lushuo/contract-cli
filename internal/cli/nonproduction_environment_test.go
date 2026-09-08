package cli

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/config"
	"cn.qfei/contract-cli/internal/credential"
)

func environmentMetadata(env nonProductionEnvironment) string {
	return strings.ReplaceAll(developmentMetadata(), developmentAccountOrigin, env.accountOrigin)
}

func environmentProfile(env nonProductionEnvironment) config.Profile {
	p := validDevelopmentProfile()
	p.Name, p.Environment = "contract-"+env.name, env.name
	p.OpenPlatformBaseURL, p.Resource = env.openOrigin, env.openOrigin
	p.AppTokenEndpoint = env.preset().AppTokenEndpoint
	p.AuthorizationServerMetadataURL = env.preset().AuthorizationServerMetadataURL
	u := &p.Identities.User
	u.DeviceClientID = env.deviceClientID
	for _, field := range []*string{&u.AuthorizationEndpoint, &u.DeviceAuthorizationEndpoint, &u.TokenEndpoint, &u.RevocationEndpoint, &u.RegistrationEndpoint} {
		*field = strings.ReplaceAll(*field, developmentAccountOrigin, env.accountOrigin)
	}
	return p
}

func TestNonProductionConfigAndDeviceLifecycle(t *testing.T) {
	for name, env := range nonProductionEnvironments {
		t.Run(name, func(t *testing.T) {
			calls := map[string]int{}
			deviceID := env.deviceClientID
			if deviceID == "" {
				deviceID = "registered-test-client-fixture"
			}
			app, credentials := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
				calls[req.URL.Path]++
				if !strings.HasPrefix(req.URL.String(), env.accountOrigin+"/") {
					t.Fatalf("foreign destination: %s", req.URL)
				}
				switch req.URL.Path {
				case "/.well-known/oauth-authorization-server/contract":
					return productionResponse(200, environmentMetadata(env)), nil
				case "/api/public/oauth/device-authorization/contract":
					if err := req.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if req.Form.Get("client_id") != deviceID || req.Form.Get("resource") != env.openOrigin {
						t.Fatal("wrong Device client/resource")
					}
					return productionResponse(200, `{"device_code":"fixture-device","user_code":"fixture-user","verification_uri_complete":"`+env.accountOrigin+`/device?user_code=fixture-user","expires_in":600}`), nil
				case "/api/public/oauth/token/contract":
					if err := req.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if req.Form.Get("client_id") != deviceID {
						t.Fatal("wrong token client")
					}
					return productionResponse(200, `{"access_token":"fixture-access","refresh_token":"fixture-refresh","expires_in":3600}`), nil
				default:
					t.Fatalf("unexpected request: %s", req.URL)
					return nil, nil
				}
			})
			prod := validProductionProfile()
			if err := app.store.UpsertProfile(prod, true); err != nil {
				t.Fatal(err)
			}
			profileName := "contract-" + name
			args := []string{"--env", name, "--name", profileName}
			if err := app.runConfigAdd(context.Background(), args); err != nil {
				t.Fatal(err)
			}
			if env.deviceClientID == "" {
				err := app.runAuthDeviceInit(context.Background(), []string{"--profile", profileName})
				if err == nil || !strings.Contains(err.Error(), "--device-client-id") {
					t.Fatalf("missing Device registration: %v", err)
				}
				if len(calls) != 1 {
					t.Fatal("missing client issued a request")
				}
				if err := app.runConfigAdd(context.Background(), append(args, "--device-client-id", deviceID)); err != nil {
					t.Fatal(err)
				}
				if err := app.runConfigAdd(context.Background(), args); err != nil {
					t.Fatal(err)
				}
			}
			if err := app.runAuthDeviceInit(context.Background(), []string{"--profile", profileName}); err != nil {
				t.Fatal(err)
			}
			if err := app.runAuthDeviceComplete(context.Background(), []string{"--profile", profileName}); err != nil {
				t.Fatal(err)
			}
			if credentials.values[profileName].Token.AccessToken != "fixture-access" {
				t.Fatal("token was not stored in selected profile")
			}
			cfg, err := app.store.Load()
			if err != nil || cfg.CurrentProfile != "contract" || cfg.Profiles["contract"].Environment != "prod" {
				t.Fatal("changed production selection")
			}
			if err := app.runConfigAdd(context.Background(), append(args, "--device-client-id", "replacement-fixture")); err != nil {
				t.Fatal(err)
			}
			if _, ok := credentials.values[profileName]; ok {
				t.Fatal("retained Device credentials after client change")
			}
		})
	}
}

func TestNonProductionIsolationAcrossEveryEnvironment(t *testing.T) {
	for name, env := range nonProductionEnvironments {
		t.Run(name, func(t *testing.T) {
			p := environmentProfile(env)
			// Supply a registered fixture for snapshot tests; production code does not
			// assume that test has the same registration as dev or blue.
			if p.Identities.User.DeviceClientID == "" {
				p.Identities.User.DeviceClientID = "registered-test-fixture"
			}
			if err := validateProductionProfile(p); err != nil {
				t.Fatal(err)
			}
			for otherName, other := range nonProductionEnvironments {
				if otherName == name {
					continue
				}
				mixed := p
				mixed.Identities.User.TokenEndpoint = other.accountOrigin + "/token"
				if validateProductionProfile(mixed) == nil {
					t.Fatalf("accepted %s token URL", otherName)
				}
				stored := credential.DeviceCredential{DeviceProfile: snapshotDeviceProfile(p), Pending: &credential.PendingTransaction{TokenEndpoint: other.accountOrigin + "/token"}}
				if validateProductionDeviceCredential(p.Name, stored) == nil {
					t.Fatalf("accepted %s pending state", otherName)
				}
				stored.DeviceProfile = snapshotDeviceProfile(environmentProfile(other))
				if validateProductionDeviceCredential(p.Name, stored) == nil {
					t.Fatalf("accepted %s snapshot", otherName)
				}
			}
			p.Identities.App.SecretRef = config.AppSecretKey("contract")
			if validateProductionProfile(p) == nil {
				t.Fatal("accepted production secret reference")
			}
			app, _ := newDevelopmentTestApp(t, func(*http.Request) (*http.Response, error) { t.Fatal("unexpected network"); return nil, nil })
			for _, badName := range []string{"contract", "custom", "contract-unknown"} {
				if err := app.runConfigAdd(context.Background(), []string{"--env", name, "--name", badName}); err == nil {
					t.Fatalf("accepted %s", badName)
				}
			}
			if err := app.store.UpsertProfile(environmentProfile(env), true); err != nil {
				t.Fatal(err)
			}
			if _, err := app.loadDeviceAwareProfile(""); err == nil {
				t.Fatal("implicitly selected nonproduction profile")
			}
			provider := appAuthProvider{secrets: config.NewSecretsStore(t.TempDir()), lookupEnv: func(key string) (string, bool) {
				if key == "CONTRACT_CLI_"+strings.ToUpper(name)+"_APP_ID" || key == "CONTRACT_CLI_"+strings.ToUpper(name)+"_APP_SECRET" {
					return "", false
				}
				return "foreign-fixture", true
			}}
			if _, err := provider.resolveCredentials(environmentProfile(env), authCommandOptions{}, true); err == nil {
				t.Fatal("inherited foreign credentials")
			}
			provider.lookupEnv = func(key string) (string, bool) {
				return "selected-fixture", key == "CONTRACT_CLI_"+strings.ToUpper(name)+"_APP_ID" || key == "CONTRACT_CLI_"+strings.ToUpper(name)+"_APP_SECRET"
			}
			if _, err := provider.resolveCredentials(environmentProfile(env), authCommandOptions{}, true); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNonProductionRedirectsAndProductionGuard(t *testing.T) {
	destinations := []string{productionOpenPlatformOrigin, productionAccountOrigin, "https://example.com"}
	for _, env := range nonProductionEnvironments {
		destinations = append(destinations, env.openOrigin, env.accountOrigin)
	}
	for name, env := range nonProductionEnvironments {
		for _, destination := range destinations {
			if destination == env.openOrigin || destination == env.accountOrigin {
				continue
			}
			t.Run(name+" to "+destination, func(t *testing.T) {
				calls := 0
				app, _ := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
					calls++
					res := productionResponse(307, "")
					res.Header.Set("Location", destination+"/redirect")
					return res, nil
				})
				client := clientForEnvironment(app.httpClient, name)
				client.Timeout = time.Second
				if _, err := client.Get(env.openOrigin + "/start"); err == nil {
					t.Fatal("allowed foreign redirect")
				}
				if calls != 1 {
					t.Fatalf("requests=%d", calls)
				}
				if _, err := app.httpClient.Get(env.openOrigin); err == nil {
					t.Fatal("shared production client reached test environment")
				}
				if calls != 1 {
					t.Fatal("production guard did not block")
				}
			})
		}
	}
}

func TestNonProductionBrowserLoginRegistersItsOwnClient(t *testing.T) {
	for name, env := range nonProductionEnvironments {
		t.Run(name, func(t *testing.T) {
			registrations, exchanges := 0, 0
			app, _ := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
				if !strings.HasPrefix(req.URL.String(), env.accountOrigin+"/") {
					t.Fatal("foreign auth request")
				}
				switch req.URL.Path {
				case "/api/public/oauth/register/contract":
					registrations++
					return productionResponse(201, `{"client_id":"dynamically-registered-fixture"}`), nil
				case "/api/public/oauth/token/contract":
					exchanges++
					if err := req.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if req.Form.Get("client_id") != "dynamically-registered-fixture" || req.Form.Get("grant_type") != "authorization_code" {
						t.Fatal("browser login used wrong client/grant")
					}
					return productionResponse(200, `{"access_token":"browser-fixture","expires_in":3600}`), nil
				default:
					t.Fatalf("unexpected path %s", req.URL.Path)
					return nil, nil
				}
			})
			p := environmentProfile(env)
			provider := userAuthProvider{
				httpClient: app.httpClient, logger: app.logger,
				openBrowser: func(url string) error {
					if !strings.HasPrefix(url, env.accountOrigin+"/") || !strings.Contains(url, "client_id=dynamically-registered-fixture") {
						t.Fatalf("wrong authorization URL %s", url)
					}
					return nil
				},
				startCallbackServer: func(string) (authorizationCallback, error) {
					return fakeAuthorizationCallback{wait: func(context.Context, string) (string, error) { return "fixture-code", nil }}, nil
				},
			}
			if _, err := provider.Login(context.Background(), &p, authCommandOptions{Timeout: time.Second}); err != nil {
				t.Fatal(err)
			}
			if registrations != 1 || exchanges != 1 || p.Identities.User.ClientID != "dynamically-registered-fixture" || p.Identities.User.DeviceClientID != env.deviceClientID {
				t.Fatal("mixed browser and Device clients")
			}
		})
	}
}
