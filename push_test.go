package firego

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zabawaba99/firetest"
)

func TestPush(t *testing.T) {
	t.Parallel()
	var (
		payload = map[string]interface{}{"foo": "bar"}
		server  = firetest.New()
	)
	server.Start()
	defer server.Close()

	fb := New(server.URL)
	childRef, err := fb.Push(payload)
	assert.NoError(t, err)

	path := strings.TrimPrefix(childRef.String(), server.URL+"/")
	v := server.Get(path)
	assert.Equal(t, payload, v)
}
