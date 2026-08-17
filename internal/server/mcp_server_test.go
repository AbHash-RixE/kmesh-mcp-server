package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbHash-RixE/kmesh-mcp-server/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Create a fake Kmesh status server for testing
func newFakeStatusServer(t *testing.T) *httptest.Server {
	t.Helper()
	// Create HTTP routes that simulate Kmesh's status API.
	mux := http.NewServeMux()
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "v1.2.0")
	})

	// Start the fake Kmesh HTTP server
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// Create an MCP client connected to the MCP server
func connect(t *testing.T, s *mcp.Server) (*mcp.ClientSession, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)

	// Create in-memory transport endpoints for the MCP client and server
	ct, st := mcp.NewInMemoryTransports()

	// Connect MCP server to the server-side transport
	if _, err := s.Connect(ctx, st, nil); err != nil {
		cancel()
		t.Fatalf("server connect failed: %v", err)
	}

	// Connect the test MCP client to the client-side transport
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		cancel()
		t.Fatalf("client connect failed: %v", err)
	}
	t.Cleanup(func() {
		cs.Close()
		cancel()
	})
	return cs, cancel
}

// Test for MCP client to discover all registered tools
func TestToolDiscovery(t *testing.T) {
	// Start fake Kmesh status server
	srv := newFakeStatusServer(t)
	//create MCP server
	s := Setup(tools.NewToolSet(srv.URL))
	cs, _ := connect(t, s)

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
	if len(res.Tools) != 1 {
		t.Errorf("expected exactly 1 tool, got %d", len(res.Tools))
	}
}

// Test for get_version tool for MCP
func TestCallGetVersion(t *testing.T) {
	srv := newFakeStatusServer(t)
	s := Setup(tools.NewToolSet(srv.URL))
	cs, _ := connect(t, s)

	//client ask MCP to execute get_version tool
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
	if res.StructuredContent == nil {
		t.Error("expected structured content in result")
	}
}

// extract text from response
func text(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}
