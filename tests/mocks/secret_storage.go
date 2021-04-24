package mocks

// NoopFileWatcherBackend is a mock surveillee.FileWatcher
type NoopSecretStorageBackend struct {
	Val     bool
	lastVal bool
}

// CheckForUpstreamChanges will return the public Val field to mock
// whether a change has occurred. Will not report a change on the second
// check unless the Val has been updated externally by the test rig
func (noop *NoopSecretStorageBackend) CheckForUpstreamChanges(fields ...string) (hasChanged, isHealthy bool) {
	if noop.lastVal != noop.Val {
		hasChanged = true
	}
	noop.lastVal = noop.Val
	isHealthy = noop.Val
	return hasChanged, isHealthy
}
