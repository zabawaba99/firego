package firego

import "testing"

func TestSetAuth(t *testing.T) {
	t.Parallel()
	var (
		token  = "token"
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()
	fb.SetAuth(token)
	fb.Value("")
	if expected, actual := 1, len(server.receivedReqs); expected != actual {
		t.Fatalf("Expected: %d\nActual: %d", expected, actual)
	}

	req := server.receivedReqs[0]
	if expected, actual := "auth="+token, req.URL.Query().Encode(); expected != actual {
		t.Fatalf("Expected: %s\nActual: %s", expected, actual)
	}
}

func TestRemoveAuth(t *testing.T) {
	t.Parallel()
	var (
		token  = "token"
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()
	fb.SetAuth(token)
	fb.RemoveAuth()
	fb.Value("")
	if expected, actual := 1, len(server.receivedReqs); expected != actual {
		t.Fatalf("Expected: %d\nActual: %d", expected, actual)
	}

	req := server.receivedReqs[0]
	if expected, actual := "", req.URL.Query().Encode(); expected != actual {
		t.Fatalf("Expected: %s\nActual: %s", expected, actual)
	}
}
