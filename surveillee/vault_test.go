package surveillee

import (
	"os"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestVaultObjectParse(t *testing.T) {
	var err error
	rawCfgMap := map[string]interface{}{
		"address": "vault:8201",
		"scheme":  "https",
		"token":   "s.smbuEimMcVx3lzk4EdyxxHi7",
	}
	_, err = NewVault(rawCfgMap)
	if err != nil {
		t.Fatalf("unable to parse config: %v", err)
	}
	rawCfgStr := "http://vault:8200"
	_, err = NewVault(rawCfgStr)
	assert.EqualError(t, err, "no vault token defined")

	os.Setenv("VAULT_TOKEN", "myTestToken")
	defer os.Unsetenv("VAULT_TOKEN")
	_, err = NewVault(rawCfgStr)
	if err != nil {
		t.Fatalf("unable to parse config: %v", err)
	}
}

func TestVaultAddressParse(t *testing.T) {
	// typical valid entries
	runParseVaultTest(t, "https://vault:8200", "vault:8200", "https")
	runParseVaultTest(t, "http://vault:8200", "vault:8200", "http")
	runParseVaultTest(t, "vault:8200", "vault:8200", "http")

	// malformed URI: we won't even try to fix these and just let them bubble up
	// to the Consul API call where it'll fail there.
	runParseVaultTest(t, "httpshttps://vault:8200", "httpshttps://vault:8200", "http")
	runParseVaultTest(t, "https://https://vault:8200", "https://vault:8200", "https")
	runParseVaultTest(t, "http://https://vault:8200", "https://vault:8200", "http")
	runParseVaultTest(t, "vault:8200https://", "vault:8200https://", "http")
	runParseVaultTest(t, "", "", "http")
}

func runParseVaultTest(t *testing.T, uri, expectedAddress, expectedScheme string) {
	address, scheme := parseRawVaultURI(uri)
	if address != expectedAddress || scheme != expectedScheme {
		t.Fatalf("Expected %s over %s but got %s over %s",
			expectedAddress, expectedScheme, address, scheme)
	}
}

func TestCheckForVaultChanges(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "myTestToken")
	defer os.Unsetenv("VAULT_TOKEN")
	v, err := NewVault("http://localhost:8200")
	if err != nil {
		t.Fatalf("unable to create vault: %v", err)
	}

	var rawData = map[string]interface{}{
		"foo":     "bar",
		"version": 1,
	}
	t0 := &vault.Secret{
		Data: rawData,
	}
	hasChanged := v.compareAndSwap("secret/data/test", "", t0)
	assert.False(t, hasChanged, "value for 'hasChanged' after t0")

	rawData = map[string]interface{}{
		"foo":     "bar",
		"version": 2,
	}
	t1 := &vault.Secret{
		Data: rawData,
	}
	hasChanged = v.compareAndSwap("secret/data/test", "", t1)
	assert.True(t, hasChanged, "value for 'hasChanged' after t1")

	hasChanged = v.compareAndSwap("secret/data/test", "", t0)
	assert.True(t, hasChanged, "value for 'hasChanged' after t0 (again)")

	hasChanged = v.compareAndSwap("secret/data/test", "", t1)
	assert.True(t, hasChanged, "value for 'hasChanged' after t1 (again)")

	rawData = map[string]interface{}{
		"foo":     "car",
		"version": 1,
	}
	t3 := &vault.Secret{
		Data: rawData,
	}
	hasChanged = v.compareAndSwap("secret/data/test", "", t3)
	assert.True(t, hasChanged, "value for 'hasChanged' after t3")

	hasChanged = v.compareAndSwap("secret/data/test", "foo", t0)
	assert.True(t, hasChanged, "value for 'hasChanged' after t0 (again with field)")
	hasChanged = v.compareAndSwap("secret/data/test", "foo", t1)
	assert.False(t, hasChanged, "value for 'hasChanged' after t1 (again with field)")
	hasChanged = v.compareAndSwap("secret/data/test", "version", t0)
	assert.True(t, hasChanged, "value for 'hasChanged' after t0 (again with field)")
}

func TestWithVault(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "myTestToken")
	defer os.Unsetenv("VAULT_TOKEN")
	testServer, err := NewTestServer(8200)
	if err != nil {
		t.Fatal(err)
	}
	defer testServer.Stop()

	testServer.WaitForAPI()

	t.Run("TestVaultCheckForChanges", testVaultCheckForChanges(testServer))
}

func testVaultCheckForChanges(testServer *TestServer) func(*testing.T) {
	return func(t *testing.T) {
		path := "secret/data/test"
		vault, err := NewVault(testServer.HTTPAddr)
		if err != nil {
			t.Errorf("unable to create vault %v", err)
		}
		if changed, _ := vault.CheckForUpstreamChanges(path, ""); changed {
			t.Fatalf("First read of %s should show `false` for change", path)
		}
		data := make(map[string]interface{})
		data["data"] = map[string]interface{}{
			"foo":     "bar",
			"version": 1,
		}
		_, err = vault.Logical().Write(path, data)
		if err != nil {
			t.Errorf("unable to write secret to vault: %v", err)
		}
		if changed, _ := vault.CheckForUpstreamChanges(path, ""); changed {
			t.Errorf("%s should not have changed after first write", path)
		}

		data["data"] = map[string]interface{}{
			"foo":     "bar",
			"version": 2,
		}
		_, err = vault.Logical().Write(path, data)
		if err != nil {
			t.Errorf("unable to write secret to vault: %v", err)
		}

		if changed, _ := vault.CheckForUpstreamChanges(path, ""); !changed {
			t.Errorf("%s should have changed after second write", path)
		}
		if changed, _ := vault.CheckForUpstreamChanges(path, "foo"); changed {
			t.Errorf("%s field 'foo' should not have changed after second write", path)
		}
	}
}

func TestWithVaultFailed(t *testing.T) {
	os.Setenv("VAULT_TOKEN", "myTestToken")
	defer os.Unsetenv("VAULT_TOKEN")
	testServer, err := NewTestServer(8200)
	if err != nil {
		t.Fatal(err)
	}
	defer testServer.Stop()
	testServer.WaitForAPI()

	path := "secret/data/test"
	vault, err := NewVault(testServer.HTTPAddr)
	if err != nil {
		t.Errorf("unable to create vault %v", err)
	}

	if changed, _ := vault.CheckForUpstreamChanges(path, ""); changed {
		t.Fatalf("First read of %s should show `false` for change", path)
	}

	data := make(map[string]interface{})
	data["data"] = map[string]interface{}{
		"foo":     "bar",
		"version": 1,
	}
	_, err = vault.Logical().Write(path, data)
	if err != nil {
		t.Errorf("unable to write secret to vault: %v", err)
	}

	if changed, _ := vault.CheckForUpstreamChanges(path, ""); changed {
		t.Errorf("%s should not have changed after first write", path)
	}

	testServer.Stop()

	if changed, healthy := vault.CheckForUpstreamChanges(path, ""); changed || healthy {
		t.Errorf("%s should not have changed or healthy while vault is down", path)
	}

	testServer, err = NewTestServer(8200)
	if err != nil {
		t.Fatal(err)
	}
	defer testServer.Stop()
	testServer.WaitForAPI()

	_, err = vault.Logical().Write(path, data)
	if err != nil {
		t.Errorf("unable to write secret to vault: %v", err)
	}

	changed, healthy := vault.CheckForUpstreamChanges(path, "")
	if changed || !healthy {
		t.Errorf("%s should not have changed or not healthy after vault is restarted", path)
	}
}
