package tools

import (
	"context"
	"testing"
)

func TestGetDaemonHealth(t *testing.T) {
	srv := newFakeStatusServer(t, "v1.2.0", 200, 200, "OK")
	ts := NewToolSet(srv.URL)

	_, res, err := ts.GetDaemonHealth(context.Background(), nil, GetDaemonHealthParams{})
	if err != nil {
		t.Fatalf("GetDaemonHealth returned an error: %v", err)
	}
	if res.Endpoint != "/debug/ready" {
		t.Errorf("expected endpoint /debug/ready, got %q", res.Endpoint)
	}
	if res.Status != 200 {
		t.Errorf("expected status 200, got %d", res.Status)
	}
	if !res.Healthy {
		t.Error("expected healthy to be true")
	}
}

func TestGetDaemonHealthNonOK(t *testing.T) {
	srv := newFakeStatusServer(t, "v1.2.0", 200, 500, "error")
	ts := NewToolSet(srv.URL)

	_, _, err := ts.GetDaemonHealth(context.Background(), nil, GetDaemonHealthParams{})
	if err == nil {
		t.Fatal("expected an error for a non-200 response")
	}
}

func TestGetDaemonHealthUnreachable(t *testing.T) {
	ts := NewToolSet("http://127.0.0.1:1")
	_, _, err := ts.GetDaemonHealth(context.Background(), nil, GetDaemonHealthParams{})
	if err == nil {
		t.Fatal("expected an error for an unreachable status server")
	}
}
