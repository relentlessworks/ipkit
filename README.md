# ipkit

Agentic-first IP address and network utility service. Parse, validate, and classify IPv4/IPv6 addresses. CIDR range checks, subnet calculations, IP type detection (private/public/loopback/reserved), reverse DNS, binary representation, and integer conversion. Plain text API, agent-driven, single Go binary.

## Quick Start

```bash
make build
./ipkit
# Listening on :7700
```

## API Reference

### IP Address Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/ip/{addr}` | Parse and classify an IP address |
| GET | `/validate?ip={addr}` | Validate if a string is a valid IP |
| GET | `/binary?ip={addr}` | Get binary representation |
| GET | `/reverse?ip={addr}` | Reverse DNS lookup |
| GET | `/toint?ip={addr}` | Convert IP to integer |
| GET | `/fromint?n={number}` | Convert integer to IPv4 |

### CIDR / Network Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/cidr/{cidr}` | Parse CIDR notation |
| GET | `/contains?ip={addr}&cidr={cidr}` | Check if IP is in CIDR range |
| GET | `/subnet?cidr={cidr}&prefix={n}` | Divide CIDR into smaller subnets |

### Comparison

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/compare?ip1={a}&ip2={b}` | Compare two IP addresses |

### MCP

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/mcp` | Model Context Protocol JSON-RPC 2.0 |

### Meta

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/help` | Operating manual |
| GET | `/.well-known/agent.md` | Same as /help |

## Response Format

Plain text by default (space-separated key=value pairs). JSON on demand via `Accept: application/json` header or `?format=json` query param.

## Examples

```bash
# Parse an IP address
curl http://localhost:7700/ip/192.168.1.1
# ip=192.168.1.1 version=4 type=private private=true ...

# Parse CIDR
curl http://localhost:7700/cidr/10.0.0.0/24
# cidr=10.0.0.0/24 network=10.0.0.0 broadcast=10.0.0.255 ...

# Check if IP is in CIDR
curl "http://localhost:7700/contains?ip=10.0.0.5&cidr=10.0.0.0/24"
# ip=10.0.0.5 cidr=10.0.0.0/24 contains=true

# Divide a subnet
curl "http://localhost:7700/subnet?cidr=192.168.1.0/24&prefix=26"
# 192.168.1.0/26
# 192.168.1.64/26
# 192.168.1.128/26
# 192.168.1.192/26

# Convert IP to integer
curl "http://localhost:7700/toint?ip=192.168.1.1"
# ip=192.168.1.1 integer=3232235777

# Get JSON response
curl -H "Accept: application/json" http://localhost:7700/ip/192.168.1.1
```

## Configuration

| Source | Variable | Default | Description |
|--------|----------|---------|-------------|
| Env | `IPKIT_ADDR` | `:7700` | Listen address |
| Env | `IPKIT_SECRET` | (auto) | Auth token secret |
| Flag | `-addr` | `:7700` | Listen address |
| Flag | `-secret` | (auto) | Auth token secret |

## Build

```bash
make build    # CGO_ENABLED=0 go build -trimpath
make test     # go test -race ./...
make vet      # go vet ./...
make run      # build and run
make clean    # remove binary
```

## License

MIT
