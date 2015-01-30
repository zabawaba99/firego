package firego

import "testing"

func TestSet(t *testing.T) {
	t.Parallel()
	var (
		response = `{"foo":"bar"}`
		server   = newTestServer(response)
		fb       = New(server.URL)
	)
	defer server.Close()

	fb.Set(response)
	if expected, actual := 1, len(server.receivedReqs); expected != actual {
		t.Fatalf("Expected: %d\nActual: %d", expected, actual)
	}

	req := server.receivedReqs[0]
	if expected, actual := "PUT", req.Method; expected != actual {
		t.Fatalf("Expected: %s\nActual: %s", expected, actual)
	}
}
