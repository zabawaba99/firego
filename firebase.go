package firego

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// Firebase represents a location in the cloud
type Firebase struct {
	url    string
	client *http.Client
}

// New creates a new Firebase reference
func New(url string) *Firebase {
	return &Firebase{
		url:    url,
		client: &http.Client{},
	}
}

// Child creates a new Firebase reference for the requested
// child string
func (fb *Firebase) Child(child string) *Firebase {
	return &Firebase{
		url:    fb.url + "/" + child,
		client: fb.client,
	}
}

func (fb *Firebase) doRequest(method string, body []byte) ([]byte, error) {
	url := fb.url + "/.json"
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := fb.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
