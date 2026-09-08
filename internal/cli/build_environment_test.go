package cli

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestBuildEnvironmentUsesTestDefaultsWithoutReadingProd(t *testing.T) {
	env := nonProductionEnvironments["test"]
	app, _ := newDevelopmentTestApp(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != env.preset().AuthorizationServerMetadataURL {
			t.Fatalf("wrong endpoint: %s", req.URL)
		}
		return productionResponse(200, environmentMetadata(env)), nil
	})
	app.buildEnvironment = "test"
	if err := app.store.UpsertProfile(validProductionProfile(), true); err != nil {
		t.Fatal(err)
	}
	if err := app.runConfigAdd(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	p, err := app.loadDeviceAwareProfile("")
	if err != nil {
		t.Fatal(err)
	}
	if p.Environment != "test" || p.Name != "contract-test" || p.Identities.User.DeviceClientID != "zscli_892efdadc11a3f53" {
		t.Fatalf("wrong defaults: %s %s", p.Name, p.Environment)
	}
	for _, name := range []string{"prod", "dev", "blue"} {
		if err := app.runConfigAdd(context.Background(), []string{"--env", name}); err == nil || !strings.Contains(err.Error(), "only supports test") {
			t.Fatalf("allowed %s: %v", name, err)
		}
	}
	if _, err := app.loadDeviceAwareProfile("contract"); err == nil {
		t.Fatal("accepted old prod profile")
	}
}

func TestBuildEnvironmentRejectsForeignProfiles(t *testing.T) {
	app, _ := newDevelopmentTestApp(t, func(*http.Request) (*http.Response, error) { t.Fatal("unexpected network"); return nil, nil })
	for _, name := range []string{"prod", "dev", "test", "blue"} {
		app.buildEnvironment = name
		p := validProductionProfile()
		if name != "prod" {
			p = environmentProfile(nonProductionEnvironments[name])
		}
		if err := app.validateBuildProfile(p); err != nil {
			t.Fatal(err)
		}
		p.Environment = "foreign"
		if err := app.validateBuildProfile(p); err == nil {
			t.Fatal("accepted foreign profile")
		}
	}
}
