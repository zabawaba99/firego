package firego

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Firebase represents a location in the cloud
type Firebase struct {
	url          string
	auth         string
	client       *http.Client
	watching     bool
	stopWatching chan struct{}
}

func sanitizeURL(url string) string {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		url = "https://" + url
	}

	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	return url
}

// New creates a new Firebase reference
func New(url string) *Firebase {
	return &Firebase{
		url:          sanitizeURL(url),
		client:       &http.Client{},
		stopWatching: make(chan struct{}),
	}
}

// String returns the string representation of the
// Firebase reference
func (fb *Firebase) String() string {
	return fb.url
}

// Child creates a new Firebase reference for the requested
// child string
func (fb *Firebase) Child(child string) *Firebase {
	return &Firebase{
		url:          fb.url + "/" + child,
		auth:         fb.auth,
		client:       fb.client,
		stopWatching: make(chan struct{}),
	}
}

func (fb *Firebase) doRequest(method string, body []byte) ([]byte, error) {
	path := fb.url + "/.json"

	v := url.Values{}
	if fb.auth != "" {
		v.Add("auth", fb.auth)
	}

	if len(v) > 0 {
		path += "?" + v.Encode()
	}
	req, err := http.NewRequest(method, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := fb.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/200 != 1 {
		return nil, errors.New(string(respBody))
	}
	return respBody, nil
}
