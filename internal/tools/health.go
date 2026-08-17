package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetDaemonHealthParams struct{}

type GetDaemonHealthResult struct {
	Endpoint string `json:"endpoint" jsonschema:"status server endpoint that was queried"`
	Status   int    `json:"status" jsonschema:"HTTP status code returned by the status server"`
	Healthy  bool   `json:"healthy" jsonschema:"whether the daemon reports healthy"`
}

func GetDaemonHealthTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_daemon_health",
		Description: "Checks Kmesh daemon health via the /debug/ready endpoint.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}
}

func (t *Toolset) GetDaemonHealth(ctx context.Context, _ *mcp.CallToolRequest, _ GetDaemonHealthParams) (*mcp.CallToolResult, GetDaemonHealthResult, error) {
	body, status, err := t.client.Get(ctx, "/debug/ready")
	if err != nil {
		return nil, GetDaemonHealthResult{}, err
	}
	if status != http.StatusOK {
		return nil, GetDaemonHealthResult{}, fmt.Errorf("status server returned HTTP %d: %s", status, strings.TrimSpace(string(body)))
	}
	healthy := strings.TrimSpace(string(body)) == "OK"
	return nil, GetDaemonHealthResult{
		Endpoint: "/debug/ready",
		Status:   status,
		Healthy:  healthy,
	}, nil
}
