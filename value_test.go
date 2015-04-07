package firego

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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
	fb.Value(&v, nil)

	if val, ok := v["foo"]; !ok || val != "bar" {
		t.Fatalf("Did not get valid response. Expected: %s\nActual: %s", response, v)
	}

	if expected, actual := 1, len(server.receivedReqs); expected != actual {
		t.Fatalf("Expected: %d\nActual: %d", expected, actual)
	}

	req := server.receivedReqs[0]
	if expected, actual := "GET", req.Method; expected != actual {
		t.Fatalf("Expected: %s\nActual: %s", expected, actual)
	}
}
