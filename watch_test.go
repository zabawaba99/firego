package firego

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testToken = "test_token"

func newSSEServer(t *testing.T, event, path, data string, stop chan struct{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		flusher, ok := w.(http.Flusher)

		if !ok {
			t.Fatal("Streaming unsupported!")
		}
		req.ParseForm()

		if req.Form.Get("auth") != testToken {
			http.Error(w, "Permission denied", http.StatusUnauthorized)
			flusher.Flush()
			<-stop
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
		<-stop
	}))
}

func TestWatch(t *testing.T) {
	t.Parallel()
	var (
		eventType, path, data = "put", "foo", "bar"
		notifications, stop   = make(chan Event), make(chan struct{})
		server                = newSSEServer(t, eventType, path, data, stop)
		fb                    = New(server.URL)
	)
	defer func() {
		close(stop)
		server.Close()
	}()

	fb.Auth(testToken)
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

func TestStopWatch(t *testing.T) {
	t.Parallel()
	var (
		eventType, path, data = "put", "foo", "bar"
		moveOn, stop          = make(chan struct{}), make(chan struct{})
		notifications         = make(chan Event)
		server                = newSSEServer(t, eventType, path, data, stop)
		fb                    = New(server.URL)
	)
	defer func() {
		close(stop)
		server.Close()
	}()

	fb.Auth(testToken)
	go func() {
		if err := fb.Watch(notifications); err != nil {
			t.Fatal(err)
		}
		<-moveOn
		fb.StopWatching()
	}()

	<-notifications
	moveOn <- struct{}{}
	if _, ok := <-notifications; ok {
		t.Fatal("notifications should be closed")
	}
}

func TestWatch_Cancel(t *testing.T) {
	var (
		eventType, path, data = "cancel", "", ""
		notifications, stop   = make(chan Event), make(chan struct{})
		server                = newSSEServer(t, eventType, path, data, stop)
		fb                    = New(server.URL)
	)

	defer func() {
		close(stop)
		server.Close()
	}()

	fb.Auth(testToken)
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

	if _, ok := <-notifications; ok {
		t.Fatal("notifications should be closed")
	}
}
