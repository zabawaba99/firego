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

	// The article below explains how Amazon's ELB's to not always send the "close_notify packet"
	// when using TLS.  By default the golang http client attempts to reuse connectons.  This
	// caused an issue because, the server closed the connection, we never received the close_notify
	// thus, golang http client tried to reuse a closed connection.  This issue does not happen all
	// the time, but it did happen once in every 5 requests that are fired off one right after the
	// other.  Note, the article below is for AWS, but seems to have the same issue with Firebase.
	// The fix for me, is to disable the KeepAlives with really causes golang not to reause the
	// connection.
	// https://code.google.com/p/go/issues/detail?id=3514
	tr := &http.Transport{DisableKeepAlives: true}

	return &Firebase{
		url:          sanitizeURL(url),
		client:       &http.Client{Transport: tr},
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
