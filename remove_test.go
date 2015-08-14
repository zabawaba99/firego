package firego

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zabawaba99/firetest"
)

func TestRemove(t *testing.T) {
	t.Parallel()
	server := firetest.New()
	server.Start()
	defer server.Close()

	server.Set("", true)

	fb := New(server.URL)
	err := fb.Remove()
	assert.NoError(t, err)

	v := server.Get("")
	assert.Nil(t, v)
}
