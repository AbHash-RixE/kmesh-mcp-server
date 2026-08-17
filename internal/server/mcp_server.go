package server

import (
	"github.com/AbHash-RixE/kmesh-mcp-server/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Setup(t *tools.Toolset) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "kmesh-mcp-server", Version: "1.0.0"},
		nil,
	)

	//add tools to mcp server (s)
	mcp.AddTool(s, tools.GetVersionTool(), t.GetVersion)
	mcp.AddTool(s, tools.GetDaemonHealthTool(), t.GetDaemonHealth)

	return s
}
