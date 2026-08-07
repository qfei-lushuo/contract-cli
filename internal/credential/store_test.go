package credential

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cn.qfei/contract-cli/internal/config"
)

func TestPendingTransactionStatusDefaultsToPendingForLegacyCredentials(t *testing.T) {
	var pending PendingTransaction
	if err := json.Unmarshal([]byte(`{"device_code":"device-a"}`), &pending); err != nil {
		t.Fatal(err)
	}
	if pending.EffectiveStatus() != PendingStatusPending {
		t.Fatalf("effective status = %q, want pending", pending.EffectiveStatus())
	}
}

func TestDoubaoStoreEncryptsCredentialAndDoesNotPersistPlaintext(t *testing.T) {
	workspace := t.TempDir()
	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	store, err := NewStore(Options{LookupEnv: envLookup(map[string]string{
		"SKILL_SESSION_WORKSPACE":        workspace,
		"CONTRACT_CLI_CREDENTIAL_KEY_V1": key,
	})})
	if err != nil {
		t.Fatal(err)
	}
	credential := DeviceCredential{
		Pending: &PendingTransaction{DeviceCode: "secret-device-code"},
		Token:   &config.Token{AccessToken: "secret-access", RefreshToken: "secret-refresh", Expiry: time.Now().Add(time.Hour)},
		DeviceProfile: &DeviceProfile{
			Name:                        "contract",
			Environment:                 "prod",
			OpenPlatformBaseURL:         "https://open.qfei.cn",
			Resource:                    "https://open.qfei.cn",
			BusinessType:                "contract",
			DeviceClientID:              "device-client",
			DeviceScope:                 "contract:full",
			DeviceAuthorizationEndpoint: "https://myaccount.qfei.cn/api/public/oauth/device-authorization/contract",
			TokenEndpoint:               "https://myaccount.qfei.cn/api/public/oauth/token/contract",
			RevocationEndpoint:          "https://myaccount.qfei.cn/api/public/oauth/revoke/contract",
		},
	}
	if err := store.Save("contract", credential); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(workspace, ".contract-cli", "credentials", profileFileName("contract"))
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{
		"secret-device-code",
		"secret-access",
		"secret-refresh",
		"https://open.qfei.cn",
		"device-client",
	} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("encrypted file contains %q", secret)
		}
	}
	loaded, err := store.Load("contract")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Token == nil || loaded.Token.AccessToken != "secret-access" || loaded.Pending == nil || loaded.Pending.DeviceCode != "secret-device-code" {
		t.Fatalf("unexpected loaded credential: %+v", loaded)
	}
	if loaded.DeviceProfile == nil || loaded.DeviceProfile.Name != "contract" || loaded.DeviceProfile.DeviceScope != "contract:full" {
		t.Fatalf("unexpected loaded device profile: %+v", loaded.DeviceProfile)
	}
}

func TestDoubaoStoreRejectsMissingOrInvalidKey(t *testing.T) {
	for _, key := range []string{"", "not-base64"} {
		_, err := NewStore(Options{LookupEnv: envLookup(map[string]string{
			"SKILL_SESSION_WORKSPACE":        t.TempDir(),
			"CONTRACT_CLI_CREDENTIAL_KEY_V1": key,
		})})
		if err == nil {
			t.Fatalf("key %q should fail", key)
		}
	}
}

func TestWorkBuddyStoreRequiresTaskIDAndUsesItAsNamespace(t *testing.T) {
	backend := &memoryKeyring{values: map[string]string{}}
	_, err := NewStore(Options{LookupEnv: envLookup(map[string]string{}), Keyring: backend})
	if err == nil || !strings.Contains(err.Error(), "SKILL_SESSION_WORKSPACE") || !strings.Contains(err.Error(), "CODEBUDDY_SESSION_ID") {
		t.Fatalf("error = %v", err)
	}

	storeA, err := NewStore(Options{LookupEnv: envLookup(map[string]string{"CODEBUDDY_SESSION_ID": "task-a"}), Keyring: backend})
	if err != nil {
		t.Fatal(err)
	}
	storeB, err := NewStore(Options{LookupEnv: envLookup(map[string]string{"CODEBUDDY_SESSION_ID": "task-b"}), Keyring: backend})
	if err != nil {
		t.Fatal(err)
	}
	if err := storeA.Save("contract", DeviceCredential{Token: &config.Token{AccessToken: "token-a"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := storeB.Load("contract"); err == nil {
		t.Fatal("task-b unexpectedly loaded task-a credential")
	}
	if _, ok := backend.values[keyringService+":task-a:contract"]; !ok {
		t.Fatal("WorkBuddy keyring account changed and would invalidate existing credentials")
	}
}

func TestDeviceRuntimeRejectsRemovedDoubaoLocalEnvironment(t *testing.T) {
	lookupEnv := envLookup(map[string]string{
		"DOUBAO_SESSION_ID": "removed-local-session",
	})
	_, err := ResolveDeviceRuntime(lookupEnv)
	if err == nil {
		t.Fatal("DOUBAO_SESSION_ID unexpectedly enabled the removed Doubao local runtime")
	}
	if !strings.Contains(err.Error(), "SKILL_SESSION_WORKSPACE") || !strings.Contains(err.Error(), "CODEBUDDY_SESSION_ID") {
		t.Fatalf("error = %v, want the two supported runtime contracts", err)
	}
	if _, err := NewStore(Options{LookupEnv: lookupEnv, Keyring: &memoryKeyring{values: map[string]string{}}}); err == nil {
		t.Fatal("DOUBAO_SESSION_ID unexpectedly created a Device CredentialStore")
	}
}

func TestDeviceRuntimeKeepsDoubaoCloudPrecedenceOverWorkBuddy(t *testing.T) {
	workspace := t.TempDir()
	runtimeContext, err := ResolveDeviceRuntime(envLookup(map[string]string{
		"SKILL_SESSION_WORKSPACE": workspace,
		"CODEBUDDY_SESSION_ID":    "workbuddy-task",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if runtimeContext.Kind != DeviceRuntimeDoubaoCloud || runtimeContext.Workspace != workspace {
		t.Fatalf("runtime = %+v, want Doubao cloud workspace", runtimeContext)
	}
}

type memoryKeyring struct{ values map[string]string }

func (m *memoryKeyring) Get(service, user string) (string, error) {
	value, ok := m.values[service+":"+user]
	if !ok {
		return "", ErrCredentialNotFound
	}
	return value, nil
}
func (m *memoryKeyring) Set(service, user, value string) error {
	m.values[service+":"+user] = value
	return nil
}
func (m *memoryKeyring) Delete(service, user string) error {
	delete(m.values, service+":"+user)
	return nil
}

func envLookup(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
