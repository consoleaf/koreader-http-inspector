package koreaderinspector_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	koreaderinspector "github.com/Consoleaf/koreader-http-inspector"
	"github.com/Consoleaf/koreader-http-inspector/testutils"
	"github.com/Consoleaf/koreader-http-inspector/utils"
)

func Test_GetVersion(t *testing.T) {
	WithDifferentClients(t, func(t *testing.T, httpInspector *koreaderinspector.HTTPInspectorClient) {
		resp, err := httpInspector.GetLuaVersion()
		if err != nil {
			t.Fatal(err)
		}
		body := string(resp)
		expected := "Lua 5.1"
		if body != expected {
			t.Errorf("Expected response %q, got %q", expected, body)
		}
	})
}

func Test_Restart(t *testing.T) {
	fakeClient := NewFakeHTTPClient(t)
	defer fakeClient.server.Close()

	inspector, _ := koreaderinspector.NewWithClient("", fakeClient)
	inspector.Logger = *testutils.MakeTestLogger(t)

	inspector.RestartKOReader()

	if !slices.ContainsFunc(fakeClient.requests, func(req http.Request) bool {
		return strings.Contains(req.URL.Path, "koreader/UIManager/restartKOReader/")
	}) {
		t.Errorf("Expected to have received a restartKOReader request, but didn't. Got: %q", utils.SliceMap(fakeClient.requests, func(a http.Request) string {
			return a.URL.Path
		}))
	}
}

func Test_FullRefresh(t *testing.T) {
	WithDifferentClients(t, func(t *testing.T, httpInspector *koreaderinspector.HTTPInspectorClient) {
		err := httpInspector.FullRefresh()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func Test_SSH(t *testing.T) {
	WithDifferentClients(t, func(t *testing.T, httpInspector *koreaderinspector.HTTPInspectorClient) {
		defer httpInspector.SSHStop()

		sshRunning, err := httpInspector.SSHIsRunning()
		if err != nil {
			t.Fatal(err)
		}

		if sshRunning {
			t.Fatalf("SSH is already running at the start of the test! Got: %v", sshRunning)
		}

		port, err := httpInspector.SSHStart()
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("SSH started successfully, with port %d", port)
	})
}

// HELPERS

var ErrNotFound = errors.New("not found")

type FakeHTTPClient struct {
	server     *httptest.Server
	requests   []http.Request
	sshRunning bool
}

func NewFakeHTTPClient(t *testing.T) *FakeHTTPClient {
	fakeClient := &FakeHTTPClient{sshRunning: false}
	fakeClient.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("[FakeHttpClient.server] Got request for %q\n", r.URL)
		fakeClient.requests = append(fakeClient.requests, *r)
		res, err := fakeClient.handleRequest(r)
		if err != nil {
			if err == ErrNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write([]byte(res))
		t.Logf("[FakeHttpClient.server] Handler finished for %q\n", r.URL) // ADD THIS LINE
	}))
	return fakeClient
}

// checkServerReachable helper function (as defined above)
func checkServerReachable(baseURL string, t *testing.T) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodHead, baseURL, nil)
	if err != nil {
		t.Logf("Error creating check request: %v\n", err)
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Server at %q is not reachable: %v\n", baseURL, err)
		return false
	}
	defer resp.Body.Close()
	return true
}

func WithDifferentClients(t *testing.T, testfn func(*testing.T, *koreaderinspector.HTTPInspectorClient)) {
	t.Helper()
	t.Run("Fake client", func(t *testing.T) {
		client := NewFakeHTTPClient(t)
		defer client.server.Close()

		httpInspector, _ := koreaderinspector.NewWithClient(
			"",
			client,
		)
		httpInspector.Logger = *testutils.MakeTestLogger(t)
		testfn(t, httpInspector)
	})
	t.Run("Real client", func(t *testing.T) {
		baseURL := "http://192.168.15.244:8080/"
		httpInspector, _ := koreaderinspector.New(baseURL)
		httpInspector.Logger = *testutils.MakeTestLogger(t)

		if !checkServerReachable(baseURL, t) {
			t.Skipf("Skipping 'Real client' tests: Server at %q is not reachable or timed out.", baseURL)
			return // Important to return after skipping
		}

		testfn(t, httpInspector)
	})
}

func (fakeClient *FakeHTTPClient) Get(path string) (*http.Response, error) {
	url, err := url.JoinPath(fakeClient.server.URL, path)
	if err != nil {
		return nil, err
	}
	return http.Get(url)
}

func (fakeClient *FakeHTTPClient) handleRequest(r *http.Request) (string, error) {
	switch r.URL.Path {
	case "/koreader/globals/_VERSION":
		return "Lua 5.1", nil
	case "/koreader/ui/SSH/isRunning/":
		return fmt.Sprintf("[%v]", fakeClient.sshRunning), nil
	case "/koreader/ui/SSH/start/":
		fakeClient.sshRunning = true
		return "", nil
	case "/koreader/ui/SSH/SSH_port":
		return "2222", nil
	case "/koreader/ui/SSH/stop/":
		fakeClient.sshRunning = false
		return "", nil
	case "/koreader/event/ToggleNightMode":
	case "/koreader/event/FullRefresh":
		return "", nil
	}
	return "", ErrNotFound
}
