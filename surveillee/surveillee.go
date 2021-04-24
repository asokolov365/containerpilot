// Package surveillee manages the configuration of various entities being monitored by watches
// and the functions used to query them for data.
package surveillee

// Backend is an interface which all surveillee entities must implement
type Backend interface {
	CheckForUpstreamChanges(fields ...string) (bool, bool) // returns hasChanged, isHealthy
}

// Services is a structure which contains all known entities that
// can be monitored for changes.
type Services struct {
	Discovery     Backend
	FileWatcher   Backend
	SecretStorage Backend
}

func NewServices(discovery Backend, fileWatcher Backend, secretStorage Backend) *Services {
	return &Services{
		Discovery:     discovery,
		FileWatcher:   fileWatcher,
		SecretStorage: secretStorage,
	}
}
