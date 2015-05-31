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

func TestShallow(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		shallow bool
	}{
		{
			true,
		},
		{
			false,
		},
	}

	for _, tt := range testCases {
		var (
			server = newTestServer("")
			fb     = New(server.URL)
		)
		defer server.Close()
		fb.Shallow(tt.shallow)
		fb.Value("")
		if expected, actual := 1, len(server.receivedReqs); expected != actual {
			t.Fatalf("Expected: %d\nActual: %d", expected, actual)
		}

		req := server.receivedReqs[0]
		expected, actual := shallowParam+"=true", req.URL.Query().Encode()

		if tt.shallow {
			if expected != actual {
				t.Fatalf("Expected: %s\nActual: %s", expected, actual)
			}
		} else {
			if expected == actual {
				t.Fatalf(`Expected: ""\nActual: %s`, actual)
			}
		}
	}
}

func TestIncludePriority(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		priority bool
	}{
		{
			true,
		},
		{
			false,
		},
	}

	for _, tt := range testCases {
		var (
			server = newTestServer("")
			fb     = New(server.URL)
		)
		defer server.Close()
		fb.IncludePriority(tt.priority)
		fb.Value("")
		if expected, actual := 1, len(server.receivedReqs); expected != actual {
			t.Fatalf("Expected: %d\nActual: %d", expected, actual)
		}

		req := server.receivedReqs[0]
		expected, actual := formatParam+"="+formatVal, req.URL.Query().Encode()

		if tt.priority {
			if expected != actual {
				t.Fatalf("Expected: %s\nActual: %s", expected, actual)
			}
		} else {
			if expected == actual {
				t.Fatalf(`Expected: ""\nActual: %s`, actual)
			}
		}
	}
}
