package surveillee

import (
	"fmt"
	"os"
	"strings"

	"github.com/asokolov365/containerpilot/config/decode"
	"github.com/hashicorp/vault/api"
)

// VaultConfig is used to configure the creation of the Hashicorp Vault client.
type VaultConfig struct {
	Address string         `mapstructure:"address"`
	Scheme  string         `mapstructure:"scheme"`
	Token   string         `mapstructure:"token"`
	TLS     VaultTLSConfig `mapstructure:"tls"` // optional TLS settings
}

// VaultTLSConfig is optional TLS settings for VaultConfig.
type VaultTLSConfig struct {
	HTTPCACert        string `mapstructure:"cafile"`
	HTTPCAPath        string `mapstructure:"capath"`
	HTTPClientCert    string `mapstructure:"clientcert"`
	HTTPClientKey     string `mapstructure:"clientkey"`
	HTTPTLSServerName string `mapstructure:"servername"`
	HTTPSSLVerify     bool   `mapstructure:"verify"`
}

// override an already-parsed VaultConfig with any options that might
// be set in the environment and then return the TLSConfig
func getVaultTLSConfig(cfg *VaultConfig) *api.TLSConfig {
	if cafile := os.Getenv("VAULT_CACERT"); cafile != "" {
		cfg.TLS.HTTPCACert = cafile
	}
	if capath := os.Getenv("VAULT_CAPATH"); capath != "" {
		cfg.TLS.HTTPCAPath = capath
	}
	if clientCert := os.Getenv("VAULT_CLIENT_CERT"); clientCert != "" {
		cfg.TLS.HTTPClientCert = clientCert
	}
	if clientKey := os.Getenv("VAULT_CLIENT_KEY"); clientKey != "" {
		cfg.TLS.HTTPClientKey = clientKey
	}
	if serverName := os.Getenv("VAULT_TLS_SERVER_NAME"); serverName != "" {
		cfg.TLS.HTTPTLSServerName = serverName
	}
	skipVerify := os.Getenv("VAULT_SKIP_VERIFY")
	switch strings.ToLower(skipVerify) {
	case "1", "true":
		cfg.TLS.HTTPSSLVerify = false
	case "0", "false":
		cfg.TLS.HTTPSSLVerify = true
	}
	tlsConfig := &api.TLSConfig{
		CACert:        cfg.TLS.HTTPCACert,
		CAPath:        cfg.TLS.HTTPCAPath,
		ClientCert:    cfg.TLS.HTTPClientCert,
		ClientKey:     cfg.TLS.HTTPClientKey,
		TLSServerName: cfg.TLS.HTTPTLSServerName,
		Insecure:      !cfg.TLS.HTTPSSLVerify,
	}
	return tlsConfig
}

func vaultConfigFromMap(raw map[string]interface{}) (*api.Config, string, error) {
	parsed := &VaultConfig{}
	if err := decode.ToStruct(raw, parsed); err != nil {
		return nil, "", err
	}
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		parsed.Token = token
	}
	parsed.Token = strings.TrimSpace(parsed.Token)
	if parsed.Token == "" {
		return nil, "", fmt.Errorf("no vault token defined")
	}
	tlsConfig := getVaultTLSConfig(parsed)
	config := api.DefaultConfig()
	config.Address = parsed.Scheme + "://" + parsed.Address
	if err := config.ConfigureTLS(tlsConfig); err != nil {
		return nil, "", err
	}
	if err := config.ReadEnvironment(); err != nil {
		return nil, "", err
	}
	return config, parsed.Token, nil
}

func vaultConfigFromURI(uri string) (*api.Config, string, error) {
	address, scheme := parseRawVaultURI(uri)
	parsed := &VaultConfig{Address: address, Scheme: scheme}
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		parsed.Token = strings.TrimSpace(token)
	}
	if parsed.Token == "" {
		return nil, "", fmt.Errorf("no vault token defined")
	}
	tlsConfig := getVaultTLSConfig(parsed)
	config := api.DefaultConfig()
	config.Address = parsed.Scheme + "://" + parsed.Address
	if err := config.ConfigureTLS(tlsConfig); err != nil {
		return nil, "", err
	}
	if err := config.ReadEnvironment(); err != nil {
		return nil, "", err
	}
	return config, parsed.Token, nil
}

// Returns the uri broken into an address and scheme portion
func parseRawVaultURI(raw string) (string, string) {

	var scheme = "http" // default
	var address = raw   // we accept bare address w/o a scheme

	// strip the scheme from the prefix and (maybe) set the scheme to https
	if strings.HasPrefix(raw, "http://") {
		address = strings.Replace(raw, "http://", "", 1)
	} else if strings.HasPrefix(raw, "https://") {
		address = strings.Replace(raw, "https://", "", 1)
		scheme = "https"
	}
	return address, scheme
}
