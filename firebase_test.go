package firego

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const URL = "https://somefirebaseapp.firebaseIO.com"

type TestServer struct {
	*httptest.Server
	receivedReqs []*http.Request
}

func newTestServer(response string) *TestServer {
	ts := &TestServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Upgrade") == "websocket" {
			log.Printf("ignoring requests that are not http\n")
			return
		}
		ts.receivedReqs = append(ts.receivedReqs, req)
		fmt.Fprint(w, response)
	}))
	return ts
}

func TestNew(t *testing.T) {
	t.Parallel()
	testURLs := []string{
		URL,
		URL + "/",
		"somefirebaseapp.firebaseIO.com",
		"somefirebaseapp.firebaseIO.com/",
	}

	for _, url := range testURLs {
		fb := New(url, nil)
		assert.Equal(t, URL, fb.repo.Host(), "givenURL: %s", url)
	}
}

func TestNewWithProvidedHttpClient(t *testing.T) {
	t.Parallel()

	var client = http.DefaultClient
	testURLs := []string{
		URL,
		URL + "/",
		"somefirebaseapp.firebaseIO.com",
		"somefirebaseapp.firebaseIO.com/",
	}

	for _, url := range testURLs {
		fb := New(url, client)
		assert.Equal(t, URL, fb.repo.Host(), "givenURL: %s", url)
		assert.Equal(t, client, fb.client)
	}
}

func TestChild(t *testing.T) {
	t.Parallel()
	var (
		parent    = New(URL, nil)
		childNode = "node"
		child     = parent.Child(childNode)
	)

	assert.Equal(t, "/"+childNode, child.path)
}

func TestChild_Issue26(t *testing.T) {
	t.Parallel()
	parent := New(URL, nil)
	child1 := parent.Child("one")
	child2 := child1.Child("two")

	child1.Shallow(true)
	assert.Len(t, child2.params, 0)
}

func TestTimeoutDuration_Headers(t *testing.T) {
	defer func(dur time.Duration) { TimeoutDuration = dur }(TimeoutDuration)
	TimeoutDuration = time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(2 * TimeoutDuration)
	}))
	defer server.Close()

	fb := New(server.URL, nil)
	err := fb.Value("")
	assert.NotNil(t, err)
	assert.IsType(t, ErrTimeout{}, err)

	// ResponseHeaderTimeout should be TimeoutDuration less the time it took to dial, and should be positive
	require.IsType(t, (*http.Transport)(nil), fb.client.Transport)
	tr := fb.client.Transport.(*http.Transport)
	assert.True(t, tr.ResponseHeaderTimeout < TimeoutDuration)
	assert.True(t, tr.ResponseHeaderTimeout > 0)
}

func TestTimeoutDuration_Dial(t *testing.T) {
	defer func(dur time.Duration) { TimeoutDuration = dur }(TimeoutDuration)
	TimeoutDuration = time.Microsecond

	fb := New("http://dialtimeouterr.or/", nil)
	err := fb.Value("")
	assert.NotNil(t, err)
	assert.IsType(t, ErrTimeout{}, err)

	// ResponseHeaderTimeout should be negative since the total duration was consumed when dialing
	require.IsType(t, (*http.Transport)(nil), fb.client.Transport)
	assert.True(t, fb.client.Transport.(*http.Transport).ResponseHeaderTimeout < 0)
}
