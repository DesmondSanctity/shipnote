package github_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeServer is an httptest.Server that responds with a fixed body and
// captures the most recent request for assertions.
type fakeServer struct {
	*httptest.Server
	lastBody    string
	lastHeaders http.Header
}

// newFakeServer starts a server returning body for every request.
// The server is closed via t.Cleanup.
func newFakeServer(t *testing.T, body string) *fakeServer {
	t.Helper()
	fs := &fakeServer{}
	fs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		fs.lastBody = string(raw)
		fs.lastHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(fs.Close)
	return fs
}
