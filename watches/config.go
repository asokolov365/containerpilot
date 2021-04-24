package watches

import (
	"fmt"
	"reflect"

	"github.com/asokolov365/containerpilot/config/decode"
	"github.com/asokolov365/containerpilot/config/services"
	"github.com/asokolov365/containerpilot/surveillee"
)

// Config configures the watch
type Config struct {
	Name           string `mapstructure:"name"`
	serviceName    string
	Source         string `mapstructure:"source"`
	Poll           int    `mapstructure:"interval"` // time in seconds
	Tag            string `mapstructure:"tag"`
	DC             string `mapstructure:"dc"` // Consul datacenter
	surveilService surveillee.Backend
}

// NewConfigs parses json config into a validated slice of Configs
func NewConfigs(raw []interface{}, survSvcs *surveillee.Services) ([]*Config, error) {
	var watches []*Config
	if raw == nil {
		return watches, nil
	}
	if err := decode.ToStruct(raw, &watches); err != nil {
		return watches, fmt.Errorf("Watch configuration error: %v", err)
	}
	for _, watch := range watches {
		if err := watch.Validate(survSvcs); err != nil {
			return watches, err
		}
	}
	return watches, nil
}

// Validate ensures Config meets all requirements and assigns surveilService
// accordingly to Source field.
func (cfg *Config) Validate(survSvcs *surveillee.Services) error {
	cfg.serviceName = cfg.Name
	cfg.Name = "watch." + cfg.Name

	if cfg.Poll < 1 {
		return fmt.Errorf("watch[%s].interval must be > 0", cfg.serviceName)
	}
	switch cfg.Source {
	case "consul":
		cfg.surveilService = survSvcs.Discovery
	case "vault":
		if survSvcs.SecretStorage == nil || reflect.ValueOf(survSvcs.SecretStorage).IsNil() {
			return fmt.Errorf("watch[%s].source is vault but vault config is not defined", cfg.serviceName)
		}
		cfg.surveilService = survSvcs.SecretStorage
	case "file":
		cfg.surveilService = survSvcs.FileWatcher
	default:
		cfg.surveilService = survSvcs.Discovery
	}
	if err := services.ValidateName(cfg.serviceName, cfg.Source); err != nil {
		return err
	}
	return nil
}

// String implements the stdlib fmt.Stringer interface for pretty-printing
func (cfg *Config) String() string {
	return "watches.Config[" + cfg.Name + "]"
}
