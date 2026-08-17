package tools

import (
	"context"
	"testing"
)

func TestGetVersion(t *testing.T) {
	srv := newFakeStatusServer(t, "v1.2.0", 200, 200, "ready")
	ts := NewToolSet(srv.URL)

	_, res, err := ts.GetVersion(context.Background(), nil, GetVersionParams{})
	if err != nil {
		t.Fatalf("GetVersion returned an error: %v", err)
	}
	if res.Endpoint != "/version" {
		t.Errorf("expected endpoint /version, got %q", res.Endpoint)
	}
	if res.Status != 200 {
		t.Errorf("expected status 200, got %d", res.Status)
	}
	if res.Version != "v1.2.0" {
		t.Errorf("expected version \"v1.2.0\", got %q", res.Version)
	}
}

/*
func TestGetVersionWithPodName(t *testing.T) {
	srv := newFakeStatusServer(t, "\"v1.3.0\"", 200, 200, "ready")
	ts := NewToolSet(srv.URL)

	_, res, err := ts.GetVersion(context.Background(), nil, GetVersionParams{PodName: "kmesh-abcde"})
	if err != nil {
		t.Fatalf("GetVersion returned an error: %v", err)
	}
	if res.Version != "\"v1.3.0\"" {
        t.Errorf("expected version \"v1.3.0\", got %q", res.Version)
    }
}
*/

func TestGetVersionNonOK(t *testing.T) {
	srv := newFakeStatusServer(t, "boom", 500, 500, "boom")
	ts := NewToolSet(srv.URL)

	_, _, err := ts.GetVersion(context.Background(), nil, GetVersionParams{})
	if err == nil {
		t.Fatal("expected an error for a non-200 response")
	}
}

func TestGetVersionUnreachable(t *testing.T) {
	ts := NewToolSet("http://127.0.0.1:1")
	_, _, err := ts.GetVersion(context.Background(), nil, GetVersionParams{})
	if err == nil {
		t.Fatal("expected an error for an unreachable status server")
	}
}
