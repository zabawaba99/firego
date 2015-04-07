package firego

import "testing"

func TestRemove(t *testing.T) {
	t.Parallel()
	var (
		server = newTestServer("")
		fb     = New(server.URL)
	)
	defer server.Close()

	fb.Remove(nil)
	if expected, actual := 1, len(server.receivedReqs); expected != actual {
		t.Fatalf("Expected: %d\nActual: %d", expected, actual)
	}

	req := server.receivedReqs[0]
	if expected, actual := "DELETE", req.Method; expected != actual {
		t.Fatalf("Expected: %s\nActual: %s", expected, actual)
	}
}
