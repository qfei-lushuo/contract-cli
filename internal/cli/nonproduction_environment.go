package cli

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cn.qfei/contract-cli/internal/config"
)

const (
	developmentEnvironment        = "dev"
	developmentProfileName        = "contract-dev"
	developmentOpenPlatformOrigin = "https://dev-open.qtech.cn"
	developmentAccountOrigin      = "https://dev-myaccount.qtech.cn"
)

type nonProductionEnvironment struct {
	name, openOrigin, accountOrigin, deviceClientID string
}

// Device client IDs follow the existing production registration in each environment.
var nonProductionEnvironments = map[string]nonProductionEnvironment{
	"dev":  {"dev", developmentOpenPlatformOrigin, developmentAccountOrigin, "zscli_892efdadc11a3f53"},
	"test": {"test", "https://test-open.qtech.cn", "https://test-myaccount.qtech.cn", "zscli_892efdadc11a3f53"},
	"blue": {"blue", "https://open-b.qfei.cn", "https://myaccount-b.qfei.cn", "zscli_892efdadc11a3f53"},
}

func nonProductionProfile(profileName string) (nonProductionEnvironment, bool) {
	env, ok := nonProductionEnvironments[strings.TrimPrefix(profileName, "contract-")]
	return env, ok && profileName == "contract-"+env.name
}

func (env nonProductionEnvironment) preset() environmentPreset {
	return environmentPreset{
		OpenPlatformBaseURL:            env.openOrigin,
		AppTokenEndpoint:               env.openOrigin + "/open-apis/auth/v3/tenant_access_token/internal",
		AuthorizationServerMetadataURL: env.accountOrigin + "/.well-known/oauth-authorization-server/contract",
		Resource:                       env.openOrigin, RedirectURL: "http://127.0.0.1:8000/callback",
		Scopes: []string{"cli:tools", "cli:resources"}, BusinessType: "contract", ClientName: "contract-cli",
		DeviceClientID: env.deviceClientID, DeviceScope: "contract:full contract-review:full",
	}
}

func profileAccountOrigin(profileName string) string {
	if env, ok := nonProductionProfile(profileName); ok {
		return env.accountOrigin
	}
	return productionAccountOrigin
}

func environmentProfileError(profileName string) error {
	if env, ok := nonProductionProfile(profileName); ok {
		return fmt.Errorf("%s integration requires an isolated %q profile with matching URLs and credentials; run `contract-cli config add --env %s --name %s` and authorize again", env.name, profileName, env.name, profileName)
	}
	return productionProfileError(profileName)
}

func (a *App) saveEnvironmentProfile(profile config.Profile) error {
	if a.buildEnvironment != "" {
		if err := a.validateBuildProfile(profile); err != nil {
			return err
		}
		return a.store.UpsertProfile(profile, true)
	}
	if _, ok := nonProductionEnvironments[profile.Environment]; !ok {
		return a.store.UpsertProfile(profile, true)
	}
	cfg, err := a.store.Load()
	if err != nil {
		return err
	}
	cfg.Profiles[profile.Name] = profile
	// A test profile must never become the implicit selection, even on first use.
	return a.store.Save(cfg)
}

func validateNonProductionProfile(profile config.Profile) error {
	env, ok := nonProductionProfile(profile.Name)
	if !ok || profile.Environment != env.name {
		return environmentProfileError(profile.Name)
	}
	if profile.Identities.App.SecretRef != "" && profile.Identities.App.SecretRef != config.AppSecretKey(profile.Name) {
		return environmentProfileError(profile.Name)
	}
	return validateProfileOrigins(profile, env.openOrigin, env.accountOrigin, environmentProfileError(profile.Name))
}

type nonProductionNetworkKey struct{}

// Restrict every request, including redirects, to the selected environment.
func clientForEnvironment(client *http.Client, environment string) *http.Client {
	env, ok := nonProductionEnvironments[environment]
	if !ok {
		return client
	}
	cloned := *client
	next := client.Transport
	if next == nil {
		next = http.DefaultTransport
	}
	cloned.Transport = nonProductionTransport{next: next, environment: env}
	return &cloned
}

type nonProductionTransport struct {
	next        http.RoundTripper
	environment nonProductionEnvironment
}

func (transport nonProductionTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	env := transport.environment
	if !isProductionOriginURL(request.URL.String(), env.openOrigin, true) && !isProductionOriginURL(request.URL.String(), env.accountOrigin, true) {
		if request.Body != nil {
			_ = request.Body.Close()
		}
		return nil, fmt.Errorf("%s integration blocks foreign destination %q", env.name, request.URL.Host)
	}
	request = request.Clone(context.WithValue(request.Context(), nonProductionNetworkKey{}, env.name))
	return transport.next.RoundTrip(request)
}
