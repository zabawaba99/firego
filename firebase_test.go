package firego

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const URL = "https://somefirebaseapp.firebaseIO.com"

func init() {
	TimeoutDuration = 10 * time.Millisecond
}

type TestServer struct {
	*httptest.Server
	receivedReqs []*http.Request
}

func newTestServer(response string) *TestServer {
	ts := &TestServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ts.receivedReqs = append(ts.receivedReqs, req)
		fmt.Fprint(w, response)
	}))
	return ts
}

func TestNew(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		givenURL string
	}{
		{
			URL,
		},
		{
			URL + "/",
		},
		{
			"somefirebaseapp.firebaseIO.com",
		},
		{
			"somefirebaseapp.firebaseIO.com/",
		},
	}

	for _, tt := range testCases {
		fb := New(tt.givenURL)
		assert.Equal(t, URL, fb.url, "givenURL: %s", tt.givenURL)
	}
}

func TestChild(t *testing.T) {
	t.Parallel()
	var (
		parent    = New(URL)
		childNode = "node"
		child     = parent.Child(childNode)
	)

	assert.Equal(t, fmt.Sprintf("%s/%s", parent.url, childNode), child.url)
}

func TestTimeoutDuration_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * TimeoutDuration)
	}))
	defer server.Close()

	fb := New(server.URL)

	err := fb.Value("")
	assert.NotNil(t, err)
	assert.IsType(t, ErrTimeout{}, err, "%s", err)
	assert.Equal(t, TimeoutDuration, fb.Client.Timeout)
}

func TestTimeoutDuration_Dial(t *testing.T) {
	fb := New("http://dialtimeouterr.or/")

	quickDialer := (&net.Dialer{Timeout: time.Nanosecond}).Dial
	fb.Client.Transport.(*http.Transport).Dial = quickDialer

	err := fb.Value("")
	assert.NotNil(t, err)
	assert.IsType(t, ErrTimeout{}, err, "%s", err)
	assert.Equal(t, TimeoutDuration, fb.Client.Timeout)
}

func TestShallow(t *testing.T) {
	t.Parallel()
	var (
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()

	fb.Shallow(true)
	fb.Value("")
	require.Len(t, server.receivedReqs, 1)

	req := server.receivedReqs[0]
	assert.Equal(t, shallowParam+"=true", req.URL.Query().Encode())

	fb.Shallow(false)
	fb.Value("")
	require.Len(t, server.receivedReqs, 2)

	req = server.receivedReqs[1]
	assert.Equal(t, "", req.URL.Query().Encode())
}

func TestIncludePriority(t *testing.T) {
	t.Parallel()
	var (
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()

	fb.IncludePriority(true)
	fb.Value("")
	require.Len(t, server.receivedReqs, 1)

	req := server.receivedReqs[0]
	assert.Equal(t, formatParam+"="+formatVal, req.URL.Query().Encode())

	fb.IncludePriority(false)
	fb.Value("")
	require.Len(t, server.receivedReqs, 2)

	req = server.receivedReqs[1]
	assert.Equal(t, "", req.URL.Query().Encode())
}
