package surveillee

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForChanges(t *testing.T) {
	fileWatcher := NewFileWatcher()

	t0 := "70fb669a01c366a0f0057f061e58346d" // md5 -s 'initial state'
	hasChanged := fileWatcher.compareAndSwap("test", t0)
	assert.False(t, hasChanged, "value for 'hasChanged' after t0")

	hasChanged = fileWatcher.compareAndSwap("test", t0)
	assert.False(t, hasChanged, "value for 'hasChanged' after t0 (again)")

	t1 := "601feaabee8a3dadd739e808c552621e" // md5 -s 'first change'
	hasChanged = fileWatcher.compareAndSwap("test", t1)
	assert.True(t, hasChanged, "value for 'hasChanged' after t1")

	hasChanged = fileWatcher.compareAndSwap("test", t0)
	assert.True(t, hasChanged, "value for 'hasChanged' after t0 (again)")

	hasChanged = fileWatcher.compareAndSwap("test", t1)
	assert.True(t, hasChanged, "value for 'hasChanged' after t1 (again)")

	t3 := "70f37faf41088aa6a9008ccabd65ecf3" // md5 -s 'second change'
	hasChanged = fileWatcher.compareAndSwap("test", t3)
	assert.True(t, hasChanged, "value for 'hasChanged' after t3")
}

func TestWithFileWatcher(t *testing.T) {
	fileWatcher := NewFileWatcher()
	testFile, err := ioutil.TempFile("/tmp", "testFileWatcher-")
	if err != nil {
		t.Fatalf("unable to create temp file: %s", err)
	}
	testFile.Close()
	defer os.Remove(testFile.Name())

	t.Run("TestFileWatcherInitialStatus", testFileWatcherInitialStatus(fileWatcher, testFile))
	t.Run("TestFileWatcherCheckForChanges", testFileWatcherCheckForChanges(fileWatcher, testFile))
}

func testFileWatcherInitialStatus(fileWatcher *FileWatcher, file *os.File) func(*testing.T) {
	return func(t *testing.T) {
		defer file.Close()
		data := []byte("initial state\n")
		err := ioutil.WriteFile(file.Name(), data, 0644)
		if err != nil {
			t.Fatalf("unable to write to %s: %s", file.Name(), err)
		}
		check, _ := fileWatcher.CheckForUpstreamChanges(file.Name())
		if check {
			t.Fatalf("status of check %s should be 'false' but is '%v'", file.Name(), check)
		}
	}
}

func testFileWatcherCheckForChanges(fileWatcher *FileWatcher, file *os.File) func(*testing.T) {
	return func(t *testing.T) {
		defer file.Close()
		data := []byte("data has been changed\n")
		err := ioutil.WriteFile(file.Name(), data, 0644)
		if err != nil {
			t.Fatalf("unable to write to %s: %s", file.Name(), err)
		}
		check, _ := fileWatcher.CheckForUpstreamChanges(file.Name())
		if !check {
			t.Fatalf("status of check %s should be 'true' but is '%v'", file.Name(), check)
		}
	}
}
