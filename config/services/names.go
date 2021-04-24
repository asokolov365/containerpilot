package services

import (
	"fmt"
	"regexp"
)

var validConsulServiceName = regexp.MustCompile(`^[a-z][a-zA-Z0-9\-]+$`)
var validVaultPathName = regexp.MustCompile(`^([a-zA-Z0-9\/\_\-\s\.]+)+$`)
var validFilePathName = regexp.MustCompile(`^(\/[a-zA-Z0-9\_\-\s\.]+)+(\.[a-zA-Z0-9]+)?$`)

// ValidateName checks if the service name passed as an argument
// is is alpha-numeric with dashes. This ensures compliance with both DNS
// and discovery backends.
func ValidateName(name, svcType string) error {
	if name == "" {
		return fmt.Errorf("'name' must not be blank")
	}
	switch svcType {
	case "vault":
		if ok := validVaultPathName.MatchString(name); !ok {
			return fmt.Errorf("service name must be valid vault path")
		}
	case "file":
		if ok := validFilePathName.MatchString(name); !ok {
			return fmt.Errorf("service name must be valid file path")
		}
	default:
		if ok := validConsulServiceName.MatchString(name); !ok {
			return fmt.Errorf("service name must be alphanumeric with dashes to comply with service discovery")
		}
	}
	return nil
}
