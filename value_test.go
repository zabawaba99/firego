package firego

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zabawaba99/firetest"
)

func TestValue(t *testing.T) {
	t.Parallel()
	var (
		response = map[string]interface{}{"foo": "bar"}
		server   = firetest.New()
	)
	server.Start()
	defer server.Close()

	fb := New(server.URL)

	server.Set("", response)

	var v map[string]interface{}
	err := fb.Value(&v)
	assert.NoError(t, err)
	assert.Equal(t, response, v)
}
