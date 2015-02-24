package firego

import "testing"

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
		if fb.url != URL {
			t.Fatalf("url not set correctly. Expected: %s\nActual: %s", URL, fb.url)
		}
	}
}

func TestChild(t *testing.T) {
	t.Parallel()
	var (
		parent    = New(URL)
		childNode = "node"
		child     = parent.Child(childNode)
	)

	if expected := parent.url + "/" + childNode; child.url != expected {
		t.Fatalf("url not set corrected. Expected: %s\nActual: %s", expected, child.url)
	}
}
