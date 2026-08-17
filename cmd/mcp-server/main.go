package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/AbHash-RixE/kmesh-mcp-server/internal/server"
	"github.com/AbHash-RixE/kmesh-mcp-server/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	host := flag.String("host", "localhost", "host to listen on")
	port := flag.String("port", "8080", "port to listen on")
	statusAddr := flag.String("status-addr", envOr("KMESH_STATUS_ADDR", tools.DefaultStatusAddr), "address of the Kmesh status server (host:port)")
	flag.Parse()

	mcpServer := server.Setup(tools.NewToolSet(*statusAddr))

	handler := mcp.NewStreamableHTTPHandler(
		//func to map mcp server
		func(*http.Request) *mcp.Server { return mcpServer },
		//mcp server option
		&mcp.StreamableHTTPOptions{Stateless: true},
	)

	//create mcp endpoint & routes client req to mcp handler
	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("Kmesh MCP server listening on http://%s/mcp (stateless streamable HTTP; status server: %s)", addr, *statusAddr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// get kmesh address from env
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
