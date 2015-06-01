package firego

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const URL = "https://somefirebaseapp.firebaseIO.com"

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
