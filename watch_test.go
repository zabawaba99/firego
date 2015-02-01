package firego

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSSEServer(t *testing.T, event, path, data string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		flusher, ok := w.(http.Flusher)

		if !ok {
			t.Fatal("Streaming unsupported!")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// write SSE goodies
		fmt.Fprintf(w, "event: %s\n", event)
		fmt.Fprintf(w, "data: {\"path\":\"%s\",\"data\":\"%s\"}\n\n", path, data)

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}))
}

func TestWatch(t *testing.T) {
	t.Parallel()
	var (
		eventType, path, data = "put", "foo", "bar"
		server                = newSSEServer(t, eventType, path, data)
		fb                    = New(server.URL)
		notifications         = make(chan Event)
	)
	defer server.Close()

	if err := fb.Watch(notifications); err != nil {
		t.Fatal(err)
	}

	event, ok := <-notifications
	if !ok {
		t.Fatal("notifications closed")
	}
	if event.Type != eventType {
		t.Fatalf("Expected: %s\nActual: %s", eventType, event.Type)
	}

	if event.Path != path {
		t.Fatalf("Expected: %s\nActual: %s", path, event.Path)
	}

	if event.Data.(string) != data {
		t.Fatalf("Expected: %s\nActual: %s", data, event.Data)
	}
}
