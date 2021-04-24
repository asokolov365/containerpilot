package surveillee

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/hashicorp/consul/sdk/testutil/retry"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
)

// TestServer represents a Vault server we can run our tests against. Depends
// on a local `vault` binary installed into our environ's PATH.
type TestServer struct {
	cmd        *exec.Cmd
	HTTPAddr   string
	HTTPClient *http.Client
}

// NewTestServer constructs a new TestServer by including the httpPort as well.
func NewTestServer(httpPort int) (*TestServer, error) {
	path, err := exec.LookPath("vault")
	if err != nil || path == "" {
		return nil, fmt.Errorf("vault not found on $PATH - download and install " +
			"vault or skip this test")
	}

	args := []string{
		"server",
		"-dev",
		"-dev-listen-address=127.0.0.1:" + strconv.Itoa(httpPort),
		"-dev-no-store-token",
		"-dev-root-token-id=myTestToken",
	}
	cmd := exec.Command("vault", args...)
	cmd.Stdout = io.Writer(os.Stdout)
	cmd.Stderr = io.Writer(os.Stderr)
	if err := cmd.Start(); err != nil {
		return nil, errors.New("failed starting command")
	}

	httpAddr := fmt.Sprintf("http://127.0.0.1:%d", httpPort)

	httpClient := cleanhttp.DefaultClient()

	return &TestServer{cmd, httpAddr, httpClient}, nil
}

// Stop stops a TestServer
func (s *TestServer) Stop() error {
	if s.cmd == nil {
		return nil
	}

	if s.cmd.Process != nil {
		if err := s.cmd.Process.Signal(os.Interrupt); err != nil {
			return errors.New("failed to kill vault server")
		}
	}

	return s.cmd.Wait()
}

// failer implements the retry.Failer interface
type failer struct {
	failed bool
}

func (f *failer) Helper()                 { fmt.Println("TODO: Helper() is not implemented") }
func (f *failer) Log(args ...interface{}) { fmt.Println(args...) }
func (f *failer) FailNow()                { f.failed = true }

// WaitForAPI waits for only the agent HTTP endpoint to start responding. This
// is an indication that the agent has started, but will likely return before a
// leader is elected.
func (s *TestServer) WaitForAPI() error {
	f := &failer{}
	retry.Run(f, func(r *retry.R) {
		resp, err := s.HTTPClient.Get(s.HTTPAddr + "/v1/sys/leader")
		if err != nil {
			r.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			r.Fatalf("bad status code %d", resp.StatusCode)
		}
	})
	if f.failed {
		return fmt.Errorf("api is unavailable")
	}
	return nil
}
