package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Request from client
type GetVersionParams struct {
	//PodName string `json:"pod_name,omitempty" jsonschema:"name of the kmesh-daemon pod to query; when empty the configured status server address is used"`
}

// Response from kmesh
type GetVersionResult struct {
	Endpoint string `json:"endpoint" jsonschema:"status server endpoint that was queried"`
	Status   int    `json:"status" jsonschema:"HTTP status code returned by the status server"`
	Version  string `json:"version" jsonschema:"Kmesh version information"`
}

// Tool Definition for mcp
func GetVersionTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "get_version",
		Description: "Retrieves the Kmesh version from the status server /version endpoint.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true},
	}
}

// Tool Handler for Toolset
// params: context, MCP request, tool parameters
func (t *Toolset) GetVersion(ctx context.Context, _ *mcp.CallToolRequest, _ GetVersionParams) (*mcp.CallToolResult, GetVersionResult, error) {
	body, status, err := t.client.Get(ctx, "/version")
	if err != nil {
		//empty GetVersionResult{}
		return nil, GetVersionResult{}, err
	}
	if status != http.StatusOK {
		return nil, GetVersionResult{}, fmt.Errorf("status server returned HTTP %d: %s", status, strings.TrimSpace(string(body)))
	}
	return nil, GetVersionResult{
		Endpoint: "/version",
		Status:   status,
		Version:  strings.TrimSpace(string(body)),
	}, nil
}
