package firego

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestValue(t *testing.T) {
	t.Parallel()
	var (
		response = `{"foo":"bar"}`
		server   = newTestServer(response)
		fb       = New(server.URL)
	)
	defer server.Close()

	var v map[string]interface{}
	fb.Value(&v)
	val, ok := v["foo"]
	assert.True(t, ok)
	assert.Equal(t, "bar", val)

	require.Len(t, server.receivedReqs, 1)

	req := server.receivedReqs[0]
	assert.Equal(t, "GET", req.Method)
}
