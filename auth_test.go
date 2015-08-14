package firego

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zabawaba99/firetest"
)

const authToken = "token"

func TestAuth(t *testing.T) {
	t.Parallel()
	server := firetest.New()
	server.Start()
	defer server.Close()

	server.RequireAuth(true)
	fb := New(server.URL)

	fb.Auth(server.Secret)
	var v interface{}
	err := fb.Value(&v)
	assert.NoError(t, err)
}

func TestUnauth(t *testing.T) {
	t.Parallel()
	server := firetest.New()
	server.Start()
	defer server.Close()

	server.RequireAuth(true)
	fb := New(server.URL)

	fb.params.Add("auth", server.Secret)
	fb.Unauth()
	err := fb.Value("")
	assert.Error(t, err)
}
