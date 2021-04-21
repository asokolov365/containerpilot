package discovery

import (
	"os"
	"strings"

	"github.com/asokolov365/containerpilot/config/decode"
	"github.com/hashicorp/consul/api"
)

type ConsulConfig struct {
	Address string          `mapstructure:"address"`
	Scheme  string          `mapstructure:"scheme"`
	Token   string          `mapstructure:"token"`
	TLS     ConsulTLSConfig `mapstructure:"tls"` // optional TLS settings
}

type ConsulTLSConfig struct {
	HTTPCAFile        string `mapstructure:"cafile"`
	HTTPCAPath        string `mapstructure:"capath"`
	HTTPClientCert    string `mapstructure:"clientcert"`
	HTTPClientKey     string `mapstructure:"clientkey"`
	HTTPTLSServerName string `mapstructure:"servername"`
	HTTPSSLVerify     bool   `mapstructure:"verify"`
}

// override an already-parsed ConsulConfig with any options that might
// be set in the environment and then return the TLSConfig
func getConsulTLSConfig(cfg *ConsulConfig) api.TLSConfig {
	if cafile := os.Getenv("CONSUL_CACERT"); cafile != "" {
		cfg.TLS.HTTPCAFile = cafile
	}
	if capath := os.Getenv("CONSUL_CAPATH"); capath != "" {
		cfg.TLS.HTTPCAPath = capath
	}
	if clientCert := os.Getenv("CONSUL_CLIENT_CERT"); clientCert != "" {
		cfg.TLS.HTTPClientCert = clientCert
	}
	if clientKey := os.Getenv("CONSUL_CLIENT_KEY"); clientKey != "" {
		cfg.TLS.HTTPClientKey = clientKey
	}
	if serverName := os.Getenv("CONSUL_TLS_SERVER_NAME"); serverName != "" {
		cfg.TLS.HTTPTLSServerName = serverName
	}
	verify := os.Getenv("CONSUL_HTTP_SSL_VERIFY")
	switch strings.ToLower(verify) {
	case "1", "true":
		cfg.TLS.HTTPSSLVerify = true
	case "0", "false":
		cfg.TLS.HTTPSSLVerify = false
	}
	tlsConfig := api.TLSConfig{
		Address:            cfg.TLS.HTTPTLSServerName,
		CAFile:             cfg.TLS.HTTPCAFile,
		CAPath:             cfg.TLS.HTTPCAPath,
		CertFile:           cfg.TLS.HTTPClientCert,
		KeyFile:            cfg.TLS.HTTPClientKey,
		InsecureSkipVerify: !cfg.TLS.HTTPSSLVerify,
	}
	return tlsConfig
}

func consulConfigFromMap(raw map[string]interface{}) (*api.Config, error) {
	parsed := &ConsulConfig{}
	if err := decode.ToStruct(raw, parsed); err != nil {
		return nil, err
	}
	config := &api.Config{
		Address:   parsed.Address,
		Scheme:    parsed.Scheme,
		Token:     parsed.Token,
		TLSConfig: getConsulTLSConfig(parsed),
	}
	return config, nil
}

func consulConfigFromURI(uri string) (*api.Config, error) {
	address, scheme := parseRawConsulURI(uri)
	parsed := &ConsulConfig{Address: address, Scheme: scheme}
	config := &api.Config{
		Address:   parsed.Address,
		Scheme:    parsed.Scheme,
		Token:     parsed.Token,
		TLSConfig: getConsulTLSConfig(parsed),
	}
	return config, nil
}

// Returns the uri broken into an address and scheme portion
func parseRawConsulURI(raw string) (string, string) {

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
