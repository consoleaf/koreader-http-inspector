/*
Package koreaderinspector
*/
package koreaderinspector

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

type HTTPInspectorClient struct {
	client  HTTPClient
	baseURL string
	Logger  log.Logger
}

type HTTPInspectorClientError struct {
	message string
}

func (error HTTPInspectorClientError) Error() string {
	return error.message
}

func (client *HTTPInspectorClient) SSHStart() (int, error) {
	_, err := client.Get("/ui/SSH/start/")
	if err != nil {
		return 0, err
	}
	isRunning, err := client.SSHIsRunning()
	if err != nil {
		return 0, err
	}
	if !isRunning {
		return 0, HTTPInspectorClientError{message: "SSH is not running after calling /ui/ssh/start/"}
	}
	return client.SSHGetPort()
}

func (client *HTTPInspectorClient) SSHGetPort() (int, error) {
	res, err := client.Get("ui/SSH/SSH_port")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(res))
}

func (client *HTTPInspectorClient) SSHGetAllowNoPassword() (bool, error) {
	res, err := client.Get("ui/SSH/allow_no_password")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(string(res))
}

func (client *HTTPInspectorClient) SSHSetAllowNoPassword(allow bool) error {
	_, err := client.Get(fmt.Sprintf("ui/SSH/allow_no_password?=%v", allow))
	return err
}

func (client *HTTPInspectorClient) SSHStop() error {
	_, err := client.Get("/ui/SSH/stop/")
	if err != nil {
		return err
	}
	isRunning, err := client.SSHIsRunning()
	if err != nil {
		return err
	}
	if isRunning {
		return HTTPInspectorClientError{message: "SSH is running after calling /ui/ssh/stop/"}
	}
	return nil
}

func (client *HTTPInspectorClient) SSHIsRunning() (bool, error) {
	res, err := client.Get("/ui/SSH/isRunning/")
	var body []bool
	if json.Unmarshal(res, &body) != nil {
		return false, err
	}
	return body[0], nil
}

func (client *HTTPInspectorClient) ToggleNightMode() error {
	_, err := client.Get("/event/ToggleNightMode")
	return err
}

func (client *HTTPInspectorClient) SetNightMode(nightMode bool) error {
	_, err := client.Get(
		fmt.Sprintf("/event/SetNightMode/%v", nightMode),
	)
	return err
}

func (client *HTTPInspectorClient) FullRefresh() error {
	_, err := client.Get("/event/FullRefresh")
	return err
}

func (client *HTTPInspectorClient) TurnPageBy(amount int) error {
	_, err := client.Get(
		fmt.Sprintf("/event/GotoViewRel/%d", amount),
	)
	return err
}

func (client *HTTPInspectorClient) NextPage() error {
	return client.TurnPageBy(1)
}

func (client *HTTPInspectorClient) PrevPage() error {
	return client.TurnPageBy(-1)
}

func New(baseURL string) (*HTTPInspectorClient, error) {
	return NewWithClient(baseURL, http.DefaultClient)
}

func NewWithClient(baseURL string, client HTTPClient) (*HTTPInspectorClient, error) {
	uri, err := url.JoinPath(baseURL, "/koreader/")
	return &HTTPInspectorClient{
		client:  client,
		baseURL: uri,
		Logger:  *log.Default(),
	}, err
}

func (client *HTTPInspectorClient) GetWithQuery(path string, query string) ([]byte, error) {
	url, err := url.JoinPath(path, "?", query)
	if err != nil {
		return nil, err
	}
	return client.Get(url)
}

func (client *HTTPInspectorClient) Get(path string) ([]byte, error) {
	client.Logger.Printf("[HTTPInspectorClient] GET %v", path)
	url, err := url.JoinPath(client.baseURL, path)
	if err != nil {
		client.Logger.Printf("[HTTPInspectorClient] ERROR on GET %v: %v", path, err)
		return nil, err
	}
	res, err := client.client.Get(url)
	if err != nil {
		client.Logger.Printf("[HTTPInspectorClient] ERROR on GET %v: %v", path, err)
		return []byte{}, err
	}
	defer res.Body.Close()
	buf, err := io.ReadAll(res.Body)
	client.Logger.Printf("[HTTPInspectorClient] RESPONSE on GET %v: %v", path, string(buf))
	return buf, err
}

func (client *HTTPInspectorClient) GetLuaVersion() (string, error) {
	version, err := client.Get("/globals/_VERSION")
	return string(version), err
}

func (client *HTTPInspectorClient) RestartKOReader() error {
	_, err := client.Get("/UIManager/restartKOReader/")
	return err
}
