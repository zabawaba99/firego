/*
Package database is a REST client for Firebase Realtime Database (https://firebase.google.com/docs/database/).
*/
package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	_url "net/url"
	"strings"
	"sync"
	"time"
)

const defaultRedirectLimit = 30

// query parameter constants
const (
	authParam         = "auth"
	shallowParam      = "shallow"
	formatParam       = "format"
	formatVal         = "export"
	orderByParam      = "orderBy"
	limitToFirstParam = "limitToFirst"
	limitToLastParam  = "limitToLast"
	startAtParam      = "startAt"
	endAtParam        = "endAt"
	equalToParam      = "equalTo"
)

const defaultHeartbeat = 2 * time.Minute

// Database represents a location in the cloud.
type Database struct {
	url    string
	params _url.Values
	client *http.Client

	eventMtx   sync.Mutex
	eventFuncs map[string]chan struct{}

	watchMtx       sync.Mutex
	watching       bool
	watchHeartbeat time.Duration
	stopWatching   chan struct{}
}

// New creates a new Firebase reference,
// if client is nil, http.DefaultClient is used.
func New(url string, client *http.Client) *Database {
	db := &Database{
		url:            sanitizeURL(url),
		params:         _url.Values{},
		stopWatching:   make(chan struct{}),
		watchHeartbeat: defaultHeartbeat,
		eventFuncs:     map[string]chan struct{}{},
	}
	if client == nil {
		client = &http.Client{
			Transport:     http.DefaultClient.Transport,
			CheckRedirect: redirectPreserveHeaders,
		}
	}

	db.client = client
	return db
}

// Auth sets the custom Firebase token used to authenticate to Firebase.
func (db *Database) Auth(token string) {
	db.params.Set(authParam, token)
}

// Unauth removes the current token being used to authenticate to Firebase.
func (db *Database) Unauth() {
	db.params.Del(authParam)
}

// Push creates a reference to an auto-generated child location.
func (db *Database) Push(v interface{}) (*Database, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	bytes, err = db.doRequest("POST", bytes)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}
	newRef := db.copy()
	newRef.url = db.url + "/" + m["name"]
	return newRef, err
}

// Remove the Firebase reference from the cloud.
func (db *Database) Remove() error {
	_, err := db.doRequest("DELETE", nil)
	if err != nil {
		return err
	}
	return nil
}

// Set the value of the Firebase reference.
func (db *Database) Set(v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = db.doRequest("PUT", bytes)
	return err
}

// Update the specific child with the given value.
func (db *Database) Update(v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = db.doRequest("PATCH", bytes)
	return err
}

// Value gets the value of the Firebase reference.
func (db *Database) Value(v interface{}) error {
	bytes, err := db.doRequest("GET", nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

// String returns the string representation of the
// Firebase reference.
func (db *Database) String() string {
	path := db.url + "/.json"

	if len(db.params) > 0 {
		path += "?" + db.params.Encode()
	}
	return path
}

// Child creates a new Firebase reference for the requested
// child with the same configuration as the parent.
func (db *Database) Child(child string) *Database {
	c := db.copy()
	c.url = c.url + "/" + child
	return c
}

func (db *Database) copy() *Database {
	c := &Database{
		url:            db.url,
		params:         _url.Values{},
		client:         db.client,
		stopWatching:   make(chan struct{}),
		watchHeartbeat: defaultHeartbeat,
		eventFuncs:     map[string]chan struct{}{},
	}

	// making sure to manually copy the map items into a new
	// map to avoid modifying the map reference.
	for k, v := range db.params {
		c.params[k] = v
	}
	return c
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

// Preserve headers on redirect.
//
// Reference https://github.com/golang/go/issues/4800
func redirectPreserveHeaders(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		// No redirects
		return nil
	}

	if len(via) > defaultRedirectLimit {
		return fmt.Errorf("%d consecutive requests(redirects)", len(via))
	}

	// mutate the subsequent redirect requests with the first Header
	for key, val := range via[0].Header {
		req.Header[key] = val
	}
	return nil
}

func (db *Database) doRequest(method string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, db.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := db.client.Do(req)
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
