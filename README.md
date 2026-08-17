# Kmesh MCP Server

A Go server that lets AI tools (Claude Desktop, Cursor, GitHub Copilot) talk to a Kmesh service mesh using natural language. Instead of learning `kmeshctl` commands or eBPF internals, you ask questions like "what version is running?" and the AI tool calls this server to get the answer.

This is a prototype for the [CNCF Kmesh MCP Server LFX project](https://github.com/kmesh-dev/kmesh/issues/1800).

## How it works

```
AI Client (Claude/Cursor/Copilot)
        |
        | MCP protocol (Streamable HTTP)
        v
   MCP Server (:8080/mcp)
        |
        | HTTP GET
        v
  Kmesh Daemon (:15200)
```

The MCP server sits between your AI tool and Kmesh. It doesn't run any LLM calls itself -- it just exposes Kmesh data as tools that AI clients can call.

## Try it yourself

Start the server:

```bash
go run ./cmd/mcp-server
```

In another terminal, ask for a tool:

```bash
curl -s http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list",
    "params": {}
  }'
```

You should see the `get_version` tool listed in the response.

## Project structure

```
cmd/mcp-server/
  main.go                 -- entry point, starts the HTTP server

internal/server/
  mcp_server.go           -- sets up the MCP server and registers tools

internal/tools/
  client.go               -- HTTP client for talking to Kmesh status server
  version.go              -- get_version tool (wraps /version endpoint)
  fake_status_test.go     -- fake Kmesh server used by tests
  client_test.go          -- tests for the HTTP client
  version_test.go         -- tests for the get_version tool
```

## Running

Start the server (defaults to localhost:8080, talks to Kmesh status server at localhost:15200):

```bash
go run ./cmd/mcp-server
```

With custom flags:

```bash
go run ./cmd/mcp-server -host 0.0.0.0 -port 9090 -status-addr my-kmesh-pod:15200
```

Or set the status server address via environment variable:

```bash
KMESH_STATUS_ADDR=kmesh-daemon:15200 go run ./cmd/mcp-server
```

The MCP endpoint is at `http://<host>:<port>/mcp`.

## Running tests

```bash
go test ./...
```

With verbose output:

```bash
go test -v ./...
```

With coverage:

```bash
go test -cover ./...
```

## Available tools

| Tool          | What it does                                                    |
| ------------- | --------------------------------------------------------------- |
| `get_version` | Returns the Kmesh version from the daemon's `/version` endpoint |

More tools will be added as the project progresses (see the [proposal](https://github.com/kmesh-dev/kmesh/issues/1800) for the full list).

## Connecting AI clients

Once the server is running, point your AI client at `http://localhost:8080/mcp`. The server uses Streamable HTTP transport in stateless mode, so there's no session management -- each request is independent.

## Known limitations

- Only `get_version` is implemented. The other 9 core tools are planned.
- No `PodName` routing yet -- the tool always hits the configured status server address.

## What's planned

| Tool                  | Endpoint                   | Status  |
| --------------------- | -------------------------- | ------- |
| `get_version`         | `/version`                 | Done    |
| `get_daemon_health`   | `/debug/ready`             | Next    |
| `config_dump`         | `/debug/config_dump/*`     | Planned |
| `get_bpf_maps`        | `/debug/config_dump/bpf/*` | Planned |
| `list_waypoints`      | Kubernetes Gateway API     | Planned |
| `get_waypoint_status` | Kubernetes Gateway API     | Planned |
| `get_authz_status`    | `/authz`                   | Planned |
| `get_logger_levels`   | `/debug/loggers`           | Planned |
| `list_daemon_pods`    | Kubernetes API             | Planned |
| `get_mesh_namespaces` | Kubernetes API             | Planned |

See the full [proposal](https://github.com/kmesh-dev/kmesh/issues/1800) for details.

## Tech stack

- Go 1.26+
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) v1.7.0 (Streamable HTTP, stateless mode)
- Talks to Kmesh daemon's status server on port 15200
