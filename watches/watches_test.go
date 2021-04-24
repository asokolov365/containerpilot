package watches

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asokolov365/containerpilot/events"
	"github.com/asokolov365/containerpilot/surveillee"
	"github.com/asokolov365/containerpilot/tests/mocks"
)

func TestWatchConsulPollOk(t *testing.T) {
	cfg := &Config{
		Name: "myWatchConsulOk",
		Poll: 1,
	}
	// this discovery backend will always return true when we check
	// it for changed
	survSvcs := surveillee.NewServices(&mocks.NoopDiscoveryBackend{Val: true}, nil, nil)
	err := cfg.Validate(survSvcs)
	if err != nil {
		t.Fatalf("unable to validate config for %s: %s", cfg.Name, err)
	}
	if cfg.surveilService == nil {
		t.Fatalf("expected surveillee.Discovery but got nil")
	}
	got := runWatchTest(cfg, 5)
	changed := events.Event{Code: events.StatusChanged, Source: "watch.myWatchConsulOk"}
	healthy := events.Event{Code: events.StatusHealthy, Source: "watch.myWatchConsulOk"}
	if got[changed] != 1 || got[healthy] != 1 {
		t.Fatalf("expected 2 successful StatusHealthy events but got %v", got)
	}
}

func TestWatchConsulPollFail(t *testing.T) {
	cfg := &Config{
		Name: "myWatchConsulFail",
		Poll: 1,
	}
	// this discovery backend will always return false when we check
	// it for changed
	survSvcs := surveillee.NewServices(&mocks.NoopDiscoveryBackend{Val: false}, nil, nil)
	err := cfg.Validate(survSvcs)
	if err != nil {
		t.Fatalf("unable to validate config for %s: %s", cfg.Name, err)
	}
	if cfg.surveilService == nil {
		t.Fatalf("expected surveillee.Discovery but got nil")
	}
	got := runWatchTest(cfg, 3)
	changed := events.Event{Code: events.StatusChanged, Source: "watch.myWatchConsulFail"}
	unhealthy := events.Event{Code: events.StatusUnhealthy, Source: "watch.myWatchConsulFail"}
	if got[changed] != 0 || got[unhealthy] != 0 {
		t.Fatalf("expected 2 failed poll events without changes, but got %v", got)
	}
}

func TestWatchFilePollOk(t *testing.T) {
	testFile, err := ioutil.TempFile("/tmp", "WatchFilePollOk-")
	if err != nil {
		t.Fatalf("unable to create temp file: %s", err)
	}
	data := []byte("initial state\n")
	err = ioutil.WriteFile(testFile.Name(), data, 0644)
	if err != nil {
		t.Fatalf("unable to write to %s: %s", testFile.Name(), err)
	}
	testFile.Close()
	defer os.Remove(testFile.Name())

	survSvcs := surveillee.NewServices(nil, &mocks.NoopFileWatcherBackend{Val: true}, nil)
	cfg := &Config{
		Name:   testFile.Name(),
		Source: "file",
		Poll:   1,
	}
	err = cfg.Validate(survSvcs)
	if err != nil {
		t.Fatalf("unable to validate config for %s: %s", testFile.Name(), err)
	}
	if cfg.surveilService == nil {
		t.Fatalf("expected surveillee.FileWatcher but got nil")
	}
	got := runWatchTest(cfg, 5)
	changed := events.Event{Code: events.StatusChanged, Source: "watch." + testFile.Name()}
	healthy := events.Event{Code: events.StatusHealthy, Source: "watch." + testFile.Name()}
	if got[changed] != 1 || got[healthy] != 1 {
		t.Fatalf("expected 2 successful StatusHealthy events but got %v", got)
	}
}

func TestWatchFilePollFail(t *testing.T) {
	testFile, err := ioutil.TempFile("/tmp", "WatchPollFileOk-")
	if err != nil {
		t.Fatalf("unable to create temp file: %s", err)
	}
	data := []byte("initial state\n")
	err = ioutil.WriteFile(testFile.Name(), data, 0644)
	if err != nil {
		t.Fatalf("unable to write to %s: %s", testFile.Name(), err)
	}
	testFile.Close()
	defer os.Remove(testFile.Name())

	survSvcs := surveillee.NewServices(nil, &mocks.NoopFileWatcherBackend{Val: false}, nil)
	cfg := &Config{
		Name:   testFile.Name(),
		Source: "file",
		Poll:   1,
	}
	err = cfg.Validate(survSvcs)
	if err != nil {
		t.Fatalf("unable to validate config for %s: %s", testFile.Name(), err)
	}
	if cfg.surveilService == nil {
		t.Fatalf("expected surveillee.FileWatcher but got nil")
	}
	got := runWatchTest(cfg, 3)
	changed := events.Event{Code: events.StatusChanged, Source: "watch." + testFile.Name()}
	unhealthy := events.Event{Code: events.StatusUnhealthy, Source: "watch." + testFile.Name()}
	if got[changed] != 0 || got[unhealthy] != 0 {
		t.Fatalf("expected 2 failed poll events without changes, but got %v", got)
	}
}

func TestWatchVaultPollOk(t *testing.T) {
	cfg := &Config{
		Name:   "secret/data/test",
		Source: "vault",
		Poll:   1,
	}
	survSvcs := surveillee.NewServices(nil, nil, &mocks.NoopSecretStorageBackend{Val: true})
	err := cfg.Validate(survSvcs)
	if err != nil {
		t.Fatalf("unable to validate config for %s: %s", cfg.Name, err)
	}
	if cfg.surveilService == nil {
		t.Fatalf("expected surveillee.SecretStorage but got <nil>")
	}
	got := runWatchTest(cfg, 5)
	changed := events.Event{Code: events.StatusChanged, Source: "watch.secret/data/test"}
	healthy := events.Event{Code: events.StatusHealthy, Source: "watch.secret/data/test"}
	if got[changed] != 1 || got[healthy] != 1 {
		t.Fatalf("expected 2 successful StatusHealthy events but got %v", got)
	}
}

func TestWatchVaultPollFail(t *testing.T) {
	cfg := &Config{
		Name:   "secret/data/test",
		Source: "vault",
		Poll:   1,
	}
	// this discovery backend will always return false when we check
	// it for changed
	survSvcs := surveillee.NewServices(nil, nil, &mocks.NoopSecretStorageBackend{Val: false})
	err := cfg.Validate(survSvcs)
	if err != nil {
		t.Fatalf("unable to validate config for %s: %s", cfg.Name, err)
	}
	if cfg.surveilService == nil {
		t.Fatalf("expected surveillee.SecretStorage but got <nil>")
	}
	got := runWatchTest(cfg, 3)
	changed := events.Event{Code: events.StatusChanged, Source: "watch.secret/data/test"}
	unhealthy := events.Event{Code: events.StatusUnhealthy, Source: "watch.secret/data/test"}
	if got[changed] != 0 || got[unhealthy] != 0 {
		t.Fatalf("expected 2 failed poll events without changes, but got %v", got)
	}
}

func runWatchTest(cfg *Config, count int) map[events.Event]int {
	bus := events.NewEventBus()
	watch := NewWatch(cfg)
	ctx := context.Background()
	watch.Run(ctx, bus)
	poll1 := events.Event{Code: events.TimerExpired, Source: fmt.Sprintf("%s.poll", cfg.Name)}
	poll2 := events.Event{Code: events.StatusChanged, Source: fmt.Sprintf("%s.poll", cfg.Name)}
	watch.Receive(poll1)
	watch.Receive(poll2) // Ensure we can run it more than once
	watch.Receive(events.QuitByTest)
	bus.Wait()
	results := bus.DebugEvents()

	got := make(map[events.Event]int)
	for _, result := range results {
		got[result]++
	}
	return got
}
