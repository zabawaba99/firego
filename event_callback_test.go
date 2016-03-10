package firego

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zabawaba99/firetest"
)

type testEvents struct {
	snapshot    DataSnapshot
	previousKey string
}

func TestChildAdded(t *testing.T) {
	server := firetest.New()
	server.Start()
	defer server.Close()

	fb := New(server.URL, nil)

	// set some existing values that should come down
	server.Set("something", true)
	server.Set("AAA", "foo")

	var results []testEvents
	fn := func(snapshot DataSnapshot, previousChildKey string) {
		results = append(results, testEvents{snapshot, previousChildKey})
	}
	err := fb.ChildAdded(fn)
	require.NoError(t, err)

	// should get regular addition events
	err = fb.Child("foo").Set(2)
	require.NoError(t, err)

	err = fb.Child("bar").Set(map[string]string{"hi": "mom"})
	require.NoError(t, err)

	fbChild, err := fb.Push("gaga oh la la")
	require.NoError(t, err)

	// should not get updates
	err = fb.Child("foo").Set(false)
	require.NoError(t, err)

	// or deletes
	err = fb.Child("bar").Remove()
	require.NoError(t, err)

	// should get a notification after adding a deleted field
	err = fb.Child("bar").Set("something-else")
	require.NoError(t, err)

	// should not get notifications for addition to a child not
	err = fb.Child("bar/child").Set(true)
	require.NoError(t, err)

	// wait for all notifications to come down
	time.Sleep(time.Millisecond)

	expected := []testEvents{
		{"foo", ""},
		{true, "AAA"},
		{float64(2), "something"},
		{map[string]interface{}{"hi": "mom"}, "foo"},
		{"gaga oh la la", "bar"},
		{"something-else", strings.TrimPrefix(fbChild.String(), fb.String()+"/")},
	}

	assert.Len(t, results, len(expected))
	for i, v := range expected {
		assert.EqualValues(t, v, results[i], "Did not receive '%#v' event", v)
	}
}

func TestRemoveEventFunc(t *testing.T) {
	server := firetest.New()
	server.Start()
	defer server.Close()

	fb := New(server.URL, nil)

	fn := func(snapshot DataSnapshot, previousChildKey string) {
		assert.Fail(t, "Should not have received anything")
	}
	err := fb.ChildAdded(fn)
	require.NoError(t, err)

	fb.RemoveEventFunc(fn)

	fb.Child("hello").Set(false)
	time.Sleep(time.Millisecond)

	assert.Len(t, fb.eventFuncs, 0)
}
