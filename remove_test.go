package firego

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemove(t *testing.T) {
	t.Parallel()
	var (
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()

	fb.Remove()
	require.Len(t, server.receivedReqs, 1)

	req := server.receivedReqs[0]
	assert.Equal(t, "DELETE", req.Method)
}
