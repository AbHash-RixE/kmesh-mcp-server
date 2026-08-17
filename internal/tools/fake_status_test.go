package tools

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Creates fake Kmesh status server with configurable test responses
func newFakeStatusServer(t *testing.T, versionBody string, versionCode, readyCode int, readyBody string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(versionCode)
		fmt.Fprint(w, versionBody)
	})
	mux.HandleFunc("/debug/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(readyCode)
		fmt.Fprint(w, readyBody)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
