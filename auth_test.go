package firego

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authToken = "token"

func TestAuth(t *testing.T) {
	t.Parallel()
	var (
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()

	fb.Auth(authToken)
	fb.Value("")
	require.Len(t, server.receivedReqs, 1)

	req := server.receivedReqs[0]
	assert.Equal(t, "auth="+authToken, req.URL.Query().Encode())
}

func TestUnauth(t *testing.T) {
	t.Parallel()
	var (
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()

	fb.Auth(authToken)
	fb.Unauth()
	fb.Value("")
	require.Len(t, server.receivedReqs, 1)

	req := server.receivedReqs[0]
	assert.Equal(t, "", req.URL.Query().Encode())
}
