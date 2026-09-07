package cli

import (
	"context"
	"fmt"
	"net/http"

	"cn.qfei/contract-cli/internal/config"
)

// Temporary dev integration support. Remove this capability before the production
// release; do not remove the production URL and credential guards with it.
const (
	developmentEnvironment        = "dev"
	developmentProfileName        = "contract-dev"
	developmentOpenPlatformOrigin = "https://dev-open.qtech.cn"
	developmentAccountOrigin      = "https://dev-myaccount.qtech.cn"
)

func developmentPreset() environmentPreset {
	return environmentPreset{
		OpenPlatformBaseURL:            developmentOpenPlatformOrigin,
		AppTokenEndpoint:               developmentOpenPlatformOrigin + "/open-apis/auth/v3/tenant_access_token/internal",
		AuthorizationServerMetadataURL: developmentAccountOrigin + "/.well-known/oauth-authorization-server/contract",
		Resource:                       developmentOpenPlatformOrigin,
		RedirectURL:                    "http://127.0.0.1:8000/callback",
		Scopes:                         []string{"cli:tools", "cli:resources"},
		BusinessType:                   "contract",
		ClientName:                     "contract-cli",
		// Public client from dev-contract-charts oauth2.device-authorization.clients.contract.
		DeviceClientID: "zscli_892efdadc11a3f53",
		DeviceScope:    "contract:full contract-review:full",
	}
}

func profileAccountOrigin(profileName string) string {
	if profileName == developmentProfileName {
		return developmentAccountOrigin
	}
	return productionAccountOrigin
}

func developmentProfileError() error {
	return fmt.Errorf("dev integration requires an isolated %q profile with dev-only URLs and credentials; run `contract-cli config add --env dev --name contract-dev` and authorize again", developmentProfileName)
}

func (a *App) saveEnvironmentProfile(profile config.Profile) error {
	if profile.Environment != developmentEnvironment {
		return a.store.UpsertProfile(profile, true)
	}
	cfg, err := a.store.Load()
	if err != nil {
		return err
	}
	cfg.Profiles[profile.Name] = profile
	// Do not select dev as the default, even for a fresh config directory.
	return a.store.Save(cfg)
}

func validateDevelopmentProfile(profile config.Profile) error {
	if profile.Environment != developmentEnvironment || profile.Name != developmentProfileName {
		return developmentProfileError()
	}
	if profile.Identities.App.SecretRef != "" && profile.Identities.App.SecretRef != config.AppSecretKey(developmentProfileName) {
		return developmentProfileError()
	}
	return validateProfileOrigins(profile, developmentOpenPlatformOrigin, developmentAccountOrigin, developmentProfileError())
}

type developmentNetworkKey struct{}

// Clone the client instead of relaxing the shared production transport. Every
// request (including a redirect) is restricted to the selected dev origins.
func clientForEnvironment(client *http.Client, environment string) *http.Client {
	if environment != developmentEnvironment {
		return client
	}
	cloned := *client
	next := client.Transport
	if next == nil {
		next = http.DefaultTransport
	}
	cloned.Transport = developmentTransport{next: next}
	return &cloned
}

type developmentTransport struct{ next http.RoundTripper }

func (transport developmentTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if !isProductionOriginURL(request.URL.String(), developmentOpenPlatformOrigin, true) &&
		!isProductionOriginURL(request.URL.String(), developmentAccountOrigin, true) {
		if request.Body != nil {
			_ = request.Body.Close()
		}
		return nil, fmt.Errorf("dev integration blocks non-dev destination %q", request.URL.Host)
	}
	request = request.Clone(context.WithValue(request.Context(), developmentNetworkKey{}, true))
	return transport.next.RoundTrip(request)
}
