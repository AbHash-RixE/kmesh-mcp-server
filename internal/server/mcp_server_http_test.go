package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbHash-RixE/kmesh-mcp-server/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Creates a test HTTP server exposing the MCP server
func newMCPTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := newFakeStatusServer(t)
	s := Setup(tools.NewToolSet(srv.URL))

	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return s },
		&mcp.StreamableHTTPOptions{Stateless: true},
	)
	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

	hs := httptest.NewServer(mux)
	t.Cleanup(hs.Close)
	return hs
}

// Connects an MCP test client to the MCP server over HTTP
func connectHTTP(t *testing.T, hs *httptest.Server) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             hs.URL + "/mcp",
		HTTPClient:           hs.Client(),
		DisableStandaloneSSE: true,
	}
	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		cancel()
		t.Fatalf("client connect failed: %v", err)
	}
	t.Cleanup(func() {
		cs.Close()
		cancel()
	})
	return cs
}

// Tests that MCP tools can be discovered over HTTP
func TestHTTPToolDiscovery(t *testing.T) {
	hs := newMCPTestServer(t)
	cs := connectHTTP(t, hs)

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	names := make(map[string]bool)
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"get_version"} {
		if !names[want] {
			t.Errorf("expected tool %q to be advertised, got %v", want, res.Tools)
		}
	}
}

// Tests calling the get_version tool over HTTP
func TestHTTPCallTool(t *testing.T) {
	hs := newMCPTestServer(t)
	cs := connectHTTP(t, hs)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_version",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %v", text(res))
	}
	if !strings.Contains(text(res), "v1.2.0") {
		t.Errorf("expected version in result, got %q", text(res))
	}
}

// Tests that the stateless MCP endpoint rejects GET requests
func TestHTTPStatelessRejectsGET(t *testing.T) {
	hs := newMCPTestServer(t)

	res, err := hs.Client().Get(hs.URL + "/mcp")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET in stateless mode, got %d", res.StatusCode)
	}
	if allow := res.Header.Get("Allow"); allow != "POST" {
		t.Errorf("expected Allow: POST, got %q", allow)
	}
}

// Tests that requests without the Accept header are rejected
func TestHTTPRejectsMissingAccept(t *testing.T) {
	hs := newMCPTestServer(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0.0.1"}}}`
	res, err := hs.Client().Post(hs.URL+"/mcp", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing Accept header, got %d", res.StatusCode)
	}
}

// Tests MCP initialization without creating a session
func TestHTTPInitializeSucceedsWithoutSession(t *testing.T) {
	hs := newMCPTestServer(t)

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"0.0.1"}}}`
	req, err := http.NewRequest(http.MethodPost, hs.URL+"/mcp", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	res, err := hs.Client().Do(req)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for initialize, got %d", res.StatusCode)
	}
	if sid := res.Header.Get("Mcp-Session-Id"); sid != "" {
		t.Errorf("stateless server must not issue a session ID, got %q", sid)
	}
}
