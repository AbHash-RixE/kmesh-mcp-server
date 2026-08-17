package tools

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestStatusClientGet(t *testing.T) {
	srv := newFakeStatusServer(t, "v1.2.0", http.StatusOK, http.StatusOK, "ready")
	client := NewStatusClient(srv.URL)

	body, status, err := client.Get(context.Background(), "/version")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if status != 200 {
		t.Errorf("expected status 200, got %d", status)
	}
	if strings.TrimSpace(string(body)) != "v1.2.0" {
		t.Errorf("expected body %q, got %q", "\"v1.2.0\"", body)
	}
}

func TestNewStatusClientNormalizesAddress(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"host:port", "localhost:15200", "http://localhost:15200"},
		{"with scheme", "http://localhost:15200", "http://localhost:15200"},
		{"with https", "https://example.com:15200", "https://example.com:15200"},
		{"trailing slash", "localhost:15200/", "http://localhost:15200"},
		{"empty defaults", "", "http://" + DefaultStatusAddr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStatusClient(tt.in).baseURL; got != tt.want {
				t.Errorf("NewStatusClient(%q).baseURL = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestStatusClientGetConnectionError(t *testing.T) {
	client := NewStatusClient("http://127.0.0.1:1")
	if _, _, err := client.Get(context.Background(), "/version"); err == nil {
		t.Fatal("expected an error for an unreachable status server")
	}
}
