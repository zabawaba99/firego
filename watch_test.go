package firego

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.True(t, ok, "notifications closed")
	assert.Equal(t, eventType, event.Type, "event type doesn't match")
	assert.Equal(t, path, event.Path, "event path doesn't match")
	assert.Equal(t, data, event.Data.(string), "event data doesn't match")
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
	_, ok := <-notifications
	assert.False(t, ok, "notifications should be closed")
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
	require.True(t, ok, "notifications closed")
	assert.Equal(t, eventType, event.Type, "event type doesn't match")

	_, ok = <-notifications
	require.False(t, ok, "notifications should be closed")
}
