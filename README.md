# 🚇 RelayGo

> **⚠️ This project is actively under development and is not yet production-ready.**

A lightweight HTTP tunneling system written in Go that exposes local services to the internet through a relay server. Similar in concept to [ngrok](https://ngrok.com) or [Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-tunnel/), RelayGo allows you to make a service running on your local machine accessible via a public URL — all routed through a central relay server using WebSockets.

---

## 📐 Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                              INTERNET / CLIENT                              │
│                                                                             │
│              GET http://relay:8080/tunnel/<agent-id>/some/path              │
└────────────────────────────────┬─────────────────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                          RELAY SERVER  (:8080)                              │
│                                                                             │
│  ┌─────────────┐   ┌───────────────┐   ┌──────────────┐   ┌─────────────┐  │
│  │   HTTP Mux   │──▶│  Proxy Handler │──▶│   Registry   │──▶│   Session   │  │
│  │  /tunnel/*   │   │  (proxy.go)   │   │(registry.go) │   │(session.go) │  │
│  └─────────────┘   └───────────────┘   └──────────────┘   └──────┬──────┘  │
│                                                                   │         │
│  ┌─────────────┐   ┌───────────────┐                              │         │
│  │  WebSocket   │──▶│  WS Handler   │◀─────── ReadLoop ◀──────────┘         │
│  │   /ws        │   │   (ws.go)     │                                       │
│  └─────────────┘   └───────────────┘                                        │
└──────────────────────────────────┬───────────────────────────────────────────┘
                                   │  WebSocket (persistent connection)
                                   │  JSON-framed protocol
                                   ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                         AGENT (runs locally)                                │
│                                                                             │
│  ┌─────────────┐   ┌───────────────┐   ┌──────────────────────────────────┐ │
│  │   Client     │──▶│  Read Loop    │──▶│  Forward HTTP to local service  │ │
│  │ (client.go)  │   │  (decodes     │   │  http://localhost:3000           │ │
│  │              │   │   frames)     │   │                                  │ │
│  └─────────────┘   └───────────────┘   └──────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────┘
```

### How It Works

1. **Agent connects** to the relay server over WebSocket (`/ws`) and sends a `REGISTER_AGENT` frame.
2. **Relay registers** the agent in its in-memory registry and responds with a `REGISTERED_AGENT` frame containing a unique tunnel ID.
3. **A client** makes an HTTP request to `http://relay:8080/tunnel/<agent-id>/some/path`.
4. **The relay's proxy handler** looks up the agent session by ID, wraps the incoming HTTP request into an `HTTP_REQUEST` frame, and sends it over the WebSocket.
5. **The agent** receives the frame, forwards the request to the local service (`http://localhost:3000`), and sends the response back as an `HTTP_RESPONSE` frame.
6. **The relay** resolves the pending request, rewrites HTML asset paths, and writes the response back to the original client.

---

## 🗂️ Project Structure

```
relaygo/
├── go.work                          # Go workspace (multi-module)
│
├── shared/                          # Shared protocol library
│   ├── go.mod
│   └── protocol/
│       ├── types.go                 # Frame types (REGISTER, PING, HTTP_REQUEST, etc.)
│       ├── frame.go                 # Frame struct definition
│       ├── codec.go                 # JSON encode/decode helpers
│       ├── http.go                  # HTTPReq / HTTPRes structs
│       └── codec_test.go           # Codec unit tests
│
├── relay/                           # Relay server (public-facing)
│   ├── go.mod
│   ├── .env
│   ├── cmd/
│   │   └── main.go                 # Entry point — starts relay on :8080
│   └── internal/
│       ├── server/
│       │   ├── server.go           # HTTP server setup
│       │   ├── route.go            # Route registration (/, /ws, /tunnel/{id}/{path...})
│       │   └── proxy.go            # Reverse proxy — HTTP ↔ WebSocket bridging
│       ├── tunnel/
│       │   ├── registry.go         # Thread-safe map of agent sessions
│       │   └── session.go          # WebSocket session, read loop, request/response matching
│       └── websocket/
│           └── ws.go               # WebSocket upgrade handler + session lifecycle
│
└── agent/                           # Agent client (runs on user's machine)
    ├── go.mod
    ├── .env
    ├── cmd/
    │   └── main.go                 # Entry point — connects to relay, registers
    └── internal/
        ├── client/
        │   └── client.go           # WebSocket client, frame handler, HTTP forwarder
        ├── server/
        │   ├── server.go           # Local HTTP server (placeholder, port 8081)
        │   └── route.go            # Local route registration
        └── tunnel/                  # (reserved for future use)
```

---

## 🔌 Protocol

Communication between the relay and agent uses a **JSON-framed protocol** over WebSocket. Every message is a `Frame`:

```json
{
  "type": "FRAME_TYPE",
  "payload": { ... }
}
```

### Frame Types

| Type               | Direction       | Description                                                |
|--------------------|-----------------|------------------------------------------------------------|
| `REGISTER_AGENT`   | Agent → Relay   | Agent requests registration with a name                    |
| `REGISTERED_AGENT` | Relay → Agent   | Relay confirms registration and returns the tunnel ID      |
| `PING`             | Relay → Agent   | Heartbeat ping from relay                                  |
| `PONG`             | Agent → Relay   | Heartbeat pong response                                    |
| `HTTP_REQUEST`     | Relay → Agent   | Proxied HTTP request to forward to the local service       |
| `HTTP_RESPONSE`    | Agent → Relay   | Response from the local service to return to the client     |

---

## 🛠️ Tech Stack

- **Language**: Go 1.25+
- **WebSocket**: [gorilla/websocket](https://github.com/gorilla/websocket)
- **Build System**: Go modules with [Go workspaces](https://go.dev/doc/tutorial/workspaces) (`go.work`)
- **Architecture**: Multi-module monorepo (`shared`, `relay`, `agent`)

---

## 🚀 Getting Started

### Prerequisites

- Go 1.25 or later
- A local service running on `http://localhost:3000` (that you want to tunnel)

### 1. Start the Relay Server

```bash
cd relay
go run ./cmd/main.go
```

The relay server starts on `:8080`.

### 2. Start the Agent

```bash
cd agent
go run ./cmd/main.go
```

The agent connects to `ws://localhost:8080/ws`, registers itself, and receives a tunnel ID (printed to the console).

### 3. Access Your Local Service

Copy the tunnel ID from the agent's console output, then open:

```
http://localhost:8080/tunnel/<tunnel-id>/
```

This routes through the relay → agent → your local service on port 3000.

---

## 🧪 Running Tests

```bash
cd shared
go test ./...
```

---

## 🚧 Work In Progress

This project is still under active development. Here's what's done and what's planned:

### ✅ Completed
- [x] WebSocket-based agent ↔ relay communication
- [x] JSON-framed protocol with typed messages
- [x] Agent registration with unique tunnel IDs
- [x] HTTP request proxying (relay → agent → local service → agent → relay → client)
- [x] HTML asset path rewriting for tunneled content
- [x] Heartbeat ping/pong mechanism
- [x] Thread-safe session registry
- [x] Request/response correlation via unique request IDs
- [x] Multi-module Go workspace

### 🔮 Future Architecture

The current implementation intentionally favours simplicity and debuggability: JSON-framed messages are transported over a persistent WebSocket connection.

As RelayGo evolves, I plan to evaluate different architectural approaches rather than assuming a single technology is the right solution. The goal is to understand and measure the trade-offs between **simplicity, performance, reliability, scalability, and operational complexity**.

Some of the areas I plan to explore:

#### Protocol Evolution

- [ ] Evaluate **Protocol Buffers (Protobuf)** vs the current JSON protocol
- [ ] Define a versioned `.proto` schema for relay ↔ agent communication
- [ ] Evaluate **gRPC** and other RPC/transport options where appropriate
- [ ] Support protocol version negotiation
- [ ] Add backward compatibility between protocol versions
- [ ] Benchmark JSON vs Protobuf for latency, throughput, and message size

#### Tunnel & Transport

- [ ] Separate tunnel transport from the application protocol
- [ ] Support multiplexing multiple requests over a single connection
- [ ] Improve connection lifecycle management
- [ ] Automatic agent reconnection
- [ ] Request timeout and cancellation propagation
- [ ] Backpressure handling
- [ ] Graceful connection draining

#### Reliability & Scalability

- [ ] Persistent agent identity
- [ ] Evaluate distributed tunnel registry
- [ ] Evaluate Redis-backed session/state management
- [ ] Multiple relay server support
- [ ] Load balancing between relay nodes
- [ ] Horizontal scaling
- [ ] Observability: structured logging, metrics, and tracing

#### Security

- [ ] TLS between agent and relay
- [ ] Agent authentication
- [ ] Token-based tunnel authentication
- [ ] Access control
- [ ] Rate limiting
- [ ] Request/connection limits
- [ ] Secure tunnel IDs and lifecycle management

#### Developer Experience

- [ ] Production-ready CLI
- [ ] Configuration files
- [ ] `relaygo start` style commands
- [ ] Custom domains/subdomains
- [ ] Dashboard for active tunnels and requests
- [ ] Docker images
- [ ] Kubernetes deployment support

#### Performance

- [ ] Benchmark concurrent tunnel connections
- [ ] Benchmark request throughput
- [ ] Benchmark memory usage under concurrent load
- [ ] Reduce protocol overhead
- [ ] Evaluate connection pooling where appropriate
- [ ] Stream request/response bodies instead of buffering entire responses

---

## 📄 License

TBD

---

> **Note:** RelayGo is currently an experimental project under active development. The architecture is intentionally evolving as I benchmark, test, and evaluate different approaches to networking, concurrency, reliability, and scalability.
