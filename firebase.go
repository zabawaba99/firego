package firego

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	_url "net/url"
	"strings"
	"time"
)

var TimeoutDuration = 30 * time.Second

type ErrTimeout struct {
	error
}

// query parameter constants
const (
	authParam    = "auth"
	formatParam  = "format"
	shallowParam = "shallow"
	formatVal    = "export"
)

// Firebase represents a location in the cloud
type Firebase struct {
	url          string
	params       _url.Values
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

	var tr *http.Transport
	tr = &http.Transport{
		DisableKeepAlives: true, // https://code.google.com/p/go/issues/detail?id=3514
		Dial: func(network, address string) (net.Conn, error) {
			start := time.Now()
			c, err := net.DialTimeout(network, address, TimeoutDuration)
			tr.ResponseHeaderTimeout = TimeoutDuration - time.Since(start)
			return c, err
		},
	}

	return &Firebase{
		url:          sanitizeURL(url),
		params:       _url.Values{},
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
		params:       fb.params,
		client:       fb.client,
		stopWatching: make(chan struct{}),
	}
}

// Shallow limits the depth of the data returned when calling Value.
// If the data at the location is a JSON primitive (string, number or boolean),
// its value will be returned. If the data is a JSON object, the values
// for each key will be truncated to true.
//
// Reference https://www.firebase.com/docs/rest/api/#section-param-shallow
func (fb *Firebase) Shallow(v bool) {
	if v {
		fb.params.Set(shallowParam, "true")
	} else {
		fb.params.Del(shallowParam)
	}
}

// IncludePriority determines whether or not to ask Firebase
// for the values priority. By default, the priority is not returned
//
//		# Include Priority
//		ref.IncludePriority(true)
//		# Exclude Priority
//		ref.IncludePriority(false)
func (fb *Firebase) IncludePriority(v bool) {
	if v {
		fb.params.Set(formatParam, formatVal)
	} else {
		fb.params.Del(formatParam)
	}
}

func (fb *Firebase) makeRequest(method string, body []byte) (*http.Request, error) {
	path := fb.url + "/.json"

	if len(fb.params) > 0 {
		path += "?" + fb.params.Encode()
	}
	return http.NewRequest(method, path, bytes.NewReader(body))
}

func (fb *Firebase) doRequest(method string, body []byte) ([]byte, error) {
	req, err := fb.makeRequest(method, body)
	if err != nil {
		return nil, err
	}

	resp, err := fb.client.Do(req)
	switch err := err.(type) {
	default:
		return nil, err
	case nil:
		// carry on
	case *_url.Error:
		e1, ok := err.Err.(net.Error)
		if !ok {
			return nil, err
		}
		if e1.Timeout() {
			return nil, ErrTimeout{err}
		}
	case net.Error:
		if err.Timeout() {
			return nil, ErrTimeout{err}
		}
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
