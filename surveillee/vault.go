package surveillee

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

// Vault wraps the secrets storage backend for the Hashicorp Vault client
// and tracks the state of all watched dependencies.
type Vault struct {
	*api.Client
	tokenFile      string
	lock           sync.RWMutex
	watchedSecrets map[string]*api.Secret
}

// NewVault creates a new secrets storage backend for Vault
func NewVault(config interface{}) (*Vault, error) {
	var vaultConfig *api.Config
	var rawToken, token, tokenFile string
	var err error

	switch t := config.(type) {
	case string:
		vaultConfig, rawToken, err = vaultConfigFromURI(t)
	case map[string]interface{}:
		vaultConfig, rawToken, err = vaultConfigFromMap(t)
	default:
		return nil, fmt.Errorf("no secrets storage backend defined")
	}
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(rawToken, "file://") {
		tokenFile = strings.Replace(rawToken, "file://", "", 1)
		tokenBytes, err := os.ReadFile(tokenFile)
		if err != nil {
			return nil, err
		}
		token = strings.TrimSpace(string(tokenBytes))
	} else {
		token = rawToken
	}

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	vault := &Vault{client, tokenFile, sync.RWMutex{}, make(map[string]*api.Secret)}
	return vault, nil
}

// CheckForUpstreamChanges requests the set of healthy instances of a
// service from Consul and checks whether there has been a change since
// the last check.
func (v *Vault) CheckForUpstreamChanges(fields ...string) (hasChanged, isHealthy bool) {
	requestPath, field := fields[0], fields[1]
	isHealthy = true
	if v.tokenFile != "" {
		v.refreshTokenFromFile()
	}
	newSecret, err := v.readSecret(requestPath)
	if err != nil {
		log.Warnf("vault read %s failed: %v", requestPath, err)
		return false, false
	}
	hasChanged = v.compareAndSwap(requestPath, field, newSecret)
	return hasChanged, isHealthy
}

// returns true if secret.field has changed and updates
// the internal state, returns false if it's a first time update
func (v *Vault) compareAndSwap(requestPath, field string, newSecret *api.Secret) bool {
	v.lock.Lock()
	defer v.lock.Unlock()
	oldSecret, ok := v.watchedSecrets[requestPath]
	v.watchedSecrets[requestPath] = newSecret
	if !ok {
		return false
	}
	oldData := getDataFromSecret(oldSecret)
	newData := getDataFromSecret(newSecret)
	if field == "" {
		// comare two maps
		if len(oldData) != len(newData) {
			return true
		}
		for k, oldValue := range oldData {
			newValue, ok := newData[k]
			if !ok {
				return true
			}
			if oldValue != newValue {
				return true
			}
		}
		return false
	}
	// comare fields of two maps
	oldValue, ok := oldData[field]
	if !ok {
		return false
	}
	newValue, ok := newData[field]
	if !ok {
		return false
	}
	return oldValue != newValue
}

func (v *Vault) refreshTokenFromFile() error {
	if v.tokenFile == "" {
		return nil
	}
	tokenBytes, err := os.ReadFile(v.tokenFile)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(string(tokenBytes))
	v.SetToken(token)
	return nil
}

func (v *Vault) readSecret(requestPath string) (*api.Secret, error) {
	secret, err := v.Logical().Read(requestPath)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, fmt.Errorf("%s not found", requestPath)
	}
	if len(secret.Warnings) > 0 {
		for _, w := range secret.Warnings {
			log.Warnf("vault read %s [warn]: %s", requestPath, w)
		}
	}
	return secret, nil
}

func getDataFromSecret(secret *api.Secret) map[string]interface{} {
	var m map[string]interface{}
	if _, ok := secret.Data["data"].(map[string]interface{}); ok {
		m = secret.Data["data"].(map[string]interface{})
	} else {
		m = secret.Data
	}
	return m
}
