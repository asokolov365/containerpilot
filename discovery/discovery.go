// Package discovery manages the configuration of the Consul clients and
// the functions used to update/query Consul with service discovery data.
package discovery

import "github.com/hashicorp/consul/api"

// Backend is an interface which all service discovery backends must implement
type Backend interface {
	CheckForUpstreamChanges(fields ...string) (bool, bool)
	CheckRegister(check *api.AgentCheckRegistration) error
	UpdateTTL(checkID, output, status string) error
	ServiceDeregister(serviceID string) error
	ServiceRegister(service *api.AgentServiceRegistration) error
}
