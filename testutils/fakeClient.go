package testutils

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

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
	}
	return "", ErrNotFound
}
