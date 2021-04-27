package surveillee

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// var collector *prometheus.GaugeVec

// func init() {
// 	collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
// 		Name: "containerpilot_watch_file_instances",
// 		Help: "gauge of instances found for each ContainerPilot watch, partitioned by service",
// 	}, []string{"service"})
// 	prometheus.MustRegister(collector)
// }

// FileWatcher wraps the surveillee backend for checking changes
// in watched files and tracks their the md5 checksums.
type FileWatcher struct {
	lock      sync.RWMutex
	checksums map[string]string
}

// NewFileWatcher returns a new FileWatcher.
func NewFileWatcher() *FileWatcher {
	return &FileWatcher{sync.RWMutex{}, make(map[string]string)}
}

func (w *FileWatcher) md5sum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CheckForUpstreamChanges computes md5sum of a file and checks
// whether there has been a change since the last check.
func (w *FileWatcher) CheckForUpstreamChanges(fields ...string) (hasChanged, isHealthy bool) {
	filepath := fields[0]
	isHealthy = true
	newMD5sum, err := w.md5sum(filepath)
	if err != nil {
		log.Warnf("failed to get md5sum of %v: %s", filepath, err)
		return false, false
	}
	hasChanged = w.compareAndSwap(filepath, newMD5sum)
	return hasChanged, isHealthy
}

// returns true if md5sum for the file has changed and updates
// the internal state, returns false if it's a first time update
func (w *FileWatcher) compareAndSwap(filepath, newMD5sum string) bool {
	w.lock.Lock()
	defer w.lock.Unlock()
	oldMD5sum, ok := w.checksums[filepath]
	w.checksums[filepath] = newMD5sum
	if !ok {
		return false
	}
	// fmt.Println(oldMD5sum, newMD5sum)
	return oldMD5sum != newMD5sum
}
