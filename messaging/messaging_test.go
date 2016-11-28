package messaging

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	var serverCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		serverCalled = true
		err := json.NewEncoder(w).Encode(&Response{Failure: 0})
		require.NoError(t, err)
	}))
	defer server.Close()

	serverKey := "hello-world"
	fcm := New(serverKey, nil)
	fcm.apiURL = server.URL

	msg := Message{
		Token:    "foo",
		Priority: "asdas",
	}
	resp, err := fcm.Send(msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, 0, resp.Failure)
	assert.True(t, serverCalled)
}
