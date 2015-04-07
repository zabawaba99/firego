package firego

import "testing"

func TestAuth(t *testing.T) {
	t.Parallel()
	var (
		token  = "token"
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()
	fb.Auth(token)
	fb.Value("", nil)
	if expected, actual := 1, len(server.receivedReqs); expected != actual {
		t.Fatalf("Expected: %d\nActual: %d", expected, actual)
	}

	req := server.receivedReqs[0]
	if expected, actual := "auth="+token, req.URL.Query().Encode(); expected != actual {
		t.Fatalf("Expected: %s\nActual: %s", expected, actual)
	}
}

func TestUnauth(t *testing.T) {
	t.Parallel()
	var (
		token  = "token"
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()
	fb.Auth(token)
	fb.Unauth()
	fb.Value("", nil)
	if expected, actual := 1, len(server.receivedReqs); expected != actual {
		t.Fatalf("Expected: %d\nActual: %d", expected, actual)
	}

	req := server.receivedReqs[0]
	if expected, actual := "", req.URL.Query().Encode(); expected != actual {
		t.Fatalf("Expected: %s\nActual: %s", expected, actual)
	}
}
